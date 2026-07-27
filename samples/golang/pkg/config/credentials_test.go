package config

import (
	"os"
	"testing"

	credentials "github.com/aliyun/credentials-go/credentials"
)

// TestLoadCredentialsFromChain_BasicCall 验证凭据链函数的基本调用逻辑
func TestLoadCredentialsFromChain_BasicCall(t *testing.T) {
	// 清除所有可能影响凭据链的环境变量
	envVarsToClean := []string{
		"ALIBABA_CLOUD_ACCESS_KEY_ID",
		"ALIBABA_CLOUD_ACCESS_KEY_SECRET",
		"ALIBABA_CLOUD_SECURITY_TOKEN",
		"ALIBABA_CLOUD_ROLE_ARN",
		"ALIBABA_CLOUD_OIDC_PROVIDER_ARN",
		"ALIBABA_CLOUD_OIDC_TOKEN_FILE",
		"ALIBABA_CLOUD_ECS_METADATA",
		"ALIBABA_CLOUD_CREDENTIALS_URI",
	}
	savedEnvs := make(map[string]string)
	for _, key := range envVarsToClean {
		savedEnvs[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, val := range savedEnvs {
			if val != "" {
				os.Setenv(key, val)
			}
		}
	}()

	// 在干净的环境下调用凭据链，验证函数能正常执行（不 panic）
	credential, err := LoadCredentialsFromChain()
	if err != nil {
		// 没有配置凭据源时，预期会返回错误
		t.Logf("凭据链返回错误（测试环境预期行为）: %v", err)
		if credential.AccessKeyID != "" || credential.AccessKeySecret != "" {
			t.Error("返回错误时 akID/akSecret 应为空")
		}
		return
	}

	// 如果有凭据源可用（如配置文件），验证返回值非空
	if credential.AccessKeyID == "" {
		t.Error("成功时 AccessKeyID 不应为空")
	}
	if credential.AccessKeySecret == "" {
		t.Error("成功时 AccessKeySecret 不应为空")
	}
}

// TestLoadCredentialsFromChain_WithEnvCredentials 验证凭据链通过环境变量获取凭证
func TestLoadCredentialsFromChain_WithEnvCredentials(t *testing.T) {
	// 设置凭据链会读取的环境变量
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "chain-test-ak-id")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "chain-test-ak-secret")
	defer func() {
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}()

	credential, err := LoadCredentialsFromChain()
	if err != nil {
		t.Fatalf("LoadCredentialsFromChain() 返回错误: %v", err)
	}

	if credential.AccessKeyID != "chain-test-ak-id" {
		t.Errorf("AccessKeyID = %q, want %q", credential.AccessKeyID, "chain-test-ak-id")
	}
	if credential.AccessKeySecret != "chain-test-ak-secret" {
		t.Errorf("AccessKeySecret = %q, want %q", credential.AccessKeySecret, "chain-test-ak-secret")
	}
}

// TestConvertCredentialModel_NilSecurityToken 验证长期 AK/SK（非 STS）场景下
// SecurityToken 为 nil 时不 panic，且返回空 token
func TestConvertCredentialModel_NilSecurityToken(t *testing.T) {
	akID := "test-ak-id"
	akSecret := "test-ak-secret"
	credValue := &credentials.CredentialModel{
		AccessKeyId:     &akID,
		AccessKeySecret: &akSecret,
		SecurityToken:   nil,
	}

	credential, err := convertCredentialModel(credValue)
	if err != nil {
		t.Fatalf("convertCredentialModel() 返回错误: %v", err)
	}
	if credential.AccessKeyID != akID {
		t.Errorf("AccessKeyID = %q, want %q", credential.AccessKeyID, akID)
	}
	if credential.AccessKeySecret != akSecret {
		t.Errorf("AccessKeySecret = %q, want %q", credential.AccessKeySecret, akSecret)
	}
	if credential.SecurityToken != "" {
		t.Errorf("SecurityToken = %q, want 空字符串", credential.SecurityToken)
	}
}

// TestConvertCredentialModel_WithSecurityToken 验证 STS 场景下 SecurityToken 正常返回
func TestConvertCredentialModel_WithSecurityToken(t *testing.T) {
	akID := "test-ak-id"
	akSecret := "test-ak-secret"
	token := "test-sts-token"
	credValue := &credentials.CredentialModel{
		AccessKeyId:     &akID,
		AccessKeySecret: &akSecret,
		SecurityToken:   &token,
	}

	credential, err := convertCredentialModel(credValue)
	if err != nil {
		t.Fatalf("convertCredentialModel() 返回错误: %v", err)
	}
	if credential.SecurityToken != token {
		t.Errorf("SecurityToken = %q, want %q", credential.SecurityToken, token)
	}
}

// TestConvertCredentialModel_NilCredential 验证凭证为空时返回错误而非 panic
func TestConvertCredentialModel_NilCredential(t *testing.T) {
	if _, err := convertCredentialModel(nil); err == nil {
		t.Error("convertCredentialModel(nil) 应返回错误")
	}

	akID := "test-ak-id"
	if _, err := convertCredentialModel(&credentials.CredentialModel{AccessKeyId: &akID}); err == nil {
		t.Error("AccessKeySecret 为 nil 时应返回错误")
	}
}
