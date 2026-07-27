package com.alibaba.cloud.starops.samples.client;

import com.alibaba.cloud.starops.samples.types.ChatEvent;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import static org.junit.jupiter.api.Assertions.*;

/**
 * RetrySupport 纯函数测试
 * Tests for RetrySupport pure utility functions
 */
class RetrySupportTest {

    private final ObjectMapper objectMapper = new ObjectMapper();

    // ============================================================
    // isNewerTimestamp 测试
    // ============================================================

    @ParameterizedTest
    @CsvSource({
        "100, 50, true",
        "50, 100, false",
        "100, 100, false",
        "'', 50, false",
        "100, '', true",
        "'', '', false",
        "b, a, true",
        "a, b, false",
        "abc, 100, false"   // base是数值但ts无法解析
    })
    void testIsNewerTimestamp(String ts, String base, boolean expected) {
        assertEquals(expected, RetrySupport.isNewerTimestamp(ts, base));
    }

    @Test
    void testIsNewerTimestamp_NullInputs() {
        assertFalse(RetrySupport.isNewerTimestamp(null, "100"));
        assertTrue(RetrySupport.isNewerTimestamp("100", null));
        assertFalse(RetrySupport.isNewerTimestamp(null, null));
    }

    // ============================================================
    // calculateBackoff 测试
    // ============================================================

    @Test
    void testCalculateBackoff_FirstAttempt() {
        RetryConfig config = RetryConfig.getDefault();
        long backoff = RetrySupport.calculateBackoff(1, config);
        assertEquals(config.getInitialBackoffMs(), backoff);
    }

    @Test
    void testCalculateBackoff_Exponential() {
        RetryConfig config = RetryConfig.getDefault();
        long first = RetrySupport.calculateBackoff(1, config);
        long second = RetrySupport.calculateBackoff(2, config);
        long third = RetrySupport.calculateBackoff(3, config);

        assertTrue(second > first, "second backoff should be greater than first");
        assertTrue(third > second, "third backoff should be greater than second");
    }

    @Test
    void testCalculateBackoff_CappedAtMax() {
        RetryConfig config = RetryConfig.getDefault();
        long backoff = RetrySupport.calculateBackoff(100, config);
        assertEquals(config.getMaxBackoffMs(), backoff);
    }

    // ============================================================
    // isStreamDoneEvent 测试
    // ============================================================

    @Test
    void testIsStreamDoneEvent_NullEvent() {
        assertFalse(RetrySupport.isStreamDoneEvent(null));
    }

    @Test
    void testIsStreamDoneEvent_NullBody() {
        ChatEvent event = new ChatEvent();
        assertFalse(RetrySupport.isStreamDoneEvent(event));
    }

    @Test
    void testIsStreamDoneEvent_WithStreamDone() throws Exception {
        String json = "{\"messages\":[{\"events\":[{\"type\":\"stream_done\"}]}]}";
        JsonNode body = objectMapper.readTree(json);
        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        assertTrue(RetrySupport.isStreamDoneEvent(event));
    }

    @Test
    void testIsStreamDoneEvent_WithoutStreamDone() throws Exception {
        String json = "{\"messages\":[{\"events\":[{\"type\":\"text\"}]}]}";
        JsonNode body = objectMapper.readTree(json);
        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        assertFalse(RetrySupport.isStreamDoneEvent(event));
    }

    @Test
    void testIsStreamDoneEvent_NoEvents() throws Exception {
        String json = "{\"messages\":[{\"role\":\"assistant\"}]}";
        JsonNode body = objectMapper.readTree(json);
        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        assertFalse(RetrySupport.isStreamDoneEvent(event));
    }

    // ============================================================
    // extractNewestTimestamp 测试
    // ============================================================

    @Test
    void testExtractNewestTimestamp_NullEvent() {
        assertEquals("", RetrySupport.extractNewestTimestamp(null, ""));
    }

    @Test
    void testExtractNewestTimestamp_NullBody() {
        ChatEvent event = new ChatEvent();
        assertEquals("", RetrySupport.extractNewestTimestamp(event, ""));
    }

    @Test
    void testExtractNewestTimestamp_NewerThanBase() throws Exception {
        String json = "{\"messages\":[{\"timestamp\":\"200\"},{\"timestamp\":\"300\"}]}";
        JsonNode body = objectMapper.readTree(json);
        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        assertEquals("300", RetrySupport.extractNewestTimestamp(event, "100"));
    }

    @Test
    void testExtractNewestTimestamp_NotNewerThanBase() throws Exception {
        String json = "{\"messages\":[{\"timestamp\":\"50\"},{\"timestamp\":\"80\"}]}";
        JsonNode body = objectMapper.readTree(json);
        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        assertEquals("", RetrySupport.extractNewestTimestamp(event, "100"));
    }

    @Test
    void testExtractNewestTimestamp_EmptyBase() throws Exception {
        String json = "{\"messages\":[{\"timestamp\":\"100\"}]}";
        JsonNode body = objectMapper.readTree(json);
        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        assertEquals("100", RetrySupport.extractNewestTimestamp(event, ""));
    }

    @Test
    void testExtractNewestTimestamp_NoTimestampField() throws Exception {
        String json = "{\"messages\":[{\"role\":\"assistant\"}]}";
        JsonNode body = objectMapper.readTree(json);
        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        assertEquals("", RetrySupport.extractNewestTimestamp(event, ""));
    }
}
