# Go-Wails 桌面应用脚手架文档

> 基于 Wails + React + TypeScript 的现代化桌面应用开发脚手架

## 目录

- [项目概述](#项目概述)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [核心功能](#核心功能)
- [组件库](#组件库)
- [数据管理](#数据管理)
- [开发指南](#开发指南)
- [构建部署](#构建部署)
- [专项方案](#专项方案)

---

## 项目概述

这是一个功能完整的桌面应用脚手架，提供了：

- ✅ 完整的项目结构和开发规范
- ✅ 24+ 个常用 UI 组件
- ✅ 数据管理模块（CRUD + 分页）
- ✅ 主题切换（亮色/暗色）
- ✅ 日志系统
- ✅ 配置管理
- ✅ 响应式布局

---

## 技术栈

### 后端
- **Go 1.21+** - 主要编程语言
- **Wails v2** - 桌面应用框架
- **SQLite** - 嵌入式数据库
- **modernc.org/sqlite** - 纯 Go SQLite 驱动

### 前端
- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **TailwindCSS** - 样式框架
- **Zustand** - 状态管理
- **React Router** - 路由管理
- **Lucide React** - 图标库

---

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- Wails CLI v2

### 安装依赖

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装项目依赖
go mod tidy
cd frontend && npm install
```

### 开发模式

```bash
# Windows
bat\dev.bat

# 或直接使用 Wails
wails dev
```

### 构建应用

```bash
# Windows
bat\build.bat

# 或直接使用 Wails
wails build
```

构建产物位于 `build/bin/` 目录。

---

## 项目结构

```
.
├── bat/                    # Windows 批处理脚本
│   ├── build.bat          # 构建脚本
│   ├── dev.bat            # 开发脚本
│   └── generate-bindings.bat  # 生成绑定脚本
├── docs/                   # 项目文档
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── config/        # 配置文件
│   │   ├── modules/       # 功能模块
│   │   │   ├── dashboard/ # 控制台
│   │   │   ├── data/      # 数据管理
│   │   │   ├── components/# 组件展示
│   │   │   └── settings/  # 系统设置
│   │   ├── shared/        # 共享资源
│   │   │   ├── components/# UI 组件库
│   │   │   ├── layout/    # 布局组件
│   │   │   └── theme/     # 主题系统
│   │   ├── store/         # 状态管理
│   │   └── wailsjs/       # Wails 绑定
│   ├── index.html
│   ├── package.json
│   └── vite.config.ts
├── internal/               # 后端内部包
│   ├── config/            # 配置管理
│   ├── data/              # 数据模块
│   │   ├── controller.go  # 控制器
│   │   ├── service.go     # 业务逻辑
│   │   ├── dao.go         # 数据访问
│   │   ├── model.go       # 数据模型
│   │   └── mock.go        # 模拟数据
│   ├── database/          # 数据库
│   └── logger/            # 日志系统
├── app.go                  # 应用主逻辑
├── main.go                 # 程序入口
├── config.yaml             # 配置文件
├── go.mod
└── wails.json             # Wails 配置
```

---

## 核心功能

### 1. 数据管理模块

**功能特性：**
- ✅ 完整的 CRUD 操作
- ✅ 分页查询（默认 10 条/页）
- ✅ 多条件筛选（关键词、分类、状态）
- ✅ 30 条模拟数据自动生成
- ✅ 表格固定高度内滚动
- ✅ 固定表头

**后端实现：**
- `internal/data/service.go` - 业务逻辑层
- `internal/data/dao.go` - 数据访问层（SQLite）
- `internal/data/mock.go` - 模拟数据生成
- `app.go` - API 绑定

**前端实现：**
- `frontend/src/modules/data/DataPage.tsx` - 数据管理页面
- `frontend/src/modules/data/api.ts` - API 调用
- 支持分页、筛选、新建、编辑、删除

**数据结构：**
```typescript
interface DataRecord {
  id: string
  name: string
  category: string
  status: string
  createdAt: string
  updatedAt: string
}
```

### 2. 指纹浏览器基础环境管理

**功能特性：**
- ✅ 多套浏览器环境配置与切换
- ✅ 默认环境一键设置
- ✅ 核心路径与连接类型配置
- ✅ 启动时自动检测环境路径

**入口位置：**
- 侧边栏：指纹浏览器 -> 基础环境

**环境字段：**
- coreId：环境标识
- coreName：环境名称
- corePath：核心路径（支持相对路径）
- connectorType：连接类型
- isDefault：是否默认

**环境示例：**
```json
{
  "coreId": "chrome-114",
  "coreName": "Chrome 114",
  "corePath": "chrome/114",
  "connectorType": "standard",
  "isDefault": true
}
```

**默认行为：**
- 默认环境用于实例启动时的核心解析
- corePath 支持相对路径，会自动解析到可执行目录或工作目录

### 3. 表格组件

**特性：**
- 固定高度容器（默认 `calc(100vh - 320px)`）
- 内容在表格内滚动
- 固定表头（sticky 定位）
- 自定义列渲染
- 加载状态
- 空数据提示

**使用示例：**
```tsx
<Table
  columns={columns}
  data={data}
  rowKey="id"
  maxHeight="500px"
  stickyHeader
/>
```

### 4. 分页组件

**特性：**
- 首页/末页快速跳转
- 上一页/下一页导航
- 智能页码显示（超过 5 页显示省略号）
- 每页条数选择（10/20/50）
- 总数显示

**使用示例：**
```tsx
<Pagination
  current={page}
  total={total}
  pageSize={pageSize}
  onChange={setPage}
  onPageSizeChange={setPageSize}
/>
```

### 5. 主题系统

**特性：**
- 亮色/暗色主题切换
- CSS 变量管理颜色
- 平滑过渡动画
- 持久化存储

**主题切换：**
```tsx
<ThemeSwitcher />
```

---

## 组件库

### 组件总览（24 个）

#### 基础组件
- **Button** - 按钮（4种变体，3种尺寸）
- **Card** - 卡片容器
- **Badge** - 徽章标签

#### 表单组件
- **Input** - 输入框
- **Textarea** - 文本域
- **Select** - 选择器
- **Switch** - 开关
- **FormItem** - 表单项容器

#### 数据展示
- **Table** - 表格
- **Pagination** - 分页器
- **StatCard** - 统计卡片
- **Progress** - 进度条
- **CircleProgress** - 圆形进度条
- **Skeleton** - 骨架屏

#### 反馈组件
- **Alert** - 警告提示
- **Toast** - 消息提示
- **Loading** - 加载状态
- **Modal** - 弹窗
- **ConfirmModal** - 确认对话框
- **Drawer** - 抽屉

#### 导航组件
- **Tabs** - 标签页

#### 浮层组件
- **Popover** - 气泡卡片
- **Dropdown** - 下拉菜单

#### 其他
- **ThemeSwitcher** - 主题切换器

### 组件使用示例

#### Alert（警告提示）

```tsx
<Alert
  type="success"
  title="成功"
  message="操作已完成"
  closable
  onClose={() => console.log('closed')}
/>
```

**类型：** success | error | warning | info

#### Toast（消息提示）

```tsx
import { toast } from '@/shared/components'

toast.success('操作成功')
toast.error('操作失败')
toast.warning('警告信息')
toast.info('提示信息', 5000) // 5秒后消失
```

#### Modal（弹窗）

```tsx
<Modal
  open={visible}
  onClose={() => setVisible(false)}
  title="标题"
  footer={
    <>
      <Button variant="secondary" onClick={onClose}>取消</Button>
      <Button onClick={onConfirm}>确定</Button>
    </>
  }
>
  <div>弹窗内容</div>
</Modal>
```

#### Drawer（抽屉）

```tsx
<Drawer
  open={visible}
  onClose={() => setVisible(false)}
  title="侧边抽屉"
  placement="right"
  width="400px"
>
  <div>抽屉内容</div>
</Drawer>
```

**方向：** left | right | top | bottom

#### Popover（气泡卡片）

```tsx
<Popover
  content={<div>提示内容</div>}
  placement="top"
  trigger="hover"
>
  <Button>悬停显示</Button>
</Popover>
```

**触发方式：** click | hover  
**位置：** top | bottom | left | right

#### Dropdown（下拉菜单）

```tsx
const items = [
  { key: 'copy', label: '复制', icon: <Copy /> },
  { key: 'divider', label: '', divider: true },
  { key: 'delete', label: '删除', danger: true },
]

<Dropdown
  items={items}
  onSelect={(key) => console.log(key)}
/>
```

#### Progress（进度条）

```tsx
// 线性进度条
<Progress percent={60} />
<Progress percent={100} status="success" size="lg" />

// 圆形进度条
<CircleProgress percent={75} size={120} />
```

**状态：** normal | success | error | warning

#### Loading & Skeleton

```tsx
// 加载器
<Loading size="md" text="加载中..." />
<Loading fullscreen />

// 骨架屏
<Skeleton width="100%" height="20px" />
<Skeleton width="48px" height="48px" circle />
```

### 组件展示页面

访问 `/components` 路由查看所有组件的实时示例和交互演示。

---

## 数据管理

### 后端 API

所有数据管理 API 都通过 Wails 绑定暴露给前端：

```go
// 获取数据列表（支持分页）
func (a *App) DataGetList(filters data.DataFilters, page, pageSize int) (*data.DataListResponse, error)

// 获取单条数据
func (a *App) DataGetByID(id string) (*data.DataRecord, error)

// 创建数据
func (a *App) DataCreate(req data.CreateDataRequest) (*data.DataRecord, error)

// 更新数据
func (a *App) DataUpdate(id string, req data.UpdateDataRequest) (*data.DataRecord, error)

// 删除数据
func (a *App) DataDelete(id string) error

// 批量删除
func (a *App) DataBatchDelete(ids []string) (int, error)

// 获取统计信息
func (a *App) DataGetStats() (*data.DataStats, error)
```

### 前端调用

```typescript
import { fetchDataList, createData, updateData, deleteData } from '@/modules/data/api'

// 获取列表
const result = await fetchDataList(filters, page, pageSize)

// 创建
await createData({ name: '名称', category: '分类', status: 'active' })

// 更新
await updateData(id, { name: '新名称' })

// 删除
await deleteData(id)
```

### 模拟数据

应用首次启动时会自动检查数据库：
- 如果数据库为空，自动生成 30 条模拟数据
- 如果已有数据，跳过初始化

模拟数据包含：
- 多种分类：技术、商业、设计、产品、其他
- 多种状态：启用、禁用、待处理
- 随机的创建和更新时间

---

## 开发指南

### 添加新页面

1. 在 `frontend/src/modules/` 创建新模块目录
2. 创建页面组件
3. 在 `App.tsx` 添加路由
4. 在 `config/project.config.ts` 添加导航菜单

### 添加新组件

1. 在 `frontend/src/shared/components/` 创建组件文件
2. 在 `index.ts` 导出组件
3. 遵循现有组件的设计模式
4. 添加 TypeScript 类型定义

### 添加后端 API

1. 在 `internal/` 创建新模块
2. 实现 Service、DAO、Model 层
3. 在 `app.go` 添加 API 绑定方法
4. 运行 `wails dev` 自动生成前端绑定

### 配置管理

编辑 `config.yaml` 修改应用配置：

```yaml
app:
  name: "应用名称"

runtime:
  max_memory_mb: 512
  gc_percent: 100

logging:
  level: "info"
  file_enabled: true
  file_path: "data/logs/app.log"

database:
  sqlite:
    path: "app.db"
```

### 主题定制

在 `frontend/src/config/project.config.ts` 修改：

```typescript
export const projectConfig = {
  name: '应用名称',
  shortName: '简称',
  primaryColor: 'primary',
}
```

---

## 构建部署

### Windows 构建

```bash
bat\build.bat
```

构建产物：`build/bin/news-platform.exe`

### 跨平台构建

```bash
# macOS
wails build -platform darwin/universal

# Linux
wails build -platform linux/amd64
```

### 构建选项

编辑 `wails.json` 配置构建选项：

```json
{
  "name": "news-platform",
  "outputfilename": "news-platform",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "wailsjsdir": "./frontend/src/wailsjs"
}
```

---

## 专项方案

- [配置发布与增量迁移方案（2026-03-09）](./CONFIG_RELEASE_AND_MIGRATION_PLAN_20260309.md)

---

## 最佳实践

### 组件使用建议

| 场景 | 推荐组件 | 说明 |
|------|---------|------|
| 操作反馈 | Toast | 轻量级，自动消失 |
| 重要提示 | Alert | 持久显示 |
| 确认操作 | ConfirmModal | 需用户确认 |
| 详细信息 | Modal | 复杂内容 |
| 侧边表单 | Drawer | 不遮挡主内容 |
| 简单提示 | Popover | 轻量级浮层 |
| 操作菜单 | Dropdown | 结构化菜单 |

### 性能优化

1. **表格分页** - 大数据量使用分页，避免一次加载过多
2. **虚拟滚动** - 超长列表考虑虚拟滚动
3. **懒加载** - 路由懒加载，按需加载模块
4. **图片优化** - 压缩图片，使用合适的格式
5. **代码分割** - Vite 自动代码分割

### 代码规范

1. **TypeScript** - 所有代码使用 TypeScript
2. **组件命名** - PascalCase（如 `DataPage`）
3. **文件命名** - PascalCase（如 `DataPage.tsx`）
4. **函数命名** - camelCase（如 `fetchData`）
5. **常量命名** - UPPER_SNAKE_CASE（如 `API_URL`）

---

## 常见问题

### Q: 如何修改应用名称？

A: 修改以下文件：
- `frontend/src/config/project.config.ts` - 前端显示名称
- `wails.json` - 构建产物名称
- `config.yaml` - 应用配置名称

### Q: 如何添加新的数据表？

A: 
1. 在 `internal/database/sqlite.go` 的 `Migrate()` 方法添加建表 SQL
2. 创建对应的 Model、DAO、Service
3. 在 `app.go` 添加 API 绑定

### Q: 如何自定义主题颜色？

A: 修改 `frontend/src/index.css` 中的 CSS 变量：

```css
:root {
  --color-accent: #your-color;
}
```

### Q: 如何禁用某个功能模块？

A: 在 `frontend/src/config/project.config.ts` 修改：

```typescript
export const featuresConfig = {
  dashboard: true,
  data: false,  // 禁用数据管理
  settings: true,
}
```

---

## 更新日志

### 2026-01-06

**新增功能：**
- ✅ 数据管理模块（CRUD + 分页）
- ✅ 30 条模拟数据自动生成
- ✅ 表格固定高度内滚动
- ✅ 分页组件
- ✅ 24 个 UI 组件
- ✅ 组件展示页面

**组件库：**
- Alert、Toast、Loading、Skeleton
- Modal、ConfirmModal、Drawer
- Popover、Dropdown
- Progress、CircleProgress
- Badge、Tabs

**优化：**
- 表格支持固定表头
- 统一的设计语言和动画
- 完整的 TypeScript 类型支持

---

## 技术支持

- **文档位置：** `docs/README.md`
- **组件展示：** 运行应用访问 `/components`
- **示例代码：** 查看 `frontend/src/modules/` 各模块

---

## 许可证

MIT License

---

**构建时间：** 2026-01-06  
**版本：** 1.0.0  
**状态：** ✅ 生产就绪
