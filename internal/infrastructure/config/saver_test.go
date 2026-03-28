package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestSavePersistsProxyAddress(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})
	viper.Reset()

	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")

	cfg := &Config{
		APIAddress:   "http://127.0.0.1:9090",
		Secret:       "test-secret",
		TestURL:      "http://www.gstatic.com/generate_204",
		Timeout:      5000,
		ProxyAddress: "http://127.0.0.1:9999",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	configFile := filepath.Join(tempHome, ".mihosh", "config.yaml")
	reader := viper.New()
	reader.SetConfigFile(configFile)
	reader.SetConfigType("yaml")
	if err := reader.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig() returned error: %v", err)
	}

	if got := reader.GetString("proxy_address"); got != cfg.ProxyAddress {
		t.Fatalf("proxy_address not persisted, got %q want %q", got, cfg.ProxyAddress)
	}
}
