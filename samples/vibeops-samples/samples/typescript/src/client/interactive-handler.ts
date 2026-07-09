/**
 * Interactive handler for STAROps SDK
 * STAROps SDK 交互处理器
 */

import * as readline from 'readline';
import { EventType, InteractionType } from '../types/index.js';
import { AgentClient, ChatEvent } from './agent-client.js';
import { SDKException, ErrorCode } from './errors.js';

/** 交互响应 / Interactive response */
export interface InteractiveResponse {
  callId: string;
  type: InteractionType;
  response: Record<string, unknown>;
  source?: Record<string, unknown>;
  modifiedData?: Record<string, unknown>;
  formData?: Record<string, unknown>;
  decision?: string;
}

/** 交互事件处理器 / Interactive event handler */
export class InteractiveHandler {
  private client: AgentClient;
  private timeout?: number;
  private rl?: readline.Interface;

  constructor(client: AgentClient, timeout?: number) {
    this.client = client;
    this.timeout = timeout;
  }

  /** 处理交互事件 / Handle interactive event */
  async handleEvent(event: Record<string, unknown>, callId: string): Promise<InteractiveResponse> {
    if (!event) {
      throw new SDKException(ErrorCode.PARSE_ERROR, '事件为空');
    }

    const eventType = event.type as string;
    if (eventType !== EventType.INTERACTIVE) {
      throw new SDKException(ErrorCode.PARSE_ERROR, `不支持的事件类型: ${eventType}`);
    }

    const payload = event.payload as Record<string, unknown>;
    if (!payload) {
      throw new SDKException(ErrorCode.PARSE_ERROR, '交互负载为空');
    }

    const interactiveType = payload.type as string;

    switch (interactiveType) {
      case InteractionType.USER_ACK:
        return this.handleUserAck(payload, callId);
      case InteractionType.USER_SELECT:
        return this.handleUserSelect(payload, callId);
      case InteractionType.USER_INPUT:
        return this.handleUserInput(payload, callId);
      default:
        throw new SDKException(ErrorCode.PARSE_ERROR, `不支持的交互类型: ${interactiveType}`);
    }
  }

  /** 处理用户确认 / Handle user acknowledgment */
  async handleUserAck(payload: Record<string, unknown>, callId: string): Promise<InteractiveResponse> {
    const title = this.getTitle(payload);
    const message = this.getDescription(payload);
    const options = this.getOptions(payload);
    const modifiedData = this.extractData(payload);

    console.log('\n🔔 确认请求');
    console.log('--------------');
    if (title) console.log(title);
    if (message) console.log(`\n${message}`);

    if (modifiedData) {
      console.log();
      for (const [key, value] of Object.entries(modifiedData)) {
        if (key === 'title' || key === 'message') continue;
        console.log(`${key}: ${value}`);
      }
    }
    console.log('--------------');

    let userInput: string;
    if (options.length > 0) {
      const parts = options.map((opt, i) =>
        `[${this.getOptionValue(opt)}] ${this.getOptionLabel(opt, i)}`);
      userInput = await this.prompt(`请输入 ${parts.join(', ')}: `);
    } else {
      userInput = await this.prompt('请输入 [y/yes] 确认，[n/no] 取消: ');
    }

    userInput = userInput.trim().toLowerCase();
    let confirmed = ['', 'y', 'yes', '是'].includes(userInput);
    let decision = confirmed ? 'yes' : 'no';

    for (const opt of options) {
      if (userInput === this.getOptionValue(opt).toLowerCase()) {
        decision = this.getOptionValue(opt);
        confirmed = true;
        break;
      }
    }

    return {
      callId,
      type: InteractionType.USER_ACK,
      response: { confirmed },
      source: this.extractSource(payload),
      modifiedData,
      decision,
    };
  }

  /** 处理用户选择 / Handle user selection */
  async handleUserSelect(payload: Record<string, unknown>, callId: string): Promise<InteractiveResponse> {
    const title = this.getTitle(payload);

    console.log('\n📋 请选择');
    if (title) console.log(`   标题: ${title}`);

    const options = this.getOptions(payload);
    if (options.length === 0) {
      throw new SDKException(ErrorCode.PARSE_ERROR, '没有可选项');
    }

    console.log('   选项:');
    options.forEach((opt, i) => {
      const label = this.getOptionLabel(opt, i);
      console.log(`   [${i + 1}] ${label}`);
    });

    const input = await this.prompt(`   请输入选项编号 (1-${options.length}): `);
    const selectedIndex = parseInt(input.trim(), 10);

    if (isNaN(selectedIndex) || selectedIndex < 1 || selectedIndex > options.length) {
      throw new SDKException(
        ErrorCode.PARSE_ERROR,
        `无效的选择: ${input}，请输入 1-${options.length} 之间的数字`
      );
    }

    const selectedOption = options[selectedIndex - 1];
    const decision = this.getOptionValue(selectedOption);

    return {
      callId,
      type: InteractionType.USER_SELECT,
      response: {
        selectedIndex: selectedIndex - 1,
        selectedValue: selectedOption,
      },
      source: this.extractSource(payload),
      modifiedData: this.extractData(payload),
      decision,
    };
  }

  /** 处理用户输入（表单模式）/ Handle user input (form mode) */
  async handleUserInput(payload: Record<string, unknown>, callId: string): Promise<InteractiveResponse> {
    const title = this.getTitle(payload);
    const description = this.getDescription(payload);
    const source = this.extractSource(payload);

    const formSpec = this.extractFormSpec(payload);
    const elements = this.getFormElements(formSpec);
    const initialValues = this.getFormInitialValues(formSpec);

    console.log(`\n✏️  ${title}`);
    if (description) console.log(`    ${description}`);
    console.log(`    ${'-'.repeat(40)}`);

    const formData: Record<string, unknown> = {};

    for (const elem of elements) {
      const field = this.getFieldKey(elem);
      const label = this.getFieldLabel(elem, field);
      const widget = this.getFieldWidget(elem);
      const placeholder = this.getFieldPlaceholder(elem);
      const defaultValue = this.getInitialValue(initialValues, field);

      if (widget === 'radio' || widget === 'segmented') {
        const enumOpts = this.getFieldEnum(formSpec, field);
        if (enumOpts.length > 0) {
          console.log(`    ${label}:`);
          enumOpts.forEach((opt, i) => {
            const marker = defaultValue === opt ? '*' : ' ';
            console.log(`      [${i + 1}]${marker} ${opt}`);
          });
          let prompt = `    请选择 (1-${enumOpts.length})`;
          if (defaultValue) prompt += ` [默认: ${defaultValue}]`;
          prompt += ': ';
          const input = (await this.prompt(prompt)).trim();
          if (!input && defaultValue) {
            formData[field] = defaultValue;
          } else {
            const idx = parseInt(input, 10);
            if (!isNaN(idx) && idx >= 1 && idx <= enumOpts.length) {
              formData[field] = enumOpts[idx - 1];
            } else {
              formData[field] = defaultValue;
            }
          }
        }
      } else {
        let prompt = `    ${label}`;
        if (placeholder) prompt += ` (${placeholder})`;
        if (defaultValue) prompt += ` [默认: ${defaultValue}]`;
        prompt += ': ';
        const input = (await this.prompt(prompt)).trim();
        if (!input && defaultValue) {
          formData[field] = defaultValue;
        } else {
          formData[field] = input;
        }
      }
    }

    console.log(`    ${'-'.repeat(40)}`);

    return {
      callId,
      type: InteractionType.USER_INPUT,
      response: { value: formData },
      source,
      formData,
      decision: 'submit',
    };
  }

  /** 使用交互响应恢复对话 / Resume chat with interactive response */
  async *resumeChat(
    threadId: string,
    response: InteractiveResponse,
    baseVariables?: Record<string, unknown>
  ): AsyncIterable<ChatEvent> {
    if (!this.client) {
      yield {
        rawJson: '',
        statusCode: 0,
        isDone: false,
        error: new SDKException(ErrorCode.CLIENT_CREATE, '客户端未初始化'),
      };
      return;
    }

    if (!response) {
      yield {
        rawJson: '',
        statusCode: 0,
        isDone: false,
        error: new SDKException(ErrorCode.PARSE_ERROR, '交互响应为空'),
      };
      return;
    }

    const userInteractive: Record<string, unknown> = {
      callId: response.callId,
      source: response.source,
      decision: response.decision,
    };
    if (response.type === InteractionType.USER_INPUT) {
      userInteractive.formData = response.formData;
    } else {
      userInteractive.modifiedData = response.modifiedData;
    }

    const uiJson = JSON.stringify(userInteractive);
    yield* this.client.interact(threadId, uiJson, baseVariables);
  }

  // =================================================================================
  // 辅助方法
  // =================================================================================

  private async prompt(question: string): Promise<string> {
    return new Promise((resolve) => {
      this.rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
      });

      this.rl.question(question, (answer) => {
        this.rl?.close();
        resolve(answer);
      });
    });
  }

  private getTitle(payload: Record<string, unknown>): string | undefined {
    const userAck = payload.userAck as Record<string, unknown> | undefined;
    if (userAck) {
      const data = userAck.data as Record<string, unknown> | undefined;
      if (data?.title) return data.title as string;
    }
    const userInput = payload.userInput as Record<string, unknown> | undefined;
    if (userInput?.title) return userInput.title as string;
    const meta = payload.meta as Record<string, unknown> | undefined;
    return meta?.title as string | undefined;
  }

  private getDescription(payload: Record<string, unknown>): string | undefined {
    const userAck = payload.userAck as Record<string, unknown> | undefined;
    if (userAck?.message) return userAck.message as string;
    const userInput = payload.userInput as Record<string, unknown> | undefined;
    if (userInput?.description) return userInput.description as string;
    const meta = payload.meta as Record<string, unknown> | undefined;
    return (meta?.description || meta?.desc) as string | undefined;
  }

  private getOptions(payload: Record<string, unknown>): Array<Record<string, unknown>> {
    const userAck = payload.userAck as Record<string, unknown> | undefined;
    if (userAck?.options) {
      const opts = userAck.options as Array<Record<string, unknown>>;
      if (opts.length > 0) return opts;
    }
    const data = payload.data as Array<Record<string, unknown>> | undefined;
    if (data && data.length > 0) return data;
    const meta = payload.meta as Record<string, unknown> | undefined;
    return (meta?.options as Array<Record<string, unknown>>) || [];
  }

  private getOptionLabel(option: Record<string, unknown>, index: number): string {
    for (const field of ['label', 'name', 'title', 'value']) {
      const value = option[field];
      if (typeof value === 'string' && value) return value;
    }
    return `选项 ${index + 1}`;
  }

  private getOptionValue(option: Record<string, unknown>): string {
    const value = option.value;
    return typeof value === 'string' ? value : '';
  }

  private extractSource(payload: Record<string, unknown>): Record<string, unknown> | undefined {
    const userAck = payload.userAck as Record<string, unknown> | undefined;
    if (userAck?.source) return userAck.source as Record<string, unknown>;
    const userInput = payload.userInput as Record<string, unknown> | undefined;
    if (userInput?.source) return userInput.source as Record<string, unknown>;
    const meta = payload.meta as Record<string, unknown> | undefined;
    return meta?.source as Record<string, unknown> | undefined;
  }

  private extractData(payload: Record<string, unknown>): Record<string, unknown> | undefined {
    const userAck = payload.userAck as Record<string, unknown> | undefined;
    if (userAck?.data) return userAck.data as Record<string, unknown>;
    const meta = payload.meta as Record<string, unknown> | undefined;
    return meta?.data as Record<string, unknown> | undefined;
  }

  // =================================================================================
  // formSpec 辅助方法 (user_input 表单模式)
  // =================================================================================

  private extractFormSpec(payload: Record<string, unknown>): Record<string, unknown> | undefined {
    const userInput = payload.userInput as Record<string, unknown> | undefined;
    return userInput?.formSpec as Record<string, unknown> | undefined;
  }

  private getFormElements(formSpec?: Record<string, unknown>): Array<Record<string, unknown>> {
    if (!formSpec) return [];
    const uiSchema = formSpec.ui_schema as Record<string, unknown> | undefined;
    return (uiSchema?.elements as Array<Record<string, unknown>>) || [];
  }

  private getFormInitialValues(formSpec?: Record<string, unknown>): Record<string, unknown> | undefined {
    return formSpec?.initialValues as Record<string, unknown> | undefined;
  }

  private getFieldKey(elem: Record<string, unknown>): string {
    return (elem.field as string) || '';
  }

  private getFieldLabel(elem: Record<string, unknown>, field: string): string {
    return (elem.label as string) || field;
  }

  private getFieldWidget(elem: Record<string, unknown>): string {
    return (elem.widget as string) || 'input';
  }

  private getFieldPlaceholder(elem: Record<string, unknown>): string {
    return (elem.placeholder as string) || '';
  }

  private getInitialValue(initialValues: Record<string, unknown> | undefined, field: string): string {
    if (!initialValues) return '';
    const val = initialValues[field];
    return val !== undefined ? String(val) : '';
  }

  private getFieldEnum(formSpec: Record<string, unknown> | undefined, field: string): string[] {
    if (!formSpec) return [];
    const schema = formSpec.schema as Record<string, unknown> | undefined;
    const properties = schema?.properties as Record<string, unknown> | undefined;
    const prop = properties?.[field] as Record<string, unknown> | undefined;
    const enumVals = prop?.enum as unknown[] | undefined;
    return (enumVals || []).map(v => String(v));
  }

  /** 检查事件是否为交互事件 / Check if event is interactive */
  static isInteractiveEvent(event: Record<string, unknown>): boolean {
    return event?.type === EventType.INTERACTIVE;
  }

  /** 从消息中提取交互事件 / Extract interactive events from message */
  static extractInteractiveEvents(
    message: Record<string, unknown>
  ): Array<Record<string, unknown>> {
    const events = message?.events as Array<Record<string, unknown>> | undefined;
    if (!events) return [];
    return events.filter((e) => InteractiveHandler.isInteractiveEvent(e));
  }
}