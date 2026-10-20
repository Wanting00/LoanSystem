# LoanSystem

> For detailed architecture, design decisions, transaction handling, and concurrency control, please refer to `Loan System Design.pdf` in this repository.

## Overview

LoanSystem is a Go service for managing the loan lifecycle, including loan creation, approval, investment, and disbursement. It exposes a Gin HTTP API and persists loan data in MySQL.

## Tech Stack

- **Go 1.22** — application language
- **Gin** — HTTP routing and request handling
- **GORM** — database ORM and model migrations
- **MySQL** — relational database, accessed through the Go MySQL driver
- **shopspring/decimal** — decimal values for loan and investment amounts
- **godotenv** — loading local environment configuration

## Project Structure

```text
.
├── cmd/
│   ├── migrate/       # Database creation and model migration command
│   └── server/        # HTTP service entry point
├── config/            # Environment configuration and MySQL connection
├── internal/
│   ├── dto/           # API request and response types
│   ├── handler/       # HTTP routes and handlers
│   ├── model/         # GORM database models
│   ├── repository/    # Database access
│   └── service/       # Loan lifecycle business logic
├── mapper/            # Model and DTO mapping
├── migrations/        # GORM AutoMigrate setup
├── .env.example       # Local configuration template
└── go.mod             # Go module and dependencies
```

## API Overview

The service registers the following routes:

| Method | Endpoint | Description | Inputs |
| --- | --- | --- | --- |
| `GET` | `/health` | Health check | None |
| `POST` | `/api/v1/loans/create` | Create a loan | JSON body: `borrowerId`, `principal`, `rate`, `roi`, `agreementLetterUrl`; header: `Idempotency-Key` |
| `GET` | `/api/v1/loans/:id` | Retrieve a loan by ID | Path parameter: `id` |
| `POST` | `/api/v1/loans/:id/approve` | Approve a loan | Path parameter: `id`; JSON body: `validatorPhotoUrl`, `employeeId`, `approvalDate` |
| `POST` | `/api/v1/loans/:id/invest` | Invest in an approved loan | Path parameter: `id`; JSON body: `investorId`, `amount` |
| `POST` | `/api/v1/loans/:id/disburse` | Disburse an invested loan | Path parameter: `id`; JSON body: `signedAgreementUrl`, `employeeId`, `disbursementDate` |

The request body fields correspond to the DTOs in `internal/dto`. The loan retrieval endpoint reads from the database. The health endpoint responds with `{"message":"ok"}`.

## Getting Started

### Prerequisites

- Go 1.22 or later
- MySQL

### Run locally

1. Clone the repository and change to its directory:

   ```sh
   git clone https://github.com/Wanting00/LoanSystem.git
   cd LoanSystem
   ```

2. Copy `.env.example` to `.env` in the project root and configure the database connection:

   | Variable | Purpose | Example |
   | --- | --- | --- |
   | `MYSQL_HOST` | MySQL server host | `127.0.0.1` |
   | `MYSQL_PORT` | MySQL server port | `3306` |
   | `MYSQL_USER` | MySQL username | `root` |
   | `MYSQL_PASSWORD` | MySQL password | Set your local password |
   | `MYSQL_DATABASE` | Database used by the service | `loan_system` |
   | `HTTP_ADDR` | HTTP listen address (optional) | Defaults to `:8080` |

   The application loads `.env` if present; existing system environment variables take precedence over values in the file.

3. Initialize the database:

   ```sh
   go run ./cmd/migrate
   ```

   This command creates the database named by `MYSQL_DATABASE` if it does not exist, then creates or updates the tables using GORM `AutoMigrate`. The configured MySQL user must have permission to create the database and tables.

4. Download Go dependencies and start the service:

   ```sh
   go mod download
   go run ./cmd/server
   ```

5. By default, the service listens on port `8080` (`:8080`). Confirm it is running by checking the startup log and requesting the health endpoint:

   ```sh
   curl http://localhost:8080/health
   ```

   A healthy service responds with `{"message":"ok"}`.
