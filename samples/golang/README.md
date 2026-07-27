# STAROps SDK Samples for Go 🐹

Go client samples for Alibaba Cloud STAROps digital employees, featuring resilient SSE streaming with
automatic reconnection, exponential backoff, and message deduplication.

## Requirements

- **Go 1.22+**
- An Alibaba Cloud account with STAROps access
- Credentials configured with Alibaba Cloud CLI, or AK/SK environment variables

## Quick Start

```bash
cd golang
cp .env.example .env   # edit .env with STAROps settings, not credentials
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
go run ./cmd/chat-from-file/ -file ../sample-requests/entity.json

# Batch-process a directory
go run ./cmd/chat-from-file/ -dir ../sample-requests/
```

## Stop on interrupt

While a chat is running, press `Ctrl+C` (or send `SIGTERM`). The client sends a `stop`
request (`action="stop"`) to interrupt the ongoing response, then exits. The request is
best-effort with a short timeout so the process never hangs.

## Mock playback

Replay a recorded SSE session offline — no credentials or network required:

```bash
make go-mock
# or: go run ./cmd/chat-from-file --file ../sample-requests/data_agent.json --mock --mock-file <mock-file> --simple
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
- **Exponential backoff**: `1s, 2s, 4s, 8s, 16s, 30s` (capped at 30s), up to `STAROPS_MAX_RETRIES` attempts.
- **Reconnect** sends `action="reconnect"` and copies the original `threadId` / variables.
- **Deduplication**: after reconnecting, messages are filtered by timestamp so none are delivered twice.

> [!NOTE]
> After exceeding the maximum retries, the client returns an error event instead of hanging.

The retry logic lives in [`pkg/client/retry.go`](pkg/client/retry.go).

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

The Go sample supports Alibaba Cloud credential configuration in this order:

1. **Recommended**: configure credentials with Alibaba Cloud CLI.
2. If you do not have the CLI, install it from the [Alibaba Cloud CLI guide](https://help.aliyun.com/zh/ros/api-operation-examples-overview).
3. If you do not want to install the CLI, configure AK/SK with environment variables:

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID=<your-access-key-id>
export ALIBABA_CLOUD_ACCESS_KEY_SECRET=<your-access-key-secret>
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STAROPS_ENDPOINT` | ✅ | — | STAROps API endpoint, e.g. `starops.cn-beijing.aliyuncs.com` |
| `STAROPS_WORKSPACE` | ❌ | — | Workspace ID. Optional. The sample runs without it. |
| `STAROPS_REGION` | ❌ | `cn-beijing` | Alibaba Cloud region where data queries are performed; independent of the service endpoint |
| `STAROPS_EMPLOYEE_NAME` | ❌ | `apsara-ops` | Digital employee name |
| `STAROPS_MAX_RETRIES` | ❌ | `10` | Max SSE reconnect attempts |
| `STAROPS_IDLE_TIMEOUT` | ❌ | `60` | Idle timeout (seconds); reconnect if no message arrives within this window |

> [!IMPORTANT]
> Configure credentials with Alibaba Cloud CLI, or use AK/SK environment variables if you do not want to install the CLI.

## Command-Line Flags

| Flag | Applies to | Description |
|------|-----------|-------------|
| `-simulate-error` | `chat`, `chat-from-file` | Simulate a network disconnection to exercise retries |
| `-file <path>` | `chat-from-file` | Load a single request from a JSON file |
| `-dir <path>` | `chat-from-file` | Batch-process every JSON request in a directory |

## Project Structure

```
golang/
├── cmd/
│   ├── chat/              # Interactive chat
│   └── chat-from-file/    # Run requests from JSON (file or directory)
├── pkg/
│   ├── client/            # Core client: chat, threads, retry, printers, errors
│   │   ├── client.go      # AgentClient, config, chat methods
│   │   ├── retry.go       # SSE reconnection, backoff, dedupe
│   │   ├── credentials.go # Credential loading
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
