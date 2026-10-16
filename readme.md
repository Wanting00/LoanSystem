# Loan Service

一个使用 Go、Gin 和 GORM 搭建的贷款服务 API 雏形。目前提供分层结构、贷款相关数据模型与 DTO，以及返回 mock 数据的 HTTP 接口，供后续检查和逐步扩展。

## 技术栈

- Go 1.22+
- Gin：HTTP 路由与请求处理
- GORM：ORM
- MySQL 驱动：供后续启用数据库连接

## 项目结构

```text
.
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口，启动 HTTP 服务
├── config/
│   └── config.go                   # 环境配置与 MySQL 初始化入口
├── internal/
│   ├── dto/
│   │   ├── loan_request.go         # API 请求 DTO
│   │   └── loan_response.go        # API 响应 DTO
│   ├── handler/
│   │   └── loan_handler.go         # Gin 路由和 HTTP 处理
│   ├── model/                      # GORM 数据模型及关联
│   ├── repository/
│   │   └── loan_repository.go      # Repository 初始结构，尚无 CRUD
│   └── service/
│       └── loan_service.go         # mock 服务响应
├── go.mod
└── go.sum
```

## 分层调用

```text
HTTP Request → Gin Router → Handler → Service → Mock Response
```

- `handler` 负责路由、JSON 请求绑定和 HTTP 响应。
- `service` 提供当前的 mock 处理结果。
- `repository` 预留 GORM 数据访问层，目前未实现数据库操作。
- `model` 定义领域/数据库结构及 GORM 字段标签。
- `dto` 定义 API 请求与响应的数据形状。
- `config` 读取 HTTP 地址和 MySQL DSN，并提供数据库打开函数。

## 数据模型

`internal/model` 中定义了以下结构：

- `Borrower`：借款人，通过 `Loans` 与贷款建立一对多关系。
- `Investor`：投资人，通过 `Investments` 与投资记录建立一对多关系。
- `Loan`：贷款，状态为 `PROPOSED`、`APPROVED`、`INVESTED` 或 `DISBURSED`；关联借款人、审批、投资记录和放款记录。
- `Approval`：贷款审批记录；`LoanID` 使用唯一索引，每笔贷款至多一条记录。
- `Investment`：贷款投资记录；关联一笔贷款和一位投资人。
- `Disbursement`：贷款放款记录；`LoanID` 使用唯一索引，每笔贷款至多一条记录。

这些结构包含相应的 GORM 标签。当前没有迁移文件，不会自动建表。

## API

服务默认监听 `:8080`。

| 方法 | 路径 | 当前行为 |
| --- | --- | --- |
| `GET` | `/health` | 返回 `{"message":"ok"}` |
| `POST` | `/api/v1/loans` | 接收创建贷款 DTO，并返回基于请求字段组装的 mock 贷款响应 |
| `GET` | `/api/v1/loans/:id` | 返回固定的 mock 贷款数据 |
| `POST` | `/api/v1/loans/:id/approve` | 接收审批 DTO，返回 `{"message":"mock response"}` |
| `POST` | `/api/v1/loans/:id/invest` | 接收投资 DTO，返回 `{"message":"mock response"}` |
| `POST` | `/api/v1/loans/:id/disburse` | 接收放款 DTO，返回 `{"message":"mock response"}` |

创建贷款请求示例：

```json
{
  "borrowerId": 1,
  "principal": 100000,
  "rate": 0.05,
  "roi": 0.08,
  "agreementLetterUrl": "https://example.com/agreement.pdf"
}
```

其余请求字段定义在 `internal/dto/loan_request.go`。请求 JSON 无法绑定时，Handler 返回 HTTP `400` 和错误信息。当前贷款 ID 不会触发查询或状态变化，接口不执行贷款业务规则。

## 配置与运行

默认情况下不需要 MySQL：

```bash
go run ./cmd/server
```

可通过环境变量设置服务地址：

```powershell
$env:HTTP_ADDR = ":8081"
go run ./cmd/server
```

`MYSQL_DSN` 用于 `config.OpenMySQL` 的数据库连接初始化入口；当前应用启动流程不调用该函数，因此不会连接数据库。调用 `OpenMySQL` 前必须设置 `MYSQL_DSN`。

## 当前范围

此版本提供可运行的 HTTP 服务、项目分层、GORM 模型、DTO、MySQL 连接初始化入口及 mock API。以下能力尚未实现：

- 数据库连接集成到应用启动流程
- 数据库迁移、建表及 CRUD
- Repository 与 Service 的数据访问接线
- 贷款状态流转及审批、投资、放款业务校验
- 身份认证、文件上传、通知等附加功能
