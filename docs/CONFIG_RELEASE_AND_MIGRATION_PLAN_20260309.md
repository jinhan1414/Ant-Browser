# 配置发布与增量迁移方案（2026-03-09）

## 1. 背景与目标

本方案解决以下问题：

1. 发布包误带开发机 `app.db`，导致用户首次安装出现开发数据。
2. 发布包直接覆盖 `config.yaml`，可能破坏用户本地配置。
3. 旧版本 `config.yaml` 缺字段时，运行时行为不稳定。
4. 版本升级时需要保证“只增量补齐，不覆盖用户已有数据”。

目标：

1. 安装包不再包含业务数据库内容。
2. 首次启动自动初始化空库并执行 schema migration。
3. 配置文件仅在首次安装时落盘，升级不覆盖。
4. 老配置自动补齐缺失字段，保留用户已配置值。

---

## 2. 最终策略（已落地）

### 2.1 发布阶段

1. 新增发布专用模板：`publish/config.init.yaml`。
2. 打包时只复制模板配置到 staging，命名为 `config.yaml`。
3. 不再复制 `data/app.db`，仅创建空 `data/` 目录。

对应文件：

- `publish/config.init.yaml`
- `bat/publish.bat`

### 2.2 安装阶段

1. `config.yaml` 仅在安装目录不存在时才复制。
2. 始终创建 `data/` 目录，但不复制 `app.db`。

对应文件：

- `publish/installer.nsi`

### 2.3 运行阶段

1. 启动时继续执行 SQLite `Migrate()`，自动创建/升级表结构。
2. 保留 `migrateToSQLite()` 从配置导入逻辑（按你要求保留导入）。
3. 由于发布模板 `cores/profiles/proxies` 默认为空，首次安装不会导入开发数据。

对应文件：

- `backend/internal/database/sqlite.go`
- `backend/app.go`

### 2.4 老配置补齐

在 `Load(configPath)` 中新增 `normalizeConfig`：

1. 缺失字段自动补默认值。
2. 已存在用户值不覆盖。
3. `max_profile_limit` 继续按 `used_cd_keys` 做保底计算。
4. 列表字段统一补成空切片，避免 `nil` 行为差异。

对应文件：

- `backend/internal/config/config.go`
- `backend/internal/config/config_test.go`

---

## 3. 配置补齐覆盖范围

已补齐的关键字段：

1. `database.type`、`database.sqlite.path`
2. `app.name`、窗口尺寸最小值
3. `runtime.max_memory_mb`、`runtime.gc_percent`
4. `logging` 基础参数与 rotation 参数
5. `logging.interceptor` 缺失时回填默认
6. `browser.user_data_root`
7. `browser.default_fingerprint_args`、`browser.default_launch_args`
8. `browser.default_bookmarks`、`cores`、`proxies`、`profiles` 空切片初始化
9. `launch_server.port` 缺失/非法时补默认

---

## 4. 测试与验收

### 4.1 自动化测试

已新增并通过：

1. `TestLoadBackfillsLegacyConfig`
2. `TestLoadPreservesExplicitConfig`

执行命令：

```bash
go test ./backend/internal/config -v
```

### 4.2 发布前人工检查清单

1. 打包后解压安装包，确认不包含 `data/app.db`。
2. 首次安装后检查安装目录：有 `data/`，无 `app.db`（启动前）。
3. 首次启动后检查：自动生成 `data/app.db`，无开发实例数据。
4. 升级安装覆盖后检查：已有 `config.yaml` 未被覆盖。
5. 用旧版本配置启动，检查缺失字段已被系统兼容处理。

---

## 5. 优先级实施方案

### P0（本次已完成）

1. 发布模板配置与打包脚本切换。
2. 停止打包 `app.db`。
3. 安装器改为 `config.yaml` 仅首次复制。
4. 老配置字段补齐与单测。

验收标准：

1. 首次安装无开发数据污染。
2. 升级安装不覆盖用户配置。
3. 旧配置可稳定启动。

### P1（建议下一阶段）

1. 在 CI 增加“发布产物内容检查”（禁止包含 `app.db`）。
2. 增加端到端升级测试：`vOld -> vNew` 后数据完整性验证。
3. 文档中补充“配置模板维护规范”与“发布检查流程”。

验收标准：

1. 任意发布分支都无法产出带 `app.db` 的安装包。
2. 升级回归测试可自动跑通。

### P2（可选优化）

1. `config.yaml` 引入显式 `config_version`，将来做版本化配置迁移。
2. 为 `normalizeConfig` 增加更细粒度字段级迁移日志。
3. 增加启动诊断页，展示“配置版本/DB schema 版本”。

验收标准：

1. 配置升级路径可追踪、可回放。
2. 用户问题定位成本显著下降。

---

## 6. 当前结论

按本方案后，发布与升级路径已经满足：

1. 不带开发机 DB 数据。
2. 首次启动自动初始化。
3. 升级只做增量迁移，不覆盖用户业务数据。
4. 老配置自动补齐，兼容历史版本。
