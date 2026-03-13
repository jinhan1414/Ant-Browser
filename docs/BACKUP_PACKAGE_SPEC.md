# 备份包规范（第一阶段）

本文档定义“初始化/导出/加载”能力的第一阶段产物：导出范围与 ZIP 包格式契约。

## 1. 目标

- 统一导出包格式，避免后续实现出现路径、版本和兼容性歧义。
- 明确“全量内容”范围：系统配置、应用数据、内核数据、浏览器数据。
- 为后续“初始化后加载/不初始化判重加载”提供稳定输入。

## 2. ZIP 目录结构

```text
<backup>.zip
├── manifest.json
└── payload/
    ├── system/
    │   ├── config.yaml
    │   └── proxies.yaml
    ├── app/
    │   ├── data/
    │   ├── database/
    │   │   ├── app.db
    │   │   ├── app.db-wal
    │   │   └── app.db-shm
    │   └── logs/
    └── browser/
        ├── user-data/
        └── cores/
            ├── chrome/
            └── external/<external-id>/
```

说明：
- `database/` 仅在数据库路径不被 `app/data` 覆盖时单独导出。
- `browser/user-data/` 仅在 `browser.user_data_root` 不被 `app/data` 覆盖时导出。
- `proxies.yaml`、`logs/`、`external/*` 按“存在即导出”处理。

## 3. manifest.json 结构

```json
{
  "format": "ant-chrome-full-backup",
  "manifestVersion": 1,
  "createdAt": "2026-03-02T12:00:00Z",
  "app": {
    "name": "Ant Browser",
    "version": "1.0.0"
  },
  "entries": [
    {
      "id": "system_config_main",
      "category": "system_config",
      "entryType": "file",
      "required": true,
      "archivePath": "payload/system/config.yaml"
    }
  ]
}
```

字段规则：
- `format`：固定为 `ant-chrome-full-backup`。
- `manifestVersion`：当前为 `1`，后续破坏性变更必须升级版本。
- `entries`：描述实际导出条目（不包含本机绝对路径）。

## 4. 范围构建规则

构建顺序：
1. 系统配置：`config.yaml`、`proxies.yaml`
2. 应用数据：`data/`
3. 浏览器数据：`browser.user_data_root`
4. 内核数据：`chrome/` + 配置中额外内核路径
5. 数据库补充：`app.db` 及 `-wal/-shm`
6. 日志目录：由 `logging.file_path` 推导

去重策略：
- 如果某条目已被现有目录覆盖，则跳过该条目（避免重复导出）。
- 相同源路径只保留一个条目，`required=true` 优先级更高。

## 5. 当前实现位置

- 范围与 manifest 契约：`internal/backup/spec.go`
- 单测：`internal/backup/spec_test.go`
- App 预览接口：
  - `BackupGetScopeDefinition()`
  - `BackupGetManifestTemplate()`
