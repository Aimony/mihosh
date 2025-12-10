package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	tea "github.com/charmbracelet/bubbletea"
)

// 命令函数
func fetchGroups(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		groups, err := client.GetGroups()
		if err != nil {
			return errMsg(err)
		}
		return groupsMsg(groups)
	}
}

func fetchProxies(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		proxies, err := client.GetProxies()
		if err != nil {
			return errMsg(err)
		}
		return proxiesMsg(proxies)
	}
}

func fetchConnections(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		conns, err := client.GetConnections()
		if err != nil {
			return errMsg(err)
		}
		return connectionsMsg(conns)
	}
}

func selectProxy(client *api.Client, group, proxy string) tea.Cmd {
	return func() tea.Msg {
		if err := client.SelectProxy(group, proxy); err != nil {
			return errMsg(err)
		}
		// 需要同时刷新groups和proxies,确保✓标记更新
		return tea.Batch(fetchGroups(client), fetchProxies(client))()
	}
}

func testProxy(client *api.Client, name, testURL string, timeout int) tea.Cmd {
	return func() tea.Msg {
		delay, err := client.TestProxyDelay(name, testURL, timeout)
		if err != nil {
			return testDoneMsg{name: name, delay: 0, err: err}
		}
		return testDoneMsg{name: name, delay: delay, err: nil}
	}
}

func testGroup(client *api.Client, group, testURL string, timeout int) tea.Cmd {
	return func() tea.Msg {
		if err := client.TestGroupDelay(group, testURL, timeout); err != nil {
			return errMsg(err)
		}
		return testDoneMsg{}
	}
}

// testAllProxies 逐个测速所有节点
func testAllProxies(client *api.Client, proxies []string, testURL string, timeout int) tea.Cmd {
	return func() tea.Msg {
		var cmds []tea.Cmd
		for _, proxyName := range proxies {
			cmds = append(cmds, testProxy(client, proxyName, testURL, timeout))
		}
		return tea.Batch(cmds...)()
	}
}

// closeConnection 关闭单个连接
func closeConnection(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := client.CloseConnection(id); err != nil {
			return errMsg(err)
		}
		return connectionClosedMsg{id: id}
	}
}

// closeAllConnections 关闭所有连接
func closeAllConnections(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		if err := client.CloseAllConnections(); err != nil {
			return errMsg(err)
		}
		return allConnectionsClosedMsg{}
	}
}

// ipInfoMsg IP信息响应消息
type ipInfoMsg struct {
	info *model.IPInfo
	err  error
}

// fetchIPInfo 获取IP地理位置信息
func fetchIPInfo(ip string) tea.Cmd {
	return func() tea.Msg {
		if ip == "" {
			return ipInfoMsg{nil, nil}
		}

		client := &http.Client{Timeout: 5 * time.Second}
		url := fmt.Sprintf("https://api.ip.sb/geoip/%s", ip)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return ipInfoMsg{nil, err}
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Accept", "*/*")

		resp, err := client.Do(req)
		if err != nil {
			return ipInfoMsg{nil, err}
		}
		defer resp.Body.Close()

		var info model.IPInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return ipInfoMsg{nil, err}
		}

		return ipInfoMsg{&info, nil}
	}
}
