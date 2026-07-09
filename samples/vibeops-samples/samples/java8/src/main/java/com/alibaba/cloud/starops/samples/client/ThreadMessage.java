package com.alibaba.cloud.starops.samples.client;

/**
 * 会话消息
 * Thread message
 */
public class ThreadMessage {
    private String role;
    private String content;
    private String timestamp;

    public ThreadMessage() {}

    public ThreadMessage(String role, String content, String timestamp) {
        this.role = role;
        this.content = content;
        this.timestamp = timestamp;
    }

    public String getRole() {
        return role;
    }

    public void setRole(String role) {
        this.role = role;
    }

    public String getContent() {
        return content;
    }

    public void setContent(String content) {
        this.content = content;
    }

    public String getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(String timestamp) {
        this.timestamp = timestamp;
    }
}
