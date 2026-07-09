package com.alibaba.cloud.starops.samples.types;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;
import java.util.Map;

/**
 * 消息条目
 * Message item representing a single interaction record
 */
public class MessageItem {
    /** 父调用ID / Parent call ID */
    @JsonProperty("parentCallId")
    private String parentCallId;

    /** 当前调用的唯一标识符 / Current call unique identifier */
    @JsonProperty("callId")
    private String callId;

    /** 消息角色 / Message role */
    @JsonProperty("role")
    private MessageRole role;

    /** 消息生成的时间戳 / Message timestamp */
    @JsonProperty("timestamp")
    private String timestamp;

    /** 文本或富媒体内容列表 / Text or rich media content list */
    @JsonProperty("contents")
    private List<ItemContent> contents;

    /** 工具调用列表 / Tool call list */
    @JsonProperty("tools")
    private List<ItemTool> tools;

    /** 事件列表 / Event list */
    @JsonProperty("events")
    private List<ItemEvent> events;

    /** 产物列表 / Artifacts list */
    @JsonProperty("artifacts")
    private List<Map<String, Object>> artifacts;

    public MessageItem() {}

    public String getParentCallId() {
        return parentCallId;
    }

    public void setParentCallId(String parentCallId) {
        this.parentCallId = parentCallId;
    }

    public String getCallId() {
        return callId;
    }

    public void setCallId(String callId) {
        this.callId = callId;
    }

    public MessageRole getRole() {
        return role;
    }

    public void setRole(MessageRole role) {
        this.role = role;
    }

    public String getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(String timestamp) {
        this.timestamp = timestamp;
    }

    public List<ItemContent> getContents() {
        return contents;
    }

    public void setContents(List<ItemContent> contents) {
        this.contents = contents;
    }

    public List<ItemTool> getTools() {
        return tools;
    }

    public void setTools(List<ItemTool> tools) {
        this.tools = tools;
    }

    public List<ItemEvent> getEvents() {
        return events;
    }

    public void setEvents(List<ItemEvent> events) {
        this.events = events;
    }

    public List<Map<String, Object>> getArtifacts() {
        return artifacts;
    }

    public void setArtifacts(List<Map<String, Object>> artifacts) {
        this.artifacts = artifacts;
    }
}
