package copilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultCopilotBase = "https://api.githubcopilot.com"
	defaultGitHubBase  = "https://api.github.com"
	integrationID      = "copilot-developer-cli"
	maxResponseBytes   = 4 << 20
)

var ErrNoLogin = errors.New("no GitHub login reaches Copilot")

type Model struct {
	ID     string  `json:"id"`
	Policy *string `json:"policy"`
	Picker bool    `json:"picker"`
}

type APIModel struct {
	ID                 string `json:"id"`
	ModelPickerEnabled bool   `json:"model_picker_enabled"`
	Policy             struct {
		State string `json:"state"`
	} `json:"policy"`
}

type Client struct {
	http       *http.Client
	token      string
	copilotURL string
	githubURL  string
}

func New(timeout time.Duration) (*Client, error) {
	token, err := Token()
	if err != nil {
		return nil, err
	}
	return &Client{
		http:       &http.Client{Timeout: timeout},
		token:      token,
		copilotURL: baseURL("VLOOP_COPILOT_API_URL", defaultCopilotBase),
		githubURL:  baseURL("VLOOP_GITHUB_API_URL", defaultGitHubBase),
	}, nil
}

func Token() (string, error) {
	for _, key := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		if token := os.Getenv(key); token != "" {
			return token, nil
		}
	}
	out, err := exec.Command("gh", "auth", "token").Output()
	if err == nil {
		if token := strings.TrimSpace(string(out)); token != "" {
			return token, nil
		}
	}
	return "", ErrNoLogin
}

func baseURL(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimRight(value, "/")
	}
	return fallback
}

func (c *Client) Models(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.copilotURL+"/models", nil)
	if err != nil {
		return nil, c.redact(err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Copilot-Integration-Id", integrationID)
	var response struct {
		Data []APIModel `json:"data"`
	}
	if err := c.do(req, &response); err != nil {
		return nil, c.redact(err)
	}
	models := make([]Model, 0, len(response.Data))
	for _, entry := range response.Data {
		model := Model{ID: entry.ID, Picker: entry.ModelPickerEnabled}
		switch entry.Policy.State {
		case "enabled", "disabled":
			state := entry.Policy.State
			model.Policy = &state
		}
		models = append(models, model)
	}
	return models, nil
}

func (c *Client) Plan(ctx context.Context) (*string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.githubURL+"/copilot_internal/user", nil)
	if err != nil {
		return nil, c.redact(err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	var response struct {
		Plan *string `json:"copilot_plan"`
	}
	if err := c.do(req, &response); err != nil {
		return nil, c.redact(err)
	}
	if response.Plan == nil || *response.Plan == "" {
		return nil, nil
	}
	return response.Plan, nil
}

func (c *Client) do(req *http.Request, target any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("cannot read JSON response: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("cannot read JSON response: multiple documents")
		}
		return fmt.Errorf("cannot read JSON response: %w", err)
	}
	return nil
}

func (c *Client) redact(err error) error {
	if c.token == "" {
		return err
	}
	return errors.New(strings.ReplaceAll(err.Error(), c.token, "[redacted]"))
}
