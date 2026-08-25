# STAROps SDK Samples for Java ☕

Java client samples for Alibaba Cloud STAROps digital employees, featuring resilient SSE streaming with
automatic reconnection, exponential backoff, and message deduplication.

## Requirements

- **Java 11+**
- **Maven 3.6+**
- An Alibaba Cloud account with STAROps access
- Credentials configured with Alibaba Cloud CLI, or AK/SK environment variables

> [!NOTE]
> Restricted to Java 8? Use the syntax-compatible fork in [`../java8/`](../java8/) — functionality is identical.

## Quick Start

```bash
cd java
cp .env.example .env   # edit .env with STAROps settings, not credentials
mvn compile
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
```

## Build

```bash
mvn clean compile   # compile
mvn test            # run tests
mvn package         # build the JAR
```

## Running the Samples

### Chat — interactive chat

```bash
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
```

Multi-turn interactive chat with context preserved within a thread.

### ChatFromFile — run requests from JSON

```bash
# Single request
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
    -Dexec.args="-file ../sample-requests/entity.json"

# Batch-process a directory
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
    -Dexec.args="-dir ../sample-requests/"
```

### Stop on interrupt

While a chat is running, press `Ctrl+C` (or send `SIGTERM`). The client sends a
`stop` request (`action="stop"`) to interrupt the ongoing response, then exits.
The request is best-effort with a short timeout so the process never hangs.

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
> After exceeding the maximum retries, the client surfaces an error instead of hanging.

### Testing the retry logic

```bash
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat" \
    -Dexec.args="-simulate-error"
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
| `--model provider:modelId` | `Chat` | Specify a model when creating a session (e.g. `--model qwen:qwen3.8-max`) |
| `--list-models` | `Chat` | List all available model values and exit (no credentials required) |
| `-simulate-error` | `Chat`, `ChatFromFile` | Simulate a network disconnection to exercise retries |
| `-file <path>` | `ChatFromFile` | Load a single request from a JSON file |
| `-dir <path>` | `ChatFromFile` | Batch-process every JSON request in a directory |
| `-mock` | `ChatFromFile` | Replay a recorded SSE session (offline, no credentials) |
| `-mock-file <path>` | `ChatFromFile` | SSE recording file for mock/record mode |
| `-simple` | `ChatFromFile` | Print only the final text result |
| `-record` | `ChatFromFile` | Record real SSE events to a JSONL file |

> [!TIP]
> Pass flags through Maven with `-Dexec.args="..."`, e.g. `-Dexec.args="-file ../sample-requests/entity.json"`.

## Project Structure

```
java/
├── src/
│   ├── main/java/com/alibaba/cloud/starops/samples/
│   │   ├── client/    # Core client: chat, threads, retry, printers
│   │   ├── types/     # Type definitions
│   │   ├── logger/    # Logging utilities
│   │   └── examples/  # Chat, ChatFromFile
│   └── test/          # Test suite
├── pom.xml
└── README.md
```

## Dependencies

- Alibaba Cloud STAROps SDK (`starops20260428`)
- Maven Exec Plugin (for `mvn exec:java`)
