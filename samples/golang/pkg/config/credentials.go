// Package config — 阿里云凭据加载
// 职责：通过默认凭据链(环境变量 > OIDC > CLI > 配置文件 > IAM角色)获取 AK/SK。
// 不做：不做凭据刷新、不管理 STS Token 生命周期。
// 依赖：github.com/aliyun/credentials-go
package config

import (
	"fmt"

	credentials "github.com/aliyun/credentials-go/credentials"
)

type Credentials struct {
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
}

// LoadCredentialsFromChain 通过阿里云默认凭据链获取凭证
// 凭据链优先级：环境变量 > OIDC > CLI配置文件 > 配置文件(~/.alibabacloud/credentials) > IAM角色
func LoadCredentialsFromChain() (credential Credentials, err error) {
	cred, err := credentials.NewCredential(nil)
	if err != nil {
		return Credentials{}, fmt.Errorf("初始化凭据失败: %w", err)
	}

	credValue, err := cred.GetCredential()
	if err != nil {
		return Credentials{}, fmt.Errorf("获取凭证失败: %w", err)
	}

	return convertCredentialModel(credValue)
}

// convertCredentialModel 将凭据链返回的凭证模型转换为 Credentials
// 使用长期 AK/SK（非 STS）时 SecurityToken 为 nil，需判空避免解引用 panic
func convertCredentialModel(credValue *credentials.CredentialModel) (Credentials, error) {
	if credValue == nil || credValue.AccessKeyId == nil || credValue.AccessKeySecret == nil {
		return Credentials{}, fmt.Errorf("凭据链返回的凭证为空")
	}

	securityToken := ""
	if credValue.SecurityToken != nil {
		securityToken = *credValue.SecurityToken
	}

	return Credentials{
		AccessKeyID:     *credValue.AccessKeyId,
		AccessKeySecret: *credValue.AccessKeySecret,
		SecurityToken:   securityToken,
	}, nil
}
