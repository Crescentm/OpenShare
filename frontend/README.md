# OpenShare Frontend

OpenShare 前端项目，基于 Vue 3 + Vite + TypeScript + Tailwind CSS 构建。

## 目录结构

```
frontend/
├── src/
│   ├── api/           # API 接口定义
│   │   ├── index.ts   # 用户端 API
│   │   └── admin.ts   # 管理端 API
│   ├── layouts/       # 布局组件
│   │   ├── UserLayout.vue   # 用户端布局
│   │   └── AdminLayout.vue  # 管理端布局
│   ├── router/        # 路由配置
│   ├── stores/        # Pinia 状态管理
│   │   ├── auth.ts    # 认证状态
│   │   └── app.ts     # 应用状态
│   ├── styles/        # 全局样式
│   ├── types/         # TypeScript 类型定义
│   ├── utils/         # 工具函数
│   │   └── http.ts    # Axios 封装
│   ├── views/         # 页面组件
│   │   ├── user/      # 用户端页面
│   │   └── admin/     # 管理端页面
│   ├── App.vue
│   └── main.ts
├── public/            # 静态资源
├── index.html
├── vite.config.ts
├── tailwind.config.js
└── package.json
```

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 开发模式

```bash
npm run dev
```

前端将在 `http://localhost:3000` 启动，API 请求会代理到 `http://localhost:8080`。

### 3. 构建

```bash
npm run build
```

构建产物在 `dist/` 目录。

## 路由结构

### 用户端

| 路径 | 页面 | 说明 |
|------|------|------|
| / | Home | 首页 |
| /files | FileList | 资料列表 |
| /files/:id | FileDetail | 资料详情 |
| /search | Search | 搜索结果 |
| /upload | Upload | 上传资料 |
| /my-uploads | MyUploads | 我的上传 |

### 管理端

| 路径 | 页面 | 说明 |
|------|------|------|
| /admin/login | Login | 登录 |
| /admin/submissions | Submissions | 审核管理 |
| /admin/files | Files | 资料管理 |
| /admin/tags | Tags | Tag 管理 |
| /admin/reports | Reports | 举报管理 |
| /admin/announcements | Announcements | 公告管理 |
| /admin/admins | Admins | 管理员管理 |
| /admin/settings | Settings | 系统设置 |
| /admin/logs | Logs | 操作日志 |

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| VITE_API_BASE_URL | API 基础路径 | /api/v1 |

## 开发规范

### 组件命名

- 页面组件：PascalCase，如 `FileList.vue`
- 通用组件：PascalCase，如 `SearchBox.vue`

### 样式

使用 Tailwind CSS，自定义样式在 `src/styles/index.css`。

预定义组件类：
- `.btn` / `.btn-primary` / `.btn-secondary` / `.btn-danger`
- `.input`
- `.card`

### API 调用

```typescript
import { fileApi } from '@/api'

// 获取列表
const { data } = await fileApi.getList({ page: 1 })

// 上传文件
const formData = new FormData()
formData.append('file', file)
await fileApi.upload(formData)
```

### 状态管理

```typescript
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// 登录
authStore.setToken(token)
authStore.setAdminInfo(info)

// 登出
authStore.logout()

// 权限检查
if (authStore.hasPermission('manage_files')) {
  // ...
}
```
