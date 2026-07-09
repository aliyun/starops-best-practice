"""
Configuration for STAROps SDK
STAROps SDK 配置
"""

import os
import sys
from dataclasses import dataclass
from typing import Optional

from dotenv import load_dotenv

from .errors import SDKException
from .retry import RetryConfig, load_retry_config_from_env


@dataclass
class Config:
    """应用配置 / Application configuration"""
    workspace: str
    endpoint: str
    access_key_id: str
    access_key_secret: str
    region: str = "cn-hangzhou"
    employee_name: str = "default"
    retry_config: Optional[RetryConfig] = None  # 重试配置，None 时使用默认配置
    simulate_network_error: bool = False  # 模拟网络断连，用于测试重试逻辑

    @classmethod
    def load_from_env(cls) -> "Config":
        """从环境变量加载配置 / Load configuration from environment variables"""
        load_dotenv()

        workspace = os.getenv("VIBEOPS_WORKSPACE", "")
        endpoint = os.getenv("VIBEOPS_ENDPOINT", "")
        region = os.getenv("VIBEOPS_REGION", "cn-hangzhou")
        access_key_id = os.getenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "")
        access_key_secret = os.getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "")

        # 环境变量为空时回退到阿里云默认凭据链
        if not access_key_id or not access_key_secret:
            try:
                from .credentials import load_credentials_from_chain
                access_key_id, access_key_secret = load_credentials_from_chain()
            except Exception as e:
                print(f"凭据链加载失败: {e}，请手动设置环境变量", file=sys.stderr)

        employee_name = os.getenv("VIBEOPS_EMPLOYEE_NAME", "default")

        # Validate required fields
        missing_vars = []
        if not endpoint:
            missing_vars.append("VIBEOPS_ENDPOINT")
        if not access_key_id:
            missing_vars.append("ALIBABA_CLOUD_ACCESS_KEY_ID")
        if not access_key_secret:
            missing_vars.append("ALIBABA_CLOUD_ACCESS_KEY_SECRET")

        if missing_vars:
            raise SDKException.config_missing(missing_vars)

        return cls(
            workspace=workspace,
            endpoint=endpoint,
            access_key_id=access_key_id,
            access_key_secret=access_key_secret,
            region=region or "cn-hangzhou",
            employee_name=employee_name or "default",
            retry_config=load_retry_config_from_env(),
        )
