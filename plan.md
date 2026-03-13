# 指纹浏览器功能规划

> 状态标志：`[ ]` 未开始 · `[~]` 进行中 · `[x]` 已完成

---

## P0 — 核心能力缺口

### [ ] P0-1 指纹参数 UI 化

**问题**：目前指纹配置只能手填原始 CLI 参数（如 `--fingerprint-brand=Chrome`），没有可视化表单，用户容易填错，也无法感知有哪些可配置项。

**目标**：在实例编辑页提供结构化的指纹配置面板，覆盖以下维度：

| 维度 | 参数示例 | 说明 |
|------|----------|------|
| 浏览器品牌 | `--fingerprint-brand=Chrome/Edge/Firefox` | UA 中的浏览器名称 |
| 操作系统 | `--fingerprint-platform=windows/mac/linux` | UA 中的平台信息 |
| 语言 | `--lang=zh-CN` | `navigator.language` |
| 时区 | `--timezone=Asia/Shanghai` | `Intl.DateTimeFormat` |
| 分辨率 | `--window-size=1920,1080` | `screen.width/height` |
| Canvas 噪声 | `--fingerprint-canvas-noise=true` | Canvas 指纹随机化 |
| WebGL 供应商 | `--fingerprint-webgl-vendor=Intel` | WebGL 渲染器信息 |
| 字体列表 | `--fingerprint-fonts=...` | 可用字体集合 |

**实现要点**：
- `BrowserEditPage.tsx` 新增"指纹配置"Tab，每个维度对应一个 Select/Input
- 保存时自动序列化为 `fingerprintArgs` 数组，与现有存储格式兼容
- 保留"高级模式"原始参数编辑入口，供高级用户使用

---

### [ ] P0-2 代理真实连通性验证

**问题**：当前 `TestConnectivity` 只做 TCP 握手（直连 server:port），不经过 Xray 桥接，无法验证 vmess/vless 节点是否真实可用，延迟数据也不准确。

**目标**：测速时真正走代理链路，通过代理请求一个固定 HTTP 端点，测量端到端延迟。

**实现要点**：
- 后端新增 `TestProxyReal(proxyId)` 接口：
  1. 调用 `EnsureBridge` 启动 Xray 获取 socks5 地址
  2. 通过该 socks5 地址发起 HTTP GET `https://www.gstatic.com/generate_204`（Google 204，轻量无内容）
  3. 记录从发起请求到收到响应的耗时
  4. 返回延迟 ms 和成功/失败状态
- 前端代理池"测速"按钮调用新接口，区分"TCP 可达"和"代理可用"两种状态

---

### [ ] P0-3 实例列表标签筛选

**问题**：实例配置里已有 `tags` 字段，但列表页没有按标签过滤的入口，实例数量多时管理困难。

**目标**：列表页顶部增加标签筛选栏，支持多标签组合过滤。

**实现要点**：
- `BrowserListPage.tsx` 顶部增加标签 Badge 列表，点击切换选中状态
- 支持多选（AND 或 OR 逻辑，默认 OR）
- 标签列表从当前所有实例的 tags 动态聚合，去重排序
- 筛选状态与关键词搜索联动

---

## P1 — 重要体验问题

### [ ] P1-1 实例批量操作

**问题**：代理池已支持批量删除，但实例列表没有批量操作能力，管理大量实例时效率低。

**目标**：实例列表支持批量选择，提供批量启动、批量停止、批量删除操作。

**实现要点**：
- 列表第一列增加 checkbox，表头全选（含 indeterminate 状态）
- 有选中时顶部操作栏出现"批量启动 (N)"、"批量停止 (N)"、"批量删除 (N)"按钮
- 批量启动：串行启动（避免端口分配竞争），每个间隔 200ms
- 批量停止：并行停止
- 批量删除：二次确认弹窗，删除前自动停止运行中的实例

---

### [ ] P1-2 实例配置导入/导出

**问题**：没有配置备份和迁移能力，换机器或分发配置时只能手动复制 `config.yaml`。

**目标**：支持将选中实例（含代理绑定关系）导出为 JSON 文件，并支持从文件导入。

**实现要点**：
- 后端新增 `ExportProfiles(ids []string) string` 返回 JSON 字符串
- 后端新增 `ImportProfiles(json string) error` 解析并合并（proxyId 冲突时提示）
- 前端列表页顶部增加"导出"和"导入"按钮
- 导出格式包含 profiles + 关联的 proxies，不含用户数据目录内容

---

### [ ] P1-3 代理绑定状态可视化

**问题**：实例列表只显示代理名称文字，看不出代理是否在线、是否已建立 Xray 桥接。

**目标**：实例列表的代理列显示实时状态指示器。

**实现要点**：
- 后端 `BrowserProfileList` 返回数据中增加 `proxyStatus` 字段：
  - `none`：无代理
  - `direct`：直连代理（http/socks5）
  - `bridged`：Xray 桥接已建立
  - `error`：桥接失败
- 前端代理列显示对应颜色圆点（灰/绿/蓝/红）+ Tooltip 说明
- 实例运行时每 30s 轮询一次状态

---

### [ ] P1-4 Xray 桥接进程自动重连

**问题**：Xray 桥接进程崩溃后没有自动恢复机制，浏览器实例会静默断网，用户无感知。

**目标**：后台监控桥接进程存活状态，崩溃后自动重启，并通知前端。

**实现要点**：
- `EnsureBridge` 已有存活检测逻辑，需要在 `waitBrowserProcess` goroutine 中定期调用
- 新增后台 goroutine，每 10s 检查所有 `Running=true` 的 bridge：
  - 进程已死 → 自动重启（复用现有 `EnsureBridge` 逻辑）
  - 重启失败 → 发送 `xray:bridge:error` 事件通知前端
- 前端监听事件，在对应实例行显示警告 Badge

---

### [ ] P1-5 启动页自定义

**问题**：默认启动页硬编码在 `connector.go` 中（`ippure.com` + `iplark.com`），无法按实例配置。

**目标**：实例配置中增加"启动页"字段，支持配置多个 URL（每行一个）。

**实现要点**：
- `BrowserProfileConfig` 新增 `StartupUrls []string` 字段
- `BrowserEditPage.tsx` 新增"启动页"Textarea（每行一个 URL）
- `BuildLaunchArgs` 读取 `profile.StartupUrls`，为空时使用全局默认值
- 全局默认启动页在"基础配置"弹窗中可配置

---

## P2 — 进阶功能

### [ ] P2-1 Cookie 管理

**问题**：电商采集场景下需要管理账号登录态，目前无法查看或清除实例的 Cookie，账号切换只能靠重置整个用户数据目录。

**目标**：实例详情页提供 Cookie 管理面板，支持查看、清除、导入/导出。

**实现要点**：
- 通过 Chrome DevTools Protocol（CDP）连接到实例的调试端口（`profile.DebugPort`）
- 调用 `Network.getAllCookies` 获取 Cookie 列表
- 调用 `Network.clearBrowserCookies` 清除全部 Cookie
- 支持按域名过滤，支持导出为 Netscape Cookie 格式（兼容 curl/wget）
- 后端新增 `BrowserGetCookies(profileId)` 和 `BrowserClearCookies(profileId, domain)` 接口

---

### [ ] P2-2 实例数据快照

**问题**：无法备份某个实例的浏览器状态（登录态、书签、扩展配置等），误操作或换机器后数据丢失。

**目标**：支持对实例用户数据目录进行快照（压缩备份）和恢复。

**实现要点**：
- 后端新增 `SnapshotProfile(profileId, snapshotName)` — 将 `userDataDir` 压缩为 zip 存入 `snapshots/` 目录
- 后端新增 `RestoreProfile(profileId, snapshotId)` — 解压覆盖（需先停止实例）
- 后端新增 `ListSnapshots(profileId)` — 列出快照列表（名称、大小、时间）
- 前端实例详情页增加"快照"Tab

---

### [ ] P2-3 Hysteria2 协议支持

**问题**：代理池中已有多个 hysteria2 节点，但当前明确不支持（Xray 不支持该协议），这些节点无法使用。

**目标**：集成 Hysteria2 客户端，支持 hysteria2 协议的代理桥接。

**实现要点**：
- 内核管理中增加"Hysteria2 二进制"配置项（类似 XrayBinaryPath）
- 新增 `HysteriaManager`，参考 `XrayManager` 实现：
  - 生成 hysteria2 JSON 配置（socks5 inbound + hysteria2 outbound）
  - 启动 hysteria2 进程，等待端口就绪
  - 进程复用和存活检测
- `RequiresBridge` 和 `EnsureBridge` 中增加 hysteria2 分支，路由到 `HysteriaManager`
- 前端代理验证提示从"不支持"改为"需要配置 Hysteria2 二进制路径"

---

### [ ] P2-4 代理热切换

**问题**：实例运行中无法更换代理，必须停止实例、修改配置、重新启动，操作繁琐。

**目标**：运行中的实例支持一键切换代理，无需重启。

**实现要点**：
- Chrome 支持通过 CDP `Network.setCustomHeaders` 和系统代理动态修改，但更可靠的方式是：
  - 停止当前实例（保留用户数据目录）
  - 用新代理重新启动（对用户透明，速度快）
- 后端新增 `BrowserInstanceSwitchProxy(profileId, proxyId)` 接口
- 前端实例详情页增加"切换代理"下拉选择器

---

### [ ] P2-5 定时任务

**问题**：电商采集场景下需要定时启动/停止实例，目前只能手动操作。

**目标**：支持为实例配置定时启动和定时停止规则（Cron 表达式）。

**实现要点**：
- `BrowserProfileConfig` 新增 `ScheduleStart string` 和 `ScheduleStop string`（Cron 格式）
- 后端引入轻量 Cron 库（如 `robfig/cron`），应用启动时注册所有定时任务
- 配置变更时动态更新 Cron 任务
- 前端实例编辑页增加"定时配置"区域，提供 Cron 表达式输入和人性化预览（"每天 09:00 启动"）

---

## P3 — 锦上添花

### [ ] P3-1 暗色/亮色主题切换

**问题**：目前只有一套主题，长时间使用视觉疲劳。

**目标**：支持暗色/亮色主题切换，偏好持久化到 localStorage。

**实现要点**：
- CSS 变量已按 `var(--color-*)` 规范定义，只需增加 `[data-theme="light"]` 变量覆盖
- 顶部导航栏增加主题切换按钮（太阳/月亮图标）
- 主题偏好存入 localStorage，刷新后恢复

---

### [ ] P3-2 实例使用统计

**问题**：不知道每个实例的使用频率和时长，无法评估哪些实例在被有效使用。

**目标**：记录每个实例的启动次数、累计运行时长、最后活跃时间。

**实现要点**：
- SQLite 新增 `instance_stats` 表（profile_id, start_count, total_seconds, last_active_at）
- 每次启动/停止时更新统计
- 实例列表增加"使用统计"列（可选显示）
- 仪表盘增加统计卡片

---

### [ ] P3-3 代理流量统计

**问题**：不知道每个代理节点消耗了多少流量，无法监控用量。

**目标**：统计每个代理节点的上行/下行流量。

**实现要点**：
- Xray 支持通过 gRPC API 查询流量统计（需在配置中开启 `stats` 和 `api`）
- 定期轮询 Xray API 获取流量数据，写入 SQLite
- 代理池列表增加"流量"列
- 支持按周期重置统计

---

## 已完成功能备忘

- [x] 浏览器实例完整生命周期管理（创建/启动/停止/重启/复制/删除）
- [x] 代理池 CRUD + YAML 批量导入
- [x] Xray 桥接（vmess/vless），端口二次验证防竞争
- [x] 代理 DNS 配置（Clash dns: YAML 格式解析）
- [x] 代理批量勾选删除
- [x] 内核管理（多内核、版本检测、默认切换）
- [x] 指纹参数原始编辑
- [x] 日志系统（异步、轮转、敏感字段过滤）
- [x] 应用关闭时优雅清理所有子进程
