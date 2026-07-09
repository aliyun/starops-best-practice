package com.alibaba.cloud.starops.samples.client;

import java.util.HashMap;
import java.util.Map;

import com.aliyun.sdk.service.starops20260428.models.CreateChatRequest;
import com.fasterxml.jackson.databind.JsonNode;

/**
 * SSE 重试支持：状态、结果枚举与纯函数工具集
 * SSE retry support: state, outcome enum and pure helper functions
 */
public final class RetrySupport {

    private RetrySupport() {}

    /**
     * 重连过程中的状态聚合
     * Aggregated state during reconnection
     */
    public static class RetryState {
        /** 最后一条已转发消息的时间戳，用于重连去重 */
        String lastTimestamp = "";
        /** true=重连后去重窗口，仅转发更新的消息 */
        boolean inDedupeWindow = false;
        /** 当前连续重试次数 */
        int retryCount = 0;
    }

    /**
     * 单次连接的结束原因
     * Outcome of a single connection
     */
    public enum ConnectionOutcome {
        /** 收到 stream_done，正常结束 */
        DONE,
        /** 连接中断，需重连 */
        INTERRUPTED,
        /** 致命错误/取消，直接返回 */
        FATAL
    }

    /**
     * 判断 ts 是否比 base 更新
     * 优先数值比较（Unix 时间戳），无法解析时 fallback 为字符串比较
     */
    public static boolean isNewerTimestamp(String ts, String base) {
        if (ts == null || ts.isEmpty()) {
            return false;
        }
        if (base == null || base.isEmpty()) {
            return true;
        }
        Long tsVal = tryParseLong(ts);
        Long baseVal = tryParseLong(base);
        if (tsVal != null && baseVal != null) {
            return tsVal > baseVal;
        }
        if (tsVal == null && baseVal != null) {
            // 基准是数值但当前 ts 无法解析，视为不更新
            return false;
        }
        return ts.compareTo(base) > 0;
    }

    private static Long tryParseLong(String s) {
        try {
            return Long.parseLong(s);
        } catch (NumberFormatException e) {
            return null;
        }
    }

    /**
     * 从事件中提取比 base 更新的最大消息 timestamp
     * 返回空字符串表示没有比 base 更新的时间戳
     */
    public static String extractNewestTimestamp(ChatEvent event, String base) {
        if (event == null || event.getBody() == null) {
            return "";
        }
        JsonNode body = event.getBody();
        JsonNode messages = body.get("messages");
        if (messages == null || !messages.isArray()) {
            return "";
        }
        String newest = base == null ? "" : base;
        for (JsonNode msg : messages) {
            if (msg == null) {
                continue;
            }
            JsonNode tsNode = msg.get("timestamp");
            String ts = tsNode != null ? tsNode.asText("") : "";
            if (isNewerTimestamp(ts, newest)) {
                newest = ts;
            }
        }
        if (newest.equals(base == null ? "" : base)) {
            return "";
        }
        return newest;
    }

    /**
     * 计算退避时间(ms)
     * default config: initialBackoffMs * backoffFactor^(retryCount-1)，上限 maxBackoffMs
     */
    public static long calculateBackoff(int retryCount, RetryConfig config) {
        double backoff = config.getInitialBackoffMs()
                * Math.pow(config.getBackoffFactor(), retryCount - 1);
        if (backoff > config.getMaxBackoffMs()) {
            return config.getMaxBackoffMs();
        }
        return (long) backoff;
    }

    /**
     * 判断事件是否为 stream_done（正常结束标志）
     * 不检查 event 字段，只判断 messages[].events[].type == stream_done
     */
    public static boolean isStreamDoneEvent(ChatEvent event) {
        if (event == null || event.getBody() == null) {
            return false;
        }
        JsonNode messages = event.getBody().get("messages");
        if (messages == null || !messages.isArray()) {
            return false;
        }
        for (JsonNode msg : messages) {
            if (msg == null) {
                continue;
            }
            JsonNode events = msg.get("events");
            if (events == null || !events.isArray()) {
                continue;
            }
            for (JsonNode evt : events) {
                JsonNode type = evt.get("type");
                if (type != null && "stream_done".equals(type.asText())) {
                    return true;
                }
            }
        }
        return false;
    }

    /**
     * 构建重连请求：action="reconnect"，复制 threadId/digitalEmployeeName/variables
     * Build reconnect request, copying threadId/digitalEmployeeName/variables
     */
    public static CreateChatRequest buildReconnectRequest(CreateChatRequest origReq) {
        Map<String, Object> variables = new HashMap<>();
        if (origReq.getVariables() != null) {
            for (Map.Entry<String, ?> entry : origReq.getVariables().entrySet()) {
                variables.put(entry.getKey(), entry.getValue());
            }
        }

        CreateChatRequest.Builder builder = CreateChatRequest.builder()
                .action("reconnect")
                .threadId(origReq.getThreadId())
                .digitalEmployeeName(origReq.getDigitalEmployeeName())
                .variables(variables);

        if (origReq.getRegionId() != null) {
            builder.regionId(origReq.getRegionId());
        }

        // 重连不需要 Messages
        return builder.build();
    }
}
