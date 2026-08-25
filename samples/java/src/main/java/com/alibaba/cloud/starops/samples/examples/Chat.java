package com.alibaba.cloud.starops.samples.examples;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.BlockingQueue;

import com.alibaba.cloud.starops.samples.client.AgentClient;
import com.alibaba.cloud.starops.samples.client.ModelsConfig;
import com.alibaba.cloud.starops.samples.types.ChatEvent;
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
        System.out.println("🚀 STAROps Chat (Java)");
        System.out.println("=".repeat(60));

        boolean simulateError = false;
        String modelFlag = null;
        boolean listModels = false;
        for (int i = 0; i < args.length; i++) {
            switch (args[i]) {
                case "-simulate-error":
                case "--simulate-error":
                    simulateError = true;
                    break;
                case "--model":
                    if (i + 1 < args.length) {
                        modelFlag = args[++i];
                    }
                    break;
                case "--list-models":
                    listModels = true;
                    break;
                default:
                    if (args[i].startsWith("--model=")) {
                        modelFlag = args[i].substring("--model=".length());
                    }
                    break;
            }
        }

        // 加载模型配置（不依赖凭据）
        ModelsConfig modelsCfg = ModelsConfig.loadModels(ModelsConfig.findConfigPath());

        if (listModels) {
            if (modelsCfg == null) {
                System.out.println("⚠️  模型配置未加载");
            } else {
                System.out.println(ModelsConfig.listModelFlags(modelsCfg));
            }
            return;
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
            Map<String, String> threadAttrs = null;
            if (modelFlag != null) {
                String[] parsed = ModelsConfig.parseModelFlag(modelFlag);
                if (parsed == null) {
                    System.out.printf("❌ 解析 --model 失败: 格式应为 provider:modelId，实际: %s%n", modelFlag);
                    System.exit(1);
                }
                String provider = parsed[0];
                String modelId = parsed[1];
                if (modelsCfg != null && !ModelsConfig.validateModel(modelsCfg, provider, modelId)) {
                    System.out.printf("❌ 未知模型: %s:%s，可用 --list-models 查看可选值%n", provider, modelId);
                    System.exit(1);
                }
                threadAttrs = new HashMap<>();
                threadAttrs.put("model", ModelsConfig.buildModelJson(provider, modelId));
                System.out.printf("🤖 已指定模型: %s:%s%n", provider, modelId);
            }
            String threadId = client.createThread(threadAttrs);
            System.out.printf("✅ ThreadID: %s%n%n", threadId);

            // 捕获中断信号，发送 stop 请求
            final AgentClient stopClient = client;
            final String stopThreadId = threadId;
            final AtomicBoolean finished = new AtomicBoolean(false);
            Runtime.getRuntime().addShutdownHook(new Thread(() -> {
                // 正常完成后跳过 stop；仅在 Ctrl+C / 异常中断时发送
                if (finished.get()) return;
                System.out.println("\n⏹️  正在停止对话...");
                stopClient.stop(stopThreadId, null);
            }));

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
                switch (input) {
                    case "":
                        continue;
                    case "quit":
                    case "exit":
                        System.out.println("👋 再见!");
                        finished.set(true);
                        client.shutdown();
                        return;
                    case "/model":
                        if (modelsCfg == null) {
                            System.out.println("⚠️  模型配置未加载");
                            continue;
                        }
                        System.out.println(ModelsConfig.displayMenu(modelsCfg));
                        System.out.print("请输入序号选择模型: ");
                        String choice = reader.readLine();
                        if (choice == null) continue;
                        int idx;
                        try {
                            idx = Integer.parseInt(choice.trim());
                        } catch (NumberFormatException e) {
                            System.out.println("❌ 无效输入");
                            continue;
                        }
                        String[] modelInfo = ModelsConfig.getModelByIndex(modelsCfg, idx - 1);
                        if (modelInfo == null) {
                            System.out.printf("❌ model index %d out of range%n", idx);
                            continue;
                        }
                        try {
                            Map<String, String> attrs = new HashMap<>();
                            attrs.put("model", ModelsConfig.buildModelJson(modelInfo[0], modelInfo[1]));
                            client.updateThread(threadId, attrs);
                            System.out.println("✅ 更新模型成功");
                        } catch (SDKException e) {
                            System.out.printf("❌ 更新失败: %s%n", e.getMessage());
                        }
                        continue;
                    default:
                        System.out.println("-".repeat(60));

                        // Send message
                        printer.reset();
                        BlockingQueue<ChatEvent> events = client.chat(threadId, input);

                        processChatEvents(events, printer, interactiveHandler, threadId);

                        System.out.println();
                        System.out.println("=".repeat(60));
                        System.out.println();
                        break;
                }
            }

            // 标记正常退出，shutdown hook 跳过 stop
            finished.set(true);
            client.shutdown();
        } catch (SDKException e) {
            System.out.printf("❌ 配置加载失败: %s%n", e.getMessage());
            System.out.println("\n请设置环境变量:");
            System.out.println("  STAROPS_ENDPOINT");
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