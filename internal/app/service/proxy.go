package service

import (
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
)

// ProxyService 代理管理服务
type ProxyService struct {
	client  *api.Client
	testURL string
	timeout int
}

// NewProxyService 创建代理服务
func NewProxyService(client *api.Client, testURL string, timeout int) *ProxyService {
	return &ProxyService{
		client:  client,
		testURL: testURL,
		timeout: timeout,
	}
}

// GetGroups 获取所有策略组
func (s *ProxyService) GetGroups() (map[string]model.Group, error) {
	return s.client.GetGroups()
}

// GetProxies 获取所有代理
func (s *ProxyService) GetProxies() (map[string]model.Proxy, error) {
	return s.client.GetProxies()
}

// SelectProxy 选择代理节点
func (s *ProxyService) SelectProxy(group, proxy string) error {
	return s.client.SelectProxy(group, proxy)
}

// TestProxyDelay 测试单个代理延迟
func (s *ProxyService) TestProxyDelay(name string) (int, error) {
	return s.client.TestProxyDelay(name, s.testURL, s.timeout)
}

// TestGroupDelay 测试策略组内所有节点延迟
func (s *ProxyService) TestGroupDelay(group string) error {
	return s.client.TestGroupDelay(group, s.testURL, s.timeout)
}

// TestAllProxies 批量测试代理延迟（返回每个代理的测试结果）
func (s *ProxyService) TestAllProxies(proxies []string) map[string]int {
	results := make(map[string]int)
	for _, proxy := range proxies {
		delay, err := s.TestProxyDelay(proxy)
		if err != nil {
			results[proxy] = 0 // 失败时记录为 0
		} else {
			results[proxy] = delay
		}
	}
	return results
}
