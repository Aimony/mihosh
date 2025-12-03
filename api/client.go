package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aimony/mihomo-cli/config"
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

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
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

// GetProxies 获取所有代理信息
func (c *Client) GetProxies() (map[string]Proxy, error) {
	data, err := c.doRequest("GET", "/proxies", nil)
	if err != nil {
		return nil, err
	}

	var resp ProxiesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Proxies, nil
}

// GetProxy 获取指定代理信息
func (c *Client) GetProxy(name string) (*Proxy, error) {
	data, err := c.doRequest("GET", "/proxies/"+name, nil)
	if err != nil {
		return nil, err
	}

	var proxy Proxy
	if err := json.Unmarshal(data, &proxy); err != nil {
		return nil, err
	}

	return &proxy, nil
}

// SelectProxy 选择代理节点
func (c *Client) SelectProxy(group, proxy string) error {
	req := SelectProxyRequest{Name: proxy}
	_, err := c.doRequest("PUT", "/proxies/"+group, req)
	return err
}

// TestProxyDelay 测试单个代理延迟
func (c *Client) TestProxyDelay(name, testURL string, timeout int) (int, error) {
	path := fmt.Sprintf("/proxies/%s/delay?url=%s&timeout=%d", name, testURL, timeout)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}

	var resp DelayTestResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, err
	}

	return resp.Delay, nil
}

// TestGroupDelay 测试策略组内所有节点延迟
func (c *Client) TestGroupDelay(group, testURL string, timeout int) error {
	path := fmt.Sprintf("/group/%s/delay?url=%s&timeout=%d", group, testURL, timeout)
	_, err := c.doRequest("GET", path, nil)
	return err
}

// GetGroups 获取所有策略组
func (c *Client) GetGroups() (map[string]Group, error) {
	proxies, err := c.GetProxies()
	if err != nil {
		return nil, err
	}

	groups := make(map[string]Group)
	for name, proxy := range proxies {
		// 策略组有 all 字段
		if len(proxy.All) > 0 {
			groups[name] = Group{
				Name:    name,
				Type:    proxy.Type,
				Now:     proxy.Now,
				All:     proxy.All,
				History: proxy.History,
			}
		}
	}

	return groups, nil
}

// GetConnections 获取连接信息
func (c *Client) GetConnections() (*ConnectionsResponse, error) {
	data, err := c.doRequest("GET", "/connections", nil)
	if err != nil {
		return nil, err
	}

	var resp ConnectionsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
