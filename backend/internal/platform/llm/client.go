package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sercherai/backend/internal/platform/config"
)

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float32       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) ChatCompletion(messages []ChatMessage) (string, error) {
	if c.cfg.LLMAPIKey == "" {
		// Fallback to mock for testing if no key provided
		return "【系统提示】由于未配置 LLM_API_KEY，当前显示为 Mock 返回。请在环境变量中配置 API Key 以启用真实推演。", nil
	}

	url := c.cfg.LLMBaseURL
	if !strings.HasSuffix(url, "/chat/completions") {
		url = strings.TrimRight(url, "/") + "/chat/completions"
	}

	reqBody := ChatCompletionRequest{
		Model:       c.cfg.LLMModelName,
		Messages:    messages,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.LLMAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w (body: %s)", err, string(bodyBytes))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("llm api error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}
