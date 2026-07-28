# STAROps SDK 示例

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](java/)
[![Java%208](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](typescript/)

面向阿里云 **STAROps** 数字员工的多语言 SSE 流式对话客户端示例。

本仓库提供 **Go**、**Java**、**Java 8**、**Python** 和 **TypeScript** 五种语言的等价示例。
每种实现都覆盖相同核心流程：加载凭据、创建或复用会话、通过 SSE 流式接收响应、处理交互事件，
并在临时网络异常时自动恢复。

[English](README.md)

## 特性

- **多语言覆盖**：提供 Go、Java、Java 8、Python 与 TypeScript 实现。
- **SSE 流式响应**：实时接收结构化 STAROps 回复。
- **自动重连**：连接中断、SSE 错误或空闲超时时自动恢复。
- **指数退避**：使用有上限的退避策略，避免过于频繁地重试。
- **消息去重**：重连后跳过已接收消息，避免重复输出。
- **中断停止**：对话运行时按 `Ctrl+C`（SIGINT/SIGTERM）会发送 `stop` 请求，干净地中断当前响应。
- **Mock 回放**：通过 `--mock-file` 离线回放录制的 SSE 会话，无需凭据和网络，便于演示与测试。
- **凭据配置**：推荐使用阿里云 CLI 配置权限；不想安装 CLI 时可使用环境变量配置 AK/SK。
- **共享请求样例**：所有语言复用同一组 JSON 请求文件。
- **一致示例程序**：每种语言都提供交互式对话、文件请求。

> [!IMPORTANT]
> 推荐使用阿里云 CLI 配置权限。如果本地没有阿里云 CLI，请访问 [阿里云 CLI 说明](https://help.aliyun.com/zh/ros/api-operation-examples-overview) 下载安装；如果不想安装 CLI，可以使用环境变量配置 AK/SK。

## 仓库结构

```text
.
├── README.md / README_zh.md
├── sample-requests/              # 共享 STAROps 请求样例（[说明文档](sample-requests/README.md)）
├── golang/                        # Go 1.22+ 示例客户端
├── java/                          # Java 11+ 示例客户端
├── java8/                         # Java 8 兼容示例客户端
├── python/                        # Python 3.8+ 示例客户端
└── typescript/                    # TypeScript / Node.js 示例客户端
```

## 快速开始

### 1. 选择语言

| 语言 | 目录 | 运行时 | 主要命令 |
| --- | --- | --- | --- |
| Go | [`golang/`](golang/) | Go 1.22+ | `go run ./cmd/chat` |
| Java | [`java/`](java/) | Java 11+ / Maven | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |
| Java 8 | [`java8/`](java8/) | Java 8+ / Maven | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |
| Python | [`python/`](python/) | Python 3.8+ | `python -m starops_sdk_samples.examples.chat` |
| TypeScript | [`typescript/`](typescript/) | Node.js 18+ | `npm run chat` |

### 2. 配置环境变量

每个语言目录都有独立的 `.env.example`。运行示例前复制为 `.env`：

```bash
cp .env.example .env
```

`.env` 必填配置：

| 变量 | 说明 |
| --- | --- |
| `STAROPS_ENDPOINT` | STAROps 端点，例如 `starops.cn-beijing.aliyuncs.com` |
| `STAROPS_WORKSPACE` | STAROps 工作空间 ID。可选，不填写时程序也可运行 |
| `STAROPS_EMPLOYEE_NAME` | 数字员工名称 |

凭据配置：

1. **推荐方式**：使用阿里云 CLI 配置权限。
2. 如果本地没有 CLI，请访问 [阿里云 CLI 说明](https://help.aliyun.com/zh/ros/api-operation-examples-overview) 下载安装。
3. 如果不想安装 CLI，可以使用环境变量配置 AK/SK：

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID=<your-access-key-id>
export ALIBABA_CLOUD_ACCESS_KEY_SECRET=<your-access-key-secret>
```

可选重试配置：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `STAROPS_REGION` | `cn-beijing` | 数据查询所在的阿里云地域，与服务接入点（endpoint）无关 |
| `STAROPS_MAX_RETRIES` | `10` | SSE 最大重连次数 |
| `STAROPS_IDLE_TIMEOUT` | `60` | 空闲超时秒数，超时未收到消息则重连 |

### 3. 运行示例

运行交互式对话：

```bash
# Go
cd golang && go run ./cmd/chat

# Python
cd python && python -m starops_sdk_samples.examples.chat

# TypeScript
cd typescript && npm install && npm run chat
```

运行基于文件的 STAROps 请求：

```bash
# Go
cd golang && go run ./cmd/chat-from-file -file ../sample-requests/data_agent.json

# Java
cd java && mvn exec:java \
  -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
  -Dexec.args="-file ../sample-requests/data_agent.json"
```

## 共享请求样例

[`sample-requests/`](sample-requests/) 目录中的请求样例可被所有语言复用。

| 文件 | 场景 |
| --- | --- |
| `data_agent.json` | 数据 Agent 分析 |
| `sls_chat.json` | SLS 相关对话 |
| `user_ack_interactive.json` | 用户确认类交互 |
| `user_input_interactive.json` | 用户输入类交互 |
| `request_with_reminder.json` | 完整请求体示例（含 metadata + dynamic 上下文） |

## 示例程序

每种语言都提供相同的两个入口：

| 程序 | 用途 |
| --- | --- |
| `chat` | 启动交互式多轮对话 |
| `chat-from-file` | 加载单个 JSON 请求或批量处理请求目录 |

## 重试与重连模型

所有语言示例都实现一致的流式恢复行为：

```text
发送请求
  └─ 流式接收 SSE 事件
       ├─ stream_done                  -> 完成
       ├─ 连接中断                     -> 重试
       ├─ SSE 错误                     -> 重试
       └─ 空闲超时                     -> 重试
              └─ 退避后重连
                    └─ 对已接收消息去重
```

关键行为：

- 收到 `stream_done` 事件表示一次响应正常结束。
- 如果流在 `stream_done` 前结束，客户端会自动重连。
- 重试延迟采用指数退避，并设置上限，避免过大的重试压力。
- 重连后的流会跳过已投递过的消息。

可使用 `-simulate-error` 验证重试链路：

```bash
cd golang
go run ./cmd/chat-from-file -file ../sample-requests/data_agent.json -simulate-error true
```

## 开发检查

按语言运行测试：

```bash
# Go
cd golang && go test ./...

# Java
cd java && mvn test

# Java 8
cd java8 && mvn test

# Python
cd python && python -m pytest

# TypeScript
cd typescript && npm test
```

## 安全说明

- 不要将真实 AK/SK 提交到 Git。
- 不要提交包含凭据的 `.env`、IDE workspace 文件或本地调试配置。
- `.env.example` 仅用于占位和说明。
- 非本地环境优先使用 RAM 角色、OIDC、阿里云 CLI 或环境变量等安全凭据配置方式。
