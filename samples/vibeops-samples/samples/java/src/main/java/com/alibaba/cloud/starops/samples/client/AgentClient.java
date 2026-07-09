package com.alibaba.cloud.starops.samples.client;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.Iterator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

import com.aliyun.auth.credentials.Credential;
import com.aliyun.auth.credentials.provider.StaticCredentialProvider;
import com.aliyun.sdk.gateway.pop.Configuration;
import com.aliyun.sdk.gateway.pop.auth.SignatureVersion;
import com.aliyun.sdk.service.starops20260428.AsyncClient;
import com.aliyun.sdk.service.starops20260428.models.CreateChatRequest;
import com.aliyun.sdk.service.starops20260428.models.CreateChatResponseBody;
import com.aliyun.sdk.service.starops20260428.models.CreateThreadRequest;
import com.aliyun.sdk.service.starops20260428.models.CreateThreadResponse;
import com.aliyun.sdk.service.starops20260428.models.DeleteThreadRequest;
import com.aliyun.sdk.service.starops20260428.models.GetThreadDataRequest;
import com.aliyun.sdk.service.starops20260428.models.GetThreadDataResponse;
import com.aliyun.sdk.service.starops20260428.models.GetThreadDataResponseBody;
import com.aliyun.sdk.service.starops20260428.models.GetThreadRequest;
import com.aliyun.sdk.service.starops20260428.models.GetThreadResponse;
import com.aliyun.sdk.service.starops20260428.models.GetThreadResponseBody;
import com.aliyun.sdk.service.starops20260428.models.ListThreadsRequest;
import com.aliyun.sdk.service.starops20260428.models.ListThreadsResponse;
import com.aliyun.sdk.service.starops20260428.models.ListThreadsResponseBody;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import com.alibaba.cloud.starops.samples.client.RetrySupport.ConnectionOutcome;
import com.alibaba.cloud.starops.samples.client.RetrySupport.RetryState;

import darabonba.core.ResponseIterable;
import darabonba.core.client.ClientOverrideConfiguration;

/**
 * Agent 客户端（异步 SDK 版本，支持 SSE 流式输出）
 * Agent client using async SDK with SSE streaming support
 */
public class AgentClient implements AutoCloseable {
    private final AsyncClient client;
    private final Config config;
    private final ObjectMapper objectMapper;
    private final ExecutorService executor;

    public AgentClient(Config config) throws SDKException {
        this.config = config;
        this.objectMapper = new ObjectMapper();
        this.executor = Executors.newCachedThreadPool();

        try {
            StaticCredentialProvider provider = StaticCredentialProvider.create(
                    Credential.builder()
                            .accessKeyId(config.getAccessKeyId())
                            .accessKeySecret(config.getAccessKeySecret())
                            .build());

            this.client = AsyncClient.builder()
                    .credentialsProvider(provider)
                    .overrideConfiguration(
                            ClientOverrideConfiguration.create()
                                    .setEndpointOverride(config.getEndpoint()))
                    .serviceConfiguration(
                            Configuration.create()
                                    .setSignatureVersion(SignatureVersion.V3))
                    .build();
        } catch (Exception e) {
            throw SDKException.clientCreate(e);
        }
    }

    public Config getConfig() {
        return config;
    }

    /**
     * 创建会话
     * Create a new thread
     */
    public String createThread() throws SDKException {
        return createThread(null);
    }

    /**
     * 创建会话（支持自定义属性）
     * Create a new thread with optional attributes
     */
    public String createThread(Map<String, String> attributes) throws SDKException {
        try {
            CreateThreadRequest.Builder builder = CreateThreadRequest.builder()
                    .name(config.getEmployeeName())
                    .title("Chat-" + Instant.now().getEpochSecond())
                    .variables(CreateThreadRequest.Variables.builder()
                            .workspace(config.getWorkspace())
                            .build());

            if (attributes != null && !attributes.isEmpty()) {
                builder.attributes(attributes);
            }

            CreateThreadResponse response = client.createThread(builder.build()).get();

            if (response.getBody() == null || response.getBody().getThreadId() == null) {
                throw new SDKException(ErrorCode.THREAD_CREATE, "无效响应: 缺少ThreadID");
            }

            return response.getBody().getThreadId();
        } catch (SDKException e) {
            throw e;
        } catch (Exception e) {
            throw SDKException.threadCreate(e);
        }
    }

    /**
     * 开始 SSE 对话（基础版本）
     * Start SSE chat (basic version)
     */
    public BlockingQueue<ChatEvent> chat(String threadId, String message) {
        Map<String, Object> variables = new HashMap<>();
        variables.put("workspace", config.getWorkspace());
        variables.put("region", config.getRegion());
        variables.put("language", "zh");
        variables.put("timeZone", "Asia/Shanghai");
        variables.put("timeStamp", String.valueOf(Instant.now().getEpochSecond()));
        return chatWithVariables(threadId, message, variables);
    }

    /**
     * 开始对话（支持自定义 variables，使用异步 SDK 的 SSE 流式迭代）
     * Start chat with custom variables using async SDK's SSE streaming
     */
    public BlockingQueue<ChatEvent> chatWithVariables(String threadId, String message, Map<String, Object> variables) {
        BlockingQueue<ChatEvent> events = new LinkedBlockingQueue<>();

        final Map<String, Object> finalVariables = variables != null ? new HashMap<>(variables) : new HashMap<>();
        finalVariables.putIfAbsent("workspace", config.getWorkspace());
        finalVariables.putIfAbsent("region", config.getRegion());
        finalVariables.putIfAbsent("language", "zh");
        finalVariables.putIfAbsent("timeZone", "Asia/Shanghai");
        finalVariables.putIfAbsent("timeStamp", String.valueOf(Instant.now().getEpochSecond()));

        executor.submit(() -> {
            try {
                // 构建消息内容
                CreateChatRequest.Contents content = CreateChatRequest.Contents.builder()
                        .type("text")
                        .value(message)
                        .build();

                CreateChatRequest.Messages msg = CreateChatRequest.Messages.builder()
                        .role("user")
                        .contents(Collections.singletonList(content))
                        .build();

                CreateChatRequest request = CreateChatRequest.builder()
                        .regionId(config.getRegion())
                        .action("create")
                        .threadId(threadId)
                        .digitalEmployeeName(config.getEmployeeName())
                        .messages(Collections.singletonList(msg))
                        .variables(finalVariables)
                        .build();

                // 带重试能力的 SSE 流处理
                streamSSE(request, events);
            } catch (Exception e) {
                try {
                    events.put(ChatEvent.error(SDKException.chatFailed(e)));
                } catch (InterruptedException ignored) {}
            }
        });

        return events;
    }

    /**
     * 发送交互响应并恢复 SSE 对话
     * Send interactive response and resume SSE chat
     * 使用 action="interact"，无 messages 字段
     */
    public BlockingQueue<ChatEvent> interact(String threadId, String userInteractive,
            Map<String, Object> baseVariables) {
        BlockingQueue<ChatEvent> events = new LinkedBlockingQueue<>();

        final Map<String, Object> variables = baseVariables != null
                ? new HashMap<>(baseVariables) : new HashMap<>();
        variables.put("userInteractive", userInteractive);
        variables.putIfAbsent("workspace", config.getWorkspace());
        variables.putIfAbsent("region", config.getRegion());
        variables.putIfAbsent("language", "zh");
        variables.putIfAbsent("timeZone", "Asia/Shanghai");
        variables.putIfAbsent("timeStamp", String.valueOf(Instant.now().getEpochSecond()));

        executor.submit(() -> {
            try {
                CreateChatRequest request = CreateChatRequest.builder()
                        .regionId(config.getRegion())
                        .action("interact")
                        .threadId(threadId)
                        .digitalEmployeeName(config.getEmployeeName())
                        .variables(variables)
                        .build();

                ResponseIterable<CreateChatResponseBody> iterable =
                        client.createChatWithResponseIterable(request);

                Iterator<CreateChatResponseBody> iterator = iterable.iterator();
                while (iterator.hasNext()) {
                    CreateChatResponseBody body = iterator.next();
                    if (body != null) {
                        Map<String, Object> map = bodyToMap(body);
                        String jsonStr = objectMapper.writeValueAsString(map);
                        JsonNode jsonNode = objectMapper.readTree(jsonStr);
                        ChatEvent event = ChatEvent.fromResponse(jsonNode, jsonStr, 200);
                        events.put(event);
                        if (event.isDone()) {
                            return;
                        }
                    }
                }

                events.put(ChatEvent.done());
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                try {
                    events.put(ChatEvent.error(SDKException.cancelled()));
                } catch (InterruptedException ignored) {}
            } catch (Exception e) {
                try {
                    events.put(ChatEvent.error(SDKException.chatFailed(e)));
                } catch (InterruptedException ignored) {}
            }
        });

        return events;
    }

    /**
     * 启动带重试能力的 SSE 流处理（默认启用重试）
     * 编排层：外层重连循环，连接中断时自动重连并通过 timestamp 去重
     */
    private void streamSSE(CreateChatRequest req, BlockingQueue<ChatEvent> events) {
        RetryConfig cfg = config.getRetryConfig();
        if (cfg == null) {
            cfg = RetryConfig.getDefault();
        }
        RetryState state = new RetryState();
        while (true) {
            ConnectionOutcome outcome = streamOnce(req, events, state, cfg);
            if (outcome == ConnectionOutcome.DONE) {
                return; // stream_done，正常结束
            }
            if (outcome == ConnectionOutcome.FATAL) {
                return; // 取消或致命错误，错误已写入
            }
            // INTERRUPTED：需重连
            if (!prepareReconnect(events, state, cfg)) {
                return; // 超过最大重试或取消，错误已写入
            }
            req = RetrySupport.buildReconnectRequest(req);
        }
    }

    /**
     * 消费单次连接的事件流，返回本次连接的结束原因
     * 使用带超时的迭代实现空闲超时：后台线程消费 SSE 迭代器，主循环按 idleTimeout 轮询
     */
    private ConnectionOutcome streamOnce(CreateChatRequest req, BlockingQueue<ChatEvent> events,
            RetryState state, RetryConfig cfg) {
        BlockingQueue<Object> raw = new LinkedBlockingQueue<>();
        AtomicBoolean cancelled = new AtomicBoolean(false);
        long start = System.currentTimeMillis();

        Future<?> future = executor.submit(() -> {
            try {
                ResponseIterable<CreateChatResponseBody> iterable =
                        client.createChatWithResponseIterable(req);
                Iterator<CreateChatResponseBody> iterator = iterable.iterator();
                while (iterator.hasNext()) {
                    CreateChatResponseBody body = iterator.next();
                    if (cancelled.get()) {
                        return;
                    }
                    if (body != null) {
                        raw.offer(body);
                    }
                }
                raw.offer(STREAM_END);
            } catch (Throwable t) {
                if (!cancelled.get()) {
                    raw.offer(new StreamError(t));
                }
            }
        });

        try {
            while (true) {
                Object item = raw.poll(cfg.getIdleTimeoutMs(), TimeUnit.MILLISECONDS);
                if (item == null) {
                    // 空闲超时，未收到消息 → 连接中断
                    cancelled.set(true);
                    future.cancel(true);
                    System.out.println("连接中断，中断原因：空闲超时，未收到消息");
                    return ConnectionOutcome.INTERRUPTED;
                }
                if (item == STREAM_END) {
                    // 通道关闭且未收到 stream_done → 连接中断
                    System.out.println("连接中断，中断原因：通道关闭且未收到 stream_done");
                    return ConnectionOutcome.INTERRUPTED;
                }
                if (item instanceof StreamError) {
                    // 区分队列中的 null/关闭标记与实际异常
                    Throwable err = ((StreamError) item).cause;
                    if (err == null) {
                        continue; // 无实际异常 → 忽略，继续循环
                    }
                    // 非 stream_done 的任何错误都视为连接中断，触发重连
                    System.err.println("SSE 连接错误: " + err + "，准备重连...");
                    System.out.println("连接中断，中断原因：SSE连接错误");
                    return ConnectionOutcome.INTERRUPTED;
                }

                // 正常事件
                CreateChatResponseBody body = (CreateChatResponseBody) item;
                Map<String, Object> map = bodyToMap(body);
                String jsonStr = objectMapper.writeValueAsString(map);
                JsonNode jsonNode = objectMapper.readTree(jsonStr);
                ChatEvent event = ChatEvent.fromResponse(jsonNode, jsonStr, 200);

                if (RetrySupport.isStreamDoneEvent(event)) { // stream_done 是唯一正常结束标志
                    event.setDone(true);
                    events.put(event);
                    cancelled.set(true);
                    future.cancel(true);
                    return ConnectionOutcome.DONE;
                }

                if (forwardEvent(event, state, events)) {
                    state.retryCount = 0;
                }

                if (config.isSimulateNetworkError()) { // 模拟断连（转发后触发）
                    if (System.currentTimeMillis() - start > 5000) {
                        config.setSimulateNetworkError(false);
                        System.err.println("模拟网络断连，触发重连...");
                        cancelled.set(true);
                        future.cancel(true);
                        return ConnectionOutcome.INTERRUPTED;
                    }
                }
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            cancelled.set(true);
            future.cancel(true);
            try {
                events.put(ChatEvent.error(SDKException.cancelled()));
            } catch (InterruptedException ignored) {}
            return ConnectionOutcome.FATAL;
        } catch (Exception e) {
            cancelled.set(true);
            future.cancel(true);
            try {
                events.put(ChatEvent.error(SDKException.chatFailed(e)));
            } catch (InterruptedException ignored) {}
            return ConnectionOutcome.FATAL;
        }
    }

    /**
     * 去重转发普通事件，返回是否实际转发了消息
     * Deduplicate and forward normal events; return whether a message was forwarded
     */
    private boolean forwardEvent(ChatEvent event, RetryState state, BlockingQueue<ChatEvent> events)
            throws InterruptedException {
        String ts = RetrySupport.extractNewestTimestamp(event, state.lastTimestamp);

        if (state.inDedupeWindow) {
            if (ts.isEmpty()) {
                return false; // 重复消息，跳过
            }
            state.inDedupeWindow = false; // 收到新消息，退出去重窗口
        }

        if (!ts.isEmpty()) {
            state.lastTimestamp = ts;
        }
        events.put(event);
        return true;
    }

    /**
     * 执行退避并判定是否继续重试；返回 false 表示应终止
     * Perform backoff and decide whether to continue retrying; false means stop
     */
    private boolean prepareReconnect(BlockingQueue<ChatEvent> events, RetryState state, RetryConfig cfg) {
        if (state.retryCount >= cfg.getMaxRetries()) {
            try {
                events.put(ChatEvent.error(new SDKException(ErrorCode.NETWORK_ERROR,
                        "超过最大重试次数 " + cfg.getMaxRetries() + " 次，连接中断")));
            } catch (InterruptedException ignored) {}
            return false;
        }
        state.retryCount++;
        long backoff = RetrySupport.calculateBackoff(state.retryCount, cfg);
        System.err.printf("连接中断，%dms 后重试 (第 %d/%d 次)%n",
                backoff, state.retryCount, cfg.getMaxRetries());
        try {
            Thread.sleep(backoff);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            try {
                events.put(ChatEvent.error(SDKException.cancelled()));
            } catch (InterruptedException ignored) {}
            return false;
        }
        state.inDedupeWindow = true; // 进入去重窗口
        return true;
    }

    /** 后台迭代结束哨兵 / Sentinel for iterator completion */
    private static final Object STREAM_END = new Object();

    /** 后台迭代异常包装 / Wrapper for iterator error */
    private static final class StreamError {
        final Throwable cause;
        StreamError(Throwable cause) {
            this.cause = cause;
        }
    }

    /**
     * 将 CreateChatResponseBody 转为 Map
     * Convert CreateChatResponseBody to Map for JSON serialization
     */
    private Map<String, Object> bodyToMap(CreateChatResponseBody body) {
        Map<String, Object> map = new LinkedHashMap<>();
        if (body.getRequestId() != null) {
            map.put("requestId", body.getRequestId());
        }
        if (body.getTraceId() != null) {
            map.put("traceId", body.getTraceId());
        }
        if (body.getMessages() != null) {
            List<Map<String, Object>> msgList = new ArrayList<>();
            for (CreateChatResponseBody.Messages msg : body.getMessages()) {
                Map<String, Object> msgMap = new LinkedHashMap<>();
                if (msg.getRole() != null) msgMap.put("role", msg.getRole());
                if (msg.getType() != null) msgMap.put("type", msg.getType());
                if (msg.getCallId() != null) msgMap.put("callId", msg.getCallId());
                if (msg.getParentCallId() != null) msgMap.put("parentCallId", msg.getParentCallId());
                if (msg.getSeq() != null) msgMap.put("seq", msg.getSeq());
                if (msg.getTimestamp() != null) msgMap.put("timestamp", msg.getTimestamp());
                if (msg.getVersion() != null) msgMap.put("version", msg.getVersion());
                if (msg.getDetail() != null) msgMap.put("detail", msg.getDetail());
                if (msg.getContents() != null) msgMap.put("contents", msg.getContents());
                if (msg.getArtifacts() != null) msgMap.put("artifacts", msg.getArtifacts());
                if (msg.getTools() != null) msgMap.put("tools", msg.getTools());
                if (msg.getAgents() != null) msgMap.put("agents", msg.getAgents());
                if (msg.getEvents() != null) msgMap.put("events", msg.getEvents());
                msgList.add(msgMap);
            }
            map.put("messages", msgList);
        }
        return map;
    }

    /**
     * 带超时的对话
     * Chat with timeout
     */
    public BlockingQueue<ChatEvent> chatWithTimeout(String threadId, String message, java.time.Duration timeout) {
        BlockingQueue<ChatEvent> events = chat(threadId, message);
        BlockingQueue<ChatEvent> timedEvents = new LinkedBlockingQueue<>();

        executor.submit(() -> {
            long deadline = System.currentTimeMillis() + timeout.toMillis();
            try {
                while (true) {
                    long remaining = deadline - System.currentTimeMillis();
                    if (remaining <= 0) {
                        timedEvents.put(ChatEvent.error(SDKException.timeout(timeout.toString())));
                        return;
                    }

                    ChatEvent event = events.poll(remaining, TimeUnit.MILLISECONDS);
                    if (event == null) {
                        timedEvents.put(ChatEvent.error(SDKException.timeout(timeout.toString())));
                        return;
                    }

                    timedEvents.put(event);
                    if (event.isDone() || event.hasError()) {
                        return;
                    }
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                try {
                    timedEvents.put(ChatEvent.error(SDKException.cancelled()));
                } catch (InterruptedException ignored) {}
            }
        });

        return timedEvents;
    }

    /**
     * 列出会话
     * List threads
     */
    public ListThreadsResult listThreads(int pageSize) throws SDKException {
        try {
            if (pageSize <= 0) pageSize = 20;
            if (pageSize > 100) pageSize = 100;

            ListThreadsRequest request = ListThreadsRequest.builder()
                    .name(config.getEmployeeName())
                    .maxResults((long) pageSize)
                    .build();

            ListThreadsResponse response = client.listThreads(request).get();

            if (response.getBody() == null) {
                throw new SDKException(ErrorCode.PARSE_ERROR, "无效响应: 响应体为空")
                        .withSuggestion("请稍后重试");
            }

            List<ThreadInfo> threads = new ArrayList<>();
            ListThreadsResponseBody body = response.getBody();
            if (body.getThreads() != null) {
                for (ListThreadsResponseBody.Threads t : body.getThreads()) {
                    threads.add(new ThreadInfo(
                            t.getThreadId(),
                            t.getTitle(),
                            t.getStatus(),
                            t.getCreateTime(),
                            t.getUpdateTime()
                    ));
                }
            }

            long total = body.getTotal() != null ? body.getTotal() : 0;
            return new ListThreadsResult(threads, total);
        } catch (SDKException e) {
            throw e;
        } catch (Exception e) {
            throw new SDKException(ErrorCode.API_ERROR, "获取会话列表失败", e)
                    .withSuggestion("请检查网络连接和 API 权限");
        }
    }

    /**
     * 获取会话详情
     * Get thread details
     */
    public ThreadInfo getThread(String threadId) throws SDKException {
        validateThreadId(threadId);

        try {
            GetThreadRequest request = GetThreadRequest.builder()
                    .name(config.getEmployeeName())
                    .threadId(threadId)
                    .build();

            GetThreadResponse response = client.getThread(request).get();

            if (response.getBody() == null) {
                throw new SDKException(ErrorCode.PARSE_ERROR, "无效响应: 响应体为空")
                        .withContext("threadId", threadId)
                        .withSuggestion("请稍后重试");
            }

            GetThreadResponseBody body = response.getBody();
            return new ThreadInfo(
                    body.getThreadId(),
                    body.getTitle(),
                    body.getStatus(),
                    body.getCreateTime(),
                    body.getUpdateTime()
            );
        } catch (SDKException e) {
            throw e;
        } catch (Exception e) {
            if (isThreadNotFoundError(e)) {
                throw SDKException.threadNotFound(threadId);
            }
            throw new SDKException(ErrorCode.API_ERROR, "获取会话详情失败: " + threadId, e)
                    .withContext("threadId", threadId)
                    .withSuggestion("请检查会话 ID 是否正确");
        }
    }

    /**
     * 删除会话
     * Delete thread
     */
    public void deleteThread(String threadId) throws SDKException {
        validateThreadId(threadId);

        try {
            DeleteThreadRequest request = DeleteThreadRequest.builder()
                    .name(config.getEmployeeName())
                    .threadId(threadId)
                    .build();

            client.deleteThread(request).get();
        } catch (Exception e) {
            if (isThreadNotFoundError(e)) {
                throw SDKException.threadNotFound(threadId);
            }
            throw new SDKException(ErrorCode.API_ERROR, "删除会话失败: " + threadId, e)
                    .withContext("threadId", threadId)
                    .withSuggestion("请检查会话 ID 是否正确");
        }
    }

    /**
     * 获取会话消息
     * Get thread messages
     */
    public List<ThreadMessage> getThreadData(String threadId, int limit) throws SDKException {
        validateThreadId(threadId);

        try {
            if (limit <= 0) limit = 50;
            if (limit > 100) limit = 100;

            GetThreadDataRequest request = GetThreadDataRequest.builder()
                    .name(config.getEmployeeName())
                    .threadId(threadId)
                    .maxResults((long) limit)
                    .build();

            GetThreadDataResponse response = client.getThreadData(request).get();

            if (response.getBody() == null) {
                throw new SDKException(ErrorCode.PARSE_ERROR, "无效响应: 响应体为空")
                        .withContext("threadId", threadId)
                        .withSuggestion("请稍后重试");
            }

            GetThreadDataResponseBody body = response.getBody();

            // 策略：优先使用 system Result，而不是 assistant 流式消息
            Map<String, ThreadMessage> messageMap = new LinkedHashMap<>();

            // 检查是否存在 system Result
            boolean hasSystemResult = false;
            if (body.getData() != null) {
                outer:
                for (GetThreadDataResponseBody.Data data : body.getData()) {
                    if (data.getMessages() != null) {
                        for (GetThreadDataResponseBody.Messages msg : data.getMessages()) {
                            if ("system".equals(msg.getRole()) && msg.getArtifacts() != null) {
                                for (Map<String, ?> artifact : msg.getArtifacts()) {
                                    if (artifact != null && "Result".equals(artifact.get("name"))) {
                                        hasSystemResult = true;
                                        break outer;
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // 处理消息
            if (body.getData() != null) {
                for (GetThreadDataResponseBody.Data data : body.getData()) {
                    if (data.getMessages() != null) {
                        for (GetThreadDataResponseBody.Messages msg : data.getMessages()) {
                            String role = msg.getRole() != null ? msg.getRole() : "";
                            String timestamp = msg.getTimestamp() != null ? msg.getTimestamp() : "";

                            // 如果存在 system Result，跳过 assistant 流式消息
                            if ("assistant".equals(role) && hasSystemResult) {
                                continue;
                            }

                            String key;
                            if ("user".equals(role)) {
                                key = "user-" + timestamp;
                            } else if ("system".equals(role)) {
                                key = "system-" + timestamp;
                            } else {
                                String callId = msg.getCallId() != null ? msg.getCallId() : "";
                                key = "assistant-" + callId;
                            }

                            String content = extractMessageContent(msg);
                            if (content == null || content.isEmpty()) {
                                continue;
                            }

                            if (messageMap.containsKey(key)) {
                                ThreadMessage existing = messageMap.get(key);
                                messageMap.put(key, new ThreadMessage(
                                        existing.getRole(),
                                        existing.getContent() + content,
                                        existing.getTimestamp()
                                ));
                            } else {
                                String displayRole = "system".equals(role) ? "assistant" : role;
                                messageMap.put(key, new ThreadMessage(displayRole, content, timestamp));
                            }
                        }
                    }
                }
            }

            return new ArrayList<>(messageMap.values());
        } catch (SDKException e) {
            throw e;
        } catch (Exception e) {
            if (isThreadNotFoundError(e)) {
                throw SDKException.threadNotFound(threadId);
            }
            throw new SDKException(ErrorCode.API_ERROR, "获取会话消息失败: " + threadId, e)
                    .withContext("threadId", threadId)
                    .withSuggestion("请检查会话 ID 是否正确");
        }
    }

    private void validateThreadId(String threadId) throws SDKException {
        if (threadId == null || threadId.isEmpty()) {
            throw new SDKException(ErrorCode.CONFIG_INVALID, "会话 ID 不能为空")
                    .withContext("threadId", threadId)
                    .withSuggestion("请提供有效的会话 ID");
        }
        if (threadId.contains(" ") || threadId.contains("\t") || threadId.contains("\n")) {
            throw new SDKException(ErrorCode.CONFIG_INVALID, "会话 ID 包含非法字符: " + threadId)
                    .withContext("threadId", threadId)
                    .withSuggestion("会话 ID 不能包含空白字符");
        }
    }

    private boolean isThreadNotFoundError(Exception e) {
        if (e == null) return false;
        String errStr = e.getMessage();
        if (errStr == null) return false;
        return errStr.contains("NotFound") || errStr.contains("not found") ||
               errStr.contains("NOT_FOUND") || errStr.contains("ThreadNotFound") ||
               errStr.contains("InvalidThreadId") || errStr.contains("does not exist");
    }

    @SuppressWarnings("unchecked")
    private String extractMessageContent(GetThreadDataResponseBody.Messages msg) {
        if (msg == null) return "";

        // 1. 尝试从 contents 提取（流式文本块）
        if (msg.getContents() != null) {
            StringBuilder result = new StringBuilder();
            for (Map<String, ?> content : msg.getContents()) {
                if (content != null) {
                    Object type = content.get("type");
                    Object value = content.get("value");
                    if ("text".equals(type) && value != null) {
                        result.append(value.toString());
                    }
                }
            }
            if (result.length() > 0) {
                return result.toString();
            }
        }

        // 2. 尝试从 artifacts 提取（最终结果）
        if (msg.getArtifacts() != null) {
            for (Map<String, ?> artifact : msg.getArtifacts()) {
                if (artifact != null && "Result".equals(artifact.get("name"))) {
                    Object partsObj = artifact.get("parts");
                    if (partsObj instanceof List) {
                        List<?> parts = (List<?>) partsObj;
                        StringBuilder textParts = new StringBuilder();
                        for (Object partObj : parts) {
                            if (partObj instanceof Map) {
                                Map<?, ?> part = (Map<?, ?>) partObj;
                                if ("text".equals(part.get("kind")) && part.get("text") != null) {
                                    textParts.append(part.get("text").toString());
                                }
                            }
                        }
                        if (textParts.length() > 0) {
                            return textParts.toString();
                        }
                    }
                }
            }
        }

        return "";
    }

    public void shutdown() {
        executor.shutdown();
        try {
            client.close();
        } catch (Exception ignored) {}
    }

    @Override
    public void close() {
        shutdown();
    }

    /**
     * 列出会话结果
     * List threads result
     */
    public static class ListThreadsResult {
        private final List<ThreadInfo> threads;
        private final long total;

        public ListThreadsResult(List<ThreadInfo> threads, long total) {
            this.threads = threads;
            this.total = total;
        }

        public List<ThreadInfo> getThreads() {
            return threads;
        }

        public long getTotal() {
            return total;
        }
    }
}
