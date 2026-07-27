package com.alibaba.cloud.starops.samples.client;

import java.util.Map;
import java.util.concurrent.BlockingQueue;

import com.alibaba.cloud.starops.samples.types.ChatEvent;

/**
 * 对话客户端接口（InteractiveHandler 依赖的最小 API）
 * Chat client interface - minimal API surface used by InteractiveHandler
 */
public interface ChatClient {

    /**
     * 发送交互响应并恢复 SSE 对话
     * Send interactive response and resume SSE chat
     */
    BlockingQueue<ChatEvent> interact(String threadId, String userInteractive,
            Map<String, Object> baseVariables);
}
