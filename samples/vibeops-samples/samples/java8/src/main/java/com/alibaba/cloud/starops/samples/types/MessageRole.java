package com.alibaba.cloud.starops.samples.types;

/**
 * 消息角色枚举
 * Message role enumeration
 */
public enum MessageRole {
    /** 用户输入 / User input */
    USER("user"),
    /** Agent 回复或操作 / Agent reply or action */
    ASSISTANT("assistant"),
    /** 系统消息 / System message */
    SYSTEM("system");

    private final String value;

    MessageRole(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }

    public static MessageRole fromValue(String value) {
        for (MessageRole role : values()) {
            if (role.value.equals(value)) {
                return role;
            }
        }
        return null;
    }
}
