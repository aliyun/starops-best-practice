package com.alibaba.cloud.starops.samples.types;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Map;

/**
 * 事件定义
 * Event definition
 */
public class ItemEvent {
    /** 事件类型 / Event type */
    @JsonProperty("type")
    private EventType type;

    /** 事件内容 / Event payload */
    @JsonProperty("payload")
    private Map<String, Object> payload;

    public ItemEvent() {}

    public ItemEvent(EventType type, Map<String, Object> payload) {
        this.type = type;
        this.payload = payload;
    }

    public EventType getType() {
        return type;
    }

    public void setType(EventType type) {
        this.type = type;
    }

    public Map<String, Object> getPayload() {
        return payload;
    }

    public void setPayload(Map<String, Object> payload) {
        this.payload = payload;
    }
}
