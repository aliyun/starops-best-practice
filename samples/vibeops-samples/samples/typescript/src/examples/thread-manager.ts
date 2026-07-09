#!/usr/bin/env tsx
/**
 * 会话管理工具
 * Thread management tool
 *
 * Usage:
 *   npm run thread-manager -- list
 *   npm run thread-manager -- get <thread-id>
 *   npm run thread-manager -- messages <thread-id>
 *   npm run thread-manager -- delete <thread-id>
 */

import { loadConfigFromEnv, AgentClient, SDKException } from '../client/index.js';

async function listThreads(client: AgentClient): Promise<void> {
  console.log('📋 会话列表');
  console.log('-'.repeat(100));

  const { threads, total } = await client.listThreads(20);
  console.log(`共 ${total} 个会话:\n`);

  if (threads.length === 0) {
    console.log('  (无会话)');
    return;
  }

  console.log(
    `${'Thread ID'.padEnd(40)} ${'标题'.padEnd(25)} ${'状态'.padEnd(10)} 创建时间`
  );
  console.log('-'.repeat(100));

  for (const t of threads) {
    let title = t.title || '';
    if (title.length > 23) {
      title = title.slice(0, 23) + '...';
    }
    console.log(
      `${t.threadId.padEnd(40)} ${title.padEnd(25)} ${(t.status || '').padEnd(10)} ${t.createTime || ''}`
    );
  }
}

async function getThread(client: AgentClient, threadId: string): Promise<void> {
  console.log(`📋 会话详情: ${threadId}`);
  console.log('-'.repeat(60));

  const detail = await client.getThread(threadId);

  console.log(`  Thread ID: ${detail.threadId}`);
  console.log(`  标题: ${detail.title}`);
  console.log(`  状态: ${detail.status}`);
  console.log(`  创建时间: ${detail.createTime}`);
  console.log(`  更新时间: ${detail.updateTime}`);
}

async function listMessages(client: AgentClient, threadId: string): Promise<void> {
  console.log(`💬 会话消息: ${threadId}`);
  console.log('-'.repeat(80));

  const messages = await client.getThreadData(threadId, 50);
  console.log(`共 ${messages.length} 条消息:\n`);

  for (let i = 0; i < messages.length; i++) {
    const m = messages[i];
    let roleIcon = '👤';
    if (m.role === 'assistant') {
      roleIcon = '🤖';
    } else if (m.role === 'system') {
      roleIcon = '⚙️';
    }

    let content = m.content || '';
    if (content.length > 100) {
      content = content.slice(0, 100) + '...';
    }

    console.log(`[${i + 1}] ${roleIcon} ${m.role}`);
    if (content) {
      console.log(`    ${content}`);
    }
    console.log(`    时间戳: ${m.timestamp}\n`);
  }
}

async function deleteThread(client: AgentClient, threadId: string): Promise<void> {
  console.log(`🗑️ 删除会话: ${threadId}`);

  await client.deleteThread(threadId);
  console.log('✅ 会话已删除');
}

function printUsage(): void {
  console.log('会话管理工具 - STAROps SDK 接口演示 (TypeScript)');
  console.log('='.repeat(50));
  console.log('\n用法:');
  console.log('  npm run thread-manager -- <command> [args]');
  console.log('\n命令:');
  console.log('  list                  列出所有会话');
  console.log('  get <thread-id>       获取会话详情');
  console.log('  messages <thread-id>  列出会话消息');
  console.log('  delete <thread-id>    删除会话');
}

async function main() {
  const args = process.argv.slice(2);

  if (args.length < 1) {
    printUsage();
    process.exit(1);
  }

  const command = args[0];

  try {
    // Load configuration
    const cfg = await loadConfigFromEnv();

    // Create client
    const client = new AgentClient(cfg);

    switch (command) {
      case 'list':
        await listThreads(client);
        break;
      case 'get':
        if (args.length < 2) {
          console.log('❌ 请指定 thread-id');
          process.exit(1);
        }
        await getThread(client, args[1]);
        break;
      case 'messages':
        if (args.length < 2) {
          console.log('❌ 请指定 thread-id');
          process.exit(1);
        }
        await listMessages(client, args[1]);
        break;
      case 'delete':
        if (args.length < 2) {
          console.log('❌ 请指定 thread-id');
          process.exit(1);
        }
        await deleteThread(client, args[1]);
        break;
      default:
        console.log(`❌ 未知命令: ${command}`);
        printUsage();
        process.exit(1);
    }
  } catch (e) {
    if (e instanceof SDKException) {
      console.log(`❌ 错误: ${e}`);
    } else {
      console.log(`❌ 错误: ${(e as Error).message}`);
    }
    process.exit(1);
  }
}

main();
