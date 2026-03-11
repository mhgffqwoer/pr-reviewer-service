# Pull Request Reviewer Assignment System

A Go-based service for automatically assigning reviewers to Pull Requests (PRs) from the author's team, managing teams and members. Interaction is carried out via HTTP API.

## Description

`pr-reviewer-service` is a microservice that automatically assigns reviewers to Pull Requests from the author's team, allows reassigning reviewers, and provides the ability to get a list of PRs assigned to a specific user, as well as manage teams and user activity.

**Key features:**

* **Automatic assignment**: When creating a PR, up to two active reviewers from the author's team are automatically assigned
* **Flexible reassignment**: Ability to replace one reviewer with a random active member from their team
* **Activity control**: Users with `isActive = false` are not assigned for review
* **Change protection**: After PR merge, changing the reviewer composition is prohibited
* **Idempotency**: The merge operation is idempotent - repeated calls do not cause errors
* **Flexible configuration**: Support for configuring logging, database, and server

## Requirements

* Go 1.22 or higher
* Docker and Docker Compose
* PostgreSQL (runs in Docker)

## Installation and Launch

```bash
# Launch service with database
task docker-up

# Stop service
task docker-down
```

## Configuration

The service supports configuration via `configs/config.yaml` file and environment variables from `.env`.

### Configuration Example

```yaml
logging:
  level: debug # debug | info | warn | error
  rotation:
    name: "./logs/app.log" # path to log file
    max_size: 10 # maximum file size in MB
    max_backups: 5 # maximum number of backups
    max_age: 7 # log retention period in days
    duplicate_to_console: true # duplicate logs to console

database:
  max_connections: 10 # maximum number of database connections

server:
  port: 8080 # server port
```

### Environment Variables

Create `.env` file based on `.env.example`:

## Architecture

The project is built on Clean Architecture principles using Dependency Injection.

```text
pr-reviewer-service/
├── cmd/
│   └── server/              # Application entry point
│       └── main.go          # Server and dependencies initialization
├── configs/                 # Configuration files
│   ├── config.yaml          # Application configuration
│   └── embed.go            # Static files embedding
├── internal/
│   ├── adapters/           # Adapters (interaction layers)
│   │   ├── http/           # HTTP handlers
│   │   │   └── handlers.go # HTTP endpoint implementation
│   │   └── postgres/       # PostgreSQL adapters
│   │       ├── db.go       # Database connection
│   │       └── repositories.go # Repository implementation
│   ├── config/             # Application configuration
│   │   └── config.go       # Configuration structures and loading logic
│   ├── domain/             # Business logic
│   │   ├── models/         # Domain models
│   │   │   └── models.go   # Team, User, PullRequest structures
│   │   └── services/       # Business services
│   │       ├── services.go # Main service with business logic
│   │       └── services_test.go # Business logic tests
│   ├── logger/             # Logging
│   │   └── logger.go       # Logger configuration and initialization
│   └── shared/             # Shared utilities
├── migrations/             # Database migrations
│   ├── 001_teams_and_users.up.sql
│   ├── 001_teams_and_users.down.sql
│   ├── 002_pull_requests.up.sql
│   ├── 002_pull_requests.down.sql
│   ├── 003_add_indexes.up.sql
│   └── 003_add_indexes.down.sql
├── deployment/             # Docker configuration
│   ├── docker-compose.yaml
│   ├── docker-compose.test.yaml
│   └── dockerfile
├── tests/                  # Integration tests
│   ├── integration_test.go
│   └── config/
│       └── config.yaml
├── api/                    # OpenAPI specification
│   └── openapi.yaml
├── docs/                   # Documentation
│   └── task_condition.md   # Task condition
└── .github/workflows/      # CI/CD
    └── ci.yml             # GitHub Actions workflow