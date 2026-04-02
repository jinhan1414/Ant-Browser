# AGENTS

## 记住动作

- 当用户明确提出“记住”某个方案或结论时，必须执行“落地记住”。
- “落地记住”指将方案写入本文件或对应 skill 文档，并在回复中给出落地文件路径与可检索关键词。
- 本动作的全局可检索关键词：`记住动作`、`落地记住`、`写入AGENTS`。

## 打包分发规则

- 可检索关键词：`便携压缩包分发`、`不走NSIS`、`portable zip`、`Windows分发`。
- 当用户要求“打包并分发”且未明确要求安装器时，默认采用 Windows 便携压缩包分发，不使用 NSIS 安装包。
- 标准产物命名为 `publish/output/AntBrowser-<version>-windows-amd64-portable.zip`。
- 便携压缩包的标准内容为：
  - `ant-chrome.exe`
  - `config.yaml`，来源于 `publish/config.init.yaml`
  - `bin/xray.exe`
  - `bin/sing-box.exe`
  - 空 `data/` 目录
- 若仓库存在 `chrome/` 目录，则随压缩包一并分发；若不存在，不阻塞便携包打包。
- 标准流程为先执行 `wails build` 生成 `build/bin/ant-chrome.exe`，再组装便携目录并压缩为 zip。
