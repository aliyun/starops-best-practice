"""默认凭据链支持 - 与 Go 实现保持一致"""
from alibabacloud_credentials.client import Client as CredentialClient


def load_credentials_from_chain() -> tuple[str, str]:
    """
    通过阿里云默认凭据链获取凭证
    凭据链优先级：环境变量 > OIDC > CLI配置文件 > 配置文件(~/.alibabacloud/credentials) > IAM角色

    Returns:
        tuple: (access_key_id, access_key_secret)
    Raises:
        Exception: 凭据链初始化或获取凭证失败
    """
    # None config triggers default credential chain
    cred = CredentialClient()
    credential = cred.get_credential()
    access_key_id = credential.access_key_id
    access_key_secret = credential.access_key_secret
    if not access_key_id or not access_key_secret:
        raise Exception("凭据链返回的凭证为空")
    return access_key_id, access_key_secret
