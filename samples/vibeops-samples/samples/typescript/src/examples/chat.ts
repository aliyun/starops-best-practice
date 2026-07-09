#!/usr/bin/env tsx
/**
 * 交互式对话示例
 * Interactive chat example
 *
 * Usage: npm run chat
 */

import * as readline from 'readline';
import { loadConfigFromEnv, AgentClient, SimplePrinter, InteractiveHandler, InteractiveResponse, SDKException } from '../client/index.js';

function prompt(rl: readline.Interface, question: string): Promise<string> {
  return new Promise((resolve) => {
    rl.question(question, resolve);
  });
}

async function extractInteractiveEvent(
  event: { rawJson?: string },
  handler: InteractiveHandler
): Promise<InteractiveResponse | null> {
  if (!event.rawJson) return null;
  try {
    const body = JSON.parse(event.rawJson);
    for (const msg of body.messages || []) {
      for (const evt of msg.events || []) {
        if (InteractiveHandler.isInteractiveEvent(evt)) {
          const callId = msg.callId || '';
          return await handler.handleEvent(evt, callId);
        }
      }
    }
  } catch (e) {
    console.log(`⚠️ 交互事件解析失败: ${(e as Error).message}`);
  }
  return null;
}

async function main() {
  console.log('🚀 VibeOps Chat (TypeScript)');
  console.log('='.repeat(60));

  try {
    // Load configuration
    const cfg = await loadConfigFromEnv();
    console.log(`📋 Employee: ${cfg.employeeName}\n`);

    if (process.argv.slice(2).includes('-simulate-error')) {
      cfg.simulateNetworkError = true;
      console.log('⚠️  已启用网络断连模拟，将在收到首个事件后触发重试');
    }

  

    // Create client
    const client = new AgentClient(cfg);

    // Create thread
    console.log('📝 创建会话...');
    const threadId = await client.createThread();
    console.log(`✅ ThreadID: ${threadId}\n`);

    // Create printer
    const printer = new SimplePrinter();
    const interactiveHandler = new InteractiveHandler(client);

    // Interactive loop
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    while (true) {
      const input = (await prompt(rl, '👤 请输入 (quit 退出): ')).trim();

      if (!input) continue;
      if (input === 'quit' || input === 'exit') {
        console.log('👋 再见!');
        break;
      }

      console.log('-'.repeat(60));

      printer.reset();
      try {
        let events = client.chat(threadId, input);
        while (events) {
          let shouldContinue = false;
          for await (const event of events) {
            if (event.error) {
              console.log(`❌ 错误: ${event.error.message}`);
              continue;
            }

            // 正常输出（先输出）
            const text = printer.processEvent(event);
            if (text) process.stdout.write(text);

            // 检测交互事件（在输出之后）
            const interactiveResp = await extractInteractiveEvent(event, interactiveHandler);
            if (interactiveResp) {
              events = interactiveHandler.resumeChat(threadId, interactiveResp);
              shouldContinue = true;
              break;
            }

            if (event.isDone) {
              shouldContinue = false;
              break;
            }
          }
          if (!shouldContinue) break;
        }
      } catch (chatError) {
        console.log(`❌ 对话异常: ${(chatError as Error).message}`);
      }

      console.log();
      console.log('='.repeat(60));
      console.log();
    }

    rl.close();
  } catch (e) {
    if (e instanceof SDKException) {
      console.log(`❌ 配置加载失败: ${e}`);
      console.log('\n请设置环境变量:');
      console.log('  VIBEOPS_ENDPOINT');
      console.log('  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET');
    } else {
      console.log(`❌ 错误: ${(e as Error).message}`);
    }
    process.exit(1);
  }
}

main();