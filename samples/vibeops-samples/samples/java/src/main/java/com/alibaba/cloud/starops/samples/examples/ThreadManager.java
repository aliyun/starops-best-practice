package com.alibaba.cloud.starops.samples.examples;

import com.alibaba.cloud.starops.samples.client.*;

import java.util.List;

/**
 * 会话管理工具
 * Thread management tool
 *
 * Usage:
 *   mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ThreadManager" -Dexec.args="list"
 *   mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ThreadManager" -Dexec.args="get <thread-id>"
 *   mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ThreadManager" -Dexec.args="messages <thread-id>"
 *   mvn exec:java -Dexec.mainClass="com.alibaba.cloud.starops.samples.examples.ThreadManager" -Dexec.args="delete <thread-id>"
 */
public class ThreadManager {
    public static void main(String[] args) {
        if (args.length < 1) {
            printUsage();
            System.exit(1);
        }

        String command = args[0];

        try {
            // Load configuration
            Config cfg = Config.loadFromEnv();

            // Create client
            AgentClient client = new AgentClient(cfg);

            switch (command) {
                case "list":
                    listThreads(client);
                    break;
                case "get":
                    if (args.length < 2) {
                        System.out.println("❌ 请指定 thread-id");
                        System.exit(1);
                    }
                    getThread(client, args[1]);
                    break;
                case "messages":
                    if (args.length < 2) {
                        System.out.println("❌ 请指定 thread-id");
                        System.exit(1);
                    }
                    listMessages(client, args[1]);
                    break;
                case "delete":
                    if (args.length < 2) {
                        System.out.println("❌ 请指定 thread-id");
                        System.exit(1);
                    }
                    deleteThread(client, args[1]);
                    break;
                default:
                    System.out.printf("❌ 未知命令: %s%n", command);
                    printUsage();
                    System.exit(1);
            }

            client.shutdown();
        } catch (SDKException e) {
            System.out.printf("❌ 错误: %s%n", e.getMessage());
            System.exit(1);
        }
    }

    private static void listThreads(AgentClient client) throws SDKException {
        System.out.println("📋 会话列表");
        System.out.println("-".repeat(100));

        AgentClient.ListThreadsResult result = client.listThreads(20);
        System.out.printf("共 %d 个会话:%n%n", result.getTotal());

        if (result.getThreads().isEmpty()) {
            System.out.println("  (无会话)");
            return;
        }

        System.out.printf("%-40s %-25s %-10s %s%n", "Thread ID", "标题", "状态", "创建时间");
        System.out.println("-".repeat(100));

        for (ThreadInfo t : result.getThreads()) {
            String title = t.getTitle();
            if (title != null && title.length() > 23) {
                title = title.substring(0, 23) + "...";
            }
            System.out.printf("%-40s %-25s %-10s %s%n",
                    t.getThreadId(),
                    title != null ? title : "",
                    t.getStatus() != null ? t.getStatus() : "",
                    t.getCreateTime() != null ? t.getCreateTime() : "");
        }
    }

    private static void getThread(AgentClient client, String threadId) throws SDKException {
        System.out.printf("📋 会话详情: %s%n", threadId);
        System.out.println("-".repeat(60));

        ThreadInfo detail = client.getThread(threadId);

        System.out.printf("  Thread ID: %s%n", detail.getThreadId());
        System.out.printf("  标题: %s%n", detail.getTitle());
        System.out.printf("  状态: %s%n", detail.getStatus());
        System.out.printf("  创建时间: %s%n", detail.getCreateTime());
        System.out.printf("  更新时间: %s%n", detail.getUpdateTime());
    }

    private static void listMessages(AgentClient client, String threadId) throws SDKException {
        System.out.printf("💬 会话消息: %s%n", threadId);
        System.out.println("-".repeat(80));

        List<ThreadMessage> messages = client.getThreadData(threadId, 50);
        System.out.printf("共 %d 条消息:%n%n", messages.size());

        for (int i = 0; i < messages.size(); i++) {
            ThreadMessage m = messages.get(i);
            String roleIcon = "👤";
            if ("assistant".equals(m.getRole())) {
                roleIcon = "🤖";
            } else if ("system".equals(m.getRole())) {
                roleIcon = "⚙️";
            }

            String content = m.getContent();
            if (content != null && content.length() > 100) {
                content = content.substring(0, 100) + "...";
            }

            System.out.printf("[%d] %s %s%n", i + 1, roleIcon, m.getRole());
            if (content != null && !content.isEmpty()) {
                System.out.printf("    %s%n", content);
            }
            System.out.printf("    时间戳: %s%n%n", m.getTimestamp());
        }
    }

    private static void deleteThread(AgentClient client, String threadId) throws SDKException {
        System.out.printf("🗑️ 删除会话: %s%n", threadId);

        client.deleteThread(threadId);
        System.out.println("✅ 会话已删除");
    }

    private static void printUsage() {
        System.out.println("会话管理工具 - STAROps SDK 接口演示 (Java)");
        System.out.println("=".repeat(50));
        System.out.println("\n用法:");
        System.out.println("  <command> [args]");
        System.out.println("\n命令:");
        System.out.println("  list                  列出所有会话");
        System.out.println("  get <thread-id>       获取会话详情");
        System.out.println("  messages <thread-id>  列出会话消息");
        System.out.println("  delete <thread-id>    删除会话");
    }
}
