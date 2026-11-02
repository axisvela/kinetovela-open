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
	"strings"
	"time"
)

// Version is the SDK semantic version, synchronized with core releases.
const Version = "1.0.0"

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

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return "kinetovela: unexpected status " + http.StatusText(e.StatusCode)
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
