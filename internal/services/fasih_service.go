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

func (s *FasihService) GetAssignmentDatatable(ctx context.Context, creds dto.FasihCredentials, req dto.FasihDatatableRequest) (*dto.FasihDatatableResponse, error) {
	body := map[string]interface{}{
		"start":   req.Start,
		"length":  req.Length,
		"columns": fasihColumns,
		"order":   []interface{}{},
		"search":  map[string]interface{}{"value": "", "regex": false},
		"assignmentExtraParam": map[string]interface{}{
			"surveyPeriodId":            req.AssignmentExtraParam.SurveyPeriodID,
			"assignmentErrorStatusType": req.AssignmentExtraParam.AssignmentErrorStatusType,
			"assignmentStatusAlias":     req.AssignmentExtraParam.AssignmentStatusAlias,
			"filterTargetType":          req.AssignmentExtraParam.FilterTargetType,
		},
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
