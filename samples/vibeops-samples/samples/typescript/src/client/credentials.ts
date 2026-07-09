/**
 * 默认凭据链支持 - 与 Go 实现保持一致
 * 凭据链优先级：环境变量 > OIDC > CLI配置文件 > 配置文件(~/.alibabacloud/credentials) > IAM角色
 */
import CredentialPkg from '@alicloud/credentials';

type CredentialValue = {
  accessKeyId?: string;
  accessKeySecret?: string;
};

type CredentialClient = {
  getCredential(): Promise<CredentialValue>;
};

type CredentialConstructor = new () => CredentialClient;

function resolveCredentialConstructor(pkg: unknown): CredentialConstructor {
  const moduleValue = pkg as Record<string, unknown> | undefined;
  const defaultValue = moduleValue?.default as Record<string, unknown> | undefined;
  const candidates = [pkg, moduleValue?.default, defaultValue?.default];

  for (const candidate of candidates) {
    if (typeof candidate === 'function') {
      return candidate as CredentialConstructor;
    }
  }

  throw new TypeError('@alicloud/credentials 未导出可用的 Credential 构造函数');
}

export async function loadCredentialsFromChain(): Promise<{ accessKeyId: string; accessKeySecret: string }> {
  const Credential = resolveCredentialConstructor(CredentialPkg);
  const cred = new Credential();
  const credential = await cred.getCredential();
  const accessKeyId = credential.accessKeyId;
  const accessKeySecret = credential.accessKeySecret;
  if (!accessKeyId || !accessKeySecret) {
    throw new Error('凭据链返回的凭证为空');
  }
  return { accessKeyId, accessKeySecret };
}
