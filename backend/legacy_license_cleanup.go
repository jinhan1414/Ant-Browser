package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const legacyLicenseStateFilename = ".ant-license.json"

func legacyLicenseStatePath(configPath string) string {
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		return legacyLicenseStateFilename
	}
	dir := filepath.Dir(configPath)
	if dir == "." || dir == "" {
		if cwd, err := os.Getwd(); err == nil {
			dir = cwd
		}
	}
	return filepath.Join(dir, legacyLicenseStateFilename)
}

func removeLegacyLicenseState(configPath string) error {
	statePath := legacyLicenseStatePath(configPath)
	err := os.Remove(statePath)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("删除旧额度状态文件失败: %w", err)
}
