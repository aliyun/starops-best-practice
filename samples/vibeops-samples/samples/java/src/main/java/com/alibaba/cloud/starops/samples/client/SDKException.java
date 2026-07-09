package com.alibaba.cloud.starops.samples.client;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * SDK 异常类
 * SDK exception class
 */
public class SDKException extends Exception {
    private final ErrorCode code;
    private final Map<String, Object> context;
    private String suggestion;

    public SDKException(ErrorCode code, String message) {
        super(message);
        this.code = code;
        this.context = new HashMap<>();
    }

    public SDKException(ErrorCode code, String message, Throwable cause) {
        super(message, cause);
        this.code = code;
        this.context = new HashMap<>();
    }

    public SDKException withContext(String key, Object value) {
        this.context.put(key, value);
        return this;
    }

    public SDKException withSuggestion(String suggestion) {
        this.suggestion = suggestion;
        return this;
    }

    public ErrorCode getCode() {
        return code;
    }

    public Map<String, Object> getContext() {
        return context;
    }

    public String getSuggestion() {
        return suggestion;
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append("[").append(code.getCode()).append("] ").append(getMessage());
        if (getCause() != null) {
            sb.append(": ").append(getCause().getMessage());
        }
        return sb.toString();
    }

    // Convenience factory methods
    public static SDKException configMissing(List<String> missingVars) {
        return new SDKException(ErrorCode.CONFIG_MISSING, 
                "缺少必需的配置项: " + String.join(", ", missingVars))
                .withContext("missingVariables", missingVars)
                .withSuggestion("请检查 .env 文件或环境变量设置");
    }

    public static SDKException configInvalid(String field, String reason) {
        return new SDKException(ErrorCode.CONFIG_INVALID,
                String.format("配置项 %s 无效: %s", field, reason))
                .withContext("field", field)
                .withContext("reason", reason)
                .withSuggestion("请检查配置值是否正确");
    }

    public static SDKException clientCreate(Throwable cause) {
        return new SDKException(ErrorCode.CLIENT_CREATE, "创建客户端失败", cause)
                .withSuggestion("请检查网络连接和认证信息");
    }

    public static SDKException threadCreate(Throwable cause) {
        return new SDKException(ErrorCode.THREAD_CREATE, "创建会话失败", cause)
                .withSuggestion("请检查 API 权限和配额");
    }

    public static SDKException threadNotFound(String threadId) {
        return new SDKException(ErrorCode.THREAD_NOT_FOUND,
                String.format("会话不存在: %s", threadId))
                .withContext("threadId", threadId)
                .withSuggestion("请检查会话 ID 是否正确，或创建新会话");
    }

    public static SDKException chatFailed(Throwable cause) {
        return new SDKException(ErrorCode.CHAT_FAILED, "对话失败", cause)
                .withSuggestion("请稍后重试");
    }

    public static SDKException timeout(String duration) {
        return new SDKException(ErrorCode.TIMEOUT,
                String.format("操作超时: %s", duration))
                .withContext("duration", duration)
                .withSuggestion("请增加超时时间或检查网络连接");
    }

    public static SDKException cancelled() {
        return new SDKException(ErrorCode.CANCELLED, "操作已取消")
                .withSuggestion("如需继续，请重新发起请求");
    }

    public static SDKException networkError(Throwable cause) {
        return new SDKException(ErrorCode.NETWORK_ERROR, "网络错误", cause)
                .withSuggestion("请检查网络连接");
    }

    public static SDKException apiError(String apiCode, String apiMessage) {
        return new SDKException(ErrorCode.API_ERROR,
                String.format("API 错误 [%s]: %s", apiCode, apiMessage))
                .withContext("apiCode", apiCode)
                .withContext("apiMessage", apiMessage)
                .withSuggestion("请参考 API 文档检查请求参数");
    }

    public static SDKException parseError(Throwable cause) {
        return new SDKException(ErrorCode.PARSE_ERROR, "解析响应失败", cause)
                .withSuggestion("请检查 SDK 版本是否最新");
    }

    public static SDKException interactiveTimeout(String duration) {
        return new SDKException(ErrorCode.INTERACTIVE_TIMEOUT,
                String.format("等待用户响应超时: %s", duration))
                .withContext("duration", duration)
                .withSuggestion("请重新操作并在规定时间内响应");
    }
}
