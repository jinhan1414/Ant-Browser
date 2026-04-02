package backend

import (
	"ant-chrome/backend/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestReloadConfigLoadsFromDisk(t *testing.T) {
	root := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.App.Name = "Reload-Test-App"
	if err := cfg.Save(filepath.Join(root, "config.yaml")); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}

	app := NewApp(root)
	app.config = config.DefaultConfig()

	if err := app.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig 失败: %v", err)
	}

	if app.config == nil {
		t.Fatalf("ReloadConfig 后 config 为空")
	}
	if app.config.App.Name != "Reload-Test-App" {
		t.Fatalf("ReloadConfig 未生效，got=%q", app.config.App.Name)
	}
}

func TestReloadConfigRemovesLegacyLicenseStateFile(t *testing.T) {
	root := t.TempDir()

	cfg := config.DefaultConfig()
	if err := cfg.Save(filepath.Join(root, "config.yaml")); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".ant-license.json"), []byte(`{"maxProfileLimit":999,"usedCdKeys":["A1"]}`), 0o644); err != nil {
		t.Fatalf("写入本机额度状态失败: %v", err)
	}

	app := NewApp(root)
	app.config = config.DefaultConfig()

	if err := app.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig 失败: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".ant-license.json")); !os.IsNotExist(err) {
		t.Fatalf("ReloadConfig 应删除旧额度状态文件: err=%v", err)
	}
}

func TestReloadConfigDoesNotExposeLegacyQuotaFields(t *testing.T) {
	root := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.App.Name = "No-License-State"
	if err := cfg.Save(filepath.Join(root, "config.yaml")); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}
	stateJSON := `{"maxProfileLimit":999,"usedCdKeys":["GITHUB_STAR_REWARD","ANT-AAAA-BBBB-CCCC-DDDD-EEEEEEEE"]}`
	if err := os.WriteFile(filepath.Join(root, ".ant-license.json"), []byte(stateJSON), 0o644); err != nil {
		t.Fatalf("写入旧额度状态失败: %v", err)
	}

	app := NewApp(root)
	app.config = config.DefaultConfig()

	if err := app.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig 失败: %v", err)
	}

	data, err := yaml.Marshal(app.config)
	if err != nil {
		t.Fatalf("序列化配置失败: %v", err)
	}
	text := string(data)
	if strings.Contains(text, "max_profile_limit:") {
		t.Fatalf("ReloadConfig 后配置不应暴露 max_profile_limit: %s", text)
	}
	if strings.Contains(text, "used_cd_keys:") {
		t.Fatalf("ReloadConfig 后配置不应暴露 used_cd_keys: %s", text)
	}
	if app.config.App.Name != "No-License-State" {
		t.Fatalf("ReloadConfig 不应影响正常配置字段: got=%q", app.config.App.Name)
	}
}
