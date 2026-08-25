/**
 * models.ts — 模型配置加载与辅助函数
 * 职责：从 config/models.json 加载模型列表、解析/校验模型标识、构建 API 参数。
 * 不做：不处理网络请求、不处理对话逻辑。
 * 依赖：Node.js fs/path
 */

import * as fs from 'fs';
import * as path from 'path';

/** 单个模型条目 */
export interface Model {
  modelId: string;
  displayName: string;
  shortName: string;
}

/** 模型提供方 */
export interface Provider {
  provider: string;
  displayName: string;
  models: Model[];
}

/** 模型配置顶层结构 */
export interface ModelsConfig {
  providers: Provider[];
}

/** 读取并反序列化模型配置文件 */
export function loadModels(configPath: string): ModelsConfig | null {
  try {
    const data = fs.readFileSync(configPath, 'utf-8');
    return JSON.parse(data) as ModelsConfig;
  } catch {
    return null;
  }
}

/**
 * 定位 config/models.json 文件路径
 * 优先级：STAROPS_MODELS_CONFIG 环境变量 > 从 cwd 向上查找 > 默认值
 */
export function findConfigPath(): string {
  const envPath = process.env.STAROPS_MODELS_CONFIG;
  if (envPath) return envPath;

  let dir = process.cwd();
  while (true) {
    const candidate = path.join(dir, 'config', 'models.json');
    if (fs.existsSync(candidate)) {
      return candidate;
    }
    const parent = path.dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }

  return 'config/models.json';
}

/** 解析 "provider:modelId" 格式 */
export function parseModelFlag(flag: string): { provider: string; modelId: string } {
  const idx = flag.indexOf(':');
  if (idx <= 0 || idx === flag.length - 1) {
    throw new Error(`invalid model flag "${flag}": format should be provider:modelId`);
  }
  return { provider: flag.substring(0, idx), modelId: flag.substring(idx + 1) };
}

/** 校验 provider+modelId 组合是否存在于配置中 */
export function validateModel(cfg: ModelsConfig, provider: string, modelId: string): boolean {
  for (const p of cfg.providers) {
    if (p.provider === provider) {
      for (const m of p.models) {
        if (m.modelId === modelId) {
          return true;
        }
      }
    }
  }
  return false;
}

/** 生成 JSON 字符串：{"provider":"x","modelID":"y"} */
export function buildModelJSON(provider: string, modelId: string): string {
  return JSON.stringify({ provider, modelID: modelId });
}

/** 生成完整配置 JSON（含 disableThreadData） */
export function buildConfigWithModel(provider: string, modelId: string): string {
  return JSON.stringify({
    model: { provider, modelID: modelId },
    disableThreadData: false,
  });
}

/** 返回带序号的模型菜单（1-based） */
export function displayMenu(cfg: ModelsConfig): string {
  const lines: string[] = ['可选模型列表：'];
  let idx = 1;
  for (const p of cfg.providers) {
    for (const m of p.models) {
      lines.push(`  ${idx}. ${p.displayName}: ${m.displayName}`);
      idx++;
    }
  }
  return lines.join('\n') + '\n';
}

/** 返回所有可传入 --model 的取值及描述 */
export function listModelFlags(cfg: ModelsConfig): string {
  const lines: string[] = ['可传入 --model 的模型取值（格式: provider:modelId）：'];
  for (const p of cfg.providers) {
    for (const m of p.models) {
      const flag = `${p.provider}:${m.modelId}`;
      lines.push(`  ${flag.padEnd(30)} ${p.displayName}: ${m.displayName}`);
    }
  }
  return lines.join('\n') + '\n';
}

/** 根据 0-based 索引获取模型的 provider 和 modelId */
export function getModelByIndex(cfg: ModelsConfig, index: number): { provider: string; modelId: string } {
  let idx = 0;
  for (const p of cfg.providers) {
    for (const m of p.models) {
      if (idx === index) {
        return { provider: p.provider, modelId: m.modelId };
      }
      idx++;
    }
  }
  throw new Error(`model index ${index} out of range (total ${idx} models)`);
}
