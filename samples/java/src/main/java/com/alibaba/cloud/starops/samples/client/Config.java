package com.alibaba.cloud.starops.samples.client;

import java.util.ArrayList;
import java.util.List;

import io.github.cdimascio.dotenv.Dotenv;

/**
 * 应用配置
 * Application configuration
 */
public class Config {
    private String workspace;
    private String endpoint;
    private String region;
    private String accessKeyId;
    private String accessKeySecret;
    private String employeeName;
    private RetryConfig retryConfig = RetryConfig.getDefault();
    private boolean simulateNetworkError = false;
    private boolean mockMode = false;
    private boolean recordMode = false;
    public String mockFile;
    public String mockInput;

    public Config() {}

    /**
     * 从环境变量加载配置
     * Load configuration from environment variables
     */
    public static Config loadFromEnv() throws SDKException {
        Dotenv dotenv = Dotenv.configure()
                .ignoreIfMissing()
                .load();

        Config cfg = new Config();
        cfg.workspace = getEnvValue(dotenv, "STAROPS_WORKSPACE");
        cfg.endpoint = getEnvValue(dotenv, "STAROPS_ENDPOINT");
        cfg.region = getEnvValue(dotenv, "STAROPS_REGION");
        cfg.accessKeyId = getEnvValue(dotenv, "ALIBABA_CLOUD_ACCESS_KEY_ID");
        cfg.accessKeySecret = getEnvValue(dotenv, "ALIBABA_CLOUD_ACCESS_KEY_SECRET");
        cfg.employeeName = getEnvValue(dotenv, "STAROPS_EMPLOYEE_NAME");

        // AK/SK 为空时回退阿里云默认凭据链（环境变量 > OIDC > CLI配置 > 配置文件 > IAM角色）
        if (cfg.accessKeyId == null || cfg.accessKeyId.isEmpty()
                || cfg.accessKeySecret == null || cfg.accessKeySecret.isEmpty()) {
            try {
                String[] creds = Credentials.loadFromChain();
                cfg.accessKeyId = creds[0];
                cfg.accessKeySecret = creds[1];
            } catch (Exception e) {
                System.err.println("凭据链加载失败: " + e.getMessage() + "，请手动设置环境变量");
            }
        }

        // Validate required fields
        List<String> missingVars = new ArrayList<>();
        if (cfg.endpoint == null || cfg.endpoint.isEmpty()) {
            missingVars.add("STAROPS_ENDPOINT");
        }
        if (cfg.accessKeyId == null || cfg.accessKeyId.isEmpty()) {
            missingVars.add("ALIBABA_CLOUD_ACCESS_KEY_ID");
        }
        if (cfg.accessKeySecret == null || cfg.accessKeySecret.isEmpty()) {
            missingVars.add("ALIBABA_CLOUD_ACCESS_KEY_SECRET");
        }

        if (!missingVars.isEmpty()) {
            throw SDKException.configMissing(missingVars);
        }

        // Set defaults
        if (cfg.employeeName == null || cfg.employeeName.isEmpty()) {
            cfg.employeeName = "apsara-ops";
        }
        if (cfg.region == null || cfg.region.isEmpty()) {
            cfg.region = "cn-beijing";
        }

        // 加载重试配置（支持环境变量优先，.env fallback）
        cfg.retryConfig = RetryConfig.loadFromEnv(dotenv);

        return cfg;
    }

    private static String getEnvValue(Dotenv dotenv, String key) {
        String value = System.getenv(key);
        if ((value == null || value.isEmpty()) && dotenv != null) {
            value = dotenv.get(key);
        }
        return value;
    }

    // Getters and Setters
    public String getWorkspace() {
        return workspace;
    }

    public void setWorkspace(String workspace) {
        this.workspace = workspace;
    }

    public String getEndpoint() {
        return endpoint;
    }

    public void setEndpoint(String endpoint) {
        this.endpoint = endpoint;
    }

    public String getRegion() {
        return region;
    }

    public void setRegion(String region) {
        this.region = region;
    }

    public String getAccessKeyId() {
        return accessKeyId;
    }

    public void setAccessKeyId(String accessKeyId) {
        this.accessKeyId = accessKeyId;
    }

    public String getAccessKeySecret() {
        return accessKeySecret;
    }

    public void setAccessKeySecret(String accessKeySecret) {
        this.accessKeySecret = accessKeySecret;
    }

    public String getEmployeeName() {
        return employeeName;
    }

    public void setEmployeeName(String employeeName) {
        this.employeeName = employeeName;
    }

    public RetryConfig getRetryConfig() {
        return retryConfig;
    }

    public void setRetryConfig(RetryConfig retryConfig) {
        this.retryConfig = retryConfig;
    }

    public boolean isSimulateNetworkError() {
        return simulateNetworkError;
    }

    public void setSimulateNetworkError(boolean simulateNetworkError) {
        this.simulateNetworkError = simulateNetworkError;
    }

    public boolean isMockMode() { return mockMode; }
    public void setMockMode(boolean mockMode) { this.mockMode = mockMode; }
    public boolean isRecordMode() { return recordMode; }
    public void setRecordMode(boolean recordMode) { this.recordMode = recordMode; }
    public String getMockFile() { return mockFile; }
    public void setMockFile(String mockFile) { this.mockFile = mockFile; }
    public String getMockInput() { return mockInput; }
    public void setMockInput(String mockInput) { this.mockInput = mockInput; }
}
