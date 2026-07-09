# STAROps SDK Samples for TypeScript 🟦

TypeScript client samples for Alibaba Cloud STAROps digital employees, featuring resilient SSE streaming
with automatic reconnection, exponential backoff, and message deduplication.

## Requirements

- **Node.js 18+**
- **TypeScript 5.0+**
- An Alibaba Cloud account with STAROps access and valid credentials

## Installation

```bash
cd samples/typescript
npm install
cp .env.example .env   # edit .env with your credentials and endpoint
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
npm run chat-from-file -- -file ../../requests/starops/entity.json

# Batch-process a directory
npm run chat-from-file -- -dir ../../requests/starops/

# Simple mode: text-only output
npm run chat-from-file -- -file ../../requests/starops/entity.json -simple
```

By default the detailed `EventPrinter` shows role, content, tool calls, agent calls, and durations.
Use `-simple` to switch to text-only output.

> [!NOTE]
> With `npm run`, pass a `--` separator before the flags so npm forwards them to the script.

### thread-manager — manage threads

```bash
npm run thread-manager -- list                # list threads
npm run thread-manager -- get <thread-id>     # thread details
npm run thread-manager -- delete <thread-id>  # delete a thread
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
- **Exponential backoff**: `1s, 2s, 4s, 8s, 16s, 30s` (capped at 30s), up to `VIBEOPS_MAX_RETRIES` attempts.
- **Reconnect** sends `action="reconnect"` to resume the session.
- **Deduplication**: after reconnecting, messages are filtered by timestamp so none are delivered twice.

> [!NOTE]
> After exceeding the maximum retries, the client throws an error instead of hanging.

### Testing the retry logic

```bash
npx tsx src/examples/chat.ts -simulate-error
npm run chat-from-file -- -file ../../requests/starops/entity.json -simulate-error
```

With `-simulate-error`, the client simulates a network disconnection, backs off, reconnects,
deduplicates by timestamp, and finishes at `stream_done`.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VIBEOPS_ENDPOINT` | ✅ | — | STAROps API endpoint, e.g. `starops.cn-beijing.aliyuncs.com` |
| `VIBEOPS_WORKSPACE` | ✅ | — | Workspace ID |
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | ✅ | — | Access Key ID |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | ✅ | — | Access Key Secret |
| `VIBEOPS_REGION` | ❌ | `cn-hangzhou` | Region (should match the endpoint) |
| `VIBEOPS_EMPLOYEE_NAME` | ❌ | `default` | Digital employee name |
| `VIBEOPS_MAX_RETRIES` | ❌ | `10` | Max SSE reconnect attempts |
| `VIBEOPS_IDLE_TIMEOUT` | ❌ | `60` | Idle timeout (seconds); reconnect if no message arrives within this window |

## Command-Line Flags

| Flag | Applies to | Description |
|------|-----------|-------------|
| `-simulate-error` | `chat`, `chat-from-file` | Simulate a network disconnection to exercise retries |
| `-file <path>` | `chat-from-file` | Load a single request from a JSON file |
| `-dir <path>` | `chat-from-file` | Batch-process every JSON request in a directory |
| `-simple` | `chat-from-file` | Text-only output |

## Project Structure

```
samples/typescript/
├── src/
│   ├── client/        # Core client: chat, threads, retry, printers
│   ├── types/         # Type definitions
│   └── examples/      # chat, chat-from-file, thread-manager
├── tests/             # Test suite (vitest)
├── package.json
├── tsconfig.json
└── README.md
```
