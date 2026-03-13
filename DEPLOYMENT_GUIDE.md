# Ant-Chrome 部署与配置指南

本指南将帮助你快速部署 Ant-Chrome 并集成 `fingerprint-chromium` 引擎。

## 1. 自动化说明
我们已经实现了**浏览器路径自动检测**功能。你只需要按照下方的目录结构放置文件，程序在启动时会自动识别并配置 `chrome.exe` 的路径，无需在系统设置中手动填写。

## 2. 目录结构
建议的部署目录结构如下：

```text
Ant-Chrome/
├── news-platform.exe (由 wails build 生成的主程序)
├── config.yaml (自动生成的配置文件)
├── chrome/ (新建此文件夹，用于存放浏览器引擎)
│   ├── chrome.exe (核心二进制文件)
│   ├── chrome_proxy.exe
│   ├── locales/
│   ├── resources.pak
│   └── ... (其他 Chromium 依赖文件)
└── profiles/ (自动创建，用于存放各实例的用户数据)
```

## 3. 如何集成 fingerprint-chromium
根据你提供的 [fingerprint-chromium 文档](https://github.com/adryfish/fingerprint-chromium/blob/142.0.7444.175/README-ZH.md)，请按以下步骤操作：

1. **下载**: 访问 GitHub Release 页面，下载适合 Windows 的 **ZIP** 版本（例如 `chrome-win.zip`）。
2. **解压**: 将 ZIP 包中的所有内容解压。
3. **复制**: 将解压出的所有文件（确保包含 `chrome.exe`）复制到 Ant-Chrome 根目录下的 `chrome/` 文件夹中。

## 4. 指纹功能使用
Ant-Chrome 已经完美对接了 `fingerprint-chromium` 的命令行参数：

- **自动种子**: 如果你在实例配置中没有指定 `--fingerprint` 参数，Ant-Chrome 会根据 `profileId` 自动生成一个固定的 32 位整数种子，确保每个实例都有唯一且稳定的指纹。
- **自定义指纹**: 你可以在“实例列表 -> 编辑 -> 指纹参数”中手动添加如下参数：
  - `--fingerprint=123456` (手动指定种子)
  - `--fingerprint-platform=windows` (模拟系统)
  - `--fingerprint-brand=Edge` (模拟浏览器品牌)

## 5. 常见问题
- **无法启动**: 请确保 `chrome/` 文件夹下包含完整的 Chromium 运行环境（不仅仅是 `chrome.exe`，还有 `.pak` 文件和 `locales` 目录）。
- **路径映射**: 如果你想使用非标准路径，依然可以在主程序的“系统设置”中手动修改。
