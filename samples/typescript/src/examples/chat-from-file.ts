#!/usr/bin/env tsx
/**
 * 从文件加载请求示例
 * Load requests from file example
 *
 * Usage:
 *   npm run chat-from-file -- --file ../../sample-requests/entity.json
 *   npm run chat-from-file -- --dir ../../sample-requests/
 */

import * as fs from 'fs';
import * as path from 'path';
import { loadConfigFromEnv, AgentClient, SimplePrinter, EventPrinter, InteractiveHandler, InteractiveResponse, SDKException } from '../client/index.js';

interface Args {
  filePath?: string;
  dirPath?: string;
  simpleMode: boolean;
  outputDir: string;
  simulateError: boolean;
  mockMode: boolean;
  recordMode: boolean;
  mockInput: string;
  mockFile: string;
}

function parseArgs(): Args {
  const args: Args = {
    simpleMode: false,
    outputDir: '../.starops-samples/logs',
    simulateError: false,
    mockMode: false,
    recordMode: false,
    mockInput: '',
    mockFile: '',
  };

  const argv = process.argv.slice(2);
  for (let i = 0; i < argv.length; i++) {
    switch (argv[i]) {
      case '-file':
      case '--file':
        args.filePath = argv[++i];
        break;
      case '-dir':
      case '--dir':
        args.dirPath = argv[++i];
        break;
      case '-simple':
      case '--simple':
        args.simpleMode = true;
        break;
      case '-output':
      case '--output':
        args.outputDir = argv[++i];
        break;
      case '-simulate-error':
      case '--simulate-error':
        args.simulateError = true;
        break;
      case '-mock':
      case '--mock':
        args.mockMode = true;
        break;
      case '-record':
      case '--record':
        args.recordMode = true;
        break;
      case '-mock-input':
      case '--mock-input':
        args.mockInput = argv[++i] || '';
        break;
      case '-mock-file':
      case '--mock-file':
        args.mockFile = argv[++i];
        break;
    }
  }

  return args;
}

async function processFile(
  client: AgentClient,
  filePath: string,
  outputDir: string,
  simpleMode: boolean,
  mockInput: string
): Promise<void> {
  try {
    // Load request file
    const content = fs.readFileSync(filePath, 'utf-8');
    const reqFile = JSON.parse(content);

    // Extract message
    const message = extractMessage(reqFile);
    if (!message) {
      console.log('⚠️ 文件中没有消息内容');
      return;
    }

    const fileName = path.basename(filePath);
    console.log(`📄 文件: ${fileName}`);
    console.log(`💬 消息: ${truncate(message, 60)}`);

    // Create thread
    const threadId = await client.createThread();

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

    // Create output file
    const outputFile = createOutputFile(filePath, outputDir);

    // Write request info
    writeOutput(outputFile, `# Request: ${fileName}`);
    writeOutput(outputFile, `# Time: ${new Date().toISOString()}`);
    writeOutput(outputFile, `# ThreadID: ${threadId}`);
    writeOutput(outputFile, `# Message: ${message}\n`);

    console.log('-'.repeat(60));

    // Extract variables
    const variables = reqFile.variables || {};

    // Send request
    const startTime = Date.now();
    const simplePrinter = simpleMode ? new SimplePrinter() : null;
    const eventPrinter = simpleMode ? null : new EventPrinter(false, true);
    const interactiveHandler = new InteractiveHandler(client);
    if (mockInput) interactiveHandler.setMockInput(mockInput);
    let eventIndex = 0;

    let events = client.chatWithVariables(threadId, message, variables);
    while (events) {
      let shouldContinue = false;
      for await (const event of events) {
        eventIndex++;

        if (event.error) {
          const err = event.error as SDKException;
          console.log(`❌ 错误: ${err.message}`);
          if (err.cause) {
            console.log(`   原因: ${(err.cause as Error).message || err.cause}`);
          }
          if ((err as any).context) {
            console.log(`   详情: ${JSON.stringify((err as any).context)}`);
          }
          writeOutput(outputFile, `[ERROR] ${err.message}`);
          shouldContinue = false;
          break;
        }

        // Write raw event
        if (event.rawJson) {
          writeOutput(outputFile, `[EVENT ${eventIndex}]\n${event.rawJson}\n`);
        }

        // 正常输出（先输出）
        if (simpleMode) {
          const text = simplePrinter!.processEvent(event);
          if (text) {
            process.stdout.write(text);
          }
        } else {
          eventPrinter!.printEvent(event, eventIndex);
        }

        // 检测交互事件（在输出之后）
        const interactiveResp = await extractInteractiveEvent(event, interactiveHandler);
        if (interactiveResp) {
          console.log('\n🔄 检测到交互事件，用户已响应...');
          events = interactiveHandler.resumeChat(threadId, interactiveResp, variables);
          eventIndex = 0;
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

    const elapsed = Date.now() - startTime;
    console.log();

    // Write final result
    if (simpleMode && simplePrinter) {
      const finalText = simplePrinter.getFinalText();
      writeOutput(outputFile, `\n# Final Result:\n${finalText}`);
      console.log(`📄 最终文本:\n${finalText}`);
    }

    writeOutput(outputFile, `\n# Duration: ${elapsed}ms`);
    console.log(`⏱️  耗时: ${elapsed}ms`);

    if (outputFile) {
      outputFile.close();
    }
  } catch (e) {
    console.log(`❌ 处理文件失败: ${(e as Error).message}`);
  }
}

async function processDirectory(
  client: AgentClient,
  dirPath: string,
  outputDir: string,
  simpleMode: boolean,
  mockInput: string
): Promise<void> {
  const files = fs.readdirSync(dirPath).filter((f) => f.endsWith('.json'));
  if (files.length === 0) {
    console.log(`⚠️ 目录中没有 JSON 文件: ${dirPath}`);
    return;
  }

  console.log(`📁 找到 ${files.length} 个请求文件\n`);

  for (let i = 0; i < files.length; i++) {
    console.log(`━━━ [${i + 1}/${files.length}] ${files[i]} ━━━`);
    await processFile(client, path.join(dirPath, files[i]), outputDir, simpleMode, mockInput);
    console.log();
  }

  console.log(`✅ 处理完成，共 ${files.length} 个文件`);
}

function extractMessage(reqFile: Record<string, unknown>): string | undefined {
  const messages = reqFile.messages as Array<Record<string, unknown>> | undefined;
  if (messages && messages[0]) {
    const contents = messages[0].contents as Array<Record<string, unknown>> | undefined;
    if (contents && contents[0]?.value) {
      return contents[0].value as string;
    }
  }
  return undefined;
}

function createOutputFile(inputFile: string, outputDir: string): fs.WriteStream | null {
  try {
    fs.mkdirSync(outputDir, { recursive: true });
    const baseName = path.basename(inputFile, '.json');
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
    const outputPath = path.join(outputDir, `${baseName}-${timestamp}.log`);
    return fs.createWriteStream(outputPath);
  } catch {
    return null;
  }
}

function writeOutput(file: fs.WriteStream | null, content: string): void {
  if (file) {
    file.write(content + '\n');
  }
}

function truncate(s: string, maxLen: number): string {
  if (s.length <= maxLen) return s;
  return s.slice(0, maxLen) + '...';
}

async function main() {
  const args = parseArgs();

  if (!args.filePath && !args.dirPath) {
    console.log('用法:');
    console.log('  --file <path>       处理单个文件');
    console.log('  --dir <path>        处理目录下所有 JSON 文件');
    console.log('  --mock-file <path>  指定 mock 文件路径');
    process.exit(1);
  }

  console.log('🚀 VibeOps Chat - 从文件加载请求 (TypeScript)');
  console.log('='.repeat(60));

  try {
    // Load configuration
    const cfg = await loadConfigFromEnv();

    if (args.simulateError) {
      cfg.simulateNetworkError = true;
      console.log('⚠️  已启用网络断连模拟，将在收到首个事件后触发重试');
    }
    if (args.mockMode) {
      cfg.mockMode = true;
      cfg.mockFile = args.mockFile || (args.filePath ? args.filePath + '.mock' : '');
      cfg.mockInput = args.mockInput;
      console.log(`🎭 Mock 模式，回放文件: ${cfg.mockFile}`);
      if (args.mockInput) console.log(`🤖 Mock 交互输入: ${args.mockInput}`);
    }
    if (args.recordMode) {
      cfg.recordMode = true;
      cfg.mockFile = args.mockFile || (args.filePath ? args.filePath + '.mock' : '');
      cfg.mockInput = args.mockInput;
      console.log(`📝 录制模式，输出文件: ${cfg.mockFile}`);
      if (args.mockInput) console.log(`🤖 录制交互输入: ${args.mockInput}`);
    }

    // Create client
    const client = new AgentClient(cfg);

    // Ensure output directory exists
    fs.mkdirSync(args.outputDir, { recursive: true });

    // Process files
    if (args.dirPath) {
      await processDirectory(client, args.dirPath, args.outputDir, args.simpleMode, args.mockInput);
    } else if (args.filePath) {
      await processFile(client, args.filePath, args.outputDir, args.simpleMode, args.mockInput);
    }
  } catch (e) {
    if (e instanceof SDKException) {
      console.log(`❌ 配置加载失败: ${e}`);
    } else {
      console.log(`❌ 错误: ${(e as Error).message}`);
    }
    process.exit(1);
  }
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

main();
