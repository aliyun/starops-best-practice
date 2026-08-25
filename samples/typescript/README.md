# STAROps SDK Samples for TypeScript 🟦

TypeScript client samples for Alibaba Cloud STAROps digital employees, featuring resilient SSE streaming
with automatic reconnection, exponential backoff, and message deduplication.

## Requirements

- **Node.js 18+**
- **TypeScript 5.0+**
- An Alibaba Cloud account with STAROps access
- Credentials configured with Alibaba Cloud CLI, or AK/SK environment variables

## Installation

```bash
cd typescript
npm install
cp .env.example .env   # edit .env with STAROps settings, not credentials
```

## Quick Start

```bash
npm run chat
```

## Build

```bash
npm run build   # compile to ./dist
npm test        # run the test suite (vitest)
```

## Running the Samples

### chat — interactive chat

```bash
npm run chat
```

Multi-turn interactive chat with context preserved within a thread.

### chat-from-file — run requests from JSON

```bash
# Single request (detailed event output by default)
npm run chat-from-file -- -file ../sample-requests/entity.json

# Batch-process a directory
npm run chat-from-file -- -dir ../sample-requests/

# Simple mode: text-only output
npm run chat-from-file -- -file ../sample-requests/entity.json -simple
```

By default the detailed `EventPrinter` shows role, content, tool calls, agent calls, and durations.
Use `-simple` to switch to text-only output.

> [!NOTE]
> With `npm run`, pass a `--` separator before the flags so npm forwards them to the script.

## Stop on interrupt

While a chat is running, press `Ctrl+C` (or send `SIGTERM`). The client sends a `stop`
request (`action="stop"`) to interrupt the ongoing response, then exits. The request is
best-effort with a short timeout so the process never hangs.

## Mock playback

Replay a recorded SSE session offline — no credentials or network required:

```bash
make ts-mock
# or: npx tsx src/examples/chat-from-file.ts -file ../sample-requests/data_agent.json -mock-file <mock-file> -mock -simple
```

## SSE Retry & Reconnection

The client streams responses over SSE and recovers transparently from interruptions.

```
create ──► stream events ──► stream_done ✅ (normal completion)
              │
              ├─ connection dropped ─┐
              ├─ idle timeout ───────┤──► backoff ──► reconnect (action="reconnect") ──► dedupe ──► resume
              └─ SSE error ──────────┘
```

- **Normal completion** is marked by a `stream_done` event; a stream ending before it triggers a reconnect.
- **Exponential backoff**: `1s, 2s, 4s, 8s, 16s, 30s` (capped at 30s), up to `STAROPS_MAX_RETRIES` attempts.
- **Reconnect** sends `action="reconnect"` to resume the session.
- **Deduplication**: after reconnecting, messages are filtered by timestamp so none are delivered twice.

> [!NOTE]
> After exceeding the maximum retries, the client throws an error instead of hanging.

### Testing the retry logic

```bash
npx tsx src/examples/chat.ts -simulate-error
npm run chat-from-file -- -file ../sample-requests/entity.json -simulate-error
```

With `-simulate-error`, the client simulates a network disconnection, backs off, reconnects,
deduplicates by timestamp, and finishes at `stream_done`.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STAROPS_ENDPOINT` | ✅ | — | STAROps API endpoint, e.g. `starops.cn-beijing.aliyuncs.com` |
| `STAROPS_WORKSPACE` | ❌ | — | Workspace ID. Optional. The sample runs without it. |
| `STAROPS_REGION` | ❌ | `cn-beijing` | Alibaba Cloud region where data queries are performed; independent of the service endpoint |
| `STAROPS_EMPLOYEE_NAME` | ❌ | `apsara-ops` | Digital employee name |
| `STAROPS_MAX_RETRIES` | ❌ | `10` | Max SSE reconnect attempts |
| `STAROPS_IDLE_TIMEOUT` | ❌ | `60` | Idle timeout (seconds); reconnect if no message arrives within this window |

Credential configuration:

1. **Recommended**: configure credentials with Alibaba Cloud CLI.
2. If you do not have the CLI, install it from the [Alibaba Cloud CLI guide](https://help.aliyun.com/zh/ros/api-operation-examples-overview).
3. If you do not want to install the CLI, configure AK/SK with environment variables:

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID=<your-access-key-id>
export ALIBABA_CLOUD_ACCESS_KEY_SECRET=<your-access-key-secret>
```

## Command-Line Flags

| Flag | Applies to | Description |
|------|-----------|-------------|
| `--model provider:modelId` | `chat` | Specify a model when creating a session (e.g. `--model qwen:qwen3.8-max`) |
| `--list-models` | `chat` | List all available model values and exit (no credentials required) |
| `-simulate-error` | `chat`, `chat-from-file` | Simulate a network disconnection to exercise retries |
| `-file <path>` | `chat-from-file` | Load a single request from a JSON file |
| `-dir <path>` | `chat-from-file` | Batch-process every JSON request in a directory |
| `-simple` | `chat-from-file` | Text-only output |

## Project Structure

```
typescript/
├── src/
│   ├── client/        # Core client: chat, threads, retry, printers
│   ├── types/         # Type definitions
│   └── examples/      # chat, chat-from-file
├── tests/             # Test suite (vitest)
├── package.json
├── tsconfig.json
└── README.md
```
