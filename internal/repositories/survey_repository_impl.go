package repositories

import (
	"jejak/internal/models"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SurveyRepositoryImpl struct {
	db *gorm.DB
}

func NewSurveyRepository(db *gorm.DB) SurveyRepository {
	return &SurveyRepositoryImpl{db: db}
}

func (r *SurveyRepositoryImpl) Create(survey *models.Survey) error {
	return r.db.Create(survey).Error
}

func (r *SurveyRepositoryImpl) Upsert(survey *models.Survey) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "survey_period_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"xsrf_token",
			"cookie",
			"log_delta_max_mins",
			"log_date_from",
			"log_date_to",
			"updated_at",
		}),
	}).Create(survey).Error
}

func (r *SurveyRepositoryImpl) UpdateBySurveyPeriodID(surveyPeriodID string, survey *models.Survey) error {
	result := r.db.Model(&models.Survey{}).
		Where("survey_period_id = ?", surveyPeriodID).
		Updates(map[string]interface{}{
			"name":               survey.Name,
			"xsrf_token":         survey.XSRFToken,
			"cookie":             survey.Cookie,
			"log_delta_max_mins": survey.LogDeltaMaxMins,
			"log_date_from":      survey.LogDateFrom,
			"log_date_to":        survey.LogDateTo,
			"updated_at":         gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SurveyRepositoryImpl) FindAll() ([]models.Survey, error) {
	var surveys []models.Survey
	if err := r.db.Find(&surveys).Error; err != nil {
		return nil, err
	}
	return surveys, nil
}

func (r *SurveyRepositoryImpl) FindByID(id string) (*models.Survey, error) {
	var survey models.Survey
	if err := r.db.Where("id = ?", id).First(&survey).Error; err != nil {
		return nil, err
	}
	return &survey, nil
}

func (r *SurveyRepositoryImpl) FindBySurveyPeriodID(surveyPeriodID string) (*models.Survey, error) {
	var survey models.Survey
	if err := r.db.Preload("Area").Where("survey_period_id = ?", surveyPeriodID).First(&survey).Error; err != nil {
		return nil, err
	}
	return &survey, nil
}

func (r *SurveyRepositoryImpl) UpdateRegionMetadata(surveyPeriodID string, groupID string, levelCount int) error {
	return r.db.Model(&models.Survey{}).
		Where("survey_period_id = ?", surveyPeriodID).
		Updates(map[string]interface{}{
			"region_group_id":    groupID,
			"region_level_count": levelCount,
			"updated_at":         gorm.Expr("NOW()"),
		}).Error
}

func (r *SurveyRepositoryImpl) UpdateSurveyRegionAssignmentCounts(surveyPeriodID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Region{}).
			Where("survey_period_id = ?", surveyPeriodID).
			Updates(map[string]interface{}{
				"assignment_count": 0,
				"usaha":            0,
				"open_count":       0,
				"draft_count":      0,
				"submitted_count":  0,
				"approved_count":   0,
				"rejected_count":   0,
				"revoked_count":    0,
			}).Error; err != nil {
			return err
		}

		return tx.Exec(`
			UPDATE regions AS r
			SET assignment_count = counts.assignment_count,
				usaha = counts.usaha,
				draft_count = counts.draft_count,
				submitted_count = counts.submitted_count,
				approved_count = counts.approved_count,
				rejected_count = counts.rejected_count,
				revoked_count = counts.revoked_count
			FROM (
				SELECT survey_period_id,
				       region_full_code,
				       COUNT(*) AS assignment_count,
				       COALESCE(SUM(COALESCE(usaha, 0)), 0) AS usaha,
				       SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS draft_count,
				       SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS submitted_count,
				       SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS approved_count,
				       SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS rejected_count,
				       SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS revoked_count
				FROM assignments
				WHERE survey_period_id = ?
				  AND region_full_code IS NOT NULL
				  AND region_full_code <> ''
				GROUP BY survey_period_id, region_full_code
			) AS counts
			WHERE r.survey_period_id = counts.survey_period_id
			  AND r.full_code = counts.region_full_code
		`,
			models.AssignmentStatusDraft,
			models.AssignmentStatusSubmitted,
			models.AssignmentStatusApproved,
			models.AssignmentStatusRejected,
			models.AssignmentStatusRevoked,
			surveyPeriodID,
		).Error
	})
}

func (r *SurveyRepositoryImpl) UpdateSurveyRegionAssignmentCountsByRegion(surveyPeriodID string, regionFullCode string) error {
	trimmedRegionFullCode := strings.TrimSpace(regionFullCode)
	if trimmedRegionFullCode == "" {
		return nil
	}

	return r.db.Model(&models.Region{}).
		Where("survey_period_id = ? AND full_code = ?", surveyPeriodID, trimmedRegionFullCode).
		Updates(map[string]interface{}{
			"assignment_count": gorm.Expr(`
				COALESCE((
					SELECT COUNT(*)
					FROM assignments a
					WHERE a.survey_period_id = regions.survey_period_id
					  AND a.region_full_code = regions.full_code
				), 0)
			`),
			"usaha": gorm.Expr(`
				COALESCE((
					SELECT COALESCE(SUM(COALESCE(a.usaha, 0)), 0)
					FROM assignments a
					WHERE a.survey_period_id = regions.survey_period_id
					  AND a.region_full_code = regions.full_code
				), 0)
			`),
			"draft_count": gorm.Expr(`
				COALESCE((
					SELECT SUM(CASE WHEN a.status = ? THEN 1 ELSE 0 END)
					FROM assignments a
					WHERE a.survey_period_id = regions.survey_period_id
					  AND a.region_full_code = regions.full_code
				), 0)
			`, models.AssignmentStatusDraft),
			"submitted_count": gorm.Expr(`
				COALESCE((
					SELECT SUM(CASE WHEN a.status = ? THEN 1 ELSE 0 END)
					FROM assignments a
					WHERE a.survey_period_id = regions.survey_period_id
					  AND a.region_full_code = regions.full_code
				), 0)
			`, models.AssignmentStatusSubmitted),
			"approved_count": gorm.Expr(`
				COALESCE((
					SELECT SUM(CASE WHEN a.status = ? THEN 1 ELSE 0 END)
					FROM assignments a
					WHERE a.survey_period_id = regions.survey_period_id
					  AND a.region_full_code = regions.full_code
				), 0)
			`, models.AssignmentStatusApproved),
			"rejected_count": gorm.Expr(`
				COALESCE((
					SELECT SUM(CASE WHEN a.status = ? THEN 1 ELSE 0 END)
					FROM assignments a
					WHERE a.survey_period_id = regions.survey_period_id
					  AND a.region_full_code = regions.full_code
				), 0)
			`, models.AssignmentStatusRejected),
			"revoked_count": gorm.Expr(`
				COALESCE((
					SELECT SUM(CASE WHEN a.status = ? THEN 1 ELSE 0 END)
					FROM assignments a
					WHERE a.survey_period_id = regions.survey_period_id
					  AND a.region_full_code = regions.full_code
				), 0)
			`, models.AssignmentStatusRevoked),
		}).Error
}

func (r *SurveyRepositoryImpl) UpdateRegionOpenCount(surveyPeriodID string, regionFullCode string, count int) error {
	trimmed := strings.TrimSpace(regionFullCode)
	if trimmed == "" {
		return nil
	}
	return r.db.Model(&models.Region{}).
		Where("survey_period_id = ? AND full_code = ?", surveyPeriodID, trimmed).
		Update("open_count", count).Error
}

func (r *SurveyRepositoryImpl) UpdateSurveyRegionContacts(surveyPeriodID string, contacts []RegionContactUpdate) (int, error) {
	if len(contacts) == 0 {
		return 0, nil
	}

	updatedRegions := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, contact := range contacts {
			fullCode := strings.TrimSpace(contact.FullCode)
			if fullCode == "" {
				continue
			}

			result := tx.Model(&models.Region{}).
				Where("survey_period_id = ? AND full_code = ?", surveyPeriodID, fullCode).
				Updates(map[string]interface{}{
					"pj":         contact.PJ,
					"pml":        contact.PML,
					"ppl":        contact.PPL,
					"updated_at": gorm.Expr("NOW()"),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				updatedRegions++
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return updatedRegions, nil
}

func (r *SurveyRepositoryImpl) ReplaceSurveyRegions(surveyPeriodID string, regions []models.Region) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("survey_period_id = ?", surveyPeriodID).Delete(&models.Region{}).Error; err != nil {
			return err
		}

		if len(regions) == 0 {
			return nil
		}

		return tx.Create(&regions).Error
	})
}

func (r *SurveyRepositoryImpl) FindSurveyRegionsByLevel(surveyPeriodID string, level int, parentFullCode string) ([]models.Region, error) {
	if level < 1 || level > 6 {
		return nil, gorm.ErrInvalidData
	}

	fullCodeExpr := "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''), COALESCE(level3, ''), COALESCE(level4, ''), COALESCE(level5, ''), COALESCE(level6, ''))"
	parentExpr := "''"
	groupFields := []string{"survey_period_id"}
	selectParts := []string{
		"MIN(id::text) AS id",
		"MIN(survey_id) AS survey_id",
		"survey_period_id",
		"MIN(region_group_id) AS region_group_id",
	}

	if level >= 1 {
		groupFields = append(groupFields, "level1")
		selectParts = append(selectParts, "level1", "MIN(level1_label) AS level1_label")
		fullCodeExpr = "CONCAT(COALESCE(level1, ''))"
	}
	if level >= 2 {
		groupFields = append(groupFields, "level2")
		selectParts = append(selectParts, "level2", "MIN(level2_label) AS level2_label")
		fullCodeExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''))"
		parentExpr = "CONCAT(COALESCE(level1, ''))"
	}
	if level >= 3 {
		groupFields = append(groupFields, "level3")
		selectParts = append(selectParts, "level3", "MIN(level3_label) AS level3_label")
		fullCodeExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''), COALESCE(level3, ''))"
		parentExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''))"
	}
	if level >= 4 {
		groupFields = append(groupFields, "level4")
		selectParts = append(selectParts, "level4", "MIN(level4_label) AS level4_label")
		fullCodeExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''), COALESCE(level3, ''), COALESCE(level4, ''))"
		parentExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''), COALESCE(level3, ''))"
	}
	if level >= 5 {
		groupFields = append(groupFields, "level5")
		selectParts = append(selectParts, "level5", "MIN(level5_label) AS level5_label")
		fullCodeExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''), COALESCE(level3, ''), COALESCE(level4, ''), COALESCE(level5, ''))"
		parentExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''), COALESCE(level3, ''), COALESCE(level4, ''))"
	}
	if level >= 6 {
		groupFields = append(groupFields, "level6")
		selectParts = append(selectParts, "level6", "MIN(level6_label) AS level6_label")
		fullCodeExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''), COALESCE(level3, ''), COALESCE(level4, ''), COALESCE(level5, ''), COALESCE(level6, ''))"
		parentExpr = "CONCAT(COALESCE(level1, ''), COALESCE(level2, ''), COALESCE(level3, ''), COALESCE(level4, ''), COALESCE(level5, ''))"
	}

	selectParts = append(selectParts, fullCodeExpr+" AS full_code")
	selectParts = append(selectParts, "COALESCE(SUM(usaha), 0) AS usaha")

	query := r.db.Model(&models.Region{}).
		Where("survey_period_id = ?", surveyPeriodID).
		Where("level1 IS NOT NULL").
		Select(strings.Join(selectParts, ", ")).
		Group(strings.Join(groupFields, ", ")).
		Order("full_code ASC")

	if parentFullCode != "" && level > 1 {
		query = query.Where(parentExpr+" = ?", parentFullCode)
	}

	var regions []models.Region
	if err := query.Find(&regions).Error; err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *SurveyRepositoryImpl) FindBySurveyPeriodIDWithFilter(surveyPeriodID string, filter AssignmentRegionFilter) ([]models.Region, error) {
	query := r.buildRegionFilterQuery(surveyPeriodID, filter)
	query = applyRegionSort(query, filter)

	var regions []models.Region
	if err := query.Find(&regions).Error; err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *SurveyRepositoryImpl) FindBySurveyPeriodIDWithFilterPaginated(surveyPeriodID string, filter AssignmentRegionFilter, limit int, offset int) ([]models.Region, error) {
	query := r.buildRegionFilterQuery(surveyPeriodID, filter)
	query = applyRegionSort(query, filter)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var regions []models.Region
	if err := query.Find(&regions).Error; err != nil {
		return nil, err
	}

	return regions, nil
}

func applyRegionSort(query *gorm.DB, filter AssignmentRegionFilter) *gorm.DB {
	sortColumn := mapRegionSortColumn(filter.SortBy)
	if sortColumn == "" {
		return query.Order("full_code ASC")
	}

	sortDir := "DESC"
	if strings.EqualFold(strings.TrimSpace(filter.SortDir), "asc") {
		sortDir = "ASC"
	}

	return query.Order(sortColumn + " " + sortDir).Order("full_code ASC")
}

func mapRegionSortColumn(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "draft":
		return "draft_count"
	case "submitted":
		return "submitted_count"
	case "approved":
		return "approved_count"
	case "rejected":
		return "rejected_count"
	case "revoked":
		return "revoked_count"
	case "total":
		return "assignment_count"
	case "usaha":
		return "usaha"
	case "open":
		return "open_count"
	case "progress":
		return "CASE WHEN assignment_count > 0 THEN (draft_count + submitted_count + approved_count + rejected_count + revoked_count)::float / assignment_count ELSE 0 END"
	default:
		return ""
	}
}

func (r *SurveyRepositoryImpl) CountBySurveyPeriodIDWithFilter(surveyPeriodID string, filter AssignmentRegionFilter) (int64, error) {
	query := r.buildRegionFilterQuery(surveyPeriodID, filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *SurveyRepositoryImpl) buildRegionFilterQuery(surveyPeriodID string, filter AssignmentRegionFilter) *gorm.DB {
	query := r.db.Model(&models.Region{}).Where("survey_period_id = ?", surveyPeriodID)

	if strings.TrimSpace(filter.RegionFullCode) != "" {
		query = query.Where("full_code = ?", strings.TrimSpace(filter.RegionFullCode))
	}
	if strings.TrimSpace(filter.RegionLevel1) != "" {
		query = query.Where("level1 = ?", strings.TrimSpace(filter.RegionLevel1))
	}
	if strings.TrimSpace(filter.RegionLevel2) != "" {
		query = query.Where("level2 = ?", strings.TrimSpace(filter.RegionLevel2))
	}
	if strings.TrimSpace(filter.RegionLevel3) != "" {
		query = query.Where("level3 = ?", strings.TrimSpace(filter.RegionLevel3))
	}
	if strings.TrimSpace(filter.RegionLevel4) != "" {
		query = query.Where("level4 = ?", strings.TrimSpace(filter.RegionLevel4))
	}
	if strings.TrimSpace(filter.RegionLevel5) != "" {
		query = query.Where("level5 = ?", strings.TrimSpace(filter.RegionLevel5))
	}
	if strings.TrimSpace(filter.RegionLevel6) != "" {
		query = query.Where("level6 = ?", strings.TrimSpace(filter.RegionLevel6))
	}
	if strings.TrimSpace(filter.PJ) != "" {
		query = query.Where("pj ILIKE ?", "%"+strings.TrimSpace(filter.PJ)+"%")
	}
	if strings.TrimSpace(filter.PML) != "" {
		query = query.Where("pml ILIKE ?", "%"+strings.TrimSpace(filter.PML)+"%")
	}
	if strings.TrimSpace(filter.PPL) != "" {
		query = query.Where("ppl ILIKE ?", "%"+strings.TrimSpace(filter.PPL)+"%")
	}

	assignment := strings.TrimSpace(filter.Assignment)
	if assignment == "has" {
		query = query.Where("assignment_count > 0")
	}
	if assignment == "none" {
		query = query.Where("assignment_count = 0")
	}

	status := strings.ToLower(strings.TrimSpace(filter.Status))
	statusFilters := parseStatusFilters(status)
	if len(statusFilters) > 0 {
		conditions := make([]string, 0, len(statusFilters))
		for _, item := range statusFilters {
			switch item {
			case "draft":
				conditions = append(conditions, "draft_count > 0")
			case "submitted":
				conditions = append(conditions, "submitted_count > 0")
			case "approved":
				conditions = append(conditions, "approved_count > 0")
			case "rejected":
				conditions = append(conditions, "rejected_count > 0")
			case "revoked":
				conditions = append(conditions, "revoked_count > 0")
			}
		}

		if len(conditions) > 0 {
			query = query.Where("(" + strings.Join(conditions, " OR ") + ")")
		}
	}

	return query
}

func parseStatusFilters(raw string) []string {
	if raw == "" {
		return nil
	}

	allowed := map[string]struct{}{
		"draft":     {},
		"submitted": {},
		"approved":  {},
		"rejected":  {},
		"revoked":   {},
	}

	seen := make(map[string]struct{})
	result := make([]string, 0, 5)
	for _, token := range strings.Split(raw, ",") {
		item := strings.ToLower(strings.TrimSpace(token))
		if item == "" {
			continue
		}
		if _, ok := allowed[item]; !ok {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}

		seen[item] = struct{}{}
		result = append(result, item)
	}

	return result
}

func (r *SurveyRepositoryImpl) getDistinctRegionLevel(surveyPeriodID string, valueCol string, labelCol string, filter RegionLevelFilter) ([]RegionLevelOption, error) {
	var options []RegionLevelOption
	query := r.db.Model(&models.Region{}).
		Where("survey_period_id = ?", surveyPeriodID).
		Where(valueCol + " IS NOT NULL AND " + valueCol + " != ''")

	// Apply parent filters
	if filter.Level1 != "" {
		query = query.Where("level1 = ?", filter.Level1)
	}
	if filter.Level2 != "" {
		query = query.Where("level2 = ?", filter.Level2)
	}
	if filter.Level3 != "" {
		query = query.Where("level3 = ?", filter.Level3)
	}
	if filter.Level4 != "" {
		query = query.Where("level4 = ?", filter.Level4)
	}
	if filter.Level5 != "" {
		query = query.Where("level5 = ?", filter.Level5)
	}

	if err := query.
		Distinct(valueCol, labelCol).
		Order(valueCol).
		Select(valueCol + " as value, " + labelCol + " as label").
		Scan(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *SurveyRepositoryImpl) GetDistinctRegionLevel1(surveyPeriodID string) ([]RegionLevelOption, error) {
	return r.getDistinctRegionLevel(surveyPeriodID, "level1", "level1_label", RegionLevelFilter{})
}

func (r *SurveyRepositoryImpl) GetDistinctRegionLevel2(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error) {
	return r.getDistinctRegionLevel(surveyPeriodID, "level2", "level2_label", filter)
}

func (r *SurveyRepositoryImpl) GetDistinctRegionLevel3(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error) {
	return r.getDistinctRegionLevel(surveyPeriodID, "level3", "level3_label", filter)
}

func (r *SurveyRepositoryImpl) GetDistinctRegionLevel4(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error) {
	return r.getDistinctRegionLevel(surveyPeriodID, "level4", "level4_label", filter)
}

func (r *SurveyRepositoryImpl) GetDistinctRegionLevel5(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error) {
	return r.getDistinctRegionLevel(surveyPeriodID, "level5", "level5_label", filter)
}

func (r *SurveyRepositoryImpl) GetDistinctRegionLevel6(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error) {
	return r.getDistinctRegionLevel(surveyPeriodID, "level6", "level6_label", filter)
}
