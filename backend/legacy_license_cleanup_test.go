package backend

import (
	"ant-chrome/backend/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigRemovesLegacyLicenseStateFile(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.yaml")

	cfg := config.DefaultConfig()
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, legacyLicenseStateFilename), []byte(`{"maxProfileLimit":999}`), 0o644); err != nil {
		t.Fatalf("写入旧额度状态失败: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 失败: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadConfig 返回空配置")
	}
	if _, err := os.Stat(filepath.Join(root, legacyLicenseStateFilename)); !os.IsNotExist(err) {
		t.Fatalf("LoadConfig 应删除旧额度状态文件: err=%v", err)
	}
}

func TestLoadConfigMissingFileDoesNotCreateLegacyLicenseState(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.yaml")

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 失败: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadConfig 返回空配置")
	}
	if _, err := os.Stat(filepath.Join(root, legacyLicenseStateFilename)); !os.IsNotExist(err) {
		t.Fatalf("LoadConfig 不应生成旧额度状态文件: err=%v", err)
	}
}
