package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSystemctlStatusParsesCGroupProcessLine(t *testing.T) {
	output := `● mihomo.service - mihomo Daemon, Another Clash Kernel.
     Loaded: loaded (/etc/systemd/system/mihomo.service; enabled; preset: enabled)
     Active: active (running) since Sun 2026-02-08 12:33:48 CST; 1 month 23 days ago
   Main PID: 902928 (mihomo)
      Tasks: 11 (limit: 4291)
     Memory: 51.3M (peak: 100.0M swap: 4.1M swap peak: 10.6M)
        CPU: 24min 16.624s
     CGroup: /system.slice/mihomo.service
             └─902928 /home/ubuntu/Apps/local/mihomo/mihomo -d /home/ubuntu/Apps/local/mihomo
`

	assert.Equal(
		t,
		"/home/ubuntu/Apps/local/mihomo/mihomo -d /home/ubuntu/Apps/local/mihomo",
		parseSystemctlStatus(output),
	)
}

func TestGetMihomoConfigPathReturnsFallbackHintWhenAutoDiscoveryFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("APPDATA", "")

	originalRunner := systemctlStatusRunner
	systemctlStatusRunner = func() ([]byte, error) {
		return []byte("mihomo.service could not be found"), nil
	}
	t.Cleanup(func() {
		systemctlStatusRunner = originalRunner
	})

	path, err := GetMihomoConfigPath()
	require.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "sudo systemctl status mihomo")
	assert.Contains(t, err.Error(), "-d")
	assert.Contains(t, err.Error(), "config.yaml")
	assert.Contains(t, err.Error(), "config.yml")
}

func TestGetMihomoConfigPathFromProcessFindsConfigInSystemctlDirectory(t *testing.T) {
	homeDir := t.TempDir()
	serviceDir := t.TempDir()
	configPath := serviceDir + string(os.PathSeparator) + "config.yaml"

	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("APPDATA", "")

	err := os.WriteFile(configPath, []byte("mixed-port: 7890\n"), 0644)
	require.NoError(t, err)

	originalRunner := systemctlStatusRunner
	systemctlStatusRunner = func() ([]byte, error) {
		return []byte("             └─902928 /home/ubuntu/Apps/local/mihomo/mihomo -d " + serviceDir), nil
	}
	t.Cleanup(func() {
		systemctlStatusRunner = originalRunner
	})

	path, err := GetMihomoConfigPathFromProcess()
	require.NoError(t, err)
	assert.Equal(t, configPath, path)
}
