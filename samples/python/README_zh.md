# STAROps SDK Python 示例 🐍

面向阿里云 STAROps 数字员工的 Python 客户端示例，内置弹性 SSE 流式能力：自动重连、指数退避与消息去重。

## 环境要求

- **Python 3.8+**
- 拥有 STAROps 访问权限的阿里云账号
- 已通过阿里云 CLI 配置权限，或已设置 AK/SK 环境变量

## 安装

```bash
cd python
python -m venv .venv && source .venv/bin/activate
pip install -e .
cp .env.example .env   # 编辑 .env 填入 STAROps 配置，不填凭据
```

## 快速开始

```bash
python -m starops_sdk_samples.examples.chat
```

## 运行示例

### chat — 交互式对话

```bash
python -m starops_sdk_samples.examples.chat
```

交互式多轮对话，在同一会话中保持上下文。

### chat_from_file — 从 JSON 文件运行请求

```bash
# 单个请求（默认显示详细事件信息）
python -m starops_sdk_samples.examples.chat_from_file -file ../sample-requests/entity.json

# 批量处理目录
python -m starops_sdk_samples.examples.chat_from_file -dir ../sample-requests/

# 简洁模式：仅输出文本
python -m starops_sdk_samples.examples.chat_from_file -file ../sample-requests/entity.json -simple
```

默认使用 `EventPrinter` 显示角色、内容、工具调用、Agent 调用与耗时等详细信息。
使用 `-simple` 可切换为仅输出文本。


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
python -m starops_sdk_samples.examples.chat -simulate-error
python -m starops_sdk_samples.examples.chat_from_file -file ../sample-requests/entity.json -simulate-error
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
| `-simulate-error` | `chat`、`chat_from_file` | 模拟网络断连，用于验证重试 |
| `-file <path>` | `chat_from_file` | 从单个 JSON 文件加载请求 |
| `-dir <path>` | `chat_from_file` | 批量处理目录下的所有 JSON 请求 |
| `-simple` | `chat_from_file` | 仅输出文本 |

## 测试

```bash
pip install -e ".[dev]"
pytest
```

## 项目结构

```
python/
├── starops_sdk_samples/
│   ├── client/        # 核心客户端：对话、会话、重试、打印器
│   ├── types/         # 类型定义
│   ├── logger/        # 日志工具
│   └── examples/      # chat、chat_from_file
├── tests/             # 测试用例
├── pyproject.toml
└── README.md
```
