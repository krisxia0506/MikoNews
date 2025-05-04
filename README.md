# Miko News 飞书机器人

Miko News 是一个飞书机器人应用，旨在简化社区新闻的收集和分享流程。用户可以通过与 Miko 机器人私聊轻松投稿，机器人会自动将内容转发到指定的飞书群聊，并同时将稿件存档到数据库中作为备份记录。

---

## 面向用户 (For Users)

### 功能简介

*   **用户**: 通过飞书私聊 Miko 机器人，使用特定格式发送内容即可完成投稿。机器人会确认收到并告知您内容已被转发。
*   **群成员**: 在预先配置好的飞书群聊中接收由机器人转发的最新投稿内容。
*   **管理员/运营**: 机器人自动完成转发和存档。您可以使用配套的 API (如果需要) 来查看历史投稿记录。

### 如何部署 (推荐使用 Docker)

推荐使用 Docker 进行部署，过程简单快捷。

**前提条件:**

*   一台可以访问互联网并安装了 [Docker](https://docs.docker.com/get-docker/) 和 [Docker Compose](https://docs.docker.com/compose/install/) 的服务器。
*   一个可用的 MySQL 数据库 (版本 5.7 或更高)。
*   拥有一个配置好的飞书应用 (获取 App ID, App Secret 等信息)。

**部署步骤:**

1.  **准备配置:**
    *   从代码库获取项目文件。
    *   在项目根目录，复制环境变量模板文件：`cp .env.example .env`
    *   编辑 `.env` 文件，填入您的飞书应用密钥、数据库连接信息和希望发布到的群聊 ID 等。

2.  **初始化数据库:**
    *   连接到您的 MySQL 数据库。
    *   执行项目 `migrations/init.sql` 文件中的 SQL 脚本，创建所需的表结构。
      ```bash
      # 示例命令 (请替换为您的实际数据库信息)
      mysql -h your-mysql-host -u your-db-user -p your-db-password your-db-name < migrations/init.sql
      ```

3.  **启动服务:**
    *   在项目根目录运行以下命令：
      ```bash
      docker-compose up -d
      ```
    *   服务将在后台启动。您可以查看 Docker 日志以确认启动状态。

**详细说明:**

*   完整的 Docker 部署细节和环境变量说明，请参考 [部署文档](docs/deployment.md)。

### 如何投稿

1.  在飞书中找到 Miko News 机器人并发起私聊。
2.  发送 **富文本 (Post)** 消息，格式如下：
    *   消息的 **标题 (Title)** 必须设置为 **`投稿`** (这两个字，没有其他字符)。
    *   富文本内容的第一行应 **加粗 (Bold)**，这将作为新闻的标题。
    *   富文本的其余内容将作为新闻正文。
    *   示例消息结构 (内部 JSON 格式，飞书客户端会为您生成):
        ```json
        {
          "title": "投稿",
          "content": [
            [{"tag":"text","text":"这是加粗的标题","style":["bold"]}],
            [{"tag":"text","text":"这是第一行正文。"}],
            [{"tag":"text","text":"这是第二行正文。"}]
          ]
        }
        ```
3.  发送成功后，机器人会回复确认消息，告知您稿件已收到并已被转发。

### 管理员操作 (通过 API)

(当前版本主要流程为自动转发，API 操作可能有限)

如果需要查看历史投稿记录，可以通过 Miko News 提供的 HTTP API 进行：

*   **查看稿件**: 获取已存档的稿件列表或特定稿件。

具体的 API 端点和使用方法，请参考开发者部分的 API 文档或直接查看代码。

---

## 面向开发者 (For Developers)

### 技术栈

*   **语言**: Go 1.24+
*   **Web 框架**: Gin
*   **数据库**: MySQL 5.7+
*   **ORM**: GORM
*   **飞书 SDK**: [larksuite/oapi-sdk-go/v3](https://github.com/larksuite/oapi-sdk-go)
*   **日志**: Zap + Lumberjack
*   **部署**: Docker & Docker Compose

### 系统概览

#### 1. 系统流程

```mermaid
graph LR
    A[用户（飞书）] -->|富文本消息（title:“投稿”）| B[飞书平台]
    B -->|事件回调 （WebSocket）| C{MikoNews 应用}
    C -->|获取用户信息| B
    B -->|用户信息| C
    C -->|存储稿件 （含作者名）| D[（MySQL 数据库）]
    C -->|构建卡片/消息内容| C
    C -->|发送卡片/消息到群聊| B
    B -->|群聊消息| E[指定群聊（飞书）]
    C -->|回复确认消息| B
    B -->|私聊确认消息| A
    F[管理员] --->|API 请求 （查看历史）| C
    C --->|查询历史稿件| D
```

#### 2. 消息处理流程 (策略模式)

机器人通过 WebSocket 接收私聊消息，并通过策略模式处理：

```mermaid
sequenceDiagram
    participant FP as 飞书平台 (WebSocket)
    participant Dispatcher as FeishuEventDispatcher
    participant MHS as MessageHandlingService
    participant SubStrat as SubmissionHandlerStrategy
    participant ContactSvc as FeishuContactService
    participant ArticleSvc as ArticleService
    participant MsgSvc as FeishuMessageService
    participant DefStrat as DefaultMessageHandlerStrategy

    FP->>Dispatcher: 收到 P2MessageReceiveV1 事件
    activate Dispatcher
    Dispatcher->>MHS: ProcessReceivedMessage(event)
    activate MHS
    MHS->>SubStrat: ShouldHandle(event)?
    activate SubStrat
    alt 消息为"投稿"富文本
        SubStrat-->>MHS: true
        MHS->>SubStrat: Handle(event)
        SubStrat->>SubStrat: parsePostContentForSubmission(content)
        SubStrat->>ContactSvc: GetUserInfoByOpenID(senderID)
        activate ContactSvc
        ContactSvc-->>SubStrat: (用户信息或错误)
        deactivate ContactSvc
        SubStrat->>ArticleSvc: SaveSubmission(..., authorName, title, content)
        activate ArticleSvc
        ArticleSvc-->>SubStrat: (创建的文章或错误)
        deactivate ArticleSvc
        alt 保存成功
            Note over SubStrat: 构建要转发到群聊的卡片/消息
            SubStrat->>MsgSvc: SendCardMessage(groupID, cardContent) // 或 SendTextMessage
            activate MsgSvc
            MsgSvc-->>SubStrat: (发送结果)
            deactivate MsgSvc
            SubStrat->>MsgSvc: ReplyTextMessage(msgID, "投稿成功并已转发...")
            activate MsgSvc
            MsgSvc-->>SubStrat: (发送结果)
            deactivate MsgSvc
        else 保存失败
            SubStrat->>MsgSvc: ReplyTextMessage(msgID, "保存投稿记录失败...")
            activate MsgSvc
            MsgSvc-->>SubStrat: (发送结果)
            deactivate MsgSvc
        end
        SubStrat-->>MHS: (处理结果)
    else 消息非"投稿"富文本
        SubStrat-->>MHS: false
    end
    deactivate SubStrat

    MHS->>DefStrat: ShouldHandle(event)?
    activate DefStrat
    alt 消息是 P2P (默认策略接管)
        DefStrat-->>MHS: true
        MHS->>DefStrat: Handle(event) (例如：回复"无法识别")
        DefStrat-->>MHS: (处理结果)
    else 消息非 P2P
        DefStrat-->>MHS: false
    end
    deactivate DefStrat
    MHS-->>Dispatcher: (返回 nil)
    deactivate MHS
    Dispatcher-->>FP: (内部处理，无需显式 ACK)
    deactivate Dispatcher
```

### 项目设计

#### 1. 设计理念

*   **分层架构**: 清晰分离 API、业务逻辑 (Service)、数据访问 (Repository)。
*   **依赖注入**: 主要在 `cmd/main.go` 和 `internal/bot/feishu_bot.go` 中组装依赖。
*   **面向接口**: Service 和 Repository 层都定义了接口。
*   **策略模式**: 用于处理不同类型的飞书消息。
*   **标准化**: 统一的日志、配置管理、错误处理和 API 响应结构。

#### 2. 项目结构

```
MikoNews/
├── cmd/                    # 应用程序入口 (main.go)
├── configs/                # 配置文件目录 (.yaml, .example)
├── docs/                   # 项目文档 (deployment.md)
├── internal/
│   ├── api/                # API 相关 (Gin)
│   │   ├── handler/        # HTTP 处理器
│   │   ├── middleware/     # HTTP 中间件
│   │   ├── router/         # 路由配置 (待完善)
│   │   └── server.go       # API 服务器启动与配置
│   ├── bot/                # Feishu Bot (WebSocket 实现)
│   │   ├── feishu_bot.go
│   │   └── feishu_event_dispatcher.go (可能存在)
│   ├── config/             # 配置加载逻辑
│   ├── database/           # 数据库连接 (GORM)
│   ├── model/              # 数据模型 (GORM 结构体)
│   ├── pkg/                # 内部公共库
│   │   ├── errors/         # 自定义错误
│   │   ├── logger/         # Zap 日志配置与全局函数
│   │   └── response/       # API 标准响应
│   ├── repository/         # 数据仓库层 (接口 + MySQL 实现)
│   │   ├── article_repository.go
│   │   └── impl/mysql/
│   └── service/            # 业务逻辑层 (接口 + 实现)
│       ├── article_service.go
│       ├── feishu_contact_service.go
│       ├── feishu_message_service.go
│       ├── message_handling_service.go (策略模式接口)
│       └── impl/           # 服务实现
│           ├── article_service.go
│           ├── feishu_contact_service.go
│           ├── feishu_message_service.go
│           └── messagehandler/ # 消息处理策略实现
│               ├── default_message_handler.go
│               ├── message_handling_service.go
│               └── submission_handler.go
├── migrations/             # 数据库迁移脚本 (init.sql)
├── scripts/                # 辅助脚本 (暂无)
├── test/                   # 测试文件 (待完善)
├── .env.example            # Docker 环境变量模板
├── .gitignore
├── .dockerignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

### 本地开发环境设置

**前提条件:**

*   Go 1.24+
*   MySQL 5.7+
*   飞书开发者账号和应用配置

**步骤:**

1.  **配置:**
    *   复制配置文件模板: `cp configs/config.yaml.example configs/config.yaml`
    *   编辑 `configs/config.yaml`，填入您的飞书应用配置 (App ID, Secret等) 和 MySQL 数据库连接信息。
    > **注意**: `configs/config.yaml` 已加入 `.gitignore`，不会提交到版本库。

2.  **数据库初始化:**
    *   连接到您的 MySQL 数据库。
    *   执行 `migrations/init.sql` 脚本创建表结构。
      ```bash
      mysql -u your-db-user -p your-db-password your-db-name < migrations/init.sql
      ```

3.  **启动服务:**
    *   在项目根目录运行:
      ```bash
      go run cmd/main.go
      ```
    *   服务将在配置的端口（默认为 8080）启动，并同时启动 WebSocket 连接监听飞书事件。

### API 接口

(当前版本下，API 主要用于查看历史记录)

*   `GET /health` - 健康检查
*   `GET /ping` - 服务可用性检查
*   (预期可能存在的接口)
    *   `GET /api/v1/articles` - 获取已存档的文章列表 (可添加过滤参数: 如按作者、时间范围)
    *   `GET /api/v1/articles/:id` - 获取特定存档文章详情

### 扩展开发

为项目添加新功能（例如，增加评论功能）的大致步骤：

1.  **模型**: 在 `internal/model/` 中定义新的数据模型 (`comment.go`)。
2.  **迁移**: 在 `migrations/` 中添加新的 SQL 脚本以创建 `comments` 表。
3.  **仓库**: 在 `internal/repository/` 定义 `CommentRepository` 接口并在 `impl/mysql/` 中实现。
4.  **服务**: 在 `internal/service/` 定义 `CommentService` 接口并在 `impl/` 中实现业务逻辑 (可能需要依赖 `CommentRepository` 和 `ArticleRepository`)。
5.  **处理器**: 在 `internal/api/handler/` 创建 `comment_handler.go` 处理 HTTP 请求。
6.  **路由**: 在 `internal/api/router/router.go` 中添加新的路由，指向 `CommentHandler` 的方法。
7.  **主程序**: 在 `cmd/main.go` 中初始化新的 Repository, Service, Handler 并注入依赖。

### 测试

(测试覆盖率可能不完整)

```bash
# 运行所有测试
go test ./...

# 查看测试覆盖率
go test -cover ./...
```

### 部署 (Docker)

本项目推荐使用 Docker 进行部署，请参考 **面向用户** 部分的部署说明或查看 [部署文档](docs/deployment.md)。

### 许可证

本项目采用 MIT 许可证 - 详见 `LICENSE` 文件。
