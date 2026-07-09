package com.alibaba.cloud.starops.samples.types;

/**
 * 交互类型枚举
 * Interaction type enumeration
 */
public enum InteractionType {
    /** 点击确认 / User acknowledgment */
    USER_ACK("user_ack"),
    /** 选择框 / User selection */
    USER_SELECT("user_select"),
    /** 用户输入 / User input */
    USER_INPUT("user_input"),
    /** SLS 查询 / SLS query */
    SLS_QUERY("sls_query");

    private final String value;

    InteractionType(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }

    public static InteractionType fromValue(String value) {
        for (InteractionType type : values()) {
            if (type.value.equals(value)) {
                return type;
            }
        }
        return null;
    }
}
