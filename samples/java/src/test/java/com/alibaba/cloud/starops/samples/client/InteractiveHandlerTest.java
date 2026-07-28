package com.alibaba.cloud.starops.samples.client;

import com.alibaba.cloud.starops.samples.types.InteractionType;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;

import java.io.StringReader;
import java.io.StringWriter;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

/**
 * InteractiveHandler 单元测试
 * 覆盖：交互事件识别、交互事件提取、user_ack / user_input 两类事件的 happy path。
 * 不依赖真实 SDK 网络调用（client 传 null，输入输出使用内存流）。
 */
class InteractiveHandlerTest {
    private final ObjectMapper mapper = new ObjectMapper();

    // =====================================================================
    // isInteractiveEvent
    // =====================================================================

    @Test
    void isInteractiveEvent_null_returnsFalse() {
        assertFalse(InteractiveHandler.isInteractiveEvent(null));
    }

    @Test
    void isInteractiveEvent_interactive_returnsTrue() throws Exception {
        JsonNode event = mapper.readTree("{\"type\":\"interactive\"}");
        assertTrue(InteractiveHandler.isInteractiveEvent(event));
    }

    @Test
    void isInteractiveEvent_thinking_returnsFalse() throws Exception {
        JsonNode event = mapper.readTree("{\"type\":\"thinking\"}");
        assertFalse(InteractiveHandler.isInteractiveEvent(event));
    }

    // =====================================================================
    // extractInteractiveEvents
    // =====================================================================

    @Test
    void extractInteractiveEvents_nullMessage_returnsEmpty() {
        List<JsonNode> result = InteractiveHandler.extractInteractiveEvents(null);
        assertTrue(result.isEmpty());
    }

    @Test
    void extractInteractiveEvents_mixedEvents_returnsOnlyInteractive() throws Exception {
        JsonNode message = mapper.readTree(
                "{\"events\":[{\"type\":\"thinking\"},{\"type\":\"interactive\"},"
                        + "{\"type\":\"error\"},{\"type\":\"interactive\"}]}");
        List<JsonNode> result = InteractiveHandler.extractInteractiveEvents(message);
        assertEquals(2, result.size());
        for (JsonNode evt : result) {
            assertEquals("interactive", evt.get("type").asText());
        }
    }

    @Test
    void extractInteractiveEvents_noInteractive_returnsEmpty() throws Exception {
        JsonNode message = mapper.readTree(
                "{\"events\":[{\"type\":\"thinking\"},{\"type\":\"error\"}]}");
        List<JsonNode> result = InteractiveHandler.extractInteractiveEvents(message);
        assertTrue(result.isEmpty());
    }

    // =====================================================================
    // handleEvent - user_ack happy path
    // =====================================================================

    @Test
    void handleEvent_userAck_confirmYes() throws Exception {
        InteractiveHandler handler = new InteractiveHandler(null, null);
        handler.setIO(new StringReader("y\n"), new StringWriter());

        JsonNode event = mapper.readTree(
                "{\"type\":\"interactive\",\"payload\":{\"type\":\"user_ack\","
                        + "\"userAck\":{\"message\":\"是否继续?\","
                        + "\"data\":{\"title\":\"确认操作\"},"
                        + "\"source\":{\"app\":\"test-app\"}}}}");

        InteractiveResponse resp = handler.handleEvent(event, "call-001");

        assertNotNull(resp);
        assertEquals("call-001", resp.getCallId());
        assertEquals(InteractionType.USER_ACK, resp.getType());
        assertEquals("yes", resp.getDecision());
        assertEquals(Boolean.TRUE, resp.getResponse().get("confirmed"));
        assertNotNull(resp.getSource());
        assertEquals("test-app", resp.getSource().get("app"));
    }

    @Test
    void handleEvent_userAck_declineNo() throws Exception {
        InteractiveHandler handler = new InteractiveHandler(null, null);
        handler.setIO(new StringReader("n\n"), new StringWriter());

        JsonNode event = mapper.readTree(
                "{\"type\":\"interactive\",\"payload\":{\"type\":\"user_ack\","
                        + "\"userAck\":{\"message\":\"是否继续?\"}}}");

        InteractiveResponse resp = handler.handleEvent(event, "call-002");

        assertEquals("no", resp.getDecision());
        assertEquals(Boolean.FALSE, resp.getResponse().get("confirmed"));
    }

    // =====================================================================
    // handleEvent - user_input happy path
    // =====================================================================

    @Test
    void handleEvent_userInput_formSubmit() throws Exception {
        InteractiveHandler handler = new InteractiveHandler(null, null);
        handler.setIO(new StringReader("hello-world\n"), new StringWriter());

        JsonNode event = mapper.readTree(
                "{\"type\":\"interactive\",\"payload\":{\"type\":\"user_input\","
                        + "\"userInput\":{\"title\":\"填写表单\",\"description\":\"请输入名称\","
                        + "\"formSpec\":{\"ui_schema\":{\"elements\":[{\"field\":\"name\",\"label\":\"名称\"}]}}}}}");

        InteractiveResponse resp = handler.handleEvent(event, "call-003");

        assertNotNull(resp);
        assertEquals("call-003", resp.getCallId());
        assertEquals(InteractionType.USER_INPUT, resp.getType());
        assertEquals("submit", resp.getDecision());
        assertNotNull(resp.getFormData());
        assertEquals("hello-world", resp.getFormData().get("name"));
    }

    // =====================================================================
    // handleEvent - 异常输入
    // =====================================================================

    @Test
    void handleEvent_nullEvent_throws() {
        InteractiveHandler handler = new InteractiveHandler(null, null);
        assertThrows(SDKException.class, () -> handler.handleEvent(null, "call-x"));
    }

    @Test
    void handleEvent_nonInteractiveType_throws() throws Exception {
        InteractiveHandler handler = new InteractiveHandler(null, null);
        JsonNode event = mapper.readTree("{\"type\":\"thinking\"}");
        assertThrows(SDKException.class, () -> handler.handleEvent(event, "call-x"));
    }

    @Test
    void handleEvent_missingPayload_throws() throws Exception {
        InteractiveHandler handler = new InteractiveHandler(null, null);
        JsonNode event = mapper.readTree("{\"type\":\"interactive\"}");
        assertThrows(SDKException.class, () -> handler.handleEvent(event, "call-x"));
    }
}
