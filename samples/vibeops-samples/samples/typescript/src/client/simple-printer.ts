/**
 * Simple printer for STAROps SDK
 * STAROps SDK 简洁打印器
 */

import { ChatEvent } from './agent-client.js';

/** 简洁模式打印器 / Simple mode printer */
export class SimplePrinter {
  private buffer: string[] = [];
  private seenArtifacts: Set<string> = new Set();

  /** 处理事件，提取文本内容 / Process event and extract text content */
  processEvent(event: ChatEvent | null | undefined): string {
    if (!event?.body) return '';
    // 利用 event 字段快速跳过非文本事件
    if (event.event && event.event !== 'text' && event.event !== 'task_finished') return '';

    const extracted: string[] = [];
    const messages = (event.body.messages as Array<Record<string, unknown>>) || [];

    for (const msg of messages) {
      // Only process system role messages
      if (msg.role === 'system') {
        const text = this.extractTextFromArtifacts(
          msg.artifacts as Array<Record<string, unknown>> | undefined
        );
        if (text && !this.seenArtifacts.has(text)) {
          this.seenArtifacts.add(text);
          extracted.push(text);
          this.buffer.push(text);
        }
      }
    }

    return extracted.join('');
  }

  /** 从 artifacts 中提取文本 / Extract text from artifacts */
  private extractTextFromArtifacts(
    artifacts: Array<Record<string, unknown>> | undefined
  ): string {
    if (!artifacts) return '';

    const result: string[] = [];
    for (const artifact of artifacts) {
      const parts = artifact.parts as Array<Record<string, unknown>> | undefined;
      if (!parts) continue;

      for (const part of parts) {
        if (part.kind === 'text' && part.text) {
          result.push(part.text as string);
        }
      }
    }

    return result.join('');
  }

  /** 获取最终文本 / Get final text */
  getFinalText(): string {
    return this.buffer.join('');
  }

  /** 重置缓冲区 / Reset buffer */
  reset(): void {
    this.buffer = [];
    this.seenArtifacts.clear();
  }
}
