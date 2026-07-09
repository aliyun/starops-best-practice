// Package client 提供 VibeOps Agent 客户端的公共实现
package client

import (
	"fmt"

	credentials "github.com/aliyun/credentials-go/credentials"
)

// LoadCredentialsFromChain 通过阿里云默认凭据链获取凭证
// 凭据链优先级：环境变量 > OIDC > CLI配置文件 > 配置文件(~/.alibabacloud/credentials) > IAM角色
func LoadCredentialsFromChain() (accessKeyID, accessKeySecret string, err error) {
	cred, err := credentials.NewCredential(nil)
	if err != nil {
		return "", "", fmt.Errorf("初始化凭据失败: %w", err)
	}

	credValue, err := cred.GetCredential()
	if err != nil {
		return "", "", fmt.Errorf("获取凭证失败: %w", err)
	}

	if credValue.AccessKeyId == nil || credValue.AccessKeySecret == nil {
		return "", "", fmt.Errorf("凭据链返回的凭证为空")
	}

	return *credValue.AccessKeyId, *credValue.AccessKeySecret, nil
}
