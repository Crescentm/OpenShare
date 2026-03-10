# OpenShare Backend

OpenShare 后端服务，基于 Go + Gin + GORM 构建。

## 目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go      # 应用入口
├── internal/
│   ├── config/          # 配置管理
│   ├── handler/         # HTTP 处理器
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   ├── router/          # 路由配置
│   └── service/         # 业务逻辑层
├── pkg/
│   ├── logger/          # 日志工具
│   ├── response/        # 响应封装
│   └── storage/         # 存储管理
├── configs/
│   └── config.yaml      # 配置文件
├── migrations/          # 数据库迁移
└── go.mod
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置数据库

编辑 `configs/config.yaml`：

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: openshare
```

### 3. 运行服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## 配置项

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| server.port | OPENSHARE_SERVER_PORT | 8080 | 服务端口 |
| server.mode | OPENSHARE_SERVER_MODE | debug | 运行模式 |
| database.host | OPENSHARE_DATABASE_HOST | localhost | 数据库主机 |
| storage.base_path | OPENSHARE_STORAGE_BASE_PATH | /data/openshare | 存储路径 |

## API 分层

- **Handler**: 处理 HTTP 请求，参数校验，调用 Service
- **Service**: 业务逻辑实现，调用 Repository
- **Repository**: 数据访问，与数据库交互
- **Model**: 数据结构定义

## 中间件

- **Recovery**: 异常恢复，防止 panic 导致服务崩溃
- **Logger**: 请求日志记录
- **CORS**: 跨域处理
- **Auth**: JWT 认证
- **RateLimiter**: 限流（待实现）

## 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

错误码：
- 0: 成功
- 400: 请求错误
- 401: 未授权
- 403: 禁止访问
- 404: 资源不存在
- 409: 冲突
- 500: 服务器错误
