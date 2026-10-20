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
- `Investor`：投资人，包含邮箱地址，通过 `Investments` 与投资记录建立一对多关系。
- `Loan`：贷款，状态为 `PROPOSED`、`APPROVED`、`INVESTED` 或 `DISBURSED`；关联借款人、审批、投资记录和放款记录。借款人、本金、利率、ROI、协议链接、状态及幂等 key 在创建时必填并由数据库禁止 `NULL`；`IdempotencyKey` 保留原有唯一索引。审批、投资和放款关联记录在各自生命周期阶段创建。
- `Approval`：贷款审批记录；`LoanID` 使用唯一索引，每笔贷款至多一条记录。
- `Investment`：贷款投资记录；关联一笔贷款和一位投资人。
- `Disbursement`：贷款放款记录；`LoanID` 使用唯一索引，每笔贷款至多一条记录。
- `NotificationLog`：投资满额后的邮件通知记录；为每笔投资保存接收邮箱、协议链接和发送状态。

这些结构包含相应的 GORM 标签。应用启动不会自动执行迁移；请先运行 `go run ./cmd/migrate` 创建或更新表结构，再启动服务。

## API

服务默认监听 `:8080`。

| 方法 | 路径 | 当前行为 |
| --- | --- | --- |
| `GET` | `/health` | 返回 `{"message":"ok"}` |
| `POST` | `/api/v1/loans/create` | 请求体必须提供非零 `borrowerId`、`principal`、`rate`、`roi` 及非空 `agreementLetterUrl`（URL 须为 HTTP(S) PDF/JPEG 链接；数值字段显式传 `0` 与未传区分，不额外限制数值范围）。`Idempotency-Key` 请求头需提供 1-255 个无空格可见 ASCII 字符；相同 key 和请求参数的重复请求返回 `409` 及 `loan already created`（不新增记录），相同 key 配不同参数同样返回 `409`（首次创建 `201`，无效请求/key `400`，借款人不存在 `404`，持久化失败 `500`） |
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

审批请求需要提供 `validatorPhotoUrl`、`employeeId` 和 `approvalDate`；审批记录创建与贷款状态更新在同一事务中执行，并通过 MySQL `SELECT ... FOR UPDATE` 锁定贷款行。投资请求使用 `investorId` 和正数 `amount`（DECIMAL，最多两位小数）；投资记录创建、贷款状态检查及筹满时的状态更新在同一事务中执行。每次投资先锁定贷款行，再汇总该贷款已有的投资金额，因此同一贷款的并发投资会串行进行。放款请求需要提供 `signedAgreementUrl`（HTTP(S) PDF/JPEG 链接）、`employeeId` 和 RFC3339 格式的 `disbursementDate`；放款记录创建与状态更新同样在单个事务中执行，先通过行锁读取最新状态，且利用 `disbursements.loan_id` 唯一索引避免重复记录。URL 扩展名校验不能证明远程文件的真实内容。相关数据库表必须使用支持事务和行级锁的存储引擎（例如 InnoDB）。`investments.loan_id` 已建立索引，Disbursement 唯一索引通过 `go run ./cmd/migrate` 的 AutoMigrate 管理；数据库结构变更后请运行迁移命令。请求 JSON 无法绑定时，Handler 返回 HTTP `400` 和错误信息。贷款查询接口当前仍返回 mock 数据。

创建贷款时，服务会先检查 key 是否已映射到同一组业务参数；新贷款和 key 在同一数据库事务中写入。`ux_loans_idempotency_key` 唯一索引是并发请求时的最终防线；唯一键竞争失败后，服务先结束失败事务，再从数据库重新读取并比较参数。数据库结构通过 `go run ./cmd/migrate` 的 AutoMigrate 更新；幂等 key 列设置为 `NOT NULL`，且保留其唯一索引。当前表为空，无需对历史数据回填。部署时应先备份数据库，再执行迁移命令。

创建贷款请求示例（key 通过 HTTP 请求头传递）：

```http
Idempotency-Key: create-loan-2026-10-19-001
```

数据库迁移通过 GORM `AutoMigrate` 将 `loans.idempotency_key`、`loans.rate`、`loans.roi` 和 `loans.agreement_letter_url` 更新为 `NOT NULL`（借款人、本金和状态字段已经有该约束）。在当前空表上，执行 `go run ./cmd/migrate` 后可用以下查询确认：

```sql
SELECT COLUMN_NAME, IS_NULLABLE
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'loans'
  AND COLUMN_NAME IN ('borrower_id', 'idempotency_key', 'principal', 'rate', 'roi', 'agreement_letter_url', 'status');
```

当一笔投资使贷款首次募集完成时，投资事务提交成功后会启动 goroutine，逐笔读取该贷款所有投资人的邮箱，为每笔投资创建独立的 `notification_logs` 记录并调用 `sendEmail`（当前为 Mock）；发送结果会分别写回 `sent` 或 `failed`。当前不接真实 SMTP。未来可增加定时任务周期性扫描 `pending` 记录补发；`failed` 记录留待人工处理或后续重试策略，本项目不实现定时补发或重试。

## 配置与运行

应用从项目根目录的 `.env.local` 读取配置；如果该文件不存在，则使用部署环境中已注入的环境变量。必需配置缺失或为空时，启动及迁移命令会报告对应的变量名并退出。`.env.example` 是配置模板；首次使用时请将 `MYSQL_PASSWORD` 替换为本机 MySQL 密码，不要把真实密码写入 README。

| 变量 | 用途 | 本地开发值 |
| --- | --- | --- |
| `MYSQL_HOST` | MySQL 主机地址 | `127.0.0.1` |
| `MYSQL_PORT` | MySQL 端口 | `3306` |
| `MYSQL_USER` | MySQL 用户名 | `root` |
| `MYSQL_PASSWORD` | MySQL 用户密码 | 在 `.env.local` 中填写 |
| `MYSQL_DATABASE` | 应用使用的数据库名称 | `loan_system` |
| `HTTP_ADDR` | HTTP 服务监听地址（可选） | 默认 `:8080` |

本地运行步骤：

1. 安装并启动 MySQL，创建一个可供本项目使用的用户，并确保该用户有权创建及使用数据库。
2. 复制 `.env.example` 为 `.env.local`（或编辑已提供的 `.env.local`），填写本地 MySQL 用户名和密码。代码实际读取的是 `MYSQL_*` 变量，而非 `DB_*` 变量。
3. 在项目根目录执行 `go run ./cmd/migrate`，创建 `loan_system` 数据库并迁移模型。
4. 执行 `go run ./cmd/server` 启动服务。可选地设置 `HTTP_ADDR` 修改监听地址。

`config.OpenMySQL` 使用以上分项连接配置；当前代码没有读取 `MYSQL_DSN`。应用启动不会自动迁移表结构。

## 当前范围

贷款创建、审批、投资和放款接口已接入 Repository 和 MySQL。以下能力仍未实现：
- `GET /api/v1/loans/:id` 的数据库查询
- 身份认证、文件上传及实际 SMTP 邮件发送（当前 `sendEmail` 仅为 Mock）

投资累计金额当前通过对 `investments.amount` 执行 `SUM` 实时计算。如果投资记录数量显著增长且聚合查询成为性能瓶颈，未来可考虑维护冗余的累计投资金额字段；该优化需要额外保证投资记录与累计金额的一致性，本次未实现。
