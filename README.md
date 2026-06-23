# Futu Agent - AI 驱动的自动交易系统

一个基于 LLM 的 7×24 小时自动交易代理系统，支持港股、美股、A股的模拟交易操作。

## 特性

- **LLM 驱动**: 使用 OpenAI 兼容的 API 进行交易决策
- **多市场支持**: 港股、美股、A股
- **分钟级交易**: 可配置的交易频率
- **完全无人值守**: 自动化运行，无需人工干预
- **Docker 一键部署**: 简单的部署流程

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go + go-chi |
| 前端 | SvelteKit + shadcn-svelte |
| 数据库 | PostgreSQL 16 |
| 交易网关 | Futu OpenD |
| 容器化 | Docker + Docker Compose |

## 快速开始

### 前置要求

- Docker 和 Docker Compose
- Futu 账号（模拟交易）
- OpenAI 兼容的 API Key

### 部署步骤

1. 克隆项目:
```bash
git clone https://github.com/your-username/futu-agent.git
cd futu-agent
```

2. 配置环境变量:
```bash
cp .env.example .env
# 编辑 .env 文件，填入你的配置
```

3. 启动服务:
```bash
docker-compose up -d
```

4. 访问系统:
- 前端 Dashboard: http://localhost:3000
- 后端 API: http://localhost:8080

## 环境变量配置

```env
# Futu 账号
FUTU_LOGIN_ACCOUNT=your_account_id
FUTU_LOGIN_PWD_MD5=your_password_md5

# LLM 配置
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=your_api_key
LLM_MODEL=gpt-4

# 数据库
DATABASE_URL=postgres://postgres:postgres@postgres:5432/futu_agent?sslmode=disable

# 代理（可选）
HTTP_PROXY=http://your-proxy:port
HTTPS_PROXY=http://your-proxy:port

# 交易
TRADING_ENABLED=false
```

## API 接口

### 账户
- `GET /api/account/funds` - 获取账户资金
- `GET /api/account/positions` - 获取持仓列表

### 决策日志
- `GET /api/decisions` - 获取决策列表
- `GET /api/decisions/:id` - 获取决策详情

### 代理配置
- `GET /api/agents` - 获取代理列表
- `POST /api/agents` - 创建代理
- `PUT /api/agents/:id` - 更新代理
- `DELETE /api/agents/:id` - 删除代理
- `POST /api/agents/:id/start` - 启动代理
- `POST /api/agents/:id/stop` - 停止代理

### 系统
- `GET /api/config` - 获取系统配置
- `PUT /api/config` - 更新系统配置
- `GET /api/status` - 获取系统状态

## 项目结构

```
futu-agent/
├── backend/                 # Go 后端
│   ├── cmd/server/         # 主程序入口
│   ├── internal/           # 内部包
│   │   ├── config/        # 配置管理
│   │   ├── database/      # 数据库操作
│   │   ├── handlers/      # HTTP 处理器
│   │   ├── models/        # 数据模型
│   │   └── services/      # 业务逻辑
│   └── Dockerfile
├── frontend/                # SvelteKit 前端
│   ├── src/
│   │   ├── routes/        # 页面路由
│   │   └── lib/           # 组件和工具
│   └── Dockerfile
├── futu-opend/              # Futu OpenD 容器
│   └── Dockerfile
├── docker-compose.yml       # Docker 编排
└── .env                     # 环境变量
```

## 开发

### 后端开发

```bash
cd backend
go mod tidy
go run ./cmd/server
```

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

### 构建镜像

```bash
docker-compose build
```

## 注意事项

1. **模拟交易**: 系统默认使用模拟交易模式，不会使用真实资金
2. **网络代理**: 如果需要访问 OpenAI API，配置 HTTP_PROXY 和 HTTPS_PROXY
3. **Futu OpenD**: 需要 Futu 账号才能登录 OpenD 网关
4. **LLM 配置**: 支持任何 OpenAI 兼容的 API 接口

## 许可证

MIT License
