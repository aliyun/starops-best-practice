package com.alibaba.cloud.starops.samples.client;

import com.alibaba.cloud.starops.samples.types.EventType;
import com.alibaba.cloud.starops.samples.types.InteractionType;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.*;
import java.time.Duration;
import java.util.*;
import java.util.concurrent.*;

/**
 * 交互事件处理器
 * Interactive event handler
 */
public class InteractiveHandler {
    private final AgentClient client;
    private final Duration timeout;
    private BufferedReader reader;
    private PrintWriter writer;
    private final ObjectMapper objectMapper;

    public InteractiveHandler(AgentClient client, Duration timeout) {
        this.client = client;
        this.timeout = timeout;
        this.reader = new BufferedReader(new InputStreamReader(System.in));
        this.writer = new PrintWriter(System.out, true);
        this.objectMapper = new ObjectMapper();
    }

    public void setIO(Reader reader, Writer writer) {
        this.reader = new BufferedReader(reader);
        this.writer = new PrintWriter(writer, true);
    }

    /**
     * 处理交互事件
     * Handle interactive event
     * callId: 来自外层 MessageItem.CallID，非 payload 内部字段
     */
    public InteractiveResponse handleEvent(JsonNode event, String callId) throws SDKException {
        if (event == null) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "事件为空");
        }

        String eventType = event.has("type") ? event.get("type").asText() : null;
        if (!EventType.INTERACTIVE.getValue().equals(eventType)) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "不支持的事件类型: " + eventType);
        }

        JsonNode payload = event.get("payload");
        if (payload == null) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "交互负载为空");
        }

        String interactiveType = payload.has("type") ? payload.get("type").asText() : null;
        InteractionType type = InteractionType.fromValue(interactiveType);

        if (type == null) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "不支持的交互类型: " + interactiveType);
        }

        switch (type) {
            case USER_ACK:
                return handleUserAck(payload, callId);
            case USER_SELECT:
                return handleUserSelect(payload, callId);
            case USER_INPUT:
                return handleUserInput(payload, callId);
            default:
                throw new SDKException(ErrorCode.PARSE_ERROR, "不支持的交互类型: " + interactiveType);
        }
    }

    /**
     * 处理用户确认
     * Handle user acknowledgment
     * 字段映射: title ← userAck.data.title, message ← userAck.message
     */
    public InteractiveResponse handleUserAck(JsonNode payload, String callId) throws SDKException {
        String title = getTitle(payload);
        String message = getDescription(payload);
        List<Map<String, Object>> options = getOptions(payload);
        Map<String, Object> modifiedData = extractData(payload);

        writer.println("\n🔔 确认请求");
        writer.println("--------------");
        if (title != null && !title.isEmpty()) {
            writer.println(title);
        }
        if (message != null && !message.isEmpty()) {
            writer.println("\n" + message);
        }

        // 展示 data 详情字段
        if (modifiedData != null) {
            writer.println();
            for (Map.Entry<String, Object> entry : modifiedData.entrySet()) {
                if ("title".equals(entry.getKey()) || "message".equals(entry.getKey())) {
                    continue;
                }
                writer.println(entry.getKey() + ": " + entry.getValue());
            }
        }
        writer.println("--------------");

        // 构建选项提示
        if (!options.isEmpty()) {
            writer.print("请输入 ");
            for (int i = 0; i < options.size(); i++) {
                if (i > 0) writer.print(", ");
                Map<String, Object> opt = options.get(i);
                writer.printf("[%s] %s", getOptionValue(opt), getOptionLabel(opt, i));
            }
            writer.print(": ");
        } else {
            writer.print("请输入 [y/yes] 确认，[n/no] 取消: ");
        }
        writer.flush();

        String input = readInputWithTimeout();
        input = input.trim().toLowerCase();
        boolean confirmed = input.isEmpty() || input.equals("y") || input.equals("yes") || input.equals("是");

        String decision = confirmed ? "yes" : "no";

        // 如果用户输入匹配某个选项的 value，使用该 value 作为 decision
        for (Map<String, Object> opt : options) {
            if (input.equals(getOptionValue(opt).toLowerCase())) {
                decision = getOptionValue(opt);
                confirmed = true;
                break;
            }
        }

        Map<String, Object> source = extractSource(payload);

        Map<String, Object> response = new HashMap<>();
        response.put("confirmed", confirmed);

        InteractiveResponse resp = new InteractiveResponse(callId, InteractionType.USER_ACK, response);
        resp.setSource(source);
        resp.setModifiedData(modifiedData);
        resp.setDecision(decision);
        return resp;
    }

    /**
     * 处理用户选择
     * Handle user selection
     */
    public InteractiveResponse handleUserSelect(JsonNode payload, String callId) throws SDKException {
        String title = getTitle(payload);

        writer.println("\n📋 请选择");
        if (title != null && !title.isEmpty()) {
            writer.println("   标题: " + title);
        }

        List<Map<String, Object>> options = getOptions(payload);
        if (options.isEmpty()) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "没有可选项");
        }

        writer.println("   选项:");
        for (int i = 0; i < options.size(); i++) {
            String label = getOptionLabel(options.get(i), i);
            writer.println("   [" + (i + 1) + "] " + label);
        }
        writer.print("   请输入选项编号 (1-" + options.size() + "): ");
        writer.flush();

        String input = readInputWithTimeout();
        input = input.trim();

        int selectedIndex;
        try {
            selectedIndex = Integer.parseInt(input);
        } catch (NumberFormatException e) {
            throw new SDKException(ErrorCode.PARSE_ERROR,
                    "无效的选择: " + input + "，请输入 1-" + options.size() + " 之间的数字");
        }

        if (selectedIndex < 1 || selectedIndex > options.size()) {
            throw new SDKException(ErrorCode.PARSE_ERROR,
                    "无效的选择: " + input + "，请输入 1-" + options.size() + " 之间的数字");
        }

        Map<String, Object> selectedOption = options.get(selectedIndex - 1);
        String decision = getOptionValue(selectedOption);

        Map<String, Object> response = new HashMap<>();
        response.put("selectedIndex", selectedIndex - 1);
        response.put("selectedValue", selectedOption);

        InteractiveResponse resp = new InteractiveResponse(callId, InteractionType.USER_SELECT, response);
        resp.setSource(extractSource(payload));
        resp.setModifiedData(extractData(payload));
        resp.setDecision(decision);
        return resp;
    }

    /**
     * 处理用户输入（表单模式）
     * Handle user input (form mode)
     * 根据 formSpec 中的 ui_schema 逐字段提示用户输入
     */
    @SuppressWarnings("unchecked")
    public InteractiveResponse handleUserInput(JsonNode payload, String callId) throws SDKException {
        String title = getTitle(payload);
        String description = getDescription(payload);
        Map<String, Object> source = extractSource(payload);

        Map<String, Object> formSpec = extractFormSpec(payload);
        List<Map<String, Object>> elements = getFormElements(formSpec);
        Map<String, Object> initialValues = getFormInitialValues(formSpec);

        writer.println("\n✏️  " + title);
        if (description != null && !description.isEmpty()) {
            writer.println("    " + description);
        }
        writer.println("    " + repeatStr("-", 40));

        Map<String, Object> formData = new HashMap<>();

        for (Map<String, Object> elem : elements) {
            String field = getFieldKey(elem);
            String label = getFieldLabel(elem, field);
            String widget = getFieldWidget(elem);
            String placeholder = getFieldPlaceholder(elem);
            String defaultValue = getInitialValue(initialValues, field);

            if ("radio".equals(widget) || "segmented".equals(widget)) {
                List<String> enumOpts = getFieldEnum(formSpec, field);
                if (!enumOpts.isEmpty()) {
                    writer.println("    " + label + ":");
                    for (int i = 0; i < enumOpts.size(); i++) {
                        String marker = enumOpts.get(i).equals(defaultValue) ? "*" : " ";
                        writer.printf("      [%d]%s %s%n", i + 1, marker, enumOpts.get(i));
                    }
                    writer.print("    请选择 (1-" + enumOpts.size() + ")");
                    if (!defaultValue.isEmpty()) {
                        writer.print(" [默认: " + defaultValue + "]");
                    }
                    writer.print(": ");
                    writer.flush();

                    String input = readInputWithTimeout().trim();
                    if (input.isEmpty() && !defaultValue.isEmpty()) {
                        formData.put(field, defaultValue);
                    } else {
                        try {
                            int idx = Integer.parseInt(input);
                            if (idx >= 1 && idx <= enumOpts.size()) {
                                formData.put(field, enumOpts.get(idx - 1));
                            } else {
                                formData.put(field, defaultValue);
                            }
                        } catch (NumberFormatException e) {
                            formData.put(field, defaultValue);
                        }
                    }
                }
            } else {
                writer.print("    " + label);
                if (!placeholder.isEmpty()) {
                    writer.print(" (" + placeholder + ")");
                }
                if (!defaultValue.isEmpty()) {
                    writer.print(" [默认: " + defaultValue + "]");
                }
                writer.print(": ");
                writer.flush();

                String input = readInputWithTimeout().trim();
                if (input.isEmpty() && !defaultValue.isEmpty()) {
                    formData.put(field, defaultValue);
                } else {
                    formData.put(field, input);
                }
            }
        }

        writer.println("    " + repeatStr("-", 40));

        Map<String, Object> response = new HashMap<>();
        response.put("value", formData);

        InteractiveResponse resp = new InteractiveResponse(callId, InteractionType.USER_INPUT, response);
        resp.setSource(source);
        resp.setFormData(formData);
        resp.setDecision("submit");
        return resp;
    }

    /**
     * 使用交互响应恢复对话
     * Resume chat with interactive response
     */
    public BlockingQueue<ChatEvent> resumeChat(String threadId, InteractiveResponse response,
            Map<String, Object> baseVariables) {
        if (client == null) {
            BlockingQueue<ChatEvent> events = new LinkedBlockingQueue<>();
            try {
                events.put(ChatEvent.error(new SDKException(ErrorCode.CLIENT_CREATE, "客户端未初始化")));
            } catch (InterruptedException ignored) {}
            return events;
        }

        if (response == null) {
            BlockingQueue<ChatEvent> events = new LinkedBlockingQueue<>();
            try {
                events.put(ChatEvent.error(new SDKException(ErrorCode.PARSE_ERROR, "交互响应为空")));
            } catch (InterruptedException ignored) {}
            return events;
        }

        Map<String, Object> userInteractive = new HashMap<>();
        userInteractive.put("callId", response.getCallId());
        userInteractive.put("source", response.getSource());
        userInteractive.put("decision", response.getDecision());
        if (response.getType() == InteractionType.USER_INPUT) {
            userInteractive.put("formData", response.getFormData());
        } else {
            userInteractive.put("modifiedData", response.getModifiedData());
        }

        String uiJson;
        try {
            uiJson = objectMapper.writeValueAsString(userInteractive);
        } catch (Exception e) {
            BlockingQueue<ChatEvent> events = new LinkedBlockingQueue<>();
            try {
                events.put(ChatEvent.error(new SDKException(ErrorCode.PARSE_ERROR, "序列化 userInteractive 失败", e)));
            } catch (InterruptedException ignored) {}
            return events;
        }

        return client.interact(threadId, uiJson, baseVariables);
    }

    // =================================================================================
    // 辅助方法
    // =================================================================================

    private String readInputWithTimeout() throws SDKException {
        if (timeout == null || timeout.isZero()) {
            try {
                return reader.readLine();
            } catch (IOException e) {
                throw new SDKException(ErrorCode.PARSE_ERROR, "读取输入失败", e);
            }
        }

        ExecutorService executor = Executors.newSingleThreadExecutor();
        Future<String> future = executor.submit(() -> reader.readLine());

        try {
            return future.get(timeout.toMillis(), TimeUnit.MILLISECONDS);
        } catch (TimeoutException e) {
            future.cancel(true);
            throw SDKException.interactiveTimeout(timeout.toString());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw SDKException.cancelled();
        } catch (ExecutionException e) {
            throw new SDKException(ErrorCode.PARSE_ERROR, "读取输入失败", e.getCause());
        } finally {
            executor.shutdownNow();
        }
    }

    private void printf(String format, Object... args) {
        writer.printf(format, args);
        writer.flush();
    }

    // ===== 取值方法 =====

    private String getTitle(JsonNode payload) {
        // userAck.data.title
        JsonNode userAck = payload.get("userAck");
        if (userAck != null) {
            JsonNode data = userAck.get("data");
            if (data != null && data.has("title")) {
                return data.get("title").asText();
            }
        }
        // userInput.title
        JsonNode userInput = payload.get("userInput");
        if (userInput != null && userInput.has("title")) {
            return userInput.get("title").asText();
        }
        // fallback: meta.title
        return getMetaField(payload, "title");
    }

    private String getDescription(JsonNode payload) {
        // userAck.message
        JsonNode userAck = payload.get("userAck");
        if (userAck != null && userAck.has("message")) {
            return userAck.get("message").asText();
        }
        // userInput.description
        JsonNode userInput = payload.get("userInput");
        if (userInput != null && userInput.has("description")) {
            return userInput.get("description").asText();
        }
        // fallback: meta
        String desc = getMetaField(payload, "description");
        if (desc != null) return desc;
        return getMetaField(payload, "desc");
    }

    private String getMetaField(JsonNode payload, String field) {
        JsonNode meta = payload.get("meta");
        if (meta != null && meta.has(field)) {
            return meta.get(field).asText();
        }
        return null;
    }

    @SuppressWarnings("unchecked")
    private List<Map<String, Object>> getOptions(JsonNode payload) {
        // userAck.options
        JsonNode userAck = payload.get("userAck");
        if (userAck != null && userAck.has("options") && userAck.get("options").isArray()) {
            List<Map<String, Object>> result = new ArrayList<>();
            for (JsonNode item : userAck.get("options")) {
                try {
                    result.add(objectMapper.convertValue(item, Map.class));
                } catch (Exception ignored) {}
            }
            if (!result.isEmpty()) return result;
        }

        // fallback: Data 字段
        if (payload.has("data") && payload.get("data").isArray()) {
            List<Map<String, Object>> result = new ArrayList<>();
            for (JsonNode item : payload.get("data")) {
                try {
                    result.add(objectMapper.convertValue(item, Map.class));
                } catch (Exception ignored) {}
            }
            if (!result.isEmpty()) return result;
        }

        // fallback: meta.options
        JsonNode meta = payload.get("meta");
        if (meta != null && meta.has("options") && meta.get("options").isArray()) {
            List<Map<String, Object>> result = new ArrayList<>();
            for (JsonNode item : meta.get("options")) {
                try {
                    result.add(objectMapper.convertValue(item, Map.class));
                } catch (Exception ignored) {}
            }
            return result;
        }

        return new ArrayList<>();
    }

    private String getOptionLabel(Map<String, Object> option, int index) {
        String[] fields = {"label", "name", "title", "value"};
        for (String field : fields) {
            Object value = option.get(field);
            if (value instanceof String && !((String) value).isEmpty()) {
                return (String) value;
            }
        }
        return "选项 " + (index + 1);
    }

    private String getOptionValue(Map<String, Object> option) {
        Object value = option.get("value");
        return value instanceof String ? (String) value : "";
    }

    @SuppressWarnings("unchecked")
    private Map<String, Object> extractSource(JsonNode payload) {
        // userAck.source
        JsonNode userAck = payload.get("userAck");
        if (userAck != null && userAck.has("source")) {
            return objectMapper.convertValue(userAck.get("source"), Map.class);
        }
        // userInput.source
        JsonNode userInput = payload.get("userInput");
        if (userInput != null && userInput.has("source")) {
            return objectMapper.convertValue(userInput.get("source"), Map.class);
        }
        // fallback: meta.source
        JsonNode meta = payload.get("meta");
        if (meta != null && meta.has("source")) {
            return objectMapper.convertValue(meta.get("source"), Map.class);
        }
        return null;
    }

    @SuppressWarnings("unchecked")
    private Map<String, Object> extractData(JsonNode payload) {
        // userAck.data
        JsonNode userAck = payload.get("userAck");
        if (userAck != null && userAck.has("data")) {
            return objectMapper.convertValue(userAck.get("data"), Map.class);
        }
        // fallback: meta.data
        JsonNode meta = payload.get("meta");
        if (meta != null && meta.has("data")) {
            return objectMapper.convertValue(meta.get("data"), Map.class);
        }
        return null;
    }

    // =================================================================================
    // formSpec 辅助方法 (user_input 表单模式)
    // =================================================================================

    @SuppressWarnings("unchecked")
    private Map<String, Object> extractFormSpec(JsonNode payload) {
        JsonNode userInput = payload.get("userInput");
        if (userInput != null && userInput.has("formSpec")) {
            return objectMapper.convertValue(userInput.get("formSpec"), Map.class);
        }
        return null;
    }

    @SuppressWarnings("unchecked")
    private List<Map<String, Object>> getFormElements(Map<String, Object> formSpec) {
        if (formSpec == null) return new ArrayList<>();
        Map<String, Object> uiSchema = (Map<String, Object>) formSpec.get("ui_schema");
        if (uiSchema == null) return new ArrayList<>();
        List<Map<String, Object>> elements = (List<Map<String, Object>>) uiSchema.get("elements");
        return elements != null ? elements : new ArrayList<>();
    }

    @SuppressWarnings("unchecked")
    private Map<String, Object> getFormInitialValues(Map<String, Object> formSpec) {
        if (formSpec == null) return null;
        return (Map<String, Object>) formSpec.get("initialValues");
    }

    private String getFieldKey(Map<String, Object> elem) {
        Object field = elem.get("field");
        return field instanceof String ? (String) field : "";
    }

    private String getFieldLabel(Map<String, Object> elem, String field) {
        Object label = elem.get("label");
        return label instanceof String ? (String) label : field;
    }

    private String getFieldWidget(Map<String, Object> elem) {
        Object widget = elem.get("widget");
        return widget instanceof String ? (String) widget : "input";
    }

    private String getFieldPlaceholder(Map<String, Object> elem) {
        Object placeholder = elem.get("placeholder");
        return placeholder instanceof String ? (String) placeholder : "";
    }

    private String getInitialValue(Map<String, Object> initialValues, String field) {
        if (initialValues == null) return "";
        Object val = initialValues.get(field);
        return val != null ? val.toString() : "";
    }

    @SuppressWarnings("unchecked")
    private List<String> getFieldEnum(Map<String, Object> formSpec, String field) {
        if (formSpec == null) return new ArrayList<>();
        Map<String, Object> schema = (Map<String, Object>) formSpec.get("schema");
        if (schema == null) return new ArrayList<>();
        Map<String, Object> properties = (Map<String, Object>) schema.get("properties");
        if (properties == null) return new ArrayList<>();
        Map<String, Object> prop = (Map<String, Object>) properties.get(field);
        if (prop == null) return new ArrayList<>();
        List<Object> enumVals = (List<Object>) prop.get("enum");
        if (enumVals == null) return new ArrayList<>();
        List<String> result = new ArrayList<>();
        for (Object v : enumVals) {
            result.add(v != null ? v.toString() : "");
        }
        return result;
    }

    private String repeatStr(String s, int count) {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < count; i++) {
            sb.append(s);
        }
        return sb.toString();
    }

    // =================================================================================
    // 便捷方法
    // =================================================================================

    /**
     * 检查事件是否为交互事件
     * Check if event is interactive
     */
    public static boolean isInteractiveEvent(JsonNode event) {
        if (event == null) return false;
        String type = event.has("type") ? event.get("type").asText() : null;
        return EventType.INTERACTIVE.getValue().equals(type);
    }

    /**
     * 从消息中提取交互事件
     * Extract interactive events from message
     */
    public static List<JsonNode> extractInteractiveEvents(JsonNode message) {
        List<JsonNode> events = new ArrayList<>();
        if (message == null || !message.has("events")) return events;

        JsonNode eventsNode = message.get("events");
        if (eventsNode.isArray()) {
            for (JsonNode event : eventsNode) {
                if (isInteractiveEvent(event)) {
                    events.add(event);
                }
            }
        }
        return events;
    }
}