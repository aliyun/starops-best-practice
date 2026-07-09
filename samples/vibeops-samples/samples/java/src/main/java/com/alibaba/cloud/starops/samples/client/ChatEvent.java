package com.alibaba.cloud.starops.samples.client;

import com.fasterxml.jackson.databind.JsonNode;

/**
 * 聊天事件
 * Chat event
 */
public class ChatEvent {
    private JsonNode body;
    private String rawJson;
    private int statusCode;
    private boolean done;
    private Exception error;
    private String id;
    private String event;

    public ChatEvent() {}

    public static ChatEvent done() {
        ChatEvent event = new ChatEvent();
        event.done = true;
        return event;
    }

    public static ChatEvent error(Exception error) {
        ChatEvent event = new ChatEvent();
        event.error = error;
        return event;
    }

    public static ChatEvent fromResponse(JsonNode body, String rawJson, int statusCode) {
        ChatEvent event = new ChatEvent();
        event.body = body;
        event.rawJson = rawJson;
        event.statusCode = statusCode;
        event.id = body != null ? body.path("id").asText(null) : null;
        event.event = body != null ? body.path("event").asText(null) : null;
        event.done = isDoneMessage(body);
        return event;
    }

    private static boolean isDoneMessage(JsonNode body) {
        if (body == null) {
            return false;
        }
        // 优先使用 response 级别的 event 字段
        JsonNode eventNode = body.get("event");
        if (eventNode != null && "done".equals(eventNode.asText())) {
            return true;
        }
        // fallback: 遍历 messages
        if (!body.has("messages")) {
            return false;
        }
        JsonNode messages = body.get("messages");
        if (messages.isArray()) {
            for (JsonNode msg : messages) {
                if (msg.has("type") && "done".equals(msg.get("type").asText())) {
                    return true;
                }
            }
        }
        return false;
    }

    public JsonNode getBody() {
        return body;
    }

    public void setBody(JsonNode body) {
        this.body = body;
    }

    public String getRawJson() {
        return rawJson;
    }

    public void setRawJson(String rawJson) {
        this.rawJson = rawJson;
    }

    public int getStatusCode() {
        return statusCode;
    }

    public void setStatusCode(int statusCode) {
        this.statusCode = statusCode;
    }

    public boolean isDone() {
        return done;
    }

    public void setDone(boolean done) {
        this.done = done;
    }

    public Exception getError() {
        return error;
    }

    public void setError(Exception error) {
        this.error = error;
    }

    public boolean hasError() {
        return error != null;
    }

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getEvent() { return event; }
    public void setEvent(String event) { this.event = event; }
}
