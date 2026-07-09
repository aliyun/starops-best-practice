package com.alibaba.cloud.starops.samples.types;

/**
 * 事件类型枚举
 * Event type enumeration
 */
public enum EventType {
    /** 会话标题更新 / Thread title updated */
    THREAD_TITLE_UPDATED("thread_title_updated"),
    /** 错误事件 / Error event */
    ERROR("error"),
    /** 思考事件 / Thinking event */
    THINKING("thinking"),
    /** 交互事件 / Interactive event */
    INTERACTIVE("interactive"),
    /** 交互响应事件 / Interactive response event */
    INTERACTIVE_RESPONSE("interactive_response"),
    /** 任务完成事件 / Task finished event */
    TASK_FINISHED("task_finished"),
    /** 取消事件 / Cancel event */
    CANCEL("cancel");

    private final String value;

    EventType(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }

    public static EventType fromValue(String value) {
        for (EventType type : values()) {
            if (type.value.equals(value)) {
                return type;
            }
        }
        return null;
    }
}
