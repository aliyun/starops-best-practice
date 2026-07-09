package com.alibaba.cloud.starops.samples.client;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

/**
 * 简洁打印器测试
 * Simple printer tests
 */
class SimplePrinterTest {
    private SimplePrinter printer;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        printer = new SimplePrinter();
        objectMapper = new ObjectMapper();
    }

    @Test
    void testProcessNullEvent() {
        String result = printer.processEvent(null);
        assertEquals("", result);
    }

    @Test
    void testProcessEventWithNullBody() {
        ChatEvent event = new ChatEvent();
        String result = printer.processEvent(event);
        assertEquals("", result);
    }

    @Test
    void testProcessEventWithSystemArtifacts() throws Exception {
        String json = "{\"messages\":[{\"role\":\"system\",\"artifacts\":[{\"parts\":[{\"kind\":\"text\",\"text\":\"Hello World\"}]}]}]}";
        JsonNode body = objectMapper.readTree(json);

        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        String result = printer.processEvent(event);

        assertEquals("Hello World", result);
        assertEquals("Hello World", printer.getFinalText());
    }

    @Test
    void testProcessEventIgnoresAssistantRole() throws Exception {
        String json = "{\"messages\":[{\"role\":\"assistant\",\"artifacts\":[{\"parts\":[{\"kind\":\"text\",\"text\":\"Should be ignored\"}]}]}]}";
        JsonNode body = objectMapper.readTree(json);

        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        String result = printer.processEvent(event);

        assertEquals("", result);
    }

    @Test
    void testDeduplication() throws Exception {
        String json = "{\"messages\":[{\"role\":\"system\",\"artifacts\":[{\"parts\":[{\"kind\":\"text\",\"text\":\"Duplicate\"}]}]}]}";
        JsonNode body = objectMapper.readTree(json);

        ChatEvent event = ChatEvent.fromResponse(body, json, 200);

        // Process same event twice
        String result1 = printer.processEvent(event);
        String result2 = printer.processEvent(event);

        assertEquals("Duplicate", result1);
        assertEquals("", result2); // Should be empty due to deduplication
        assertEquals("Duplicate", printer.getFinalText());
    }

    @Test
    void testReset() throws Exception {
        String json = "{\"messages\":[{\"role\":\"system\",\"artifacts\":[{\"parts\":[{\"kind\":\"text\",\"text\":\"Test\"}]}]}]}";
        JsonNode body = objectMapper.readTree(json);

        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        printer.processEvent(event);

        assertEquals("Test", printer.getFinalText());

        printer.reset();

        assertEquals("", printer.getFinalText());

        // After reset, same content should be processed again
        String result = printer.processEvent(event);
        assertEquals("Test", result);
    }

    @Test
    void testMultipleArtifacts() throws Exception {
        String json = "{\"messages\":[{\"role\":\"system\",\"artifacts\":[" +
                "{\"parts\":[{\"kind\":\"text\",\"text\":\"Part1\"}]}," +
                "{\"parts\":[{\"kind\":\"text\",\"text\":\"Part2\"}]}" +
                "]}]}";
        JsonNode body = objectMapper.readTree(json);

        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        String result = printer.processEvent(event);

        assertEquals("Part1Part2", result);
    }

    @Test
    void testIgnoresNonTextKind() throws Exception {
        String json = "{\"messages\":[{\"role\":\"system\",\"artifacts\":[{\"parts\":[{\"kind\":\"image\",\"url\":\"http://example.com\"}]}]}]}";
        JsonNode body = objectMapper.readTree(json);

        ChatEvent event = ChatEvent.fromResponse(body, json, 200);
        String result = printer.processEvent(event);

        assertEquals("", result);
    }
}
