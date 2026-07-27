# 请求样例

各语言实现共享的 STAROps 请求样例文件。

## 请求结构

```json
{
  "action": "create",
  "messages": [
    {
      "role": "user",
      "contents": [{ "type": "text", "value": "your message" }]
    }
  ],
  "variables": {
    "userContext": "<UserContext[] 的 JSON 字符串>",
    "timeStamp": "<Unix 时间戳，秒>",
    "config": "<config 对象的 JSON 字符串>",
    "skill": "<skill name>"
  }
}
```

## 变量说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `userContext` | JSON 字符串 | `UserContext` 对象数组，序列化为字符串。详见下方 [UserContext](#usercontext)。 |
| `timeStamp` | string | 当前 Unix 时间戳，单位秒。 |
| `config` | JSON 字符串 | 运行时配置，例如 `{"disableThreadData": false}`。 |
| `skill` | string | （可选）指定要调用的技能，例如 `"sql_generation"`。 |

## UserContext

`userContext` 是一个 JSON 字符串，表示上下文对象数组。每个对象包含 `type` 和 `data` 两个字段：

```json
[
  { "type": "<上下文类型>", "data": { ... } }
]
```

### 支持的类型

#### `metadata`

为请求提供时间范围上下文。

```json
{ "type": "metadata", "data": { "fromTime": 1784600207, "toTime": 1784601107 } }
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `fromTime` | number | 起始时间（Unix 时间戳，秒） |
| `toTime` | number | 结束时间（Unix 时间戳，秒） |

#### `dynamic`

向 Agent 的系统提示词注入**动态系统提醒**。

```json
{ "type": "dynamic", "data": { "data": "以云资源运维专家身份回答，输出语言为中文。" } }
```

**行为：**

- `data.data` 中的内容会追加到 Agent 的 **System Reminder**，作为系统提示词的一部分。
- 该内容**对终端用户不可见** —— 不会在对话页面展示。

由于这是一段在运行时注入、用户无感知的系统提示，调用方可以按本次请求的上下文动态塑造 Agent 的表现，无需改动 Agent 配置或代码。典型用途：

- **业务集成**：按调用方身份或权限注入约束（例如"仅回答 ECS 相关问题"），或将实时业务数据作为背景提供给 Agent 参考。
- **产品与运营**：统一品牌话术、合规免责声明、固定开场白，或注入限时活动引导、灰度文案。
- **输出控制**：强制结构化输出格式（如 JSON、Markdown 表格），便于下游系统直接解析。
- **行为微调**：快速调整语气、回答长度、语言风格，适合 A/B 测试不同提示词效果。

**示例：** 注入运维场景约束 —— 限定 Agent 角色边界并要求结构化输出，便于下游系统解析：

```json
{
  "type": "dynamic",
  "data": {
    "data": "你是云资源运维助手，仅回答阿里云 ECS / RDS / SLB 相关的运维问题，超出范围时拒绝并提示用户。进行告警分析时，必须按以下 JSON 结构输出，不要附加额外说明：{\"summary\":\"告警概要\",\"severity\":\"P0|P1|P2\",\"suggestion\":\"处置建议\"}。"
  }
}
```

#### `entity`

提供实体上下文（例如某个具体的云资源）。

```json
{
  "type": "entity",
  "data": {
    "entityId": "i-bp1234",
    "entityDomain": "ecs",
    "entityType": "instance",
    "fromTime": 1784600207,
    "toTime": 1784601107,
    "title": "我的 ECS 实例"
  }
}
```

## 样例文件

| 文件 | 说明 |
| --- | --- |
| [`data_agent.json`](data_agent.json) | 基础数据查询，含时间范围 metadata。 |
| [`sls_chat.json`](sls_chat.json) | SLS 日志查询，使用 `sql_generation` 技能。 |
| [`request_with_reminder.json`](request_with_reminder.json) | 演示 `type: dynamic` 系统提醒注入。 |
| [`user_ack_interactive.json`](user_ack_interactive.json) | 触发 `user_ack` 交互事件（确认对话框）。 |
| [`user_input_interactive.json`](user_input_interactive.json) | 触发 `user_input` 交互事件（表单输入）。 |
