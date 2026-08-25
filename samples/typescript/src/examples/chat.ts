#!/usr/bin/env tsx
/**
 * 交互式对话示例
 * Interactive chat example
 *
 * Usage: npm run chat [-- --model provider:modelId] [-- --list-models]
 */

import * as readline from 'readline';
import { loadConfigFromEnv, AgentClient, SimplePrinter, InteractiveHandler, InteractiveResponse, SDKException } from '../client/index.js';
import { loadModels, findConfigPath, parseModelFlag, validateModel, buildModelJSON, displayMenu, listModelFlags, getModelByIndex } from '../client/models.js';

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

/** 解析命令行参数 */
function parseArgs(argv: string[]): { simulateError: boolean; model: string; listModels: boolean } {
  const args = argv.slice(2);
  let simulateError = false;
  let model = '';
  let listModels = false;

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '-simulate-error':
      case '--simulate-error':
        simulateError = true;
        break;
      case '--model':
        model = args[++i] || '';
        break;
      case '--list-models':
        listModels = true;
        break;
    }
  }
  return { simulateError, model, listModels };
}

async function main() {
  console.log('🚀 STAROps Chat (TypeScript)');
  console.log('='.repeat(60));

  const opts = parseArgs(process.argv);

  // 加载模型配置（不依赖凭据）
  const modelsCfg = loadModels(findConfigPath());

  if (opts.listModels) {
    if (!modelsCfg) {
      console.log('⚠️  模型配置未加载');
    } else {
      console.log(listModelFlags(modelsCfg));
    }
    return;
  }

  try {
    // Load configuration
    const cfg = await loadConfigFromEnv();
    console.log(`📋 Employee: ${cfg.employeeName}\n`);

    if (opts.simulateError) {
      cfg.simulateNetworkError = true;
      console.log('⚠️  已启用网络断连模拟，将在收到首个事件后触发重试');
    }

    // Create client
    const client = new AgentClient(cfg);

    // Create thread
    console.log('📝 创建会话...');
    let threadAttrs: Record<string, string> | undefined;
    if (opts.model) {
      const { provider, modelId } = parseModelFlag(opts.model);
      if (modelsCfg && !validateModel(modelsCfg, provider, modelId)) {
        console.error(`❌ 未知模型: ${provider}:${modelId}，可用 --list-models 查看可选值`);
        process.exit(1);
      }
      threadAttrs = { model: buildModelJSON(provider, modelId) };
      console.log(`🤖 已指定模型: ${provider}:${modelId}`);
    }
    const threadId = await client.createThread(threadAttrs);
    console.log(`✅ ThreadID: ${threadId}\n`);

    // 捕获中断信号，发送 stop 请求
    const stopWithTimeout = async () => {
      console.log('\n⏹️  正在停止对话...');
      try {
        await Promise.race([
          client.stop(threadId),
          new Promise((_, reject) => setTimeout(() => reject(new Error('timeout')), 6000)),
        ]);
      } catch (e) {
        console.error(`⚠️  stop 请求超时或失败: ${(e as Error).message}`);
      }
      process.exit(0);
    };
    process.on('SIGINT', stopWithTimeout);
    process.on('SIGTERM', stopWithTimeout);

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

      switch (input) {
        case '':
          continue;
        case 'quit':
        case 'exit':
          console.log('👋 再见!');
          rl.close();
          return;
        case '/model': {
          if (!modelsCfg) {
            console.log('⚠️  模型配置未加载');
            continue;
          }
          console.log(displayMenu(modelsCfg));
          const choice = (await prompt(rl, '请输入序号选择模型: ')).trim();
          const idx = parseInt(choice, 10);
          if (isNaN(idx)) {
            console.log('❌ 无效输入');
            continue;
          }
          try {
            const { provider, modelId } = getModelByIndex(modelsCfg, idx - 1);
            await client.updateThread(threadId, { model: buildModelJSON(provider, modelId) });
            console.log('✅ 更新模型成功');
          } catch (e) {
            console.log(`❌ ${(e as Error).message}`);
          }
          continue;
        }
        default: {
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
      }
    }
  } catch (e) {
    if (e instanceof SDKException) {
      console.log(`❌ 配置加载失败: ${e}`);
      console.log('\n请设置环境变量:');
      console.log('  STAROPS_ENDPOINT');
      console.log('  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET');
    } else {
      console.log(`❌ 错误: ${(e as Error).message}`);
    }
    process.exit(1);
  }
}

main();
