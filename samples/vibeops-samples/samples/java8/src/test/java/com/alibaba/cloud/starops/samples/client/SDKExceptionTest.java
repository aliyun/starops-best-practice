package com.alibaba.cloud.starops.samples.client;

import org.junit.jupiter.api.Test;
import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

/**
 * SDK 异常测试
 * SDK exception tests
 */
class SDKExceptionTest {

    @Test
    void testConfigMissing() {
        List<String> missingVars = Arrays.asList("VAR1", "VAR2");
        SDKException ex = SDKException.configMissing(missingVars);

        assertEquals(ErrorCode.CONFIG_MISSING, ex.getCode());
        assertTrue(ex.getMessage().contains("VAR1"));
        assertTrue(ex.getMessage().contains("VAR2"));
        assertNotNull(ex.getSuggestion());
        assertEquals(missingVars, ex.getContext().get("missingVariables"));
    }

    @Test
    void testConfigInvalid() {
        SDKException ex = SDKException.configInvalid("endpoint", "invalid URL");

        assertEquals(ErrorCode.CONFIG_INVALID, ex.getCode());
        assertTrue(ex.getMessage().contains("endpoint"));
        assertTrue(ex.getMessage().contains("invalid URL"));
        assertEquals("endpoint", ex.getContext().get("field"));
    }

    @Test
    void testThreadNotFound() {
        SDKException ex = SDKException.threadNotFound("thread-123");

        assertEquals(ErrorCode.THREAD_NOT_FOUND, ex.getCode());
        assertTrue(ex.getMessage().contains("thread-123"));
        assertEquals("thread-123", ex.getContext().get("threadId"));
    }

    @Test
    void testTimeout() {
        SDKException ex = SDKException.timeout("30s");

        assertEquals(ErrorCode.TIMEOUT, ex.getCode());
        assertTrue(ex.getMessage().contains("30s"));
    }

    @Test
    void testWithContext() {
        SDKException ex = new SDKException(ErrorCode.API_ERROR, "test error")
                .withContext("key1", "value1")
                .withContext("key2", 123);

        assertEquals("value1", ex.getContext().get("key1"));
        assertEquals(123, ex.getContext().get("key2"));
    }

    @Test
    void testWithSuggestion() {
        SDKException ex = new SDKException(ErrorCode.API_ERROR, "test error")
                .withSuggestion("Try again later");

        assertEquals("Try again later", ex.getSuggestion());
    }

    @Test
    void testToString() {
        SDKException ex = new SDKException(ErrorCode.API_ERROR, "test error");
        String str = ex.toString();

        assertTrue(str.contains("API_ERROR"));
        assertTrue(str.contains("test error"));
    }

    @Test
    void testWithCause() {
        Exception cause = new RuntimeException("root cause");
        SDKException ex = new SDKException(ErrorCode.NETWORK_ERROR, "network failed", cause);

        assertEquals(cause, ex.getCause());
        assertTrue(ex.toString().contains("root cause"));
    }
}
