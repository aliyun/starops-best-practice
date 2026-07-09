package com.alibaba.cloud.starops.samples.examples;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.BlockingQueue;

import com.alibaba.cloud.starops.samples.client.AgentClient;
import com.alibaba.cloud.starops.samples.client.ChatEvent;
import com.alibaba.cloud.starops.samples.client.Config;
import com.alibaba.cloud.starops.samples.client.InteractiveHandler;
import com.alibaba.cloud.starops.samples.client.InteractiveResponse;
import com.alibaba.cloud.starops.samples.client.SDKException;
import com.alibaba.cloud.starops.samples.client.SimplePrinter;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * 交互式对话示例
 * Interactive chat example
 *
 * Usage: mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.Chat"
 */
public class Chat {
    public static void main(String[] args) {
        System.out.println("🚀 VibeOps Chat (Java)");
        System.out.println("=".repeat(60));

        boolean simulateError = false;
        for (String arg : args) {
            if ("-simulate-error".equals(arg)) {
                simulateError = true;
            }
        }

        try {
            // Load configuration
            Config cfg = Config.loadFromEnv();
            System.out.printf("📋 Employee: %s%n%n", cfg.getEmployeeName());

            if (simulateError) {
                cfg.setSimulateNetworkError(true);
                System.out.println("⚠️  已启用网络断连模拟，将在收到首个事件后触发重试");
            }

            // Create client
            AgentClient client = new AgentClient(cfg);

            // Create thread
            System.out.println("📝 创建会话...");
            String threadId = client.createThread();
            System.out.printf("✅ ThreadID: %s%n%n", threadId);

            // Create printer
            SimplePrinter printer = new SimplePrinter();
            InteractiveHandler interactiveHandler = new InteractiveHandler(client, null);

            // Interactive loop
            BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
            while (true) {
                System.out.print("👤 请输入 (quit 退出): ");
                String input = reader.readLine();

                if (input == null) {
                    System.out.println("\n👋 再见!");
                    break;
                }

                input = input.trim();
                if (input.isEmpty()) {
                    continue;
                }
                if (input.equals("quit") || input.equals("exit")) {
                    System.out.println("👋 再见!");
                    break;
                }

                System.out.println("-".repeat(60));

                // Send message
                printer.reset();
                BlockingQueue<ChatEvent> events = client.chat(threadId, input);

                processChatEvents(events, printer, interactiveHandler, threadId);

                System.out.println();
                System.out.println("=".repeat(60));
                System.out.println();
            }

            client.shutdown();
        } catch (SDKException e) {
            System.out.printf("❌ 配置加载失败: %s%n", e.getMessage());
            System.out.println("\n请设置环境变量:");
            System.out.println("  VIBEOPS_ENDPOINT");
            System.out.println("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET");
            System.exit(1);
        } catch (Exception e) {
            System.out.printf("❌ 错误: %s%n", e.getMessage());
            System.exit(1);
        }
    }

    private static void processChatEvents(
            BlockingQueue<ChatEvent> events,
            SimplePrinter printer,
            InteractiveHandler handler,
            String threadId) throws InterruptedException {
        while (events != null) {
            ChatEvent event = events.take();

            if (event.hasError()) {
                System.out.printf("❌ 错误: %s%n", event.getError().getMessage());
                break;
            }

            // 正常输出（先输出）
            String text = printer.processEvent(event);
            if (!text.isEmpty()) {
                System.out.print(text);
            }

            // 检测交互事件（在输出之后）
            InteractiveResponse interactiveResp = extractChatInteractiveEvent(event, handler);
            if (interactiveResp != null) {
                Map<String, Object> variables = new HashMap<>();
                events = handler.resumeChat(threadId, interactiveResp, variables);
                continue;
            }

            if (event.isDone()) {
                break;
            }
        }
    }

    private static InteractiveResponse extractChatInteractiveEvent(ChatEvent event, InteractiveHandler handler) {
        if (event.getRawJson() == null || event.getRawJson().isEmpty()) {
            return null;
        }
        try {
            ObjectMapper mapper = new ObjectMapper();
            JsonNode root = mapper.readTree(event.getRawJson());
            JsonNode messages = root.get("messages");
            if (messages == null || !messages.isArray()) return null;

            for (JsonNode msg : messages) {
                JsonNode events = msg.get("events");
                if (events == null || !events.isArray()) continue;

                for (JsonNode evt : events) {
                    if (InteractiveHandler.isInteractiveEvent(evt)) {
                        String callId = msg.has("callId") ? msg.get("callId").asText() : "";
                        return handler.handleEvent(evt, callId);
                    }
                }
            }
        } catch (Exception e) {
            System.out.printf("⚠️ 交互事件解析失败: %s%n", e.getMessage());
        }
        return null;
    }
}