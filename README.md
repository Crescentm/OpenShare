# OpenShare

轻量级、自托管的资料共享平台，让知识自由流动。

## 项目简介

OpenShare 是一个面向内网环境的学习资料共享平台，适用于校园网、实验室或组织内部网络，实现学习资料的共建与共享。

### 特性

- 📚 **资料共享** - 支持文件上传、下载、搜索
- 🔍 **智能搜索** - 支持关键词搜索、模糊匹配、Tag 过滤
- 📤 **简单上传** - 无需注册，回执码追踪投稿状态
- 👮 **审核机制** - 管理员审核确保内容质量
- 🏷️ **Tag 系统** - 灵活的标签分类管理
- 📢 **公告系统** - 系统公告及时通知

## 技术栈

### 前端
- Vue 3
- Vite
- TypeScript
- Tailwind CSS
- Pinia
- Vue Router

### 后端
- Go
- Gin
- GORM
- PostgreSQL

## 项目结构

```
OpenShare/
├── frontend/          # 前端项目
│   ├── src/
│   │   ├── api/       # API 接口
│   │   ├── layouts/   # 布局组件
│   │   ├── router/    # 路由配置
│   │   ├── stores/    # 状态管理
│   │   ├── styles/    # 全局样式
│   │   ├── types/     # 类型定义
│   │   ├── utils/     # 工具函数
│   │   └── views/     # 页面组件
│   └── ...
├── backend/           # 后端项目
│   ├── cmd/           # 应用入口
│   ├── internal/      # 内部模块
│   │   ├── config/    # 配置管理
│   │   ├── handler/   # 请求处理
│   │   ├── middleware/# 中间件
│   │   ├── model/     # 数据模型
│   │   └── router/    # 路由配置
│   ├── pkg/           # 公共包
│   │   ├── logger/    # 日志
│   │   ├── response/  # 响应封装
│   │   └── storage/   # 存储管理
│   ├── configs/       # 配置文件
│   └── migrations/    # 数据库迁移
└── 文档/              # 项目文档
```

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

### 后端启动

```bash
cd backend

# 安装依赖
go mod tidy

# 配置数据库
# 编辑 configs/config.yaml

# 运行
go run cmd/server/main.go
```

### 前端启动

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build
```

### 数据库准备

```sql
-- 创建数据库
CREATE DATABASE openshare;

-- 启用必要扩展（后续搜索功能需要）
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

## 配置说明

### 后端配置 (configs/config.yaml)

```yaml
server:
  port: 8080
  mode: debug  # debug, release

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: openshare

storage:
  base_path: /data/openshare  # 文件存储路径

jwt:
  secret: your-secret-key
  expire_hour: 24
```

### 环境变量

支持通过环境变量覆盖配置：

```bash
export OPENSHARE_SERVER_PORT=8080
export OPENSHARE_DATABASE_HOST=localhost
```

## API 说明

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /health | 健康检查 |
| GET | /api/v1/files | 资料列表 |
| GET | /api/v1/files/:id | 资料详情 |
| GET | /api/v1/files/:id/download | 下载文件 |
| POST | /api/v1/files/upload | 上传文件 |
| GET | /api/v1/search | 搜索资料 |
| GET | /api/v1/submissions | 查询投稿 |
| GET | /api/v1/tags | Tag 列表 |

### 管理接口

所有 `/api/v1/admin/*` 接口需要携带 `Authorization: Bearer <token>` 头。

## 文件存储

系统采用本地文件存储，目录结构：

```
/data/openshare/
├── repository/   # 已审核文件
├── staging/      # 待审核文件
└── trash/        # 回收站
```

## License

MIT