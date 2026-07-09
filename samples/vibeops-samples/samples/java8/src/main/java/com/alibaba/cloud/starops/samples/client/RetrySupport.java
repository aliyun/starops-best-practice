package com.alibaba.cloud.starops.samples.client;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.BlockingQueue;

import com.aliyun.sdk.service.starops20260428.models.CreateChatRequest;
import com.fasterxml.jackson.databind.JsonNode;

/**
 * SSE 重试支持工具（Java 8 版本，严格 JDK 8 语法）
 * SSE retry support helpers
 *
 * 跨语言对齐 Go 参考实现（samples/golang/internal/client/retry.go）：
 * - 重试触发：所有非 stream_done 的中断均触发重连
 * - 唯一结束：stream_done 是正常结束的唯一标志
 * - 不区分错误：不判断错误类型，一律重试
 * - 去重机制：inDedupeWindow + lastTimestamp，重连后仅转发比上次更新的消息
 */
public final class RetrySupport {

    /** SSE 流正常结束标志 / Normal stream termination marker */
    public static final String STREAM_DONE = "stream_done";

    private RetrySupport() {
    }

    /**
     * 聚合重连过程中的状态
     * Retry state aggregated across reconnect attempts
     */
    public static class RetryState {
        /** 最后一条已转发消息的时间戳，用于重连去重 */
        private String lastTimestamp = "";
        /** true=重连后去重窗口，仅转发更新的消息 */
        private boolean inDedupeWindow = false;
        /** 当前连续重试次数 */
        private int retryCount = 0;

        public String getLastTimestamp() {
            return lastTimestamp;
        }

        public void setLastTimestamp(String lastTimestamp) {
            this.lastTimestamp = lastTimestamp;
        }

        public boolean isInDedupeWindow() {
            return inDedupeWindow;
        }

        public void setInDedupeWindow(boolean inDedupeWindow) {
            this.inDedupeWindow = inDedupeWindow;
        }

        public int getRetryCount() {
            return retryCount;
        }

        public void setRetryCount(int retryCount) {
            this.retryCount = retryCount;
        }
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
        /** 不可恢复，直接结束 */
        FATAL
    }

    /**
     * 判断事件是否为 stream_done（正常结束标志）
     * 不检查 event.event 字段，只判断 messages[].events[].type == stream_done
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
                if (evt == null) {
                    continue;
                }
                JsonNode type = evt.get("type");
                if (type != null && STREAM_DONE.equals(type.asText())) {
                    return true;
                }
            }
        }
        return false;
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
            return tsVal.longValue() > baseVal.longValue();
        }
        if (tsVal == null && baseVal != null) {
            return false; // 基准是数值但当前 ts 无法解析，视为不更新
        }
        return ts.compareTo(base) > 0;
    }

    private static Long tryParseLong(String s) {
        try {
            return Long.valueOf(s.trim());
        } catch (NumberFormatException e) {
            return null;
        }
    }

    /**
     * 从事件中提取比 base 更新的最大消息 timestamp
     * 返回空字符串表示没有比 base 更新的时间戳
     */
    public static String extractNewestTimestamp(ChatEvent event, String base) {
        String baseVal = base != null ? base : "";
        if (event == null || event.getBody() == null) {
            return "";
        }
        JsonNode messages = event.getBody().get("messages");
        if (messages == null || !messages.isArray()) {
            return "";
        }
        String newest = baseVal;
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
        if (newest.equals(baseVal)) {
            return "";
        }
        return newest;
    }

    /**
     * 去重转发普通事件，返回是否实际转发了消息
     * Forward a normal event with dedupe; returns whether it was actually forwarded
     */
    public static boolean forwardEvent(ChatEvent event, RetryState state,
            BlockingQueue<ChatEvent> events) throws InterruptedException {
        String ts = extractNewestTimestamp(event, state.getLastTimestamp());

        if (state.isInDedupeWindow()) {
            if (ts.isEmpty()) {
                return false; // 重复消息，跳过
            }
            state.setInDedupeWindow(false); // 收到新消息，退出去重窗口
        }

        if (!ts.isEmpty()) {
            state.setLastTimestamp(ts);
        }
        events.put(event);
        return true;
    }

    /**
     * 构建重连请求
     * action="reconnect"，复制 threadId/digitalEmployeeName/variables，不携带 messages
     */
    public static CreateChatRequest buildReconnectRequest(CreateChatRequest origReq) {
        Map<String, Object> variables = new HashMap<String, Object>();
        if (origReq != null && origReq.getVariables() != null) {
            variables.putAll(origReq.getVariables());
        }

        CreateChatRequest.Builder builder = CreateChatRequest.builder()
                .action("reconnect")
                .variables(variables);

        if (origReq != null) {
            if (origReq.getRegionId() != null) {
                builder.regionId(origReq.getRegionId());
            }
            if (origReq.getThreadId() != null) {
                builder.threadId(origReq.getThreadId());
            }
            if (origReq.getDigitalEmployeeName() != null) {
                builder.digitalEmployeeName(origReq.getDigitalEmployeeName());
            }
        }

        return builder.build();
    }
}
