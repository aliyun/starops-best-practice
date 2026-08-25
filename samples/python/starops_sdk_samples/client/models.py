"""
models.py — 模型选择配置加载与辅助函数
职责：从共享的 config/models.json 读取模型清单，提供解析/校验/展示等纯函数。
不做：不调用 SDK、不处理交互（→examples/chat.py）。
依赖：标准库 json/os。
"""

import json
import os
from dataclasses import dataclass, field
from typing import List, Tuple


@dataclass
class Model:
    """单个模型条目 / A single model entry."""

    model_id: str
    display_name: str
    short_name: str = ""


@dataclass
class Provider:
    """模型提供方及其模型列表 / A model provider with its models."""

    provider: str
    display_name: str
    models: List[Model] = field(default_factory=list)


@dataclass
class ModelsConfig:
    """顶层配置结构 / Top-level config structure."""

    providers: List[Provider] = field(default_factory=list)


def load_models(config_path: str) -> ModelsConfig:
    """读取并反序列化 JSON 配置文件 / Read and deserialize the JSON config file."""
    try:
        with open(config_path, "r", encoding="utf-8") as f:
            data = json.load(f)
    except OSError as e:
        raise RuntimeError(f"failed to read models config: {e}") from e
    except json.JSONDecodeError as e:
        raise RuntimeError(f"failed to parse models config: {e}") from e

    providers: List[Provider] = []
    for p in data.get("providers", []) or []:
        models = [
            Model(
                model_id=m.get("modelId", ""),
                display_name=m.get("displayName", ""),
                short_name=m.get("shortName", ""),
            )
            for m in (p.get("models", []) or [])
        ]
        providers.append(
            Provider(
                provider=p.get("provider", ""),
                display_name=p.get("displayName", ""),
                models=models,
            )
        )
    return ModelsConfig(providers=providers)


def find_config_path() -> str:
    """定位 config/models.json 文件。
    优先级：STAROPS_MODELS_CONFIG 环境变量 > 从 cwd 向上查找 > 默认值。
    """
    env_path = os.getenv("STAROPS_MODELS_CONFIG")
    if env_path:
        return env_path

    try:
        cwd = os.getcwd()
    except OSError:
        return "config/models.json"

    dir_ = cwd
    while True:
        candidate = os.path.join(dir_, "config", "models.json")
        if os.path.exists(candidate):
            return candidate
        parent = os.path.dirname(dir_)
        if parent == dir_:
            break
        dir_ = parent

    return "config/models.json"


def parse_model_flag(flag: str) -> Tuple[str, str]:
    """解析 "provider:modelId" 格式 / Parse "provider:modelId" format."""
    parts = flag.split(":", 1)
    if len(parts) != 2 or parts[0] == "" or parts[1] == "":
        raise ValueError(
            f'invalid model flag "{flag}": format should be provider:modelId'
        )
    return parts[0], parts[1]


def validate_model(cfg: ModelsConfig, provider: str, model_id: str) -> bool:
    """校验 provider+modelId 组合是否存在于配置中。"""
    for p in cfg.providers:
        if p.provider == provider:
            for m in p.models:
                if m.model_id == model_id:
                    return True
    return False


def build_model_json(provider: str, model_id: str) -> str:
    """生成 JSON 字符串：{"provider":"x","modelID":"y"}。"""
    return json.dumps(
        {"provider": provider, "modelID": model_id},
        separators=(",", ":"),
        ensure_ascii=False,
    )


def build_config_with_model(provider: str, model_id: str) -> str:
    """生成含 model 与 disableThreadData 的完整配置 JSON 字符串。"""
    return json.dumps(
        {
            "model": {"provider": provider, "modelID": model_id},
            "disableThreadData": False,
        },
        separators=(",", ":"),
        ensure_ascii=False,
    )


def display_menu(cfg: ModelsConfig) -> str:
    """返回所有模型的带序号（1-based）菜单字符串。"""
    lines = ["可选模型列表：\n"]
    idx = 1
    for p in cfg.providers:
        for m in p.models:
            lines.append(f"  {idx}. {p.display_name}: {m.display_name}\n")
            idx += 1
    return "".join(lines)


def list_model_flags(cfg: ModelsConfig) -> str:
    """返回所有合法的 --model 取值及其描述。"""
    lines = ["可传入 --model 的模型取值（格式: provider:modelId）：\n"]
    for p in cfg.providers:
        for m in p.models:
            flag = f"{p.provider}:{m.model_id}"
            lines.append(f"  {flag:<30} {p.display_name}: {m.display_name}\n")
    return "".join(lines)


def get_model_by_index(cfg: ModelsConfig, index: int) -> Tuple[str, str]:
    """返回给定 0-based 序号处的 provider 与 modelId。
    顺序与 display_menu 的扁平顺序一致。
    """
    idx = 0
    for p in cfg.providers:
        for m in p.models:
            if idx == index:
                return p.provider, m.model_id
            idx += 1
    raise ValueError(f"model index {index} out of range (total {idx} models)")
