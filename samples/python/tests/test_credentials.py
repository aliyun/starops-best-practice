"""
Tests for credentials — 凭据链加载
"""

import os

import pytest

from starops_sdk_samples.client import credentials as credentials_module
from starops_sdk_samples.client.credentials import load_credentials_from_chain


class _FakeCredential:
    def __init__(self, ak, sk, token=None):
        self.access_key_id = ak
        self.access_key_secret = sk
        self.security_token = token


class _FakeCredentialClient:
    """模拟 alibabacloud_credentials 客户端"""

    def __init__(self, credential=None, init_error=None):
        if init_error:
            raise init_error
        self._credential = credential

    def get_credential(self):
        return self._credential


def test_load_from_env_vars(monkeypatch):
    """设置 AK/SK 环境变量时，凭据链（环境变量 provider）应返回对应凭证

    mock CredentialClient 模拟环境变量 provider 行为，
    避免受本机 CLI/配置文件等其他凭据源干扰。
    """
    monkeypatch.setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "env-test-ak")
    monkeypatch.setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "env-test-sk")
    monkeypatch.delenv("ALIBABA_CLOUD_SECURITY_TOKEN", raising=False)

    def _env_provider_client(*args, **kwargs):
        return _FakeCredentialClient(
            credential=_FakeCredential(
                os.getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"),
                os.getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"),
            )
        )

    monkeypatch.setattr(credentials_module, "CredentialClient", _env_provider_client)

    access_key_id, access_key_secret = load_credentials_from_chain()
    assert access_key_id == "env-test-ak"
    assert access_key_secret == "env-test-sk"


def test_no_credential_source_raises(monkeypatch):
    """无任何凭据源时（客户端初始化失败），应抛出异常"""
    def _raise_client(*args, **kwargs):
        raise Exception("unable to find credentials")

    monkeypatch.setattr(credentials_module, "CredentialClient", _raise_client)

    with pytest.raises(Exception, match="unable to find credentials"):
        load_credentials_from_chain()


def test_empty_credential_raises(monkeypatch):
    """凭据链返回空凭证时，应抛出明确异常"""
    fake = _FakeCredentialClient(credential=_FakeCredential(None, None))
    monkeypatch.setattr(
        credentials_module, "CredentialClient", lambda *a, **k: fake
    )

    with pytest.raises(Exception, match="凭据链返回的凭证为空"):
        load_credentials_from_chain()


def test_credential_without_security_token(monkeypatch):
    """长期 AK/SK 场景（security_token 为 None）不应报错"""
    fake = _FakeCredentialClient(
        credential=_FakeCredential("chain-ak", "chain-sk", token=None)
    )
    monkeypatch.setattr(
        credentials_module, "CredentialClient", lambda *a, **k: fake
    )

    access_key_id, access_key_secret = load_credentials_from_chain()
    assert access_key_id == "chain-ak"
    assert access_key_secret == "chain-sk"
