package dto

import "time"

type CreateSurveyRequest struct {
	Name            string  `json:"name" validate:"required"`
	SurveyID        string  `json:"survey_id" validate:"required"`
	SurveyPeriodID  string  `json:"survey_period_id" validate:"required"`
	XSRFToken       string  `json:"xsrf_token" validate:"required"`
	Cookie          string  `json:"cookie" validate:"required"`
	RegionLevel1    string  `json:"region_level_1" validate:"required"`
	RegionLevel2    string  `json:"region_level_2" validate:"required"`
	LogDeltaMaxMins *int    `json:"log_delta_max_minutes,omitempty" validate:"omitempty,min=1"`
	LogDateFrom     *string `json:"log_date_from,omitempty"`
	LogDateTo       *string `json:"log_date_to,omitempty"`
	AreaID          string  `json:"area_id" validate:"required"`
	GeoJSONKey      string  `json:"geojson_key" validate:"required"`
}

type UpdateSurveyRequest struct {
	Name            string  `json:"name" validate:"required"`
	SurveyID        string  `json:"survey_id" validate:"required"`
	XSRFToken       string  `json:"xsrf_token" validate:"required"`
	Cookie          string  `json:"cookie" validate:"required"`
	RegionLevel1    string  `json:"region_level_1" validate:"required"`
	RegionLevel2    string  `json:"region_level_2" validate:"required"`
	LogDeltaMaxMins *int    `json:"log_delta_max_minutes,omitempty" validate:"omitempty,min=1"`
	LogDateFrom     *string `json:"log_date_from,omitempty"`
	LogDateTo       *string `json:"log_date_to,omitempty"`
	AreaID          string  `json:"area_id" validate:"required"`
	GeoJSONKey      string  `json:"geojson_key" validate:"required"`
}

type SurveyResponse struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	SurveyID         string        `json:"survey_id"`
	SurveyPeriodID   string        `json:"survey_period_id"`
	XSRFToken        string        `json:"xsrf_token"`
	Cookie           string        `json:"cookie"`
	RegionLevel1     string        `json:"region_level_1,omitempty"`
	RegionLevel2     string        `json:"region_level_2,omitempty"`
	LogDeltaMaxMins  *int          `json:"log_delta_max_minutes,omitempty"`
	LogDateFrom      *string       `json:"log_date_from,omitempty"`
	LogDateTo        *string       `json:"log_date_to,omitempty"`
	RegionGroupID    *string       `json:"region_group_id,omitempty"`
	RegionLevelCount *int          `json:"region_level_count,omitempty"`
	AreaID           string        `json:"area_id,omitempty"`
	GeoJSONKey       string        `json:"geojson_key,omitempty"`
	Area             *AreaResponse `json:"area,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type AssignmentResponse struct {
	ID             string               `json:"id"`
	SurveyPeriodID string               `json:"survey_period_id"`
	AssignmentID   string               `json:"assignment_id"`
	Status         *int                 `json:"status,omitempty"`
	RegionFullCode *string              `json:"region_full_code,omitempty"`
	RegionLevel1   *string              `json:"region_level_1,omitempty"`
	RegionLevel2   *string              `json:"region_level_2,omitempty"`
	RegionLevel3   *string              `json:"region_level_3,omitempty"`
	RegionLevel4   *string              `json:"region_level_4,omitempty"`
	RegionLevel5   *string              `json:"region_level_5,omitempty"`
	RegionLevel6   *string              `json:"region_level_6,omitempty"`
	Latitude       float64              `json:"latitude"`
	Longitude      float64              `json:"longitude"`
	Usaha          *int                 `json:"usaha"`
	OpenedAt       *time.Time           `json:"opened_at,omitempty"`
	StartedAt      *time.Time           `json:"started_at,omitempty"`
	SubmittedAt    time.Time            `json:"submitted_at"`
	RevisedAt      time.Time            `json:"revised_at"`
	IsViolation    bool                 `json:"is_violation"`
	ViolationNote  *string              `json:"violation_note,omitempty"`
	ViolationScore *float64             `json:"violation_score,omitempty"`
	Locations      []LocationAnswerStat `json:"locations,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type AssignmentLogPointResponse struct {
	ID           string    `json:"id"`
	AssignmentID string    `json:"assignment_id"`
	Action       string    `json:"action"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	ActionedAt   time.Time `json:"actioned_at"`
}

type AssignmentRegionFilterQuery struct {
	RegionFullCode string `json:"region_full_code,omitempty"`
	RegionLevel1   string `json:"region_level_1,omitempty"`
	RegionLevel2   string `json:"region_level_2,omitempty"`
	RegionLevel3   string `json:"region_level_3,omitempty"`
	RegionLevel4   string `json:"region_level_4,omitempty"`
	RegionLevel5   string `json:"region_level_5,omitempty"`
	RegionLevel6   string `json:"region_level_6,omitempty"`
	PJ             string `json:"pj,omitempty"`
	PML            string `json:"pml,omitempty"`
	PPL            string `json:"ppl,omitempty"`
	Assignment     string `json:"assignment_filter,omitempty"`
	Status         string `json:"status_filter,omitempty"`
	SortBy         string `json:"sort_by,omitempty"`
	SortDir        string `json:"sort_dir,omitempty"`
	Page           int    `json:"page,omitempty"`
	PerPage        int    `json:"per_page,omitempty"`
}

type SyncSurveyAssignmentsResponse struct {
	TotalAssignments int `json:"total_assignments"`
	SavedAssignments int `json:"saved_assignments"`
	SavedLogs        int `json:"saved_logs"`
	SavedAnswers     int `json:"saved_answers"`
}

type SyncSurveyAssignmentsRequest struct {
	RegionFullCode string `json:"region_full_code,omitempty"`
}

type QueuedSurveyTaskResponse struct {
	TaskID         string `json:"task_id"`
	Queue          string `json:"queue"`
	Type           string `json:"type"`
	SurveyPeriodID string `json:"survey_period_id"`
	RegionFullCode string `json:"region_full_code,omitempty"`
}

type SyncSurveyRegionsRequest struct {
	RegionGroupID string `json:"region_group_id,omitempty"`
}

type SurveyRegionMetadataLevelResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SurveyRegionMetadataResponse struct {
	RegionGroupID       string                              `json:"region_group_id"`
	LevelCount          int                                 `json:"level_count"`
	SmallestRegionLevel string                              `json:"smallest_region_level"`
	GroupName           string                              `json:"group_name"`
	IsActive            bool                                `json:"is_active"`
	IsPublic            bool                                `json:"is_public"`
	Level               []SurveyRegionMetadataLevelResponse `json:"level"`
}

type SurveyRegionResponse struct {
	ID              string  `json:"id"`
	SurveyID        string  `json:"survey_id"`
	SurveyPeriodID  string  `json:"survey_period_id"`
	RegionGroupID   string  `json:"region_group_id"`
	AssignmentCount int     `json:"assignment_count"`
	Usaha           int     `json:"usaha"`
	DraftCount      int     `json:"draft_count"`
	SubmittedCount  int     `json:"submitted_count"`
	ApprovedCount   int     `json:"approved_count"`
	RejectedCount   int     `json:"rejected_count"`
	RevokedCount    int     `json:"revoked_count"`
	Level1          *string `json:"level_1,omitempty"`
	Level1Label     *string `json:"level_1_label,omitempty"`
	Level2          *string `json:"level_2,omitempty"`
	Level2Label     *string `json:"level_2_label,omitempty"`
	Level3          *string `json:"level_3,omitempty"`
	Level3Label     *string `json:"level_3_label,omitempty"`
	Level4          *string `json:"level_4,omitempty"`
	Level4Label     *string `json:"level_4_label,omitempty"`
	Level5          *string `json:"level_5,omitempty"`
	Level5Label     *string `json:"level_5_label,omitempty"`
	Level6          *string `json:"level_6,omitempty"`
	Level6Label     *string `json:"level_6_label,omitempty"`
	PJ              *string `json:"pj,omitempty"`
	PML             *string `json:"pml,omitempty"`
	PPL             *string `json:"ppl,omitempty"`
	FullCode        string  `json:"full_code"`
}

type SurveyRegionFilterQuery struct {
	Level          int    `json:"level"`
	ParentFullCode string `json:"parent_full_code,omitempty"`
}

type SyncSurveyRegionsResponse struct {
	RegionGroupID string `json:"region_group_id"`
	LevelCount    int    `json:"level_count"`
	SavedRegions  int    `json:"saved_regions"`
}

type ImportSurveyRegionContactsResponse struct {
	TotalRows      int `json:"total_rows"`
	UpdatedRegions int `json:"updated_regions"`
	SkippedRows    int `json:"skipped_rows"`
}

type ImportedRegionItem struct {
	SurveyID       string `json:"survey_id"`
	SurveyPeriodID string `json:"survey_period_id"`
	RegionGroupID  string `json:"region_group_id"`
	Level1         string `json:"level_1"`
	Level1Label    string `json:"level_1_label"`
	Level2         string `json:"level_2"`
	Level2Label    string `json:"level_2_label"`
	Level3         string `json:"level_3"`
	Level3Label    string `json:"level_3_label"`
	Level4         string `json:"level_4"`
	Level4Label    string `json:"level_4_label"`
	Level5         string `json:"level_5"`
	Level5Label    string `json:"level_5_label"`
	Level6         string `json:"level_6"`
	Level6Label    string `json:"level_6_label"`
	FullCode       string `json:"full_code"`
}

type RegionImportPayload struct {
	Type           string               `json:"type"`
	SurveyID       string               `json:"survey_id"`
	SurveyPeriodID string               `json:"survey_period_id"`
	RegionGroupID  string               `json:"region_group_id"`
	LevelCount     int                  `json:"level_count"`
	Regions        []ImportedRegionItem `json:"regions"`
}

type ImportedAssignmentItem struct {
	AssignmentID   string  `json:"assignment_id"`
	SurveyPeriodID string  `json:"survey_period_id"`
	Status         *int    `json:"status,omitempty"`
	RegionFullCode string  `json:"region_full_code"`
	RegionLevel1   string  `json:"region_level_1"`
	RegionLevel2   string  `json:"region_level_2"`
	RegionLevel3   string  `json:"region_level_3"`
	RegionLevel4   string  `json:"region_level_4"`
	RegionLevel5   string  `json:"region_level_5"`
	RegionLevel6   string  `json:"region_level_6"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	OpenedAt       string  `json:"opened_at"`
	StartedAt      string  `json:"started_at"`
	SubmittedAt    string  `json:"submitted_at"`
	RevisedAt      string  `json:"revised_at"`
}

type ImportedLogItem struct {
	AssignmentID string  `json:"assignment_id"`
	Action       string  `json:"action"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ActionedAt   string  `json:"actioned_at"`
}

type ImportedAnswerItem struct {
	AssignmentID string `json:"assignment_id"`
	Name         string `json:"name"`
	Value        string `json:"value,omitempty"`
	AnsweredAt   string `json:"answered_at"`
	RevisedAt    string `json:"revised_at"`
}

type AssignmentImportPayload struct {
	Type             string                   `json:"type"`
	SurveyPeriodID   string                   `json:"survey_period_id"`
	TotalHit         int                      `json:"total_hit"`
	Assignments      []ImportedAssignmentItem `json:"assignments"`
	Logs             []ImportedLogItem        `json:"logs"`
	Answers          []ImportedAnswerItem     `json:"answers"`
	SavedAssignments int                      `json:"saved_assignments"`
	SavedLogs        int                      `json:"saved_logs"`
	SavedAnswers     int                      `json:"saved_answers"`
}

type LocationAnswerStat struct {
	CanonicalID            string  `json:"canonical_id"`
	Latitude               float64 `json:"latitude"`
	Longitude              float64 `json:"longitude"`
	AnswerCount            int     `json:"answer_count"`
	Proportion             float64 `json:"proportion"`
	DistanceToSampleMeters float64 `json:"distance_to_sample_meters"`
	WithinSampleAreaRadius bool    `json:"within_sample_area_radius"`
}

type AssignmentSurveyAnalysis struct {
	AssignmentID          string               `json:"assignment_id"`
	SurveyPeriodID        string               `json:"survey_period_id"`
	TotalAnswers          int                  `json:"total_answers"`
	Locations             []LocationAnswerStat `json:"locations"`
	OutsideAreaProportion float64              `json:"outside_area_proportion"`
	IsViolation           bool                 `json:"is_violation"`
	ViolationScore        *float64             `json:"violation_score,omitempty"`
}

type SurveyFraudAnalysisResult struct {
	SurveyPeriodID      string                     `json:"survey_period_id"`
	TotalAssignments    int                        `json:"total_assignments"`
	AnalyzedAssignments int                        `json:"analyzed_assignments"`
	GeneratedAt         time.Time                  `json:"generated_at"`
	Assignments         []AssignmentSurveyAnalysis `json:"assignments"`
}

type RegionFilterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type RegionFilterOptionsResponse struct {
	Level1 []RegionFilterOption `json:"level_1"`
	Level2 []RegionFilterOption `json:"level_2"`
	Level3 []RegionFilterOption `json:"level_3"`
	Level4 []RegionFilterOption `json:"level_4"`
	Level5 []RegionFilterOption `json:"level_5"`
	Level6 []RegionFilterOption `json:"level_6"`
}
