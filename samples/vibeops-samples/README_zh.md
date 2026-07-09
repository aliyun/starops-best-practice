# STAROps SDK 示例 🚀

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](samples/golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](samples/java/)
[![Java8](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](samples/java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](samples/python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](samples/typescript/)

**面向阿里云 STAROps 数字员工的多语言 SSE 流式对话客户端示例，开箱即用。**

[English](README.md)

用自然语言与 STAROps 数字员工对话，并实时流式接收它的回答。
本仓库以 **同一套客户端** 提供五种语言实现——Go、Python、TypeScript、Java、Java 8——
每种实现都内置了断线重连、消息去重和一致的命令行体验。

---

## 特性

- **SSE 流式**——基于 HTTP POST + Server-Sent Events 的实时分块响应。
- **自动重连**——连接中断或空闲超时时透明恢复会话。
- **指数退避**——重试间隔按 `1s → 2s → 4s → 8s → 16s → 30s` 增长，最多 10 次。
- **消息去重**——重连后基于 timestamp 跳过已接收的消息。
- **空闲超时检测**——在可配置的时间窗口内未收到消息则重连（默认 60s）。
- **凭据链支持**——支持直接配置 AK/SK 或使用阿里云默认凭据链。
- **一致的命令行**——每种语言都提供交互式对话、文件批量对话与会话管理。
- **五种语言，行为一致**——所有 SDK 的语义与配置完全统一。

---

## 项目结构

```
.
├── README.md / README_zh.md         # 本文档
├── requests/starops/                # 共享请求 JSON 文件（多语言复用）
│   ├── general_chat.json            # 通用问答
│   ├── entity.json                  # 实体指标查询
│   ├── sql_generation.json          # 自然语言 → SQL
│   └── data_agent.json              # 数据 Agent 分析
└── samples/
    ├── golang/                      # Go 1.21+
    ├── java/                        # Java 11+
    ├── java8/                       # Java 8+（语法兼容分支）
    ├── python/                      # Python 3.8+
    └── typescript/                  # TypeScript 5.0+（Node 18+）
```

---

## 快速开始

### 前置条件

- 拥有 STAROps 访问权限的阿里云账号。
- Access Key ID 和 Secret（或已配置的[凭据链](#凭据管理)）。
- 目标语言对应的工具链（见下方表格）。

### 配置

所有示例都从 `.env` 文件读取相同的环境变量：

```bash
cd samples/<language>
cp .env.example .env   # 编辑 .env 填入你的凭证与端点
```

### 运行

| 语言 | 命令 |
|------|------|
| **Go** | `go run ./cmd/chat/` |
| **Python** | `python -m starops_sdk_samples.examples.chat` |
| **TypeScript** | `npm run chat` |
| **Java** | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |
| **Java 8** | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |

> [!TIP]
> 第一次上手？建议从 **Go** 示例开始——它的「改-跑」循环最快。

---

## 各语言示例

| 语言 | 最低版本 | 目录 | 文档 |
|------|----------|------|------|
| **Go** | Go 1.21+ | [`samples/golang/`](samples/golang/) | [README](samples/golang/README.md) · [中文](samples/golang/README_zh.md) |
| **Python** | Python 3.8+ | [`samples/python/`](samples/python/) | [README](samples/python/README.md) · [中文](samples/python/README_zh.md) |
| **TypeScript** | Node 18+ | [`samples/typescript/`](samples/typescript/) | [README](samples/typescript/README.md) · [中文](samples/typescript/README_zh.md) |
| **Java** | Java 11+ | [`samples/java/`](samples/java/) | [README](samples/java/README.md) · [中文](samples/java/README_zh.md) |
| **Java 8** | Java 8+ | [`samples/java8/`](samples/java8/) | `java/` 的语法兼容分支 |

> [!NOTE]
> `samples/java8/` 是 Java 11 项目的语法兼容分支，适用于只能使用 Java 8 的环境，功能完全一致。

---

## 环境变量

所有语言共享相同的变量（从 `.env` 加载）：

| 变量 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `VIBEOPS_ENDPOINT` | ✅ | — | STAROps API 端点，如 `starops.cn-beijing.aliyuncs.com` |
| `VIBEOPS_WORKSPACE` | ✅ | — | 工作空间 ID |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ✅* | — | Access Key ID |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ✅* | — | Access Key Secret |
| `VIBEOPS_REGION` | ❌ | `cn-hangzhou` | 地域（需与端点匹配） |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | `default` | 数字员工名称 |
| `VIBEOPS_MAX_RETRIES` | ❌ | `10` | SSE 最大重连次数 |
| `VIBEOPS_IDLE_TIMEOUT` | ❌ | `60` | 空闲超时秒数，超时未收到消息则重连 |

> [!IMPORTANT]
> \*仅当存在可用的凭据链时，AK/SK 才是可选的（见[凭据管理](#凭据管理)）。
> 两者皆无时客户端无法完成鉴权。

### 凭据管理

按优先级提供两种鉴权方式：

1. **环境变量**——直接设置 `ALIBABA_CLOUD_ACCESS_KEY_ID` 与 `ALIBABA_CLOUD_ACCESS_KEY_SECRET`。
2. **默认凭据链**——AK/SK 缺失时，回退到阿里云默认凭据链
   （环境变量 → 配置文件 `~/.alibabacloud/credentials` → ECS RAM 角色 → OIDC → IMDSv2）。

> [!TIP]
> 生产环境建议使用凭据链，避免硬编码密钥。（凭据链回退已在 Go 示例中完整实现。）

---

## SSE 重试与重连

每种语言的示例都实现了 **相同** 的弹性流式行为：

```
create ──► 流式接收事件 ──► stream_done ✅（正常结束）
              │
              ├─ 连接中断 ─────┐
              ├─ 空闲超时 ─────┤──► 退避 ──► 重连（action="reconnect"）──► 去重 ──► 续传
              └─ SSE 错误 ─────┘
```

- `stream_done` 事件表示 **正常结束**。
- 如果流在 `stream_done` **之前** 结束，客户端会自动重连。
- **退避** 按指数增长：`1s, 2s, 4s, 8s, 16s, 30s`（上限 30s），最多重试 `VIBEOPS_MAX_RETRIES` 次。
- 重连时客户端发送 `action="reconnect"`，并基于消息 timestamp **去重**，确保消息不重复投递。

> [!NOTE]
> 超过最大重试次数后，客户端会抛出错误，而不会一直挂起。

---

## 命令行参数

所有对话示例接受相同的参数：

| 参数 | 适用命令 | 说明 |
|------|----------|------|
| `-simulate-error` | `chat`、`chat-from-file` | 模拟网络断连，用于验证重试逻辑 |
| `-file <path>` | `chat-from-file` | 从单个 JSON 文件加载请求 |
| `-dir <path>` | `chat-from-file` | 批量处理目录下的所有 JSON 请求 |
| `-simple` | `chat-from-file` | 仅输出文本（Python / TypeScript） |

> [!TIP]
> 使用 `-simulate-error` 可端到端验证重连链路，例如 `go run ./cmd/chat/ -simulate-error`。
> 客户端会主动断开首个流、执行退避、重连、去重，并最终在 `stream_done` 处完成。

---

## 示例程序

每种语言都提供相同的三个程序：

| 程序 | 说明 |
|------|------|
| `chat` | 交互式多轮对话，跨轮次保持上下文 |
| `chat-from-file` | 从 JSON 文件运行请求，或批量处理整个目录 |
| `thread-manager` | 列出、查看、删除会话 |

### 共享请求文件

`requests/starops/` 存放可复用的请求 JSON 文件。推荐用法：复制最接近的示例，
修改参数后通过 `chat-from-file` 运行。

```bash
cd samples/golang
go run ./cmd/chat-from-file/ -file ../../requests/starops/entity.json
go run ./cmd/chat-from-file/ -dir  ../../requests/starops/
```

| 文件 | 场景 |
|------|------|
| `general_chat.json` | 通用问答 |
| `entity.json` | 实体指标查询（带实体上下文） |
| `sql_generation.json` | 自然语言 → SQL（`skill: sql_generation`） |
| `data_agent.json` | 数据 Agent 分析（`skill: data-agent-pro`） |

---

## 工作原理

核心交互模式是 **自然语言对话**：你描述需求，数字员工理解语义后自动完成查询、分析与回答。

```
┌───────────────┐       HTTP POST         ┌──────────────────────┐
│   你的应用     │  ───────────────────►   │  STAROps 数字员工      │
│  (SDK 客户端)  │  ◄── SSE 流式响应 ──    │  API                  │
└───────────────┘                         └──────────────────────┘
```

1. **创建会话**，获取用于多轮上下文的 `threadId`。
2. **发送消息**，携带自然语言问题与上下文变量。
3. **流式接收** 服务端通过 SSE 推送的结构化 `MessageItem`。
4. **续聊**，复用同一 `threadId` 继续对话。

完整的请求/响应结构与各场景参数，请参考上方的语言文档与
[`requests/starops/`](requests/starops/) 目录下的文件。
