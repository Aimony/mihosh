package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aimony/mihomo-cli/internal/infrastructure/config"
)

// Client mihomo API 客户端
type Client struct {
	baseURL    string
	secret     string
	httpClient *http.Client
}

// NewClient 创建新的 API 客户端
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.APIAddress,
		secret:  cfg.Secret,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Millisecond,
		},
	}
}

// DoRequest 执行 HTTP 请求（导出供 endpoints 使用）
func (c *Client) DoRequest(method, path string, body interface{}) ([]byte, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if c.secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.secret)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("API 请求失败: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
