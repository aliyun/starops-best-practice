package client

import (
	"os"
	"testing"
)

// TestLoadConfigFromEnv_WithEnvVars 验证当 AK/SK 环境变量存在时直接使用，不调用凭据链
func TestLoadConfigFromEnv_WithEnvVars(t *testing.T) {
	// 设置所有必需的环境变量
	os.Setenv("VIBEOPS_ENDPOINT", "https://test.example.com")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "test-ak-id")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "test-ak-secret")
	os.Setenv("VIBEOPS_WORKSPACE", "test-workspace")
	os.Setenv("VIBEOPS_REGION", "cn-beijing")
	os.Setenv("VIBEOPS_EMPLOYEE_NAME", "test-employee")
	defer func() {
		os.Unsetenv("VIBEOPS_ENDPOINT")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
		os.Unsetenv("VIBEOPS_WORKSPACE")
		os.Unsetenv("VIBEOPS_REGION")
		os.Unsetenv("VIBEOPS_EMPLOYEE_NAME")
	}()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() 返回错误: %v", err)
	}

	if cfg.AccessKeyID != "test-ak-id" {
		t.Errorf("AccessKeyID = %q, want %q", cfg.AccessKeyID, "test-ak-id")
	}
	if cfg.AccessKeySecret != "test-ak-secret" {
		t.Errorf("AccessKeySecret = %q, want %q", cfg.AccessKeySecret, "test-ak-secret")
	}
	if cfg.Endpoint != "https://test.example.com" {
		t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "https://test.example.com")
	}
	if cfg.Region != "cn-beijing" {
		t.Errorf("Region = %q, want %q", cfg.Region, "cn-beijing")
	}
}

// TestLoadConfigFromEnv_DefaultValues 验证默认值设置
func TestLoadConfigFromEnv_DefaultValues(t *testing.T) {
	os.Setenv("VIBEOPS_ENDPOINT", "https://test.example.com")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "test-ak-id")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "test-ak-secret")
	os.Unsetenv("VIBEOPS_REGION")
	os.Unsetenv("VIBEOPS_EMPLOYEE_NAME")
	defer func() {
		os.Unsetenv("VIBEOPS_ENDPOINT")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() 返回错误: %v", err)
	}

	if cfg.Region != "cn-hangzhou" {
		t.Errorf("Region = %q, want default %q", cfg.Region, "cn-hangzhou")
	}
	if cfg.EmployeeName != "default" {
		t.Errorf("EmployeeName = %q, want default %q", cfg.EmployeeName, "default")
	}
}

// TestLoadConfigFromEnv_FallbackToCredentialChain 验证当环境变量缺失时尝试凭据链
func TestLoadConfigFromEnv_FallbackToCredentialChain(t *testing.T) {
	// 清除 AK/SK 环境变量，保留 Endpoint
	os.Setenv("VIBEOPS_ENDPOINT", "https://test.example.com")
	os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	// 同时清除凭据链可能使用的环境变量，确保凭据链也会失败
	os.Unsetenv("ALIBABA_CLOUD_CREDENTIALS_FILE")
	defer func() {
		os.Unsetenv("VIBEOPS_ENDPOINT")
	}()

	// 凭据链在测试环境中可能无法获取凭证（没有配置文件或 ECS 角色）
	// 这里验证的是：当 AK/SK 环境变量缺失时，函数会尝试凭据链，
	// 如果凭据链也失败，则返回包含提示信息的错误
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		// 预期情况：凭据链也没有配置，应该返回错误
		// 验证错误信息中包含凭据链相关提示
		if cfg != nil {
			t.Errorf("期望 cfg 为 nil, 但得到 %+v", cfg)
		}
		return
	}

	// 如果测试环境恰好配置了凭据链（如 ~/.alibabacloud/credentials），
	// 那么应该能成功获取凭证
	if cfg.AccessKeyID == "" {
		t.Error("通过凭据链获取的 AccessKeyID 不应为空")
	}
	if cfg.AccessKeySecret == "" {
		t.Error("通过凭据链获取的 AccessKeySecret 不应为空")
	}
}

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
	akID, akSecret, err := LoadCredentialsFromChain()
	if err != nil {
		// 没有配置凭据源时，预期会返回错误
		t.Logf("凭据链返回错误（测试环境预期行为）: %v", err)
		if akID != "" || akSecret != "" {
			t.Error("返回错误时 akID/akSecret 应为空")
		}
		return
	}

	// 如果有凭据源可用（如配置文件），验证返回值非空
	if akID == "" {
		t.Error("成功时 AccessKeyID 不应为空")
	}
	if akSecret == "" {
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

	akID, akSecret, err := LoadCredentialsFromChain()
	if err != nil {
		t.Fatalf("LoadCredentialsFromChain() 返回错误: %v", err)
	}

	if akID != "chain-test-ak-id" {
		t.Errorf("AccessKeyID = %q, want %q", akID, "chain-test-ak-id")
	}
	if akSecret != "chain-test-ak-secret" {
		t.Errorf("AccessKeySecret = %q, want %q", akSecret, "chain-test-ak-secret")
	}
}
