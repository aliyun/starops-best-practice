# STAROps SDK Samples for Java ☕

Java client samples for Alibaba Cloud STAROps digital employees, featuring resilient SSE streaming with
automatic reconnection, exponential backoff, and message deduplication.

## Requirements

- **Java 11+**
- **Maven 3.6+**
- An Alibaba Cloud account with STAROps access and valid credentials

> [!NOTE]
> Restricted to Java 8? Use the syntax-compatible fork in [`../java8/`](../java8/) — functionality is identical.

## Quick Start

```bash
cd samples/java
cp .env.example .env   # edit .env with your credentials and endpoint
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
    -Dexec.args="-file ../../requests/starops/entity.json"

# Batch-process a directory
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
    -Dexec.args="-dir ../../requests/starops/"
```

### ThreadManager — manage threads

```bash
# List threads
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ThreadManager" \
    -Dexec.args="list"

# Thread details
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ThreadManager" \
    -Dexec.args="get <thread-id>"

# Delete a thread
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ThreadManager" \
    -Dexec.args="delete <thread-id>"
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
| `-simulate-error` | `Chat`, `ChatFromFile` | Simulate a network disconnection to exercise retries |
| `-file <path>` | `ChatFromFile` | Load a single request from a JSON file |
| `-dir <path>` | `ChatFromFile` | Batch-process every JSON request in a directory |

> [!TIP]
> Pass flags through Maven with `-Dexec.args="..."`, e.g. `-Dexec.args="-file ../../requests/starops/entity.json"`.

## Project Structure

```
samples/java/
├── src/
│   ├── main/java/com/alibaba/cloud/starops/samples/
│   │   ├── client/    # Core client: chat, threads, retry, printers
│   │   ├── types/     # Type definitions
│   │   ├── logger/    # Logging utilities
│   │   └── examples/  # Chat, ChatFromFile, ThreadManager
│   └── test/          # Test suite
├── pom.xml
└── README.md
```

## Dependencies

- Alibaba Cloud STAROps SDK (`starops20260428`)
- Maven Exec Plugin (for `mvn exec:java`)
