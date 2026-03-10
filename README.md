# OpenShare

轻量级、自托管的资料共享平台，面向校园网、实验室和组织内部网络场景。

## 项目状态

当前仓库处于规划和基础搭建阶段，核心文档已经整理完成，后续会按开发规划逐步落地代码。

## 项目定位

OpenShare 用于在内网环境中共享学习资料或组织内部资料，核心目标是：

- 游客可浏览、搜索、上传和下载资料
- 管理员可审核上传内容并治理资源质量
- 系统保持轻量、自托管、易部署

## 当前技术基线

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
- SQLite

### 其他基础能力

- 文件存储：Linux Local File System
- 管理端认证：Session（Cookie）
- 密码加密：bcrypt
- 搜索：SQLite FTS5

## 计划能力范围

- 资料浏览
- 单文件下载
- 批量打包下载
- 游客上传
- 回执码查询
- 管理员审核
- 搜索与 Tag
- 举报系统
- 公告系统
- 本地目录导入
- 文件预览
- 操作日志和限流

## 目录说明

当前仓库以文档为主，后续代码目录按以下结构落地：

```text
OpenShare/
├── backend/
│   ├── cmd/
│   ├── configs/
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── router/
│   │   ├── service/
│   │   └── storage/
│   ├── migrations/
│   └── pkg/
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── layouts/
│   │   ├── router/
│   │   ├── stores/
│   │   ├── styles/
│   │   ├── types/
│   │   ├── utils/
│   │   └── views/
└── 文档/
```

## 开发顺序

建议严格按开发规划推进：

1. 项目初始化与基础架构
2. 数据库设计与迁移
3. Session 认证与管理员权限
4. 上传、审核、浏览、下载主链路
5. 搜索与 Tag
6. 后台管理能力
7. 举报、审计、限流
8. 预览、批量下载、本地导入
9. 联调、测试、部署

详细步骤见 [OpenShare 开发规划.md](/Users/quan/Desktop/OpenShare/文档/OpenShare%20开发规划.md)。

## 运行环境预期

项目正式开发后，默认运行环境预期如下：

- Go 1.21+
- Node.js 18+
- npm 9+
- 支持 SQLite 的本地开发环境
- Linux 文件系统目录用于资料存储

## 配置约定

后端配置将集中在 `backend/configs/config.yaml`，预计至少包含：

- 服务端口和运行模式
- SQLite 数据库路径
- 文件存储根目录
- Session 配置
- 限流配置

## 文件存储结构

系统采用本地文件存储，默认目录结构为：

```text
/data/openshare/
├── repository/   # 审核通过后的正式文件
├── staging/      # 用户上传后的待审核文件
└── trash/        # 删除或下架后的回收文件
```

## 当前说明

- 当前 README 主要用于说明项目方向和开发约束，不代表代码已经全部落地
- 后续当后端和前端基础工程恢复后，再补充具体启动命令、接口说明和部署步骤

## License

MIT
