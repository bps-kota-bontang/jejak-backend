package dto

import (
	"encoding/json"
)

type FasihJSON json.RawMessage

func (j *FasihJSON) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*j = nil
		return nil
	}

	if len(data) > 0 && data[0] == '"' {
		var decoded string
		if err := json.Unmarshal(data, &decoded); err != nil {
			return err
		}
		data = []byte(decoded)
	}

	if !json.Valid(data) {
		return nil
	}

	*j = FasihJSON(append((*j)[:0], data...))
	return nil
}

func (j FasihJSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

type FasihCredentials struct {
	Cookie    string
	XSRFToken string
}

type FasihDatatableRequest struct {
	Start                int                       `json:"start"`
	Length               int                       `json:"length"`
	AssignmentExtraParam FasihAssignmentExtraParam `json:"assignmentExtraParam"`
}

type FasihAssignmentByIDRequest struct {
	AssignmentID string
}

type FasihRegionMetadataRequest struct {
	GroupID string
}

type FasihRegionListRequest struct {
	GroupID        string
	Level          int
	ParentFullCode string
}

type FasihSurveyByIDRequest struct {
	SurveyID string
}

type FasihAssignmentExtraParam struct {
	SurveyPeriodID string `json:"surveyPeriodId"`
}

type FasihParadataPayload struct {
	ActionLogEntities []FasihActionLogEntity `json:"actionLogEntities"`
}

type FasihActionLogEntity struct {
	Action    string  `json:"action"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp string  `json:"timestamp"`
}

type FasihAnswerPayload struct {
	Answers   []FasihAnswerItem `json:"answers"`
	CreatedAt interface{}       `json:"createdAt"`
	UpdatedAt interface{}       `json:"updatedAt"`
}

type FasihAnswerItem struct {
	DataKey   string      `json:"dataKey"`
	Answer    interface{} `json:"answer"`
	CreatedAt interface{} `json:"createdAt"`
	UpdatedAt interface{} `json:"updatedAt"`
}

type FasihDatatableResponse struct {
	SearchData        []FasihAssignmentRow     `json:"searchData"`
	SearchAggregation []FasihSearchAggregation `json:"searchAggregation"`
	TotalHit          int                      `json:"totalHit"`
}

type FasihSearchAggregation struct {
	KeyAggregation string `json:"keyAggregation"`
	DocCount       int    `json:"docCount"`
}

type FasihAssignmentRow struct {
	ID                                 string                          `json:"id"`
	SurveyPeriodID                     string                          `json:"surveyPeriodId"`
	Mode                               []string                        `json:"mode"`
	AssignmentErrorStatusType          int                             `json:"assignmentErrorStatusType"`
	UserIDResponsibility               string                          `json:"userIdResponsibility"`
	ApprovedByCreator                  bool                            `json:"approvedByCreator"`
	CodeIdentity                       string                          `json:"codeIdentity"`
	AssignmentStatusID                 int                             `json:"assignmentStatusId"`
	AssignmentStatusAlias              string                          `json:"assignmentStatusAlias"`
	IsTarikSample                      bool                            `json:"isTarikSample"`
	Data1                              string                          `json:"data1"`
	Data2                              string                          `json:"data2"`
	Data3                              string                          `json:"data3"`
	Data4                              string                          `json:"data4"`
	Data5                              string                          `json:"data5"`
	Data6                              string                          `json:"data6"`
	Data7                              string                          `json:"data7"`
	Data8                              string                          `json:"data8"`
	Data9                              string                          `json:"data9"`
	Data10                             string                          `json:"data10"`
	DateCreated                        string                          `json:"dateCreated"`
	IsActive                           bool                            `json:"isActive"`
	SumError                           int                             `json:"sumError"`
	SumRemark                          int                             `json:"sumRemark"`
	SumClean                           int                             `json:"sumClean"`
	Done                               bool                            `json:"done"`
	Secondary                          bool                            `json:"secondary"`
	Strata                             string                          `json:"strata"`
	Longitude                          float64                         `json:"longitude"`
	Latitude                           float64                         `json:"latitude"`
	CopyFromID                         string                          `json:"copyFromId"`
	ExternalDone                       bool                            `json:"externalDone"`
	CurrentUserID                      string                          `json:"currentUserId"`
	CurrentUserUsername                string                          `json:"currentUserUsername"`
	CurrentUserFullname                string                          `json:"currentUserFullname"`
	CurrentUserSurveyRoleID            string                          `json:"currentUserSurveyRoleId"`
	CurrentUserSurveyRoleName          string                          `json:"currentUserSurveyRoleName"`
	CurrentUserSurveyRoleIsPencacah    bool                            `json:"currentUserSurveyRoleIsPencacah"`
	CurrentUserSurveyRoleCanPullSample bool                            `json:"currentUserSurveyRoleCanPullSample"`
	SourceFrom                         string                          `json:"sourceFrom"`
	Listing                            bool                            `json:"listing"`
	DateModified                       string                          `json:"dateModified"`
	AssignmentResponsibility           []FasihAssignmentResponsibility `json:"assignmentResponsibility"`
	AssignmentResponsibilityAdmin      []interface{}                   `json:"assignmentResponsibilityAdmin"`
	Region                             FasihRegion                     `json:"region"`
	RegionMetadata                     FasihRegionMetadata             `json:"regionMetadata"`
	SampleType                         int                             `json:"sampleType"`
	IsTarget                           bool                            `json:"isTarget"`
	ReferencedTo                       []interface{}                   `json:"referencedTo"`
	LockedByUser                       bool                            `json:"lockedByUser"`
	LockedByAnother                    bool                            `json:"lockedByAnother"`
}

type FasihAssignmentResponsibility struct {
	ID                               string `json:"id"`
	SurveyUserBeforeID               string `json:"surveyUserBeforeId"`
	SurveyUserCurrentID              string `json:"surveyUserCurrentId"`
	AssignmentResponsibilityStatusID string `json:"assignmentResponsibilityStatusId"`
	DateCreated                      string `json:"dateCreated"`
	IsActive                         bool   `json:"isActive"`
	BeforeUserID                     string `json:"beforeUserId"`
	BeforeSurveyRoleID               string `json:"beforeSurveyRoleId"`
	BeforeSurveyRoleName             string `json:"beforeSurveyRoleName"`
	BeforeSurveyRoleRoleID           string `json:"beforeSurveyRoleRoleId"`
	BeforeSurveyRoleCanPullSample    bool   `json:"beforeSurveyRoleCanPullSample"`
	BeforeSurveyRoleIsPencacah       bool   `json:"beforeSurveyRoleIsPencacah"`
	BeforeSurveyRoleSequence         int    `json:"beforeSurveyRoleSequence"`
	CurrentUserID                    string `json:"currentUserId"`
	CurrentSurveyRoleID              string `json:"currentSurveyRoleId"`
	CurrentSurveyRoleName            string `json:"currentSurveyRoleName"`
	CurrentSurveyRoleRoleID          string `json:"currentSurveyRoleRoleId"`
	CurrentSurveyRoleCanPullSample   bool   `json:"currentSurveyRoleCanPullSample"`
	CurrentSurveyRoleIsPencacah      bool   `json:"currentSurveyRoleIsPencacah"`
	CurrentSurveyRoleSequence        int    `json:"currentSurveyRoleSequence"`
	SurveyPeriodID                   string `json:"surveyPeriodId"`
}

type FasihRegion struct {
	ID          string           `json:"id"`
	GroupID     string           `json:"groupId"`
	VersionCode float64          `json:"versionCode"`
	DateCreated string           `json:"dateCreated"`
	IsActive    bool             `json:"isActive"`
	Level1      FasihRegionLevel `json:"level1"`
}

type FasihRegionLevel struct {
	ID       string            `json:"id"`
	FullCode string            `json:"fullCode"`
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	Level2   *FasihRegionLevel `json:"level2,omitempty"`
	Level3   *FasihRegionLevel `json:"level3,omitempty"`
	Level4   *FasihRegionLevel `json:"level4,omitempty"`
	Level5   *FasihRegionLevel `json:"level5,omitempty"`
	Level6   *FasihRegionLevel `json:"level6,omitempty"`
}

type FasihRegionMetadata struct {
	ID                  string                     `json:"id"`
	LevelCount          int                        `json:"levelCount"`
	SmallestRegionLevel string                     `json:"smallestRegionLevel"`
	GroupName           string                     `json:"groupName"`
	IsActive            bool                       `json:"isActive"`
	IsPublic            bool                       `json:"isPublic"`
	Level               []FasihRegionMetadataLevel `json:"level"`
}

type FasihRegionMetadataLevel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type FasihRegionMetadataByGroupResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Data      FasihRegionMetadata `json:"data"`
	ErrorCode *string             `json:"errorCode"`
}

type FasihRegionListResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	Data      []FasihRegionItem `json:"data"`
	ErrorCode *string           `json:"errorCode"`
}

type FasihRegionItem struct {
	ID       string `json:"id"`
	FullCode string `json:"fullCode"`
	Code     string `json:"code"`
	Name     string `json:"name"`
}

type FasihSurveyByIDResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Data      FasihSurveyByIDData `json:"data"`
	ErrorCode *string             `json:"errorCode"`
}

type FasihSurveyByIDData struct {
	ID            string `json:"id"`
	RegionGroupID string `json:"regionGroupId"`
}

type FasihAssignmentByIDResponse struct {
	Success   bool                  `json:"success"`
	Message   string                `json:"message"`
	Data      FasihAssignmentDetail `json:"data"`
	ErrorCode *string               `json:"errorCode"`
}

type FasihAssignmentHistoryByIDResponse struct {
	Success   bool                     `json:"success"`
	Message   string                   `json:"message"`
	Data      []FasihAssignmentHistory `json:"data"`
	ErrorCode *string                  `json:"errorCode"`
}

type FasihAssignmentHistory struct {
	ID              *string   `json:"_id"`
	AssignmentID    string    `json:"assignment_id"`
	Remark          *string   `json:"remark"`
	Mode            []string  `json:"mode"`
	UserID          string    `json:"user_id"`
	DateCreated     string    `json:"date_created"`
	StatusID        *int      `json:"status_id"`
	StatusAlias     string    `json:"status_alias"`
	BasePath        string    `json:"base_path"`
	Comment         FasihJSON `json:"comment"`
	BasePathComment *string   `json:"base_path_comment"`
	Paradata        FasihJSON `json:"paradata"`
	SumErrorDelta   *int      `json:"sum_error_delta"`
	SumRemarkDelta  *int      `json:"sum_remark_delta"`
	SumCleanDelta   *int      `json:"sum_clean_delta"`
}

type FasihAssignmentDetail struct {
	ID                                 string                        `json:"_id"`
	SurveyPeriodID                     string                        `json:"survey_period_id"`
	Mode                               []string                      `json:"mode"`
	AssignmentErrorStatusType          int                           `json:"assignment_error_status_type"`
	PreDefinedData                     string                        `json:"pre_defined_data"`
	Data                               FasihJSON                     `json:"data"`
	UserIDResponsibility               string                        `json:"user_id_responsibility"`
	ApprovedByCreator                  bool                          `json:"approved_by_creator"`
	CodeMaster                         *string                       `json:"code_master"`
	CodeIdentity                       string                        `json:"code_identity"`
	AssignmentStatusID                 int                           `json:"assignment_status_id"`
	AssignmentStatusAlias              string                        `json:"assignment_status_alias"`
	SurveyFileID                       *string                       `json:"survey_file_id"`
	IsTarikSample                      bool                          `json:"is_tarik_sample"`
	Data1                              string                        `json:"data1"`
	Data2                              string                        `json:"data2"`
	Data3                              string                        `json:"data3"`
	Data4                              string                        `json:"data4"`
	Data5                              string                        `json:"data5"`
	Data6                              string                        `json:"data6"`
	Data7                              string                        `json:"data7"`
	Data8                              string                        `json:"data8"`
	Data9                              string                        `json:"data9"`
	Data10                             string                        `json:"data10"`
	DateCreated                        string                        `json:"date_created"`
	IsActive                           bool                          `json:"is_active"`
	Username                           *string                       `json:"username"`
	SumError                           int                           `json:"sum_error"`
	SumRemark                          int                           `json:"sum_remark"`
	SumClean                           int                           `json:"sum_clean"`
	Comment                            FasihJSON                     `json:"comment"`
	Done                               bool                          `json:"done"`
	Secondary                          bool                          `json:"secondary"`
	ReplaceID                          *string                       `json:"replace_id"`
	Longitude                          float64                       `json:"longitude"`
	Latitude                           float64                       `json:"latitude"`
	CopyFromID                         string                        `json:"copy_from_id"`
	Strata                             *string                       `json:"strata"`
	ExternalDone                       bool                          `json:"external_done"`
	CurrentUserID                      string                        `json:"current_user_id"`
	CurrentUserUsername                string                        `json:"current_user_username"`
	CurrentUserFullname                string                        `json:"current_user_fullname"`
	CurrentUserSurveyRoleID            string                        `json:"current_user_survey_role_id"`
	CurrentUserSurveyRoleName          string                        `json:"current_user_survey_role_name"`
	CurrentUserSurveyRoleIsPencacah    bool                          `json:"current_user_survey_role_is_pencacah"`
	CurrentUserSurveyRoleCanPullSample bool                          `json:"current_user_survey_role_can_pull_sample"`
	Email                              *string                       `json:"email"`
	LandLineNumber                     *string                       `json:"land_line_number"`
	PhoneNumber                        *string                       `json:"phone_number"`
	SourceFrom                         string                        `json:"source_from"`
	Remark                             *string                       `json:"remark"`
	ListingData                        *string                       `json:"listing_data"`
	DateModified                       string                        `json:"date_modified"`
	Panel1Data                         *string                       `json:"panel1_data"`
	Panel2Data                         *string                       `json:"panel2_data"`
	PendingUploadAssignmentID          *string                       `json:"pending_upload_assignment_id"`
	PendingAssignmentCreatedDate       *string                       `json:"pending_assignment_created_date"`
	Region                             FasihAssignmentRegion         `json:"region"`
	Listing                            bool                          `json:"listing"`
	SubmitVersionCode                  int                           `json:"submit_version_code"`
	RegionMetadata                     FasihAssignmentRegionMetadata `json:"region_metadata"`
	StorageType                        *string                       `json:"storage_type"`
	StorageKey                         *string                       `json:"storage_key"`
	DataSize                           *int                          `json:"data_size"`
}

type FasihAssignmentRegion struct {
	ID          string                     `json:"_id"`
	GroupID     string                     `json:"group_id"`
	VersionCode float64                    `json:"version_code"`
	DateCreated string                     `json:"date_created"`
	IsActive    bool                       `json:"is_active"`
	Level1      FasihAssignmentRegionLevel `json:"level_1"`
}

type FasihAssignmentRegionLevel struct {
	ID       string                      `json:"_id"`
	FullCode string                      `json:"full_code"`
	Code     string                      `json:"code"`
	Name     string                      `json:"name"`
	Level2   *FasihAssignmentRegionLevel `json:"level_2,omitempty"`
	Level3   *FasihAssignmentRegionLevel `json:"level_3,omitempty"`
	Level4   *FasihAssignmentRegionLevel `json:"level_4,omitempty"`
	Level5   *FasihAssignmentRegionLevel `json:"level_5,omitempty"`
	Level6   *FasihAssignmentRegionLevel `json:"level_6,omitempty"`
	Level7   *FasihAssignmentRegionLevel `json:"level_7,omitempty"`
}

type FasihAssignmentRegionMetadata struct {
	ID                  string                           `json:"_id"`
	LevelCount          int                              `json:"level_count"`
	SmallestRegionLevel string                           `json:"smallest_region_level"`
	GroupName           string                           `json:"group_name"`
	Level               []FasihAssignmentRegionMetaLevel `json:"level"`
}

type FasihAssignmentRegionMetaLevel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
