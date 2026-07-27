package config

import (
	"os"
	"testing"
)

// TestLoadFromEnv_WithEnvVars 验证当 AK/SK 环境变量存在时直接使用
func TestLoadFromEnv_WithEnvVars(t *testing.T) {
	// 设置所有必需的环境变量
	os.Setenv("STAROPS_ENDPOINT", "https://test.example.com")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "test-ak-id")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "test-ak-secret")
	os.Setenv("STAROPS_WORKSPACE", "test-workspace")
	os.Setenv("STAROPS_REGION", "cn-beijing")
	os.Setenv("STAROPS_EMPLOYEE_NAME", "test-employee")
	defer func() {
		os.Unsetenv("STAROPS_ENDPOINT")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
		os.Unsetenv("STAROPS_WORKSPACE")
		os.Unsetenv("STAROPS_REGION")
		os.Unsetenv("STAROPS_EMPLOYEE_NAME")
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

// TestLoadFromEnv_DefaultValues 验证默认值设置
func TestLoadFromEnv_DefaultValues(t *testing.T) {
	os.Setenv("STAROPS_ENDPOINT", "https://test.example.com")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "test-ak-id")
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "test-ak-secret")
	os.Unsetenv("STAROPS_REGION")
	os.Unsetenv("STAROPS_EMPLOYEE_NAME")
	defer func() {
		os.Unsetenv("STAROPS_ENDPOINT")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
		os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() 返回错误: %v", err)
	}

	if cfg.Region != "cn-beijing" {
		t.Errorf("Region = %q, want default %q", cfg.Region, "cn-beijing")
	}
	if cfg.EmployeeName != "apsara-ops" {
		t.Errorf("EmployeeName = %q, want %q", cfg.EmployeeName, "apsara-ops")
	}
}

// TestLoadFromEnv_FallbackToCredentialChain 验证当环境变量缺失时尝试凭据链
func TestLoadFromEnv_FallbackToCredentialChain(t *testing.T) {
	// 清除 AK/SK 环境变量，保留 Endpoint
	os.Setenv("STAROPS_ENDPOINT", "https://test.example.com")
	os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	os.Unsetenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	// 同时清除凭据链可能使用的环境变量，确保凭据链也会失败
	os.Unsetenv("ALIBABA_CLOUD_CREDENTIALS_FILE")
	defer func() {
		os.Unsetenv("STAROPS_ENDPOINT")
	}()

	// 凭据链在测试环境中可能无法获取凭证（没有配置文件或 ECS 角色）
	// 这里验证的是：当 AK/SK 环境变量缺失时，函数会尝试凭据链，
	// 如果凭据链也失败，则返回包含提示信息的错误
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		// 预期情况：凭据链也没有配置，应该返回错误
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
