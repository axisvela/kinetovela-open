// Package kinetovela is the public Go client for the KINETOVELA API.
package kinetovela

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Version is the SDK semantic version, synchronized with core releases.
const Version = "3.2.0"

type Config struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

type Client struct {
	base   string
	token  string
	client *http.Client
}

type FleetMetrics struct {
	FleetSize                int       `json:"fleet_size"`
	AvailableRobots          int       `json:"available_robots"`
	Availability             float64   `json:"availability"`
	MissionsAttempted        int       `json:"missions_attempted"`
	MissionsCompleted        int       `json:"missions_completed"`
	MissionCompletionRate    float64   `json:"mission_completion_rate"`
	EffectsClassified        int       `json:"effects_classified"`
	EffectsUnclassified      int       `json:"effects_unclassified"`
	EffectClassificationRate float64   `json:"effect_classification_rate"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type EdgeRegistration struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Version  string `json:"version,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type Edge struct {
	ID              string            `json:"id"`
	TenantID        string            `json:"tenant_id"`
	Version         string            `json:"version"`
	Endpoint        string            `json:"endpoint"`
	RegisteredAt    time.Time         `json:"registered_at"`
	LastHeartbeatAt *time.Time        `json:"last_heartbeat_at"`
	Lag             int64             `json:"lag"`
	Missions        map[string]string `json:"missions"`
	SyncPending     int               `json:"sync_pending"`
	LastEventSeq    int64             `json:"last_event_seq"`
	SyncedAt        *time.Time        `json:"synced_at"`
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return "kinetovela: unexpected status " + http.StatusText(e.StatusCode)
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	if e.Message != "" {
		return e.Message
	}
	return "kinetovela: unexpected status " + http.StatusText(e.StatusCode)
}

type Mission struct {
	SchemaVersion     string     `json:"schema_version"`
	MissionID         string     `json:"mission_id"`
	PlanRef           string     `json:"plan_ref"`
	CapabilityRefs    []string   `json:"capability_refs"`
	AuthorityRef      string     `json:"authority_ref"`
	SafetyEnvelopeRef string     `json:"safety_envelope_ref"`
	ExpectedEffects   []string   `json:"expected_effects"`
	DeadlineAt        *time.Time `json:"deadline_at,omitempty"`
}

type MissionAssignment struct {
	SchemaVersion    string `json:"schema_version"`
	AssignmentID     string `json:"assignment_id"`
	MissionRef       string `json:"mission_ref"`
	RobotRef         string `json:"robot_ref"`
	LeaseRef         string `json:"lease_ref"`
	FencingToken     int64  `json:"fencing_token"`
	AssignmentReason string `json:"assignment_reason"`
}

type MissionOperation struct {
	SchemaVersion      string          `json:"schema_version"`
	OperationID        string          `json:"operation_id"`
	MissionID          string          `json:"mission_id"`
	AssignmentID       string          `json:"assignment_id"`
	EdgeID             string          `json:"edge_id"`
	FencingToken       int64           `json:"fencing_token"`
	CapabilityRef      string          `json:"capability_ref"`
	DelegatedOperation json.RawMessage `json:"delegated_operation"`
}

type MissionView struct {
	Mission         Mission            `json:"mission"`
	Status          string             `json:"status"`
	Assignment      *MissionAssignment `json:"assignment,omitempty"`
	Operations      []MissionOperation `json:"operations"`
	CancelRequested bool               `json:"cancel_requested"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type MissionList struct {
	Missions []MissionView `json:"missions"`
}

type MissionOperationList struct {
	Operations []MissionOperation `json:"operations"`
}

type AssignmentRequest struct {
	EdgeID           string `json:"edge_id"`
	RobotRef         string `json:"robot_ref"`
	LeaseRef         string `json:"lease_ref"`
	FencingToken     int64  `json:"fencing_token"`
	AssignmentReason string `json:"assignment_reason"`
}

type CancelRequest struct {
	Reason string `json:"reason"`
}

type EffectExecution struct {
	EffectStatus    string `json:"effect_status,omitempty"`
	CancelStatus    string `json:"cancel_status,omitempty"`
	StopConfirmed   *bool  `json:"stop_confirmed,omitempty"`
	ContractMode    string `json:"contract_mode,omitempty"`
	ExecutionStatus string `json:"execution_status,omitempty"`
	ReceiptRef      string `json:"receipt_ref,omitempty"`
}

type EffectReport struct {
	SchemaVersion string           `json:"schema_version"`
	ReportID      string           `json:"report_id"`
	MissionID     string           `json:"mission_id"`
	AssignmentID  string           `json:"assignment_id"`
	OperationKey  string           `json:"operation_key"`
	AttemptID     string           `json:"attempt_id"`
	NodeID        string           `json:"node_id"`
	BootEpoch     int64            `json:"boot_epoch"`
	FencingToken  int64            `json:"fencing_token"`
	Status        string           `json:"status"`
	ObservedAt    time.Time        `json:"observed_at"`
	EvidenceRef   string           `json:"evidence_ref,omitempty"`
	Provenance    string           `json:"provenance"`
	Execution     *EffectExecution `json:"execution,omitempty"`
	Details       map[string]any   `json:"details,omitempty"`
}

type EffectReportResult struct {
	Accepted      bool   `json:"accepted"`
	EffectStatus  string `json:"effect_status"`
	MissionStatus string `json:"mission_status"`
}

type MissionFilter struct {
	EdgeID string
	Status string
	Limit  int
	Cursor string
}

type MissionPage struct {
	Missions   []MissionView `json:"missions"`
	NextCursor string        `json:"next_cursor"`
}

func New(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("kinetovela: BaseURL is required")
	}
	u, err := url.Parse(cfg.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, errors.New("kinetovela: BaseURL must be http or https")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{base: strings.TrimRight(cfg.BaseURL, "/"), token: cfg.Token, client: cfg.HTTPClient}, nil
}

func (c *Client) Health(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/healthz", nil, http.StatusOK)
}

func (c *Client) Version(ctx context.Context) (string, error) {
	var out struct {
		Version string `json:"version"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/version", nil, &out); err != nil {
		return "", err
	}
	return out.Version, nil
}

func (c *Client) FleetSLO(ctx context.Context) (*FleetMetrics, error) {
	var out FleetMetrics
	if err := c.doJSON(ctx, http.MethodGet, "/fleet/slo", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RegisterEdge(ctx context.Context, reg EdgeRegistration) error {
	return c.do(ctx, http.MethodPost, "/edges", reg, http.StatusOK, http.StatusCreated)
}

func (c *Client) Heartbeat(ctx context.Context, edgeID string, lag int64) error {
	return c.do(ctx, http.MethodPost, "/edges/"+url.PathEscape(edgeID)+"/heartbeat",
		map[string]int64{"lag": lag}, http.StatusOK, http.StatusNoContent)
}

func (c *Client) BindMission(ctx context.Context, edgeID, missionID, instanceID string) error {
	return c.do(ctx, http.MethodPost, "/edges/"+url.PathEscape(edgeID)+"/missions",
		map[string]string{"mission_id": missionID, "instance_id": instanceID},
		http.StatusOK, http.StatusCreated, http.StatusNoContent)
}

func (c *Client) UnbindMission(ctx context.Context, edgeID, missionID string) error {
	return c.do(ctx, http.MethodDelete,
		"/edges/"+url.PathEscape(edgeID)+"/missions/"+url.PathEscape(missionID),
		nil, http.StatusOK, http.StatusNoContent)
}

func (c *Client) ReportSync(ctx context.Context, edgeID string, pending int, lastEventSeq int64) error {
	return c.do(ctx, http.MethodPost, "/edges/"+url.PathEscape(edgeID)+"/sync",
		map[string]any{"pending": pending, "last_event_seq": lastEventSeq},
		http.StatusOK, http.StatusNoContent)
}

func (c *Client) SubmitMission(ctx context.Context, mission Mission) (*MissionView, error) {
	var out MissionView
	if err := c.doMissionJSON(ctx, http.MethodPost, "/missions", mission, &out, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetMission(ctx context.Context, missionID string) (*MissionView, error) {
	var out MissionView
	if err := c.doMissionJSON(ctx, http.MethodGet, "/missions/"+url.PathEscape(missionID), nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListMissionsPage(ctx context.Context, filter MissionFilter) (MissionPage, error) {
	query := url.Values{}
	if filter.EdgeID != "" {
		query.Set("edge_id", filter.EdgeID)
	}
	if filter.Status != "" {
		query.Set("status", filter.Status)
	}
	if filter.Limit > 0 {
		query.Set("limit", strconv.Itoa(filter.Limit))
	}
	if filter.Cursor != "" {
		query.Set("cursor", filter.Cursor)
	}
	path := "/missions"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out MissionPage
	if err := c.doMissionJSON(ctx, http.MethodGet, path, nil, &out, http.StatusOK); err != nil {
		return MissionPage{}, err
	}
	return out, nil
}

func (c *Client) ListMissions(ctx context.Context, filter MissionFilter) ([]MissionView, error) {
	page, err := c.ListMissionsPage(ctx, filter)
	if err != nil {
		return nil, err
	}
	return page.Missions, nil
}

func (c *Client) AssignMission(ctx context.Context, missionID string, request AssignmentRequest) (*MissionView, error) {
	var out MissionView
	path := "/missions/" + url.PathEscape(missionID) + "/assignments"
	if err := c.doMissionJSON(ctx, http.MethodPost, path, request, &out, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SubmitMissionOperation(ctx context.Context, missionID string, operation MissionOperation) (*MissionView, error) {
	var out MissionView
	path := "/missions/" + url.PathEscape(missionID) + "/operations"
	if err := c.doMissionJSON(ctx, http.MethodPost, path, operation, &out, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetMissionOperations(ctx context.Context, missionID string) ([]MissionOperation, error) {
	var out MissionOperationList
	path := "/missions/" + url.PathEscape(missionID) + "/operations"
	if err := c.doMissionJSON(ctx, http.MethodGet, path, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return out.Operations, nil
}

func (c *Client) CancelMission(ctx context.Context, missionID, reason string) (*MissionView, error) {
	var out MissionView
	path := "/missions/" + url.PathEscape(missionID) + "/cancel"
	if err := c.doMissionJSON(ctx, http.MethodPost, path, CancelRequest{Reason: reason}, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ReportEffect(ctx context.Context, missionID string, report EffectReport) (*EffectReportResult, error) {
	var out EffectReportResult
	path := "/missions/" + url.PathEscape(missionID) + "/effect-reports"
	if err := c.doMissionJSON(ctx, http.MethodPost, path, report, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetEdge(ctx context.Context, edgeID string) (*Edge, error) {
	var out Edge
	if err := c.doMissionJSON(ctx, http.MethodGet, "/edges/"+url.PathEscape(edgeID), nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) doMissionJSON(ctx context.Context, method, path string, payload, out any, okStatuses ...int) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ok := false
	if len(okStatuses) == 0 {
		ok = resp.StatusCode >= 200 && resp.StatusCode < 300
	} else {
		for _, status := range okStatuses {
			if resp.StatusCode == status {
				ok = true
				break
			}
		}
	}
	if !ok {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		var apiErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(raw, &apiErr) == nil && apiErr.Code != "" && apiErr.Message != "" {
			return &APIError{StatusCode: resp.StatusCode, Code: apiErr.Code, Message: apiErr.Message}
		}
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(raw)}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload, out any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(snippet)}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) do(ctx context.Context, method, path string, payload any, okStatuses ...int) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for _, ok := range okStatuses {
		if resp.StatusCode == ok {
			return nil
		}
	}
	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return &HTTPError{StatusCode: resp.StatusCode, Body: string(snippet)}
}


