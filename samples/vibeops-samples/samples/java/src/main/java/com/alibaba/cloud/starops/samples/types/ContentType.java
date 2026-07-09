package com.alibaba.cloud.starops.samples.types;

/**
 * 内容类型枚举
 * Content type enumeration
 */
public enum ContentType {
    /** 纯文本 / Plain text */
    TEXT("text"),
    /** 旋转文本，用于显示工作和思考过程 / Spin text for showing work and thinking process */
    SPIN_TEXT("spin_text"),
    /** 图片 / Image */
    IMAGE("image");

    private final String value;

    ContentType(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }

    public static ContentType fromValue(String value) {
        for (ContentType type : values()) {
            if (type.value.equals(value)) {
                return type;
            }
        }
        return null;
    }
}
