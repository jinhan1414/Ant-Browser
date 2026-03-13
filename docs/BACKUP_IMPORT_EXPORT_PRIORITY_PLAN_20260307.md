# Backup/Import-Export Priority Plan (2026-03-07)

## P0 - 必须先完成（功能正确性）

1. 导出内容与加载入口对齐
- 导出包必须包含：
  - `payload/system/config.yaml`
  - `payload/system/proxies.yaml`（存在时）
  - `payload/app/database/app.db`（存在时）
  - `payload/app/data/*`（排除运行中的 db 文件冲突）
  - `payload/browser/user-data/*`
  - `payload/browser/cores/chrome/*`
  - `manifest.json`
- 后端入口：
  - `BackupExportPackage()`
  - `BackupImportPackage(resetFirst bool)`

2. 加载模式明确化（全量 / 增量）
- 前端弹窗必须让用户二选一：
  - 全量覆盖（先初始化，再导入）
  - 增量导入（不初始化，按规则判重）
- 对应后端参数：
  - `resetFirst=true` => 全量
  - `resetFirst=false` => 增量

3. 全量覆盖行为
- 全量模式下先执行初始化：
  - 清业务表
  - 清数据目录（保留运行数据库必要文件）
  - 重写默认配置
- 再导入备份包内容。

4. 增量判重行为
- 配置层：
  - 按 `id/path/url/userDataDir/code` 等关键字段判重。
- 数据库层：
  - `insertSafe` 语句确保重复数据跳过，不中断整体流程。
- 文件层：
  - 同名同内容 => `skipped`
  - 同名不同内容 => `conflicts`
  - 不自动覆盖用户现有文件。

## P1 - 稳定性与可维护性

1. 维护操作互斥
- 初始化/导出/导入必须通过互斥锁串行执行，避免并发污染。

2. 导入导出后统一重载
- 重新加载配置、DAO、浏览器管理器、代理管理器、测速调度器与 launch code 缓存。

3. 错误透明
- 所有失败场景返回可读错误：
  - 缺失 manifest
  - 包格式不兼容
  - 数据库附加失败
  - 文件树同步失败

## P2 - 验证与回归

1. 单测
- `backupEnsureZipSuffix`
- 配置合并判重
- 文件同步冲突/覆盖计数

2. 手工验证用例
- 全量：空目录恢复 + 非空目录覆盖恢复
- 增量：重复 profile/proxy/core/bookmark 跳过
- 异常：取消文件选择、损坏 zip、缺 manifest

## 交付标准

- 用户可在设置页完成：
  - 初始化
  - 导出
  - 加载（含全量/增量确认）
- “导出内容可加载”闭环可复现，且统计字段（imported/skipped/conflicts）可见。

