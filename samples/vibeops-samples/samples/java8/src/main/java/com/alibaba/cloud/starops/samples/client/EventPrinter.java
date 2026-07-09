package com.alibaba.cloud.starops.samples.client;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;

/**
 * 事件打印器（详细模式，Java 8 兼容版本）
 * Event printer - displays detailed SSE event information (Java 8 compatible)
 */
public class EventPrinter {
    private final boolean printRawBody;
    private final boolean printParsed;
    private final ObjectMapper objectMapper;

    public EventPrinter(boolean printRawBody, boolean printParsed) {
        this.printRawBody = printRawBody;
        this.printParsed = printParsed;
        this.objectMapper = new ObjectMapper().enable(SerializationFeature.INDENT_OUTPUT);
    }

    /**
     * 打印事件
     * Print event
     */
    public void printEvent(ChatEvent event, int eventIndex) {
        if (event == null) {
            return;
        }

        if (event.hasError()) {
            System.out.printf("%n❌ 错误: %s%n", event.getError().getMessage());
            return;
        }

        if (event.isDone() && event.getBody() == null) {
            System.out.println("\n✅ 对话完成");
            return;
        }

        if (event.getBody() == null) {
            return;
        }

        String separator = repeatStr("=", 30);
        System.out.printf("%n%s 事件 #%d %s%n", separator, eventIndex, separator);

        if (printRawBody && event.getRawJson() != null && !event.getRawJson().isEmpty()) {
            System.out.println("\n📦 原始 Body:");
            try {
                Object json = objectMapper.readValue(event.getRawJson(), Object.class);
                System.out.println(objectMapper.writeValueAsString(json));
            } catch (Exception e) {
                System.out.println(event.getRawJson());
            }
        }

        if (printParsed) {
            printParsedEvent(event);
        }
    }

    private void printParsedEvent(ChatEvent event) {
        JsonNode body = event.getBody();
        if (body == null || !body.has("messages")) {
            return;
        }

        System.out.println("\n📋 解析详情:");

        JsonNode messages = body.get("messages");
        if (!messages.isArray()) {
            return;
        }

        for (JsonNode msg : messages) {
            System.out.printf("  原始消息: %s%n", msg.toString());
            printMessageItem(msg);
        }
    }

    private void printMessageItem(JsonNode item) {
        if (item.has("role") && !item.get("role").isNull()) {
            System.out.printf("  📌 角色: %s%n", item.get("role").asText());
        }
        if (item.has("callId") && !item.get("callId").isNull()) {
            System.out.printf("  🔗 CallID: %s%n", item.get("callId").asText());
        }
        if (item.has("parentCallId") && !item.get("parentCallId").isNull()) {
            System.out.printf("  🔗 ParentCallID: %s%n", item.get("parentCallId").asText());
        }

        // 内容 / Contents
        if (item.has("contents") && item.get("contents").isArray()) {
            System.out.println("  📝 内容:");
            int i = 0;
            for (JsonNode content : item.get("contents")) {
                String type = content.has("type") ? content.get("type").asText() : "";
                System.out.printf("    [%d] 类型: %s%n", i, type);
                if (content.has("value") && !content.get("value").isNull()) {
                    String value = content.get("value").asText();
                    if (value.length() > 200) {
                        value = value.substring(0, 200) + "...";
                    }
                    System.out.printf("        值: %s%n", value);
                }
                if (content.has("append") && content.get("append").asBoolean()) {
                    System.out.println("        追加: true");
                }
                if (content.has("lastChunk") && content.get("lastChunk").asBoolean()) {
                    System.out.println("        最后块: true");
                }
                i++;
            }
        }

        // 工具调用 / Tools
        if (item.has("tools") && item.get("tools").isArray() && item.get("tools").size() > 0) {
            System.out.println("  🔧 工具调用:");
            int i = 0;
            for (JsonNode tool : item.get("tools")) {
                String name = tool.has("name") ? tool.get("name").asText() : "";
                String status = tool.has("status") ? tool.get("status").asText() : "";
                System.out.printf("    [%d] 名称: %s, 状态: %s%n", i, name, status);
                if (tool.has("toolCallId") && !tool.get("toolCallId").isNull()) {
                    System.out.printf("        ToolCallID: %s%n", tool.get("toolCallId").asText());
                }
                if (tool.has("arguments") && !tool.get("arguments").isNull()) {
                    String argsStr = tool.get("arguments").toString();
                    if (argsStr.length() > 200) {
                        argsStr = argsStr.substring(0, 200) + "...";
                    }
                    System.out.printf("        参数: %s%n", argsStr);
                }
                i++;
            }
        }

        // Agent 调用 / Agents
        if (item.has("agents") && item.get("agents").isArray() && item.get("agents").size() > 0) {
            System.out.println("  🤖 Agent调用:");
            int i = 0;
            for (JsonNode agent : item.get("agents")) {
                String name = agent.has("name") ? agent.get("name").asText() : "";
                String status = agent.has("status") ? agent.get("status").asText() : "";
                System.out.printf("    [%d] 名称: %s, 状态: %s%n", i, name, status);
                i++;
            }
        }

        // 事件 / Events
        if (item.has("events") && item.get("events").isArray() && item.get("events").size() > 0) {
            System.out.println("  📢 事件:");
            int i = 0;
            for (JsonNode evt : item.get("events")) {
                String type = evt.has("type") ? evt.get("type").asText() : "";
                System.out.printf("    [%d] 类型: %s%n", i, type);
                if (evt.has("payload") && !evt.get("payload").isNull()) {
                    printEventPayload(type, evt.get("payload"));
                }
                i++;
            }
        }
    }

    private void printEventPayload(String type, JsonNode payload) {
        if ("thinking".equals(type)) {
            if (payload.has("reasoningDelta")) {
                String delta = payload.get("reasoningDelta").asText();
                if (delta.length() > 100) {
                    delta = delta.substring(0, 100) + "...";
                }
                System.out.printf("        思考: %s%n", delta);
            }
        } else if ("error".equals(type)) {
            if (payload.has("code")) {
                System.out.printf("        错误码: %s%n", payload.get("code").asText());
            }
            if (payload.has("message")) {
                System.out.printf("        消息: %s%n", payload.get("message").asText());
            }
        } else if ("task_finished".equals(type)) {
            if (payload.has("success")) {
                System.out.printf("        成功: %s%n", payload.get("success").asBoolean());
            }
            if (payload.has("statistics") && payload.get("statistics").has("duration")) {
                long durationNs = payload.get("statistics").get("duration").asLong();
                System.out.printf("        耗时: %dms%n", durationNs / 1000000);
            }
        } else {
            String payloadStr = payload.toString();
            if (payloadStr.length() > 200) {
                payloadStr = payloadStr.substring(0, 200) + "...";
            }
            System.out.printf("        负载: %s%n", payloadStr);
        }
    }

    /**
     * Java 8 兼容：String.repeat() 是 JDK 11+ 的 API，这里用辅助方法替代
     * Java 8 compat: String.repeat() is JDK 11+, use helper method instead
     */
    private static String repeatStr(String s, int count) {
        StringBuilder sb = new StringBuilder(s.length() * count);
        for (int i = 0; i < count; i++) {
            sb.append(s);
        }
        return sb.toString();
    }
}
