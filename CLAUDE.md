# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build & Run
```bash
# Build the application
go build -o miko_news ./cmd/main.go

# Run locally (requires .env configuration)
go run ./cmd/main.go

# Build with Docker
docker build -t miko-news .

# Run with Docker Compose (includes environment setup)
docker-compose up --build -d
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test file
go test ./test/article_repository_test.go
```

### Database
```bash
# Initialize database (run SQL migration)
mysql -h <host> -u <user> -p<password> <database> < migrations/init.sql
```

## Architecture Overview

MikoNews is a Go-based Feishu (Lark) bot that processes article submissions via private messages and forwards them to group chats. The application uses a layered architecture with dependency injection.

### Core Components

**Entry Point**: `cmd/main.go` - Initializes all components and starts both the Feishu bot (WebSocket) and API server (HTTP) concurrently.

**Bot Layer** (`internal/bot/`):
- `feishu_bot.go` - Main bot coordinator that starts WebSocket connection
- `feishu_event_dispatcher.go` - Routes incoming Feishu events to appropriate handlers

**Message Processing** (Strategy Pattern):
- `service/message_handling_service.go` - Interface for message processing strategies
- `service/impl/messagehandler/submission_handler.go` - Handles "投稿" (submission) messages
- `service/impl/messagehandler/default_message_handler.go` - Fallback for unrecognized messages

**Service Layer** (`internal/service/`):
- Article management: `article_service.go` and implementation
- Feishu API integration: `feishu_contact_service.go`, `feishu_message_service.go`
- Strategy-based message routing

**Repository Layer** (`internal/repository/`):
- Data access abstraction with MySQL implementation in `impl/mysql/`
- Uses GORM for database operations

**API Layer** (`internal/api/`):
- HTTP endpoints for article management (health checks, potential CRUD operations)
- Uses Gin framework

### Message Flow

1. User sends rich text message with title "投稿" via Feishu private chat
2. WebSocket receives `P2MessageReceiveV1` event
3. `FeishuEventDispatcher` routes to `MessageHandlingService`
4. `SubmissionHandler` parses message, extracts title (first bold line) and content
5. User info retrieved via `FeishuContactService`
6. Article saved to database via `ArticleService`
7. Content forwarded to configured group chats via `FeishuMessageService`
8. Confirmation sent back to user

### Configuration

- Environment variables override config file settings
- Key variables: `DB_HOST`, `FEISHU_APP_ID`, `FEISHU_APP_SECRET`, `FEISHU_GROUP_CHATS`
- Group chat IDs in `FEISHU_GROUP_CHATS` are comma-separated
- Database initialization required via `migrations/init.sql`

### Dependencies

- **Web Framework**: Gin
- **Database**: MySQL 5.7+ with GORM
- **Feishu SDK**: larksuite/oapi-sdk-go/v3
- **Logging**: Zap + Lumberjack
- **Config**: YAML + environment variables