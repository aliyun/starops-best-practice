package com.alibaba.cloud.starops.samples.examples;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.atomic.AtomicBoolean;

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
        System.out.println("🚀 STAROps Chat (Java 8)");
        System.out.println(repeatStr("=", 60));

        boolean simulateError = false;
        String modelFlag = null;
        boolean listModels = false;

        for (int i = 0; i < args.length; i++) {
            String arg = args[i];
            if ("-simulate-error".equals(arg)) {
                simulateError = true;
            } else if ("--model".equals(arg) && i + 1 < args.length) {
                modelFlag = args[++i];
            } else if (arg.startsWith("--model=")) {
                modelFlag = arg.substring("--model=".length());
            } else if ("--list-models".equals(arg)) {
                listModels = true;
            }
        }

        // 加载模型配置（不依赖凭据）
        ModelsConfig modelsCfg = ModelsConfig.loadModels(ModelsConfig.findConfigPath());

        // --list-models: 短路，打印后退出（不需要凭据）
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

            // 解析 --model 参数
            Map<String, String> threadAttrs = null;
            if (modelFlag != null && !modelFlag.isEmpty()) {
                String[] parsed = ModelsConfig.parseModelFlag(modelFlag);
                String provider = parsed[0];
                String modelId = parsed[1];
                if (modelsCfg != null && !ModelsConfig.validateModel(modelsCfg, provider, modelId)) {
                    System.out.printf("❌ 未知模型: %s:%s，可用 --list-models 查看可选值%n", provider, modelId);
                    System.exit(1);
                    return;
                }
                threadAttrs = new HashMap<String, String>();
                threadAttrs.put("model", ModelsConfig.buildModelJson(provider, modelId));
                System.out.printf("🤖 已指定模型: %s:%s%n", provider, modelId);
            }

            // Create thread
            System.out.println("📝 创建会话...");
            String threadId = client.createThread(threadAttrs);
            System.out.printf("✅ ThreadID: %s%n%n", threadId);

            // 捕获中断信号，发送 stop 请求
            final AgentClient stopClient = client;
            final String stopThreadId = threadId;
            final AtomicBoolean finished = new AtomicBoolean(false);
            Runtime.getRuntime().addShutdownHook(new Thread(new Runnable() {
                @Override
                public void run() {
                    // 正常完成后跳过 stop；仅在 Ctrl+C / 异常中断时发送
                    if (finished.get()) return;
                    System.out.println("\n⏹️  正在停止对话...");
                    stopClient.stop(stopThreadId, null);
                }
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
                        if (choice == null) {
                            continue;
                        }
                        try {
                            int idx = Integer.parseInt(choice.trim());
                            String[] selected = ModelsConfig.getModelByIndex(modelsCfg, idx - 1);
                            Map<String, String> attrs = new HashMap<String, String>();
                            attrs.put("model", ModelsConfig.buildModelJson(selected[0], selected[1]));
                            client.updateThread(threadId, attrs);
                            System.out.println("✅ 更新模型成功");
                        } catch (NumberFormatException e) {
                            System.out.println("❌ 无效输入");
                        } catch (IllegalArgumentException e) {
                            System.out.println("❌ " + e.getMessage());
                        } catch (SDKException e) {
                            System.out.printf("❌ 更新失败: %s%n", e.getMessage());
                        }
                        continue;
                    default:
                        break;
                }

                System.out.println(repeatStr("-", 60));

                // Send message
                printer.reset();
                BlockingQueue<ChatEvent> events = client.chat(threadId, input);

                processChatEvents(events, printer, interactiveHandler, threadId);

                System.out.println();
                System.out.println(repeatStr("=", 60));
                System.out.println();
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
                Map<String, Object> variables = new HashMap<String, Object>();
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

    private static String repeatStr(String s, int count) {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < count; i++) {
            sb.append(s);
        }
        return sb.toString();
    }
}
