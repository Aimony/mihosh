package api

import (
	"encoding/json"
	"fmt"

	"github.com/aimony/mihomo-cli/internal/domain/model"
)

// GetProxies 获取所有代理信息
func (c *Client) GetProxies() (map[string]model.Proxy, error) {
	data, err := c.DoRequest("GET", "/proxies", nil)
	if err != nil {
		return nil, err
	}

	var resp model.ProxiesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Proxies, nil
}

// GetProxy 获取指定代理信息
func (c *Client) GetProxy(name string) (*model.Proxy, error) {
	data, err := c.DoRequest("GET", "/proxies/"+name, nil)
	if err != nil {
		return nil, err
	}

	var proxy model.Proxy
	if err := json.Unmarshal(data, &proxy); err != nil {
		return nil, err
	}

	return &proxy, nil
}

// SelectProxy 选择代理节点
func (c *Client) SelectProxy(group, proxy string) error {
	req := model.SelectProxyRequest{Name: proxy}
	_, err := c.DoRequest("PUT", "/proxies/"+group, req)
	return err
}

// TestProxyDelay 测试单个代理延迟
func (c *Client) TestProxyDelay(name, testURL string, timeout int) (int, error) {
	path := fmt.Sprintf("/proxies/%s/delay?url=%s&timeout=%d", name, testURL, timeout)
	data, err := c.DoRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}

	var resp model.DelayTestResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, err
	}

	return resp.Delay, nil
}

// TestGroupDelay 测试策略组内所有节点延迟
func (c *Client) TestGroupDelay(group, testURL string, timeout int) error {
	path := fmt.Sprintf("/group/%s/delay?url=%s&timeout=%d", group, testURL, timeout)
	_, err := c.DoRequest("GET", path, nil)
	return err
}

// GetGroups 获取所有策略组
func (c *Client) GetGroups() (map[string]model.Group, error) {
	proxies, err := c.GetProxies()
	if err != nil {
		return nil, err
	}

	groups := make(map[string]model.Group)
	for name, proxy := range proxies {
		// 策略组有 all 字段
		if len(proxy.All) > 0 {
			groups[name] = model.Group{
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
func (c *Client) GetConnections() (*model.ConnectionsResponse, error) {
	data, err := c.DoRequest("GET", "/connections", nil)
	if err != nil {
		return nil, err
	}

	var resp model.ConnectionsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
