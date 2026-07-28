# STAROps SDK Java 示例 ☕

面向阿里云 STAROps 数字员工的 Java 客户端示例，内置弹性 SSE 流式能力：自动重连、指数退避与消息去重。

## 环境要求

- **Java 11+**
- **Maven 3.6+**
- 拥有 STAROps 访问权限的阿里云账号
- 已通过阿里云 CLI 配置权限，或已设置 AK/SK 环境变量

> [!NOTE]
> 只能使用 Java 8？请使用语法兼容分支 [`../java8/`](../java8/)，功能完全一致。

## 快速开始

```bash
cd java
cp .env.example .env   # 编辑 .env 填入 STAROps 配置，不填凭据
mvn compile
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
```

## 构建

```bash
mvn clean compile   # 编译
mvn test            # 运行测试
mvn package         # 打包 JAR
```

## 运行示例

### Chat — 交互式对话

```bash
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
```

交互式多轮对话，在同一会话中保持上下文。

### ChatFromFile — 从 JSON 文件运行请求

```bash
# 单个请求
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
    -Dexec.args="-file ../sample-requests/entity.json"

# 批量处理目录
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ChatFromFile" \
    -Dexec.args="-dir ../sample-requests/"
```


## 停止对话

对话运行时按 `Ctrl+C`（或发送 `SIGTERM`），客户端会发送 `stop` 请求（`action="stop"`）中断当前响应后退出。请求带短超时，确保进程不会卡住。

## SSE 重试与重连

客户端通过 SSE 流式接收响应，并在连接中断时透明恢复。

```
create ──► 流式接收事件 ──► stream_done ✅（正常结束）
              │
              ├─ 连接中断 ─────┐
              ├─ 空闲超时 ─────┤──► 退避 ──► 重连（action="reconnect"）──► 去重 ──► 续传
              └─ SSE 错误 ─────┘
```

- `stream_done` 事件表示 **正常结束**；在此之前结束的流会触发重连。
- **指数退避**：`1s, 2s, 4s, 8s, 16s, 30s`（上限 30s），最多重试 `STAROPS_MAX_RETRIES` 次。
- **重连** 发送 `action="reconnect"` 恢复会话。
- **去重**：重连后按 timestamp 过滤消息，确保不重复投递。

> [!NOTE]
> 超过最大重试次数后，客户端会抛出错误，而不会一直挂起。

### 测试重试逻辑

```bash
mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat" \
    -Dexec.args="-simulate-error"
```

启用 `-simulate-error` 后，客户端会模拟网络断连、执行退避、重连、按 timestamp 去重，
并最终在 `stream_done` 处完成。

## 环境变量

| 变量 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `STAROPS_ENDPOINT` | ✅ | — | STAROps API 端点，如 `starops.cn-beijing.aliyuncs.com` |
| `STAROPS_WORKSPACE` | ❌ | — | 工作空间 ID，可选，不填写时程序也可运行 |
| `STAROPS_REGION` | ❌ | `cn-beijing` | 数据查询所在的阿里云地域，与服务接入点（endpoint）无关 |
| `STAROPS_EMPLOYEE_NAME` | ❌ | `apsara-ops` | 数字员工名称 |
| `STAROPS_MAX_RETRIES` | ❌ | `10` | SSE 最大重连次数 |
| `STAROPS_IDLE_TIMEOUT` | ❌ | `60` | 空闲超时秒数，超时未收到消息则重连 |

凭据配置：

1. **推荐方式**：使用阿里云 CLI 配置权限。
2. 如果本地没有 CLI，请访问 [阿里云 CLI 说明](https://help.aliyun.com/zh/ros/api-operation-examples-overview) 下载安装。
3. 如果不想安装 CLI，可以使用环境变量配置 AK/SK：

```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID=<your-access-key-id>
export ALIBABA_CLOUD_ACCESS_KEY_SECRET=<your-access-key-secret>
```

## 命令行参数

| 参数 | 适用命令 | 说明 |
|------|----------|------|
| `-simulate-error` | `Chat`、`ChatFromFile` | 模拟网络断连，用于验证重试 |
| `-file <path>` | `ChatFromFile` | 从单个 JSON 文件加载请求 |
| `-dir <path>` | `ChatFromFile` | 批量处理目录下的所有 JSON 请求 |

> [!TIP]
> 通过 Maven 传参需使用 `-Dexec.args="..."`，例如 `-Dexec.args="-file ../sample-requests/entity.json"`。

## 项目结构

```
java/
├── src/
│   ├── main/java/com/alibaba/cloud/starops/samples/
│   │   ├── client/    # 核心客户端：对话、会话、重试、打印器
│   │   ├── types/     # 类型定义
│   │   ├── logger/    # 日志工具
│   │   └── examples/  # Chat、ChatFromFile
│   └── test/          # 测试用例
├── pom.xml
└── README.md
```

## 依赖

- 阿里云 STAROps SDK（`starops20260428`）
- Maven Exec Plugin（用于 `mvn exec:java`）
