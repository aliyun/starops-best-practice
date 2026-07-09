# STAROps SDK Go 语言示例 🐹

面向阿里云 STAROps 数字员工的 Go 客户端示例，内置弹性 SSE 流式能力：自动重连、指数退避与消息去重。

## 环境要求

- **Go 1.21+**
- 拥有 STAROps 访问权限的阿里云账号及有效凭证

## 快速开始

```bash
cd samples/golang
cp .env.example .env   # 编辑 .env 填入凭证与端点
go run ./cmd/chat/
```

## 构建

```bash
make build        # 构建所有程序到 ./bin
make test         # 运行测试
make lint         # go vet + gofmt 检查
# 或直接构建单个命令：
go build -o bin/chat ./cmd/chat/
```

## 运行示例

### chat — 交互式对话

```bash
go run ./cmd/chat/
```

交互式多轮对话，在同一会话中保持上下文。

### chat-from-file — 从 JSON 文件运行请求

```bash
# 单个请求
go run ./cmd/chat-from-file/ -file ../../requests/starops/entity.json

# 批量处理目录
go run ./cmd/chat-from-file/ -dir ../../requests/starops/
```

### thread-manager — 会话管理

```bash
go run ./cmd/thread-manager/ list                # 列出会话
go run ./cmd/thread-manager/ get <thread-id>     # 查看详情
go run ./cmd/thread-manager/ delete <thread-id>  # 删除会话
```

## SSE 重试与重连

客户端通过 SSE 流式接收响应，并在连接中断时透明恢复。

```
create ──► 流式接收事件 ──► stream_done ✅（正常结束）
              │
              ├─ 通道关闭 ─────┐
              ├─ 空闲超时 ─────┤──► 退避 ──► 重连（action="reconnect"）──► 去重 ──► 续传
              └─ SSE 错误 ─────┘
```

- `stream_done` 事件表示 **正常结束**；在此之前结束的流会触发重连。
- **指数退避**：`1s, 2s, 4s, 8s, 16s, 30s`（上限 30s），最多重试 `VIBEOPS_MAX_RETRIES` 次。
- **重连** 发送 `action="reconnect"`，并复制原始 `threadId` / variables。
- **去重**：重连后按 timestamp 过滤消息，确保不重复投递。

> [!NOTE]
> 超过最大重试次数后，客户端会返回错误事件，而不会一直挂起。

重试逻辑位于 [`internal/client/retry.go`](internal/client/retry.go)。

### 测试重试逻辑

```bash
go run ./cmd/chat/ -simulate-error
```

启用 `-simulate-error` 后，客户端会：

1. 正常创建会话并发送消息。
2. 接收到首批事件后，主动模拟网络断连。
3. 输出重试日志并执行指数退避。
4. 重连后按 timestamp 去重并继续接收。
5. 完整接收所有消息直至 `stream_done`。

## 凭据管理

Go 示例按优先级支持两种凭证来源：

1. **环境变量**——`ALIBABA_CLOUD_ACCESS_KEY_ID` 与 `ALIBABA_CLOUD_ACCESS_KEY_SECRET`。
2. **默认凭据链**——AK/SK 缺失时，`LoadConfigFromEnv()` 会通过 `credentials-go` 回退到阿里云默认凭据链：
   环境变量 → 配置文件（`~/.alibabacloud/credentials`）→ ECS RAM 角色 → OIDC → IMDSv2。

> [!TIP]
> 生产环境建议使用凭据链，避免硬编码密钥。

## 环境变量

| 变量 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `VIBEOPS_ENDPOINT` | ✅ | — | STAROps API 端点，如 `starops.cn-beijing.aliyuncs.com` |
| `VIBEOPS_WORKSPACE` | ✅ | — | 工作空间 ID |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ❌* | — | Access Key ID（使用凭据链时可选） |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ❌* | — | Access Key Secret（使用凭据链时可选） |
| `VIBEOPS_REGION` | ❌ | `cn-hangzhou` | 地域（需与端点匹配） |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | `default` | 数字员工名称 |
| `VIBEOPS_MAX_RETRIES` | ❌ | `10` | SSE 最大重连次数 |
| `VIBEOPS_IDLE_TIMEOUT` | ❌ | `60` | 空闲超时秒数，超时未收到消息则重连 |

> [!IMPORTANT]
> \*当 AK/SK 未设置时，客户端使用阿里云默认凭据链。

## 命令行参数

| 参数 | 适用命令 | 说明 |
|------|----------|------|
| `-simulate-error` | `chat`、`chat-from-file` | 模拟网络断连，用于验证重试 |
| `-file <path>` | `chat-from-file` | 从单个 JSON 文件加载请求 |
| `-dir <path>` | `chat-from-file` | 批量处理目录下的所有 JSON 请求 |

## 项目结构

```
samples/golang/
├── cmd/
│   ├── chat/              # 交互式对话
│   ├── chat-from-file/    # 从 JSON 运行请求（文件或目录）
│   └── thread-manager/    # 会话管理
├── internal/
│   ├── client/            # 核心客户端：对话、会话、重试、打印器、错误
│   │   ├── client.go      # AgentClient、配置、对话方法
│   │   ├── retry.go       # SSE 重连、退避、去重
│   │   ├── credentials.go # 默认凭据链
│   │   ├── thread.go      # 会话管理 API
│   │   └── ...
│   └── logger/            # 结构化日志
├── types/                 # 事件与输入类型定义
├── Makefile
├── go.mod
└── go.sum
```

## SDK 依赖

- `github.com/alibabacloud-go/starops-20260428` — 阿里云 STAROps SDK
- `github.com/alibabacloud-go/darabonba-openapi/v2` — OpenAPI 客户端
- `github.com/alibabacloud-go/tea` — Tea 运行时
- `github.com/aliyun/credentials-go` — 阿里云凭据链
