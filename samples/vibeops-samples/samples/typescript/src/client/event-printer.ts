/**
 * 事件打印器 - 打印每个 SSE 事件的详细信息
 * Event printer - prints detailed information for each SSE event
 */

import { ChatEvent } from './agent-client.js';

/** 事件打印器 / Event printer */
export class EventPrinter {
  private printRawBody: boolean;
  private printParsed: boolean;
  private printSeparator: boolean;

  constructor(printRawBody = false, printParsed = true) {
    this.printRawBody = printRawBody;
    this.printParsed = printParsed;
    this.printSeparator = true;
  }

  /** 打印事件 / Print event */
  printEvent(event: ChatEvent, eventIndex: number): void {
    if (event.error) {
      console.log(`\n❌ 错误: ${event.error}`);
      return;
    }

    if (event.isDone && !event.body) {
      console.log('\n✅ 对话完成');
      return;
    }

    if (!event.body) return;

    if (this.printSeparator) {
      const sep = '='.repeat(30);
      console.log(`\n${sep} 事件 #${eventIndex} ${sep}`);
    }

    if (this.printRawBody && event.rawJson) {
      console.log('\n📦 原始 Body:');
      console.log(prettyJson(event.rawJson));
    }

    if (this.printParsed) {
      this.printParsedEvent(event.body);
    }
  }

  private printParsedEvent(body: Record<string, unknown>): void {
    console.log('\n📋 解析详情:');

    const messages = (body.messages as Array<Record<string, unknown>>) || [];
    for (const msg of messages) {
      if (typeof msg !== 'object' || !msg) continue;
      console.log(`  原始消息: ${JSON.stringify(msg)}`);
      this.printMessageItem(msg);
    }
  }

  private printMessageItem(item: Record<string, unknown>): void {
    const role = item.role as string;
    if (role) console.log(`  📌 角色: ${role}`);

    const callId = item.callId as string;
    if (callId) console.log(`  🔗 CallID: ${callId}`);

    const parentCallId = item.parentCallId as string;
    if (parentCallId) console.log(`  🔗 ParentCallID: ${parentCallId}`);

    // 内容
    const contents = item.contents as Array<Record<string, unknown>> | undefined;
    if (contents?.length) {
      console.log('  📝 内容:');
      contents.forEach((content, i) => {
        if (typeof content !== 'object' || !content) return;
        console.log(`    [${i}] 类型: ${content.type || ''}`);
        let value = (content.value as string) || '';
        if (value) {
          if (value.length > 200) value = value.slice(0, 200) + '...';
          console.log(`        值: ${value}`);
        }
        if (content.append) console.log('        追加: true');
        if (content.lastChunk) console.log('        最后块: true');
      });
    }

    // 工具调用
    const tools = item.tools as Array<Record<string, unknown>> | undefined;
    if (tools?.length) {
      console.log('  🔧 工具调用:');
      tools.forEach((tool, i) => {
        if (typeof tool !== 'object' || !tool) return;
        console.log(`    [${i}] 名称: ${tool.name || ''}, 状态: ${tool.status || ''}`);
        if (tool.toolCallId) console.log(`        ToolCallID: ${tool.toolCallId}`);
        if (tool.arguments != null) {
          let argsStr = JSON.stringify(tool.arguments);
          if (argsStr.length > 200) argsStr = argsStr.slice(0, 200) + '...';
          console.log(`        参数: ${argsStr}`);
        }
      });
    }

    // Agent 调用
    const agents = item.agents as Array<Record<string, unknown>> | undefined;
    if (agents?.length) {
      console.log('  🤖 Agent调用:');
      agents.forEach((agent, i) => {
        if (typeof agent !== 'object' || !agent) return;
        console.log(`    [${i}] 名称: ${agent.name || ''}, 状态: ${agent.status || ''}`);
      });
    }

    // 事件
    const events = item.events as Array<Record<string, unknown>> | undefined;
    if (events?.length) {
      console.log('  📢 事件:');
      events.forEach((evt, i) => {
        if (typeof evt !== 'object' || !evt) return;
        const evtType = (evt.type as string) || '';
        console.log(`    [${i}] 类型: ${evtType}`);
        if (evt.payload != null) {
          this.printEventPayload(evtType, evt.payload as Record<string, unknown>);
        }
      });
    }
  }

  private printEventPayload(evtType: string, payload: unknown): void {
    if (typeof payload !== 'object' || !payload) return;
    const p = payload as Record<string, unknown>;

    switch (evtType) {
      case 'thinking': {
        let delta = (p.reasoningDelta as string) || '';
        if (delta) {
          if (delta.length > 100) delta = delta.slice(0, 100) + '...';
          console.log(`        思考: ${delta}`);
        }
        break;
      }
      case 'error':
        console.log(`        错误码: ${p.code || ''}`);
        console.log(`        消息: ${p.message || ''}`);
        break;
      case 'task_finished': {
        console.log(`        成功: ${p.success ?? false}`);
        const statistics = p.statistics as Record<string, unknown> | undefined;
        if (statistics) {
          const durationNs = (statistics.duration as number) || 0;
          console.log(`        耗时: ${Math.floor(durationNs / 1000000)}ms`);
        }
        break;
      }
      default: {
        let payloadStr = JSON.stringify(payload);
        if (payloadStr.length > 200) payloadStr = payloadStr.slice(0, 200) + '...';
        console.log(`        负载: ${payloadStr}`);
        break;
      }
    }
  }
}

/** 格式化 JSON 输出 / Pretty print JSON */
function prettyJson(jsonStr: string): string {
  try {
    return JSON.stringify(JSON.parse(jsonStr), null, 2);
  } catch {
    return jsonStr;
  }
}
