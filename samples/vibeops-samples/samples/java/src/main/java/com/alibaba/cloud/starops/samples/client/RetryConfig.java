package com.alibaba.cloud.starops.samples.client;

/**
 * SSE 重试配置
 * SSE retry configuration
 */
public class RetryConfig {
    /** 最大重试次数，默认10 */
    private int maxRetries = 10;
    /** 初始退避时间(ms)，默认1s */
    private long initialBackoffMs = 1000;
    /** 最大退避时间(ms)，默认30s */
    private long maxBackoffMs = 30000;
    /** 退避系数，默认2.0 */
    private double backoffFactor = 2.0;
    /** 空闲超时(ms)：超过此时长未收到任何消息视为连接中断，默认60s */
    private long idleTimeoutMs = 60000;

    public RetryConfig() {}

    /**
     * 返回默认重试配置
     * Return default retry configuration
     */
    public static RetryConfig getDefault() {
        return new RetryConfig();
    }

    /**
     * 从环境变量加载重试配置
     * Load retry configuration from environment variables
     */
    public static RetryConfig loadFromEnv() {
        RetryConfig cfg = getDefault();

        String maxRetries = System.getenv("VIBEOPS_MAX_RETRIES");
        if (maxRetries != null && !maxRetries.isEmpty()) {
            try {
                int n = Integer.parseInt(maxRetries.trim());
                if (n > 0) {
                    cfg.maxRetries = n;
                }
            } catch (NumberFormatException ignored) {}
        }

        String idleTimeout = System.getenv("VIBEOPS_IDLE_TIMEOUT");
        if (idleTimeout != null && !idleTimeout.isEmpty()) {
            try {
                long n = Long.parseLong(idleTimeout.trim());
                if (n > 0) {
                    cfg.idleTimeoutMs = n * 1000L;
                }
            } catch (NumberFormatException ignored) {}
        }

        return cfg;
    }

    // Getters and Setters
    public int getMaxRetries() {
        return maxRetries;
    }

    public void setMaxRetries(int maxRetries) {
        this.maxRetries = maxRetries;
    }

    public long getInitialBackoffMs() {
        return initialBackoffMs;
    }

    public void setInitialBackoffMs(long initialBackoffMs) {
        this.initialBackoffMs = initialBackoffMs;
    }

    public long getMaxBackoffMs() {
        return maxBackoffMs;
    }

    public void setMaxBackoffMs(long maxBackoffMs) {
        this.maxBackoffMs = maxBackoffMs;
    }

    public double getBackoffFactor() {
        return backoffFactor;
    }

    public void setBackoffFactor(double backoffFactor) {
        this.backoffFactor = backoffFactor;
    }

    public long getIdleTimeoutMs() {
        return idleTimeoutMs;
    }

    public void setIdleTimeoutMs(long idleTimeoutMs) {
        this.idleTimeoutMs = idleTimeoutMs;
    }
}
