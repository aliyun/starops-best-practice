# STAROps SDK Samples

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](java/)
[![Java%208](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](typescript/)

Production-ready multi-language samples for building SSE streaming chat clients with Alibaba Cloud **STAROps** digital employees.

This repository provides equivalent sample clients in **Go**, **Java**, **Java 8**, **Python**, and **TypeScript**.
Each implementation demonstrates the same core workflow: load credentials, create or reuse a conversation thread,
stream responses over SSE, handle interactive events, and recover from transient stream failures.

[中文文档](README_zh.md)

## Features

- **Multi-language coverage**: Go, Java, Java 8, Python, and TypeScript implementations.
- **SSE streaming**: Stream structured STAROps responses in real time.
- **Automatic reconnection**: Resume after dropped connections, SSE errors, or idle timeouts.
- **Exponential backoff**: Retry with bounded backoff to avoid aggressive reconnect loops.
- **Message deduplication**: Skip already received messages after reconnecting.
- **Stop on interrupt**: Press `Ctrl+C` (SIGINT/SIGTERM) during a running chat to send a `stop` request and cleanly interrupt the ongoing response.
- **Mock playback**: Replay recorded SSE sessions offline via `--mock-file` (no credentials or network needed), useful for demos and tests.
- **Credential configuration**: Prefer Alibaba Cloud CLI-based credential configuration; environment variables can be used when CLI installation is not desired.
- **Shared request fixtures**: Reuse the same JSON request files across languages.
- **Model selection**: Specify a model at session start (`--model`) or switch interactively (`/model`) during a conversation.
- **Consistent examples**: Run interactive chat, file-based chat, and thread management in every language.

> [!IMPORTANT]
> We recommend configuring credentials with Alibaba Cloud CLI. If you do not have the CLI, install it from the [Alibaba Cloud CLI guide](https://help.aliyun.com/zh/ros/api-operation-examples-overview). If you do not want to install it, configure AK/SK with environment variables.

## Repository layout

```text
.
├── README.md / README_zh.md
├── config/                           # Shared model configuration
├── sample-requests/              # Shared STAROps request examples ([docs](sample-requests/README.md))
├── golang/                        # Go 1.22+ sample client
├── java/                          # Java 11+ sample client
├── java8/                         # Java 8-compatible sample client
├── python/                        # Python 3.8+ sample client
└── typescript/                    # TypeScript / Node.js sample client
```

## Quick start

### 1. Choose a language

| Language | Directory | Runtime | Main commands |
| --- | --- | --- | --- |
| Go | [`golang/`](golang/) | Go 1.22+ | `go run ./cmd/chat` |
| Java | [`java/`](java/) | Java 11+ / Maven | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |
| Java 8 | [`java8/`](java8/) | Java 8+ / Maven | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |
| Python | [`python/`](python/) | Python 3.8+ | `python -m starops_sdk_samples.examples.chat` |
| TypeScript | [`typescript/`](typescript/) | Node.js 18+ | `npm run chat` |

### 2. Configure environment variables

Each language has its own `.env.example`. Copy it to `.env` before running a sample:

```bash
cp .env.example .env
```

Required `.env` values:

| Variable | Description |
| --- | --- |
| `STAROPS_ENDPOINT` | STAROps endpoint, for example `starops.cn-beijing.aliyuncs.com` |
| `STAROPS_WORKSPACE` | STAROps workspace ID. Optional. The sample runs without it. |
| `STAROPS_EMPLOYEE_NAME` | Digital employee name |

Credential configuration:

1. **Recommended**: configure credentials with Alibaba Cloud CLI.
2. If you do not have the CLI, install it from the [Alibaba Cloud CLI guide](https://help.aliyun.com/zh/ros/api-operation-examples-overview).
3. If you do not want to install the CLI, configure AK/SK with environment variables:

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID=<your-access-key-id>
export ALIBABA_CLOUD_ACCESS_KEY_SECRET=<your-access-key-secret>
```

Optional retry settings:

| Variable | Default | Description |
| --- | --- | --- |
| `STAROPS_REGION` | `cn-beijing` | Alibaba Cloud region where data queries are performed; independent of the service endpoint |
| `STAROPS_MAX_RETRIES` | `10` | Maximum SSE reconnect attempts |
| `STAROPS_IDLE_TIMEOUT` | `60` | Seconds to wait before treating a stream as idle |
| `STAROPS_LOG_LEVEL` | `info` | Log level: `debug` / `info` / `warn` / `error` |

> [!TIP]
> `.env` is only for STAROps sample settings. Configure credentials with Alibaba Cloud CLI or AK/SK environment variables.

### 3. Run a sample

Run an interactive chat:

```bash
# Go
cd golang && go run ./cmd/chat

# Python
cd python && python -m starops_sdk_samples.examples.chat

# TypeScript
cd typescript && npm install && npm run chat
```

Run a file-based STAROps request:

```bash
# Go
cd golang && go run ./cmd/chat-from-file -file ../sample-requests/data_agent.json

# Java
cd java && mvn exec:java \
  -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
  -Dexec.args="-file ../sample-requests/data_agent.json"
```

## Shared request examples

The [`sample-requests/`](sample-requests/) directory contains request fixtures that can be reused by every language sample.

| File | Scenario |
| --- | --- |
| `data_agent.json` | Data-agent analysis |
| `sls_chat.json` | SLS-focused chat |
| `user_ack_interactive.json` | User acknowledgement interaction |
| `user_input_interactive.json` | User input interaction |
| `data_agent_with_model.json` | Data-agent analysis with model selection |
| `request_with_reminder.json` | Complete request body with metadata + dynamic context |

## Example programs

Each language provides the same two entry points:

| Program | Purpose |
| --- | --- |
| `chat` | Start an interactive multi-turn conversation |
| `chat-from-file` | Load one JSON request or process a directory of requests |

## Model selection

Support specifying a model when creating a conversation and dynamically switching models during a conversation.
All languages share a single model registry at `config/models.json`.

### Shared configuration

| Item | Description |
| --- | --- |
| Config file | `config/models.json` — single source of truth for all languages |
| Override path | Set `STAROPS_MODELS_CONFIG` environment variable to use a custom path |

### CLI flags

| Flag | Description |
| --- | --- |
| `--model provider:modelId` | Specify a model when creating a session (e.g. `--model qwen:qwen3.8-max`) |
| `--list-models` | List all available model values and exit (no credentials required) |

### Interactive command

Type `/model` during a conversation to switch models interactively.
The client displays a menu, you select a number, and it calls `UpdateThread` to apply the change.

### Language examples

```bash
# Go
cd golang && go run ./cmd/chat/ --model qwen:qwen3.8-max

# Python
cd python && python -m starops_sdk_samples.examples.chat --model qwen:qwen3.8-max

# TypeScript
cd typescript && npm run chat -- --model qwen:qwen3.8-max

# Java / Java 8
cd java && mvn exec:java \
  -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat" \
  -Dexec.args="--model qwen:qwen3.8-max"
```

### Priority

Model selection priority (highest to lowest):

1. `CreateChat` request `variables.config.model`
2. Thread `attributes.model` (set via `UpdateThread` / `/model` command)
3. Agent default configuration

### File-based chat with model

See [`sample-requests/data_agent_with_model.json`](sample-requests/data_agent_with_model.json) for passing model configuration in a file request via `variables.config`.

## Retry and reconnection model

The samples implement the same stream recovery behavior across languages:

```text
send request
  └─ stream SSE events
       ├─ stream_done                  -> finish
       ├─ connection dropped           -> retry
       ├─ SSE error                    -> retry
       └─ idle timeout                 -> retry
              └─ reconnect with backoff
                    └─ deduplicate received messages
```

Key behaviors:

- A normal response finishes when a `stream_done` event is received.
- If a stream ends before `stream_done`, the client reconnects automatically.
- Retry delays use exponential backoff and are capped to avoid excessive retry pressure.
- Reconnected streams deduplicate messages that were already delivered.

You can exercise the retry path with `-simulate-error`:

```bash
cd golang
go run ./cmd/chat-from-file -file ../sample-requests/data_agent.json -simulate-error true
```

## Development checks

Run tests for a specific language:

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

## Security notes

- Keep real AK/SK values out of Git.
- Do not commit `.env`, IDE workspace files, or local debug launch files containing credentials.
- Use `.env.example` only for placeholders and documentation.
- Prefer RAM roles, OIDC, or the Alibaba Cloud credential chain for non-local environments.
