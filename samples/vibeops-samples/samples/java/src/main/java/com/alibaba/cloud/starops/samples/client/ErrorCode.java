package com.alibaba.cloud.starops.samples.client;

/**
 * 错误码枚举
 * Error code enumeration
 */
public enum ErrorCode {
    /** 配置缺失 / Configuration missing */
    CONFIG_MISSING("CONFIG_MISSING"),
    /** 配置无效 / Configuration invalid */
    CONFIG_INVALID("CONFIG_INVALID"),
    /** 客户端创建失败 / Client creation failed */
    CLIENT_CREATE("CLIENT_CREATE"),
    /** 会话创建失败 / Thread creation failed */
    THREAD_CREATE("THREAD_CREATE"),
    /** 会话不存在 / Thread not found */
    THREAD_NOT_FOUND("THREAD_NOT_FOUND"),
    /** 对话失败 / Chat failed */
    CHAT_FAILED("CHAT_FAILED"),
    /** 超时 / Timeout */
    TIMEOUT("TIMEOUT"),
    /** 已取消 / Cancelled */
    CANCELLED("CANCELLED"),
    /** 网络错误 / Network error */
    NETWORK_ERROR("NETWORK_ERROR"),
    /** API 错误 / API error */
    API_ERROR("API_ERROR"),
    /** 解析错误 / Parse error */
    PARSE_ERROR("PARSE_ERROR"),
    /** 交互超时 / Interactive timeout */
    INTERACTIVE_TIMEOUT("INTERACTIVE_TIMEOUT");

    private final String code;

    ErrorCode(String code) {
        this.code = code;
    }

    public String getCode() {
        return code;
    }
}
