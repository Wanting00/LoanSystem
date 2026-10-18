# Loan Service

一个使用 Go、Gin、GORM 和 MySQL 搭建的贷款服务。贷款创建、审批、投资和放款接口通过分层调用将数据持久化到 MySQL；贷款查询接口仍为 mock 实现。

## 技术栈

- Go 1.22+
- Gin：HTTP 路由与请求处理
- GORM：ORM
- MySQL 驱动：贷款数据持久化

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
│   │   └── loan_repository.go      # GORM 数据访问
│   └── service/
│       └── loan_service.go         # 贷款创建、审批业务逻辑及其他 mock 响应
├── go.mod
└── go.sum
```

## 分层调用

```text
HTTP Request → Gin Router → Handler → Service → Repository → MySQL
```

- `handler` 负责路由、JSON 请求绑定和 HTTP 响应。
- `service` 设置新贷款的 `PROPOSED` 初始状态，并将创建结果映射为响应 DTO。
- `repository` 使用注入的 GORM 数据库连接保存贷款。
- `model` 定义领域/数据库结构及 GORM 字段标签。
- `dto` 定义 API 请求与响应的数据形状。
- `config` 读取 HTTP 地址及 MySQL 连接配置，并提供数据库打开函数。

## 数据模型

`internal/model` 中定义了以下结构：

- `Borrower`：借款人，通过 `Loans` 与贷款建立一对多关系。
- `Investor`：投资人，通过 `Investments` 与投资记录建立一对多关系。
- `Loan`：贷款，状态为 `PROPOSED`、`APPROVED`、`INVESTED` 或 `DISBURSED`；关联借款人、审批、投资记录和放款记录。
- `Approval`：贷款审批记录；`LoanID` 使用唯一索引，每笔贷款至多一条记录。
- `Investment`：贷款投资记录；关联一笔贷款和一位投资人。
- `Disbursement`：贷款放款记录；`LoanID` 使用唯一索引，每笔贷款至多一条记录。

这些结构包含相应的 GORM 标签。应用启动不会自动建表或执行迁移；请确保数据库和数据表已预先初始化。

## API

服务默认监听 `:8080`。

| 方法 | 路径 | 当前行为 |
| --- | --- | --- |
| `GET` | `/health` | 返回 `{"message":"ok"}` |
| `POST` | `/api/v1/loans/create` | 接收创建贷款 DTO，持久化新贷款并返回数据库记录（创建成功 `201`，请求绑定错误 `400`，持久化失败 `500`） |
| `GET` | `/api/v1/loans/:id` | 返回固定的 mock 贷款数据 |
| `POST` | `/api/v1/loans/:id/approve` | 审批 `PROPOSED` 贷款；成功返回 `200`，贷款不存在返回 `404`，状态不允许或审批记录冲突返回 `409`，ID 或请求错误返回 `400` |
| `POST` | `/api/v1/loans/:id/invest` | 投资已审批的贷款；成功返回 `200`，Loan 或 Investor 不存在返回 `404`，金额无效返回 `400`，贷款状态不允许或超出剩余本金返回 `409` |
| `POST` | `/api/v1/loans/:id/disburse` | 为状态为 `INVESTED` 的贷款创建放款记录并更新为 `DISBURSED`；成功返回 `201`，Loan 不存在返回 `404`，请求无效返回 `400`，状态不允许或重复放款返回 `409` |

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

审批请求需要提供 `validatorPhotoUrl`、`employeeId` 和 `approvalDate`；审批记录创建与贷款状态更新在同一事务中执行，并通过 MySQL `SELECT ... FOR UPDATE` 锁定贷款行。投资请求使用 `investorId` 和正数 `amount`；投资记录创建、贷款状态检查及筹满时的状态更新在同一事务中执行。每次投资先锁定贷款行，再汇总该贷款已有的投资金额，因此同一贷款的并发投资会串行进行。放款请求需要提供 `signedAgreementUrl`（HTTP(S) PDF/JPEG 链接）、`employeeId` 和 RFC3339 格式的 `disbursementDate`；放款记录创建与状态更新同样在单个事务中执行，先通过行锁读取最新状态，且利用 `disbursements.loan_id` 唯一索引避免重复记录。URL 扩展名校验不能证明远程文件的真实内容。相关数据库表必须使用支持事务和行级锁的存储引擎（例如 InnoDB）。`investments.loan_id` 已建立索引，Disbursement 唯一索引通过 `go run ./cmd/migrate` 的 AutoMigrate 管理；数据库结构变更后请运行迁移命令。请求 JSON 无法绑定时，Handler 返回 HTTP `400` 和错误信息。贷款查询接口当前仍返回 mock 数据。

当一笔投资使贷款首次募集完成时，投资事务提交成功后会启动 goroutine，并在最多 30 秒的独立上下文中调用 `NotificationService.CreateTasks`。当前该方法仅为 Mock，直接返回成功，不访问数据库、不创建通知任务，也不发送邮件。

## 配置与运行

```powershell
$env:HTTP_ADDR = ":8081"
$env:MYSQL_HOST = "127.0.0.1"
$env:MYSQL_PORT = "3306"
$env:MYSQL_DATABASE = "loan_system"
$env:MYSQL_USER = "your-user"
$env:MYSQL_PASSWORD = "your-password"
go run ./cmd/server
```

`config.OpenMySQL` 使用以上分项连接配置；当前代码没有读取 `MYSQL_DSN`。应用不会在启动时创建数据库或迁移数据表。

## 当前范围

贷款创建、审批、投资和放款接口已接入 Repository 和 MySQL。以下能力仍未实现：
- `GET /api/v1/loans/:id` 的数据库查询
- 身份认证、文件上传及实际通知处理等附加功能（通知服务当前为 Mock）

投资累计金额当前通过对 `investments.amount` 执行 `SUM` 实时计算。如果投资记录数量显著增长且聚合查询成为性能瓶颈，未来可考虑维护冗余的累计投资金额字段；该优化需要额外保证投资记录与累计金额的一致性，本次未实现。
