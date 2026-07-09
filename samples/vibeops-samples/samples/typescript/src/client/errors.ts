/**
 * Error handling for STAROps SDK
 * STAROps SDK 错误处理
 */

/** 错误码 / Error code */
export enum ErrorCode {
  CONFIG_MISSING = 'CONFIG_MISSING',
  CONFIG_INVALID = 'CONFIG_INVALID',
  CLIENT_CREATE = 'CLIENT_CREATE',
  THREAD_CREATE = 'THREAD_CREATE',
  THREAD_NOT_FOUND = 'THREAD_NOT_FOUND',
  CHAT_FAILED = 'CHAT_FAILED',
  TIMEOUT = 'TIMEOUT',
  CANCELLED = 'CANCELLED',
  NETWORK_ERROR = 'NETWORK_ERROR',
  API_ERROR = 'API_ERROR',
  PARSE_ERROR = 'PARSE_ERROR',
  INTERACTIVE_TIMEOUT = 'INTERACTIVE_TIMEOUT',
}

/** SDK 异常类 / SDK exception class */
export class SDKException extends Error {
  code: ErrorCode;
  context: Record<string, unknown>;
  suggestion?: string;
  cause?: Error;

  constructor(code: ErrorCode, message: string, cause?: Error) {
    super(message);
    this.name = 'SDKException';
    this.code = code;
    this.context = {};
    this.cause = cause;
  }

  withContext(key: string, value: unknown): this {
    this.context[key] = value;
    return this;
  }

  withSuggestion(suggestion: string): this {
    this.suggestion = suggestion;
    return this;
  }

  override toString(): string {
    let result = `[${this.code}] ${this.message}`;
    if (this.cause) {
      result += `: ${this.cause.message}`;
    }
    return result;
  }

  // Convenience factory methods
  static configMissing(missingVars: string[]): SDKException {
    return new SDKException(
      ErrorCode.CONFIG_MISSING,
      `缺少必需的配置项: ${missingVars.join(', ')}`
    )
      .withContext('missingVariables', missingVars)
      .withSuggestion('请检查 .env 文件或环境变量设置');
  }

  static configInvalid(field: string, reason: string): SDKException {
    return new SDKException(
      ErrorCode.CONFIG_INVALID,
      `配置项 ${field} 无效: ${reason}`
    )
      .withContext('field', field)
      .withContext('reason', reason)
      .withSuggestion('请检查配置值是否正确');
  }

  static clientCreate(cause: Error): SDKException {
    return new SDKException(ErrorCode.CLIENT_CREATE, '创建客户端失败', cause)
      .withSuggestion('请检查网络连接和认证信息');
  }

  static threadCreate(cause: Error): SDKException {
    return new SDKException(ErrorCode.THREAD_CREATE, '创建会话失败', cause)
      .withSuggestion('请检查 API 权限和配额');
  }

  static threadNotFound(threadId: string): SDKException {
    return new SDKException(ErrorCode.THREAD_NOT_FOUND, `会话不存在: ${threadId}`)
      .withContext('threadId', threadId)
      .withSuggestion('请检查会话 ID 是否正确，或创建新会话');
  }

  static chatFailed(cause: Error): SDKException {
    return new SDKException(ErrorCode.CHAT_FAILED, '对话失败', cause)
      .withSuggestion('请稍后重试');
  }

  static timeout(duration: string): SDKException {
    return new SDKException(ErrorCode.TIMEOUT, `操作超时: ${duration}`)
      .withContext('duration', duration)
      .withSuggestion('请增加超时时间或检查网络连接');
  }

  static cancelled(): SDKException {
    return new SDKException(ErrorCode.CANCELLED, '操作已取消')
      .withSuggestion('如需继续，请重新发起请求');
  }

  static networkError(cause: Error): SDKException {
    return new SDKException(ErrorCode.NETWORK_ERROR, '网络错误', cause)
      .withSuggestion('请检查网络连接');
  }

  static apiError(apiCode: string, apiMessage: string): SDKException {
    return new SDKException(
      ErrorCode.API_ERROR,
      `API 错误 [${apiCode}]: ${apiMessage}`
    )
      .withContext('apiCode', apiCode)
      .withContext('apiMessage', apiMessage)
      .withSuggestion('请参考 API 文档检查请求参数');
  }

  static parseError(cause: Error): SDKException {
    return new SDKException(ErrorCode.PARSE_ERROR, '解析响应失败', cause)
      .withSuggestion('请检查 SDK 版本是否最新');
  }

  static interactiveTimeout(duration: string): SDKException {
    return new SDKException(
      ErrorCode.INTERACTIVE_TIMEOUT,
      `等待用户响应超时: ${duration}`
    )
      .withContext('duration', duration)
      .withSuggestion('请重新操作并在规定时间内响应');
  }
}
