package tui

import (
	"github.com/aimony/mihomo-cli/internal/infrastructure/api"
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
