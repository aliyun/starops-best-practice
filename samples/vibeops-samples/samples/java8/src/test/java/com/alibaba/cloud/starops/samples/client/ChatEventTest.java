package com.alibaba.cloud.starops.samples.client;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.JsonNode;
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

/**
 * ChatEvent 单元测试
 */
class ChatEventTest {
    private final ObjectMapper mapper = new ObjectMapper();

    @Test
    void fromResponse_eventFieldDone_isDoneTrue() throws Exception {
        JsonNode body = mapper.readTree("{\"event\":\"done\",\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertTrue(event.isDone());
    }

    @Test
    void fromResponse_messagesFallback_isDoneTrue() throws Exception {
        JsonNode body = mapper.readTree("{\"messages\":[{\"type\":\"done\"}]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertTrue(event.isDone());
    }

    @Test
    void fromResponse_notDone() throws Exception {
        JsonNode body = mapper.readTree("{\"event\":\"text\",\"messages\":[{\"type\":\"text\"}]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertFalse(event.isDone());
    }

    @Test
    void fromResponse_extractsIdAndEvent() throws Exception {
        JsonNode body = mapper.readTree("{\"id\":\"evt-123\",\"event\":\"text\",\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertEquals("evt-123", event.getId());
        assertEquals("text", event.getEvent());
    }

    @Test
    void fromResponse_missingIdEvent_returnsNull() throws Exception {
        JsonNode body = mapper.readTree("{\"messages\":[]}");
        ChatEvent event = ChatEvent.fromResponse(body, "{}", 200);
        assertNull(event.getId());
        assertNull(event.getEvent());
    }

    @Test
    void simplePrinter_skipsNonTextEvent() throws Exception {
        SimplePrinter printer = new SimplePrinter();
        ChatEvent event = new ChatEvent();
        event.setEvent("thinking");
        event.setBody(mapper.readTree("{\"messages\":[]}"));
        assertEquals("", printer.processEvent(event));
    }

    @Test
    void simplePrinter_allowsTextEvent() throws Exception {
        SimplePrinter printer = new SimplePrinter();
        ChatEvent event = new ChatEvent();
        event.setEvent("text");
        event.setBody(mapper.readTree("{\"messages\":[]}"));
        assertEquals("", printer.processEvent(event));
    }

    @Test
    void simplePrinter_allowsNullEvent() throws Exception {
        SimplePrinter printer = new SimplePrinter();
        ChatEvent event = new ChatEvent();
        event.setBody(mapper.readTree("{\"messages\":[]}"));
        assertEquals("", printer.processEvent(event));
    }
}
