# STAROps SDK Samples 🚀

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](samples/golang/)
[![Java](https://img.shields.io/badge/Java-11+-orange?style=flat&logo=openjdk)](samples/java/)
[![Java8](https://img.shields.io/badge/Java-8+-orange?style=flat&logo=openjdk)](samples/java8/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB?style=flat&logo=python)](samples/python/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](samples/typescript/)

**Production-ready, multi-language SSE streaming chat clients for Alibaba Cloud STAROps digital employees.**

[中文文档](README_zh.md)

Talk to a STAROps digital employee in plain language and stream its answers in real time.
This repository ships the **same client** in five languages — Go, Python, TypeScript, Java, and Java 8 —
each with resilient reconnection, message deduplication, and a consistent CLI.

---

## Features

- **SSE streaming** — real-time, chunked responses over HTTP POST + Server-Sent Events.
- **Automatic reconnection** — dropped connections and idle timeouts trigger a transparent resume.
- **Exponential backoff** — retry delays grow `1s → 2s → 4s → 8s → 16s → 30s`, up to 10 attempts.
- **Message deduplication** — already-received messages are skipped after a reconnect (timestamp-based).
- **Idle-timeout detection** — reconnect if no message arrives within a configurable window (default 60s).
- **Credential chain** — direct AK/SK or the Alibaba Cloud default credential chain.
- **Consistent CLI** — interactive chat, file-driven batch chat, and thread management in every language.
- **Five languages, one behavior** — identical semantics and configuration across all SDKs.

---

## Project Structure

```
.
├── README.md / README_zh.md         # This document
├── requests/starops/                # Shared request JSON files (reusable across languages)
│   ├── general_chat.json            # General Q&A
│   ├── entity.json                  # Entity metric query
│   ├── sql_generation.json          # Natural language → SQL
│   └── data_agent.json              # Data-agent analysis
└── samples/
    ├── golang/                      # Go 1.21+
    ├── java/                        # Java 11+
    ├── java8/                       # Java 8+ (syntax-compatible fork)
    ├── python/                      # Python 3.8+
    └── typescript/                  # TypeScript 5.0+ (Node 18+)
```

---

## Quick Start

### Prerequisites

- An Alibaba Cloud account with STAROps access.
- An Access Key ID and Secret (or a configured [credential chain](#credentials)).
- The toolchain for your chosen language (see the tables below).

### Configure

Every sample reads the same environment variables from a `.env` file:

```bash
cd samples/<language>
cp .env.example .env   # then edit .env with your credentials and endpoint
```

### Run

| Language | Command |
|----------|---------|
| **Go** | `go run ./cmd/chat/` |
| **Python** | `python -m starops_sdk_samples.examples.chat` |
| **TypeScript** | `npm run chat` |
| **Java** | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |
| **Java 8** | `mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"` |

> [!TIP]
> New here? Start with the **Go** sample — it has the fastest edit-run loop.

---

## Language Samples

| Language | Min Version | Directory | Guide |
|----------|-------------|-----------|-------|
| **Go** | Go 1.21+ | [`samples/golang/`](samples/golang/) | [README](samples/golang/README.md) · [中文](samples/golang/README_zh.md) |
| **Python** | Python 3.8+ | [`samples/python/`](samples/python/) | [README](samples/python/README.md) · [中文](samples/python/README_zh.md) |
| **TypeScript** | Node 18+ | [`samples/typescript/`](samples/typescript/) | [README](samples/typescript/README.md) · [中文](samples/typescript/README_zh.md) |
| **Java** | Java 11+ | [`samples/java/`](samples/java/) | [README](samples/java/README.md) · [中文](samples/java/README_zh.md) |
| **Java 8** | Java 8+ | [`samples/java8/`](samples/java8/) | Syntax-compatible fork of `java/` |

> [!NOTE]
> `samples/java8/` is a syntax-compatible fork of the Java 11 project for environments limited to Java 8.
> Its functionality is identical.

---

## Environment Variables

All languages share the same variables (loaded from `.env`):

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VIBEOPS_ENDPOINT` | ✅ | — | STAROps API endpoint, e.g. `starops.cn-beijing.aliyuncs.com` |
| `VIBEOPS_WORKSPACE` | ✅ | — | Workspace ID |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ✅* | — | Access Key ID |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ✅* | — | Access Key Secret |
| `VIBEOPS_REGION` | ❌ | `cn-hangzhou` | Region (should match the endpoint) |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | `default` | Digital employee name |
| `VIBEOPS_MAX_RETRIES` | ❌ | `10` | Max SSE reconnect attempts |
| `VIBEOPS_IDLE_TIMEOUT` | ❌ | `60` | Idle timeout (seconds); reconnect if no message arrives within this window |

> [!IMPORTANT]
> \*AK/SK are optional **only** when a credential chain is available (see [Credentials](#credentials)).
> Without either, the client cannot authenticate.

### Credentials

Two ways to authenticate, in priority order:

1. **Environment variables** — set `ALIBABA_CLOUD_ACCESS_KEY_ID` and `ALIBABA_CLOUD_ACCESS_KEY_SECRET` directly.
2. **Default credential chain** — when AK/SK are absent, the client falls back to the Alibaba Cloud
   default chain (environment → config file `~/.alibabacloud/credentials` → ECS RAM role → OIDC → IMDSv2).

> [!TIP]
> Prefer the credential chain in production to avoid hardcoding secrets.
> (Credential-chain fallback is fully implemented in the Go sample.)

---

## SSE Retry & Reconnection

Every sample implements the **same** resilient streaming behavior:

```
create ──► stream events ──► stream_done ✅ (normal completion)
              │
              ├─ connection dropped ─┐
              ├─ idle timeout ───────┤──► backoff ──► reconnect (action="reconnect") ──► dedupe ──► resume
              └─ SSE error ──────────┘
```

- **Normal completion** is marked by a `stream_done` event.
- If the stream ends **before** `stream_done`, the client reconnects automatically.
- **Backoff** grows exponentially: `1s, 2s, 4s, 8s, 16s, 30s` (capped at 30s), up to `VIBEOPS_MAX_RETRIES` times.
- On reconnect the client sends `action="reconnect"` and **deduplicates** by message timestamp, so no
  message is delivered twice.

> [!NOTE]
> After exceeding the maximum retries, the client surfaces an error instead of hanging.

---

## Command-Line Flags

All chat samples accept the same flags:

| Flag | Applies to | Description |
|------|-----------|-------------|
| `-simulate-error` | `chat`, `chat-from-file` | Simulate a network disconnection to exercise the retry logic |
| `-file <path>` | `chat-from-file` | Load a single request from a JSON file |
| `-dir <path>` | `chat-from-file` | Batch-process every JSON request in a directory |
| `-simple` | `chat-from-file` | Text-only output (Python / TypeScript) |

> [!TIP]
> Verify the reconnection path end-to-end with `-simulate-error`, e.g. `go run ./cmd/chat/ -simulate-error`.
> The client drops the first stream, backs off, reconnects, deduplicates, and finishes at `stream_done`.

---

## Sample Programs

Each language provides the same three programs:

| Program | Description |
|---------|-------------|
| `chat` | Interactive multi-turn chat with context preserved across turns |
| `chat-from-file` | Run a request from a JSON file, or batch-process a directory |
| `thread-manager` | List, inspect, and delete conversation threads |

### Shared request files

`requests/starops/` holds reusable request JSON files. The recommended workflow: copy the closest sample,
edit your parameters, then run it via `chat-from-file`.

```bash
cd samples/golang
go run ./cmd/chat-from-file/ -file ../../requests/starops/entity.json
go run ./cmd/chat-from-file/ -dir  ../../requests/starops/
```

| File | Scenario |
|------|----------|
| `general_chat.json` | General Q&A |
| `entity.json` | Entity metric query (with entity context) |
| `sql_generation.json` | Natural language → SQL (`skill: sql_generation`) |
| `data_agent.json` | Data-agent analysis (`skill: data-agent-pro`) |

---

## How It Works

The core interaction model is **natural language conversation**: you describe a need, and the digital
employee understands, queries, analyzes, and answers.

```
┌───────────────┐       HTTP POST         ┌──────────────────────┐
│   Your App    │  ───────────────────►   │  STAROps Digital      │
│  (SDK Client) │  ◄── SSE Streaming ──   │  Employee API         │
└───────────────┘                         └──────────────────────┘
```

1. **Create a thread** to obtain a `threadId` for multi-turn context.
2. **Send a message** with your natural-language question plus context variables.
3. **Stream the response** as structured `MessageItem` objects arrive over SSE.
4. **Continue** by reusing the same `threadId`.

For the full request/response schema and per-scenario parameters, see the language guides above and the
files under [`requests/starops/`](requests/starops/).
