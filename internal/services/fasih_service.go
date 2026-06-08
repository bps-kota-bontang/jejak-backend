package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"jejak/config"
	"jejak/internal/dto"
)

const fasihDatatablePath = "/app/api/analytic/api/v2/assignment/datatable-all-user-survey-periode"
const fasihAssignmentByIDPath = "/app/api/assignment-general/api/assignment/get-by-assignment-id"
const fasihAssignmentHistoryByIDPath = "/app/api/assignment-general/api/assignment-history/get-by-assignment-id"
const fasihRegionMetadataPath = "/app/api/region/api/v1/region-metadata"
const fasihSurveyByIDPath = "/app/api/survey/api/v1/surveys"

var fasihColumns = []map[string]interface{}{
	{"data": "id", "orderable": true},
	{"data": "codeIdentity", "orderable": true},
	{"data": "data1", "orderable": true},
	{"data": "data2", "orderable": true},
	{"data": "data3", "orderable": true},
	{"data": "data4", "orderable": true},
	{"data": "data5", "orderable": true},
	{"data": "data6", "orderable": true},
	{"data": "data7", "orderable": true},
	{"data": "data8", "orderable": true},
	{"data": "data9", "orderable": true},
	{"data": "data10", "orderable": true},
}

type FasihService struct {
	cfg *config.FasihConfig
}

func NewFasihService(cfg *config.FasihConfig) *FasihService {
	return &FasihService{cfg: cfg}
}

func (s *FasihService) IsAvailable(ctx context.Context) bool {
	endpoint, err := url.Parse(s.cfg.BaseURL)
	if err != nil {
		return false
	}

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(checkCtx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)

	return true
}

func (s *FasihService) GetAssignmentDatatable(ctx context.Context, creds dto.FasihCredentials, req dto.FasihDatatableRequest) (*dto.FasihDatatableResponse, error) {
	assignmentExtraParam := map[string]interface{}{
		"surveyPeriodId":            req.AssignmentExtraParam.SurveyPeriodID,
		"assignmentStatusAlias":     req.AssignmentExtraParam.AssignmentStatusAlias,
		"assignmentErrorStatusType": -1,
		"filterTargetType":          "TARGET_ONLY",
	}

	if req.AssignmentExtraParam.Region1ID != nil {
		assignmentExtraParam["region1Id"] = *req.AssignmentExtraParam.Region1ID
	}
	if req.AssignmentExtraParam.Region2ID != nil {
		assignmentExtraParam["region2Id"] = *req.AssignmentExtraParam.Region2ID
	}
	if req.AssignmentExtraParam.Region3ID != nil {
		assignmentExtraParam["region3Id"] = *req.AssignmentExtraParam.Region3ID
	}
	if req.AssignmentExtraParam.Region4ID != nil {
		assignmentExtraParam["region4Id"] = *req.AssignmentExtraParam.Region4ID
	}
	if req.AssignmentExtraParam.Region5ID != nil {
		assignmentExtraParam["region5Id"] = *req.AssignmentExtraParam.Region5ID
	}
	if req.AssignmentExtraParam.Region6ID != nil {
		assignmentExtraParam["region6Id"] = *req.AssignmentExtraParam.Region6ID
	}

	body := map[string]interface{}{
		"start":                req.Start,
		"length":               req.Length,
		"columns":              fasihColumns,
		"order":                []interface{}{},
		"search":               map[string]interface{}{"value": "", "regex": false},
		"assignmentExtraParam": assignmentExtraParam,
	}

	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.BaseURL+fasihDatatablePath, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	setFasihHeaders(httpReq, creds)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.FasihDatatableResponse
	if err := decodeFasihJSONResponse(resp, &result, fasihDatatablePath); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *FasihService) GetAssignmentByID(ctx context.Context, creds dto.FasihCredentials, req dto.FasihAssignmentByIDRequest) (*dto.FasihAssignmentByIDResponse, error) {
	if req.AssignmentID == "" {
		return nil, errors.New("fasih: assignment id is required")
	}

	endpoint, err := url.Parse(s.cfg.BaseURL + fasihAssignmentByIDPath)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("assignmentId", req.AssignmentID)
	endpoint.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	setFasihHeaders(httpReq, creds)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.FasihAssignmentByIDResponse
	if err := decodeFasihJSONResponse(resp, &result, fasihAssignmentByIDPath); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *FasihService) GetAssignmentHistoryByID(ctx context.Context, creds dto.FasihCredentials, req dto.FasihAssignmentByIDRequest) (*dto.FasihAssignmentHistoryByIDResponse, error) {
	if req.AssignmentID == "" {
		return nil, errors.New("fasih: assignment id is required")
	}

	endpoint, err := url.Parse(s.cfg.BaseURL + fasihAssignmentHistoryByIDPath)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("assignmentId", req.AssignmentID)
	endpoint.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	setFasihHeaders(httpReq, creds)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.FasihAssignmentHistoryByIDResponse
	if err := decodeFasihJSONResponse(resp, &result, fasihAssignmentHistoryByIDPath); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *FasihService) GetRegionMetadata(ctx context.Context, creds dto.FasihCredentials, req dto.FasihRegionMetadataRequest) (*dto.FasihRegionMetadataByGroupResponse, error) {
	if req.GroupID == "" {
		return nil, errors.New("fasih: region group id is required")
	}

	endpoint, err := url.Parse(s.cfg.BaseURL + fasihRegionMetadataPath)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("id", req.GroupID)
	endpoint.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	setFasihHeaders(httpReq, creds)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.FasihRegionMetadataByGroupResponse
	if err := decodeFasihJSONResponse(resp, &result, fasihRegionMetadataPath); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *FasihService) GetSurveyByID(ctx context.Context, creds dto.FasihCredentials, req dto.FasihSurveyByIDRequest) (*dto.FasihSurveyByIDResponse, error) {
	if req.SurveyID == "" {
		return nil, errors.New("fasih: survey id is required")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.BaseURL+fasihSurveyByIDPath+"/"+req.SurveyID, nil)
	if err != nil {
		return nil, err
	}

	setFasihHeaders(httpReq, creds)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.FasihSurveyByIDResponse
	if err := decodeFasihJSONResponse(resp, &result, fasihSurveyByIDPath+"/{surveyId}"); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *FasihService) GetRegionsByLevel(ctx context.Context, creds dto.FasihCredentials, req dto.FasihRegionListRequest) (*dto.FasihRegionListResponse, error) {
	if req.GroupID == "" {
		return nil, errors.New("fasih: region group id is required")
	}

	if req.Level < 1 {
		return nil, errors.New("fasih: level must be greater than zero")
	}

	path := fmt.Sprintf("/app/api/region/api/v1/region/level%d", req.Level)
	endpoint, err := url.Parse(s.cfg.BaseURL + path)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("groupId", req.GroupID)
	if req.Level > 1 {
		if req.ParentFullCode == "" {
			return nil, errors.New("fasih: parent full code is required for level greater than one")
		}
		query.Set(fmt.Sprintf("level%dFullCode", req.Level-1), req.ParentFullCode)
	}
	endpoint.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	setFasihHeaders(httpReq, creds)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.FasihRegionListResponse
	if err := decodeFasihJSONResponse(resp, &result, path); err != nil {
		return nil, err
	}

	return &result, nil
}

func setFasihHeaders(req *http.Request, creds dto.FasihCredentials) {
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("x-xsrf-token", creds.XSRFToken)
	req.Header.Set("Cookie", creds.Cookie)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
}

func decodeFasihJSONResponse(resp *http.Response, out interface{}, endpoint string) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bodyPreview := previewResponseBody(body)
	contentType := resp.Header.Get("Content-Type")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fasih: request to %s failed with status %s (content-type=%q, body=%q)", endpoint, resp.Status, contentType, bodyPreview)
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return fmt.Errorf("fasih: empty response from %s", endpoint)
	}

	if strings.HasPrefix(string(trimmed), "<") {
		return fmt.Errorf("fasih: non-JSON response from %s (content-type=%q, body=%q)", endpoint, contentType, bodyPreview)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("fasih: decode JSON failed for %s: %w (content-type=%q, body=%q)", endpoint, err, contentType, bodyPreview)
	}

	return nil
}

func previewResponseBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	const maxLen = 240
	if len(trimmed) <= maxLen {
		return trimmed
	}
	return trimmed[:maxLen] + "..."
}
