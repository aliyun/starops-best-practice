# STAROps SDK Samples for Go 🐹

Go client samples for Alibaba Cloud STAROps digital employees, featuring resilient SSE streaming with
automatic reconnection, exponential backoff, and message deduplication.

## Requirements

- **Go 1.21+**
- An Alibaba Cloud account with STAROps access and valid credentials

## Quick Start

```bash
cd samples/golang
cp .env.example .env   # edit .env with your credentials and endpoint
go run ./cmd/chat/
```

## Build

```bash
make build        # build all programs into ./bin
make test         # run tests
make lint         # go vet + gofmt check
# or build a single command directly:
go build -o bin/chat ./cmd/chat/
```

## Running the Samples

### chat — interactive chat

```bash
go run ./cmd/chat/
```

Multi-turn interactive chat with context preserved within a thread.

### chat-from-file — run requests from JSON

```bash
# Single request
go run ./cmd/chat-from-file/ -file ../../requests/starops/entity.json

# Batch-process a directory
go run ./cmd/chat-from-file/ -dir ../../requests/starops/
```

### thread-manager — manage threads

```bash
go run ./cmd/thread-manager/ list                # list threads
go run ./cmd/thread-manager/ get <thread-id>     # thread details
go run ./cmd/thread-manager/ delete <thread-id>  # delete a thread
```

## SSE Retry & Reconnection

The client streams responses over SSE and recovers transparently from interruptions.

```
create ──► stream events ──► stream_done ✅ (normal completion)
              │
              ├─ channel closed ─────┐
              ├─ idle timeout ───────┤──► backoff ──► reconnect (action="reconnect") ──► dedupe ──► resume
              └─ SSE error ──────────┘
```

- **Normal completion** is marked by a `stream_done` event; a stream ending before it triggers a reconnect.
- **Exponential backoff**: `1s, 2s, 4s, 8s, 16s, 30s` (capped at 30s), up to `VIBEOPS_MAX_RETRIES` attempts.
- **Reconnect** sends `action="reconnect"` and copies the original `threadId` / variables.
- **Deduplication**: after reconnecting, messages are filtered by timestamp so none are delivered twice.

> [!NOTE]
> After exceeding the maximum retries, the client returns an error event instead of hanging.

The retry logic lives in [`internal/client/retry.go`](internal/client/retry.go).

### Testing the retry logic

```bash
go run ./cmd/chat/ -simulate-error
```

With `-simulate-error`, the client:

1. Creates a thread and sends a message normally.
2. After the first events arrive, simulates a network disconnection.
3. Logs the retry and performs exponential backoff.
4. Reconnects, deduplicates by timestamp, and continues.
5. Receives all messages through to `stream_done`.

## Credentials

The Go sample supports two credential sources, in priority order:

1. **Environment variables** — `ALIBABA_CLOUD_ACCESS_KEY_ID` and `ALIBABA_CLOUD_ACCESS_KEY_SECRET`.
2. **Default credential chain** — when AK/SK are absent, `LoadConfigFromEnv()` falls back to the Alibaba
   Cloud default chain via `credentials-go`:
   environment → config file (`~/.alibabacloud/credentials`) → ECS RAM role → OIDC → IMDSv2.

> [!TIP]
> Prefer the credential chain in production to avoid hardcoding secrets.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VIBEOPS_ENDPOINT` | ✅ | — | STAROps API endpoint, e.g. `starops.cn-beijing.aliyuncs.com` |
| `VIBEOPS_WORKSPACE` | ✅ | — | Workspace ID |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ❌* | — | Access Key ID (optional when using the credential chain) |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ❌* | — | Access Key Secret (optional when using the credential chain) |
| `VIBEOPS_REGION` | ❌ | `cn-hangzhou` | Region (should match the endpoint) |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | `default` | Digital employee name |
| `VIBEOPS_MAX_RETRIES` | ❌ | `10` | Max SSE reconnect attempts |
| `VIBEOPS_IDLE_TIMEOUT` | ❌ | `60` | Idle timeout (seconds); reconnect if no message arrives within this window |

> [!IMPORTANT]
> \*When AK/SK are not set, the client uses the Alibaba Cloud default credential chain.

## Command-Line Flags

| Flag | Applies to | Description |
|------|-----------|-------------|
| `-simulate-error` | `chat`, `chat-from-file` | Simulate a network disconnection to exercise retries |
| `-file <path>` | `chat-from-file` | Load a single request from a JSON file |
| `-dir <path>` | `chat-from-file` | Batch-process every JSON request in a directory |

## Project Structure

```
samples/golang/
├── cmd/
│   ├── chat/              # Interactive chat
│   ├── chat-from-file/    # Run requests from JSON (file or directory)
│   └── thread-manager/    # Thread management
├── internal/
│   ├── client/            # Core client: chat, threads, retry, printers, errors
│   │   ├── client.go      # AgentClient, config, chat methods
│   │   ├── retry.go       # SSE reconnection, backoff, dedupe
│   │   ├── credentials.go # Default credential chain
│   │   ├── thread.go      # Thread management API
│   │   └── ...
│   └── logger/            # Structured logging
├── types/                 # Event & input type definitions
├── Makefile
├── go.mod
└── go.sum
```

## SDK Dependencies

- `github.com/alibabacloud-go/starops-20260428` — Alibaba Cloud STAROps SDK
- `github.com/alibabacloud-go/darabonba-openapi/v2` — OpenAPI client
- `github.com/alibabacloud-go/tea` — Tea runtime
- `github.com/aliyun/credentials-go` — Alibaba Cloud credential chain
