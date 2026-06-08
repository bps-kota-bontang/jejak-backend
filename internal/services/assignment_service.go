package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"jejak/internal/dto"
	"jejak/internal/helpers"
	"jejak/internal/models"
	"jejak/internal/repositories"
)

type AssignmentService struct {
	assignmentRepo repositories.AssignmentRepository
	logRepo        repositories.LogRepository
	answerRepo     repositories.AnswerRepository
	locationRepo   repositories.LocationRepository
	surveyRepo     repositories.SurveyRepository
	areaRepo       repositories.AreaRepository
}

type geoJSONCollection struct {
	Type     string           `json:"type"`
	Features []geoJSONFeature `json:"features"`
}

type geoJSONFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   geoJSONGeometry        `json:"geometry"`
}

type geoJSONGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type sampleAreaInfo struct {
	HasBoundary bool
	CenterLat   float64
	CenterLon   float64
	RadiusMeter float64
	Polygons    [][][][2]float64
}

var (
	geoJSONCacheMu        sync.Mutex
	geoJSONAreaByConfig   map[string]map[string]sampleAreaInfo
	defaultGeoJSONKeyName = "idsubsls"
)

func NewAssignmentService(
	assignmentRepo repositories.AssignmentRepository,
	logRepo repositories.LogRepository,
	answerRepo repositories.AnswerRepository,
	locationRepo repositories.LocationRepository,
	surveyRepo repositories.SurveyRepository,
	areaRepo repositories.AreaRepository,
) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		logRepo:        logRepo,
		answerRepo:     answerRepo,
		locationRepo:   locationRepo,
		surveyRepo:     surveyRepo,
		areaRepo:       areaRepo,
	}
}

// AnalyzeAssignment analyzes one assignment, persists location proportions,
// and maps each answer row to the inferred location_id.
func (s *AssignmentService) AnalyzeAssignment(ctx context.Context, assignmentID string) (*dto.AssignmentSurveyAnalysis, error) {
	select {
	case <-ctx.Done():
		fmt.Printf("Context done while starting analysis for assignment %s: %v\n", assignmentID, ctx.Err())
		return nil, ctx.Err()
	default:
	}

	assignment, err := s.assignmentRepo.FindByAssignmentID(assignmentID)
	if err != nil {
		return nil, err
	}

	analysis, err := s.analyzeAssignmentCore(*assignment)
	if err != nil {
		return nil, fmt.Errorf("analyze assignment %s: %w", assignmentID, err)
	}

	return &analysis, nil
}

func (s *AssignmentService) GetByAssignmentID(assignmentID string) (*models.Assignment, error) {
	return s.assignmentRepo.FindByAssignmentID(assignmentID)
}

func (s *AssignmentService) GetLogsByAssignmentID(assignmentID string) ([]models.Log, error) {
	logs, err := s.logRepo.FindByAssignmentID(assignmentID)
	if err != nil {
		return nil, err
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].ActionedAt.Before(logs[j].ActionedAt)
	})

	return logs, nil
}

func (s *AssignmentService) analyzeAssignmentCore(assignment models.Assignment) (dto.AssignmentSurveyAnalysis, error) {
	analysis := dto.AssignmentSurveyAnalysis{
		AssignmentID:   assignment.AssignmentID,
		SurveyPeriodID: assignment.SurveyPeriodID,
		Locations:      make([]dto.LocationAnswerStat, 0),
	}

	sampleArea := sampleAreaInfo{}
	if assignment.RegionFullCode != nil && strings.TrimSpace(*assignment.RegionFullCode) != "" {
		survey, err := s.surveyRepo.FindBySurveyPeriodID(assignment.SurveyPeriodID)
		if err != nil {
			return analysis, fmt.Errorf("load survey for assignment %s: %w", assignment.AssignmentID, err)
		}

		area, err := s.areaRepo.FindByID(survey.AreaID)
		if err != nil {
			return analysis, fmt.Errorf("load area for assignment %s: %w", assignment.AssignmentID, err)
		}

		geoJSONKey := strings.TrimSpace(survey.GeoJSONKey)
		if geoJSONKey == "" {
			geoJSONKey = defaultGeoJSONKeyName
		}

		resolvedArea, err := loadSampleAreaByFullCode(*assignment.RegionFullCode, area.GeoJSONFilePath, geoJSONKey)
		if err != nil {
			fmt.Printf("Failed to load sample area for assignment %s with region code %s: %v\n", assignment.AssignmentID, *assignment.RegionFullCode, err)
			return analysis, err
		}
		sampleArea = resolvedArea
	}

	logs, err := s.logRepo.FindByAssignmentID(assignment.AssignmentID)
	if err != nil {
		return analysis, err
	}
	if len(logs) == 0 {
		if err := s.locationRepo.ReplaceByAssignmentID(assignment.AssignmentID, nil); err != nil {
			return analysis, err
		}
		return analysis, nil
	}

	answers, err := s.answerRepo.FindByAssignmentID(assignment.AssignmentID)
	if err != nil {
		return analysis, err
	}
	if len(answers) == 0 {
		if err := s.locationRepo.ReplaceByAssignmentID(assignment.AssignmentID, nil); err != nil {
			return analysis, err
		}
		return analysis, nil
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].ActionedAt.Before(logs[j].ActionedAt)
	})

	locationCounts := map[string]*dto.LocationAnswerStat{}
	type locationAccumulator struct {
		stat         *dto.LocationAnswerStat
		rawLatSum    float64
		rawLonSum    float64
		outsideCount int
	}
	locationAccumulators := map[string]*locationAccumulator{}
	answerCanonicalID := map[string]string{}

	for _, answer := range answers {
		answerTime := getAnswerEventTime(answer)
		closestLog := nearestLogAtOrBefore(logs, answerTime)
		rawLat := closestLog.Latitude
		rawLon := closestLog.Longitude

		lat := helpers.RoundCoordinate(rawLat)
		lon := helpers.RoundCoordinate(rawLon)
		canonicalID := helpers.CanonicalIDFromCoordinates(lat, lon)

		stat, ok := locationCounts[canonicalID]
		if !ok {
			stat = &dto.LocationAnswerStat{
				CanonicalID: canonicalID,
				Latitude:    lat,
				Longitude:   lon,
			}
			locationCounts[canonicalID] = stat
		}
		acc, ok := locationAccumulators[canonicalID]
		if !ok {
			acc = &locationAccumulator{stat: stat}
			locationAccumulators[canonicalID] = acc
		}

		stat.AnswerCount++
		acc.rawLatSum += rawLat
		acc.rawLonSum += rawLon
		if sampleArea.HasBoundary && !isPointInsideSampleArea(rawLat, rawLon, sampleArea) {
			acc.outsideCount++
		}
		analysis.TotalAnswers++
		answerCanonicalID[answer.ID] = canonicalID
	}

	for canonicalID, stat := range locationCounts {
		acc := locationAccumulators[canonicalID]
		if acc == nil || acc.stat == nil || acc.stat.AnswerCount == 0 {
			continue
		}

		if analysis.TotalAnswers > 0 {
			stat.Proportion = float64(stat.AnswerCount) / float64(analysis.TotalAnswers)
		}

		if sampleArea.HasBoundary {
			avgLat := acc.rawLatSum / float64(acc.stat.AnswerCount)
			avgLon := acc.rawLonSum / float64(acc.stat.AnswerCount)
			stat.DistanceToSampleMeters = helpers.HaversineMeters(avgLat, avgLon, sampleArea.CenterLat, sampleArea.CenterLon)
			stat.WithinSampleAreaRadius = stat.DistanceToSampleMeters <= sampleArea.RadiusMeter
			analysis.OutsideAreaProportion += float64(acc.outsideCount) / float64(analysis.TotalAnswers)
		}

		analysis.Locations = append(analysis.Locations, *stat)
	}

	if analysis.OutsideAreaProportion > 0.5 {
		analysis.IsViolation = true
		score := analysis.OutsideAreaProportion
		analysis.ViolationScore = &score

		note := fmt.Sprintf("outside sample area proportion %.4f exceeds threshold 0.5000", score)
		noteCopy := note
		if err := s.assignmentRepo.UpdateViolation(assignment.AssignmentID, true, &noteCopy, &score); err != nil {
			return analysis, err
		}
	} else {
		if err := s.assignmentRepo.UpdateViolation(assignment.AssignmentID, false, nil, nil); err != nil {
			return analysis, err
		}
	}

	sort.Slice(analysis.Locations, func(i, j int) bool {
		return analysis.Locations[i].AnswerCount > analysis.Locations[j].AnswerCount
	})

	locations := make([]models.Location, 0, len(analysis.Locations))
	for _, stat := range analysis.Locations {
		locations = append(locations, models.Location{
			AssignmentID:           analysis.AssignmentID,
			CanonicalID:            stat.CanonicalID,
			Latitude:               stat.Latitude,
			Longitude:              stat.Longitude,
			AnswerCount:            stat.AnswerCount,
			Proportion:             stat.Proportion,
			DistanceToSampleMeters: stat.DistanceToSampleMeters,
			WithinSampleAreaRadius: stat.WithinSampleAreaRadius,
		})
	}

	if err := s.locationRepo.ReplaceByAssignmentID(assignment.AssignmentID, locations); err != nil {
		return analysis, err
	}

	persistedLocations, err := s.locationRepo.FindByAssignmentID(assignment.AssignmentID)
	if err != nil {
		return analysis, err
	}

	locationIDByCanonicalID := map[string]string{}
	for _, item := range persistedLocations {
		locationIDByCanonicalID[item.CanonicalID] = item.ID
	}

	for _, answer := range answers {
		canonicalID, ok := answerCanonicalID[answer.ID]
		if !ok {
			continue
		}
		locationID, ok := locationIDByCanonicalID[canonicalID]
		if !ok {
			continue
		}
		locationIDCopy := locationID
		if err := s.answerRepo.UpdateLocationID(answer.ID, &locationIDCopy); err != nil {
			return analysis, err
		}
	}

	return analysis, nil
}

func getAnswerEventTime(answer models.Answer) time.Time {
	if !answer.AnsweredAt.IsZero() {
		return answer.AnsweredAt
	}
	if !answer.RevisedAt.IsZero() {
		return answer.RevisedAt
	}
	return answer.CreatedAt
}

func nearestLogAtOrBefore(logs []models.Log, t time.Time) models.Log {
	if len(logs) == 0 {
		return models.Log{}
	}

	candidate := logs[0]
	for _, lg := range logs {
		if !lg.ActionedAt.After(t) {
			candidate = lg
			continue
		}
		break
	}

	return candidate
}

func loadSampleAreaByFullCode(fullCode, geoJSONPath, geoJSONKey string) (sampleAreaInfo, error) {
	trimmedCode := strings.TrimSpace(fullCode)
	if trimmedCode == "" {
		return sampleAreaInfo{}, nil
	}

	trimmedGeoJSONPath := strings.TrimSpace(geoJSONPath)
	if trimmedGeoJSONPath == "" {
		return sampleAreaInfo{}, nil
	}

	trimmedGeoJSONKey := strings.TrimSpace(geoJSONKey)
	if trimmedGeoJSONKey == "" {
		trimmedGeoJSONKey = defaultGeoJSONKeyName
	}

	cacheConfigKey := trimmedGeoJSONPath + "|" + trimmedGeoJSONKey

	geoJSONCacheMu.Lock()
	defer geoJSONCacheMu.Unlock()

	if geoJSONAreaByConfig != nil {
		if areaByKey, ok := geoJSONAreaByConfig[cacheConfigKey]; ok {
			if area, ok := areaByKey[trimmedCode]; ok {
				return area, nil
			}
			return sampleAreaInfo{}, nil
		}
	}

	path, err := resolveGeoJSONDiskPath(trimmedGeoJSONPath)
	if err != nil {
		return sampleAreaInfo{}, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return sampleAreaInfo{}, fmt.Errorf("read geojson sample area: %w", err)
	}

	var collection geoJSONCollection
	if err := json.Unmarshal(raw, &collection); err != nil {
		return sampleAreaInfo{}, fmt.Errorf("parse geojson sample area: %w", err)
	}

	result := make(map[string]sampleAreaInfo, len(collection.Features))
	for _, feature := range collection.Features {
		keyValue := strings.TrimSpace(fmt.Sprint(feature.Properties[trimmedGeoJSONKey]))
		if keyValue == "" || keyValue == "<nil>" {
			continue
		}

		area, ok := buildSampleAreaFromFeature(feature)
		if !ok {
			continue
		}
		result[keyValue] = area
	}

	if geoJSONAreaByConfig == nil {
		geoJSONAreaByConfig = make(map[string]map[string]sampleAreaInfo)
	}
	geoJSONAreaByConfig[cacheConfigKey] = result

	if area, ok := geoJSONAreaByConfig[cacheConfigKey][trimmedCode]; ok {
		return area, nil
	}

	return sampleAreaInfo{}, nil
}

func resolveGeoJSONDiskPath(geoJSONPath string) (string, error) {
	trimmedPath := strings.TrimSpace(geoJSONPath)
	if trimmedPath == "" {
		return "", fmt.Errorf("geojson path is empty")
	}

	if strings.HasPrefix(trimmedPath, "public/") {
		cleaned := filepath.Clean(filepath.FromSlash(trimmedPath))
		if strings.HasPrefix(cleaned, "..") {
			return "", fmt.Errorf("invalid geojson path")
		}
		return cleaned, nil
	}

	trimmedPath = strings.TrimPrefix(trimmedPath, "/static/")
	trimmedPath = strings.TrimPrefix(trimmedPath, "static/")
	trimmedPath = strings.TrimPrefix(trimmedPath, "/")

	cleaned := filepath.Clean(filepath.FromSlash(trimmedPath))
	if strings.HasPrefix(cleaned, "..") {
		return "", fmt.Errorf("invalid geojson path")
	}

	return filepath.Join("public", cleaned), nil
}

func buildSampleAreaFromFeature(feature geoJSONFeature) (sampleAreaInfo, bool) {
	polygons := parseFeaturePolygons(feature.Geometry)
	if len(polygons) == 0 {
		return sampleAreaInfo{}, false
	}

	centerLat, centerLon, radiusMeter, ok := calcSampleAreaCircle(polygons)
	if !ok {
		return sampleAreaInfo{}, false
	}

	return sampleAreaInfo{
		HasBoundary: true,
		CenterLat:   centerLat,
		CenterLon:   centerLon,
		RadiusMeter: radiusMeter,
		Polygons:    polygons,
	}, true
}

func parseFeaturePolygons(geometry geoJSONGeometry) [][][][2]float64 {
	switch geometry.Type {
	case "Polygon":
		var polygon [][][2]float64
		if err := json.Unmarshal(geometry.Coordinates, &polygon); err != nil {
			return nil
		}
		if len(polygon) == 0 {
			return nil
		}
		return [][][][2]float64{polygon}
	case "MultiPolygon":
		var multipolygon [][][][2]float64
		if err := json.Unmarshal(geometry.Coordinates, &multipolygon); err != nil {
			return nil
		}
		if len(multipolygon) == 0 {
			return nil
		}
		return multipolygon
	default:
		return nil
	}
}

func calcSampleAreaCircle(polygons [][][][2]float64) (float64, float64, float64, bool) {
	totalLat := 0.0
	totalLon := 0.0
	count := 0.0

	for _, polygon := range polygons {
		if len(polygon) == 0 {
			continue
		}
		outerRing := polygon[0]
		for _, coord := range outerRing {
			lon := coord[0]
			lat := coord[1]
			totalLat += lat
			totalLon += lon
			count++
		}
	}

	if count == 0 {
		return 0, 0, 0, false
	}

	centerLat := totalLat / count
	centerLon := totalLon / count

	radiusMeter := 0.0
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			continue
		}
		outerRing := polygon[0]
		for _, coord := range outerRing {
			lon := coord[0]
			lat := coord[1]
			distance := helpers.HaversineMeters(centerLat, centerLon, lat, lon)
			if distance > radiusMeter {
				radiusMeter = distance
			}
		}
	}

	return centerLat, centerLon, radiusMeter, true
}

func isPointInsideSampleArea(lat, lon float64, area sampleAreaInfo) bool {
	point := [2]float64{lon, lat}

	for _, polygon := range area.Polygons {
		if len(polygon) == 0 {
			continue
		}

		outerRing := polygon[0]
		if !isPointInsideRing(point, outerRing) {
			continue
		}

		insideHole := false
		for i := 1; i < len(polygon); i++ {
			if isPointInsideRing(point, polygon[i]) {
				insideHole = true
				break
			}
		}

		if !insideHole {
			return true
		}
	}

	return false
}

func isPointInsideRing(point [2]float64, ring [][2]float64) bool {
	if len(ring) < 3 {
		return false
	}

	inside := false
	px := point[0]
	py := point[1]

	for i, j := 0, len(ring)-1; i < len(ring); j, i = i, i+1 {
		xi := ring[i][0]
		yi := ring[i][1]
		xj := ring[j][0]
		yj := ring[j][1]

		intersects := (yi > py) != (yj > py) &&
			(px < (xj-xi)*(py-yi)/(yj-yi+math.SmallestNonzeroFloat64)+xi)
		if intersects {
			inside = !inside
		}
	}

	return inside
}
