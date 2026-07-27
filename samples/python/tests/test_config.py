"""
Tests for Config — 默认值与必填环境变量校验
"""

import pytest

from starops_sdk_samples.client.config import Config
from starops_sdk_samples.client.errors import SDKException, ErrorCode


ENV_VARS = [
    "STAROPS_WORKSPACE",
    "STAROPS_ENDPOINT",
    "STAROPS_REGION",
    "STAROPS_EMPLOYEE_NAME",
    "ALIBABA_CLOUD_ACCESS_KEY_ID",
    "ALIBABA_CLOUD_ACCESS_KEY_SECRET",
]


@pytest.fixture()
def clean_env(monkeypatch):
    """屏蔽 .env 加载并清理相关环境变量，保证测试确定性"""
    monkeypatch.setattr(
        "starops_sdk_samples.client.config.load_dotenv", lambda *a, **k: None
    )
    for var in ENV_VARS:
        monkeypatch.delenv(var, raising=False)
    return monkeypatch


class TestConfigDefaults:
    def test_default_region_is_cn_beijing(self):
        config = Config(
            workspace="ws", endpoint="ep", access_key_id="ak", access_key_secret="sk"
        )
        assert config.region == "cn-beijing"

    def test_default_employee_name_is_apsara_ops(self):
        config = Config(
            workspace="ws", endpoint="ep", access_key_id="ak", access_key_secret="sk"
        )
        assert config.employee_name == "apsara-ops"


class TestLoadFromEnv:
    def test_load_with_all_vars(self, clean_env):
        clean_env.setenv("STAROPS_WORKSPACE", "test-ws")
        clean_env.setenv("STAROPS_ENDPOINT", "starops.cn-beijing.aliyuncs.com")
        clean_env.setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "test-ak")
        clean_env.setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "test-sk")

        config = Config.load_from_env()
        assert config.workspace == "test-ws"
        assert config.endpoint == "starops.cn-beijing.aliyuncs.com"
        assert config.access_key_id == "test-ak"
        assert config.access_key_secret == "test-sk"
        # 未设置 STAROPS_REGION / STAROPS_EMPLOYEE_NAME 时使用默认值
        assert config.region == "cn-beijing"
        assert config.employee_name == "apsara-ops"

    def test_region_env_overrides_default(self, clean_env):
        clean_env.setenv("STAROPS_ENDPOINT", "ep")
        clean_env.setenv("STAROPS_REGION", "cn-shanghai")
        clean_env.setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "ak")
        clean_env.setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "sk")

        config = Config.load_from_env()
        assert config.region == "cn-shanghai"

    def test_missing_endpoint_raises(self, clean_env):
        clean_env.setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "ak")
        clean_env.setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "sk")

        with pytest.raises(SDKException) as exc_info:
            Config.load_from_env()

        exc = exc_info.value
        assert exc.code == ErrorCode.CONFIG_MISSING
        assert "STAROPS_ENDPOINT" in exc.message
        assert "STAROPS_ENDPOINT" in exc.context.get("missingVariables", [])
