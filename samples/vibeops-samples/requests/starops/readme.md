# STAROps 请求示例 / STAROps Request Examples

本目录包含 STAROps SDK 的请求示例文件。

## 文件列表 / File List

| 文件名 | 场景 | 描述 |
|--------|------|------|
| general_chat.json | 普通对话 | 通用查询对话 |
| entity.json | 实体查询 | APM 服务实体延迟查询（带上下文） |
| sql_generation.json | SQL 生成 | 自然语言转 SQL 查询（skill: sql_generation） |
| data_agent.json | 数据代理 | 自然语言数据查询（skill: data-agent-pro） |

---

## 文件说明 / File Details

### general_chat.json - 普通对话

通用查询对话，无特定 skill 配置。

```
统计错误数量
```

### entity.json - 实体查询

查询特定 APM 服务实体的性能指标，包含实体上下文信息。

```
cart 实体延迟多少
```

关键配置：
- `userContext.type`: `entity`
- `entity_domain`: `apm`
- `entity_type`: `apm.service`

### sql_generation.json - SQL 生成

使用 sql_generation skill 生成 SQL 查询语句。

```
帮我查看里面有多少种类的 admin_emails
```

关键配置：
- `skill`: `sql_generation`
- `skill_name`: `sql_generation`

### data_agent.json - 数据代理

使用 data-agent-pro skill 进行自然语言数据查询。

```
查询最近1小时的告警数量
```

关键配置：
- `skill`: `data-agent-pro`
- `skill_name`: `data-agent-pro`

---

## 使用方法 / Usage

```bash
cd samples/golang

# 加载环境变量
export $(cat .env | xargs)

# 执行单个请求
go run ./cmd/chat-from-file/ -file ../../requests/starops/general_chat.json -simple

# 批量执行目录下所有请求
go run ./cmd/chat-from-file/ -dir ../../requests/starops -simple
```

---

## 通用参数 / Common Parameters

| 字段 | 说明 |
|------|------|
| `region` | 地域，如 `cn-hongkong` |
| `digitalEmployeeName` | 数字员工名称，默认 `apsara-ops` |
| `threadId` | 会话 ID，空字符串创建新会话 |
| `action` | 操作类型，通常为 `create` |
| `messages` | 用户消息 |
| `variables` | 请求变量 |

### variables 字段

| 字段 | 说明 |
|------|------|
| `workspace` | 工作空间 |
| `project` | 项目名称 |
| `language` | 语言，`zh` 或 `en` |
| `timeZone` | 时区 |
| `userContext` | 用户上下文 JSON |
| `config` | 配置 JSON |
| `skill` | 技能标识（可选） |
| `skill_name` | 技能名称（可选） |
