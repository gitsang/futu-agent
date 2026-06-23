# Futu Agent 架构设计

## 1. 系统概述

Futu Agent 是一个 LLM 驱动的自动交易系统，支持港股、美股、A股的模拟交易操作。系统 7×24 小时运行，分钟级交易频率，完全无人值守。

## 2. 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| 后端 | Go + go-chi | REST API 服务器 |
| 前端 | SvelteKit + shadcn | Dashboard 界面 |
| 数据库 | PostgreSQL 18 | 数据持久化 |
| 交易网关 | Futu OpenD | 连接富途交易系统 |
| AI 模型 | OpenAI 兼容 API | 交易决策生成 |
| 容器化 | Docker + Docker Compose | 一键部署 |

## 3. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Docker Compose                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │   Frontend   │    │   Backend    │    │  Futu OpenD  │   │
│  │  (SvelteKit) │    │   (Go API)   │    │   (Gateway)  │   │
│  │   Port 3000  │    │  Port 8080   │    │  Port 11111  │   │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘   │
│         │                   │                   │           │
│         │ HTTP/REST         │ TCP/Protobuf      │           │
│         │                   │                   │           │
│         └───────────────────┼───────────────────┘           │
│                             │                               │
│                      ┌──────┴───────┐                       │
│                      │  PostgreSQL  │                       │
│                      │   Port 5432  │                       │
│                      └──────────────┘                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ HTTPS (Proxy)
                              ▼
                     ┌──────────────────┐
                     │   OpenAI API    │
                     │  (LLM Service)  │
                     └──────────────────┘
```

## 4. 核心模块

### 4.1 交易代理引擎 (Trading Agent Engine)

**职责**：
- 定时获取市场数据
- 调用 LLM 生成交易决策
- 执行交易操作
- 记录决策日志

**工作流程**：
```
1. 获取行情数据 → 2. 分析持仓状态 → 3. 构建 Prompt
      ↓
4. 调用 LLM API → 5. 解析交易决策 → 6. 执行交易
      ↓
7. 记录日志 → 8. 等待下一周期
```

### 4.2 Futu API 客户端

**职责**：
- 连接 Futu OpenD 网关
- 获取实时行情
- 执行交易操作
- 查询持仓和账户信息

**支持的市场**：
- 港股 (HK)
- 美股 (US)
- A股 (CN)

### 4.3 数据管理

**数据库表设计**：

```sql
-- 交易决策日志
CREATE TABLE trade_decisions (
    id SERIAL PRIMARY KEY,
    agent_id VARCHAR(50) NOT NULL,
    stock_code VARCHAR(20) NOT NULL,
    market VARCHAR(10) NOT NULL,
    action VARCHAR(10) NOT NULL, -- BUY, SELL, HOLD
    quantity INTEGER,
    price DECIMAL(15, 4),
    reason TEXT,
    llm_response JSONB,
    executed BOOLEAN DEFAULT FALSE,
    executed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 持仓信息
CREATE TABLE positions (
    id SERIAL PRIMARY KEY,
    stock_code VARCHAR(20) NOT NULL,
    market VARCHAR(10) NOT NULL,
    quantity INTEGER NOT NULL,
    avg_cost DECIMAL(15, 4) NOT NULL,
    current_price DECIMAL(15, 4),
    unrealized_pnl DECIMAL(15, 4),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(stock_code, market)
);

-- 账户资金
CREATE TABLE account_funds (
    id SERIAL PRIMARY KEY,
    total_assets DECIMAL(15, 4),
    cash DECIMAL(15, 4),
    market_value DECIMAL(15, 4),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 代理配置
CREATE TABLE agent_configs (
    id SERIAL PRIMARY KEY,
    agent_id VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100),
    description TEXT,
    llm_model VARCHAR(100),
    llm_endpoint VARCHAR(255),
    trading_strategy TEXT,
    risk_parameters JSONB,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 系统配置
CREATE TABLE system_configs (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4.4 前端 Dashboard

**页面结构**：
1. **Dashboard** - 总览页面
   - 账户资产概览
   - 持仓分布图
   - 最近交易记录
   - 系统状态监控

2. **决策日志** - AI 决策历史
   - 决策时间线
   - 决策详情（LLM 响应）
   - 执行状态

3. **持仓管理** - 当前持仓
   - 持仓列表
   - 盈亏分析
   - 持仓分布

4. **配置管理** - 系统设置
   - 代理配置
   - LLM 配置
   - 交易参数
   - 风控设置

## 5. API 设计

### 5.1 REST API 端点

```
# 账户相关
GET    /api/account/funds          # 获取账户资金
GET    /api/account/positions      # 获取持仓列表

# 交易决策
GET    /api/decisions              # 获取决策日志
GET    /api/decisions/:id          # 获取决策详情

# 代理配置
GET    /api/agents                 # 获取代理列表
POST   /api/agents                 # 创建代理
PUT    /api/agents/:id             # 更新代理
DELETE /api/agents/:id             # 删除代理

# 系统配置
GET    /api/config                 # 获取系统配置
PUT    /api/config                 # 更新系统配置

# 系统状态
GET    /api/status                 # 获取系统状态
POST   /api/agents/:id/start       # 启动代理
POST   /api/agents/:id/stop        # 停止代理
```

## 6. 部署架构

### 6.1 Docker Compose 服务

```yaml
services:
  postgres:
    image: postgres:18
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  futu-opend:
    build:
      context: ./futu-opend
      dockerfile: Dockerfile
    ports:
      - "11111:11111"
    environment:
      - FUTU_LOGIN_ACCOUNT=36474237
      - FUTU_LOGIN_PWD_MD5=d63b1d22c9141885b5809965be2e9d64

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - futu-opend
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/futu_agent
      - FUTU_OPEND_HOST=futu-opend
      - FUTU_OPEND_PORT=11111
      - HTTP_PROXY=http://ops.yl.c8g.top:7890
      - HTTPS_PROXY=http://ops.yl.c8g.top:7890

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    depends_on:
      - backend
```

## 7. 网络代理配置

系统需要通过代理访问 OpenAI API：

```bash
# 环境变量
HTTP_PROXY=http://ops.yl.c8g.top:7890
HTTPS_PROXY=http://ops.yl.c8g.top:7890

# Go 代码中使用
transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
}
client := &http.Client{Transport: transport}
```

## 8. 安全考虑

1. **Futu 账号安全**：使用 MD5 加密的密码，不存储明文
2. **API 认证**：后端 API 可以添加简单的 API Key 认证
3. **数据库访问**：使用强密码，限制网络访问
4. **容器隔离**：每个服务运行在独立容器中

## 9. 扩展性考虑

1. **多代理支持**：可以配置多个交易代理，每个代理有不同的策略
2. **策略扩展**：通过配置文件或数据库定义不同的交易策略
3. **监控告警**：可以添加 Prometheus + Grafana 监控
4. **日志收集**：可以添加 ELK 或 Loki 日志系统
