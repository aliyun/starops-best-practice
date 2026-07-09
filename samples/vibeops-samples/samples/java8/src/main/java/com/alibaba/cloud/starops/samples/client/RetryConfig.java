package com.alibaba.cloud.starops.samples.client;

/**
 * SSE 重试配置（Java 8 版本，严格 JDK 8 语法）
 * SSE retry configuration
 *
 * 跨语言对齐 Go 参考实现（samples/golang/internal/client/retry.go），时间单位统一为毫秒。
 * Cross-language aligned with the Go reference; all durations are in milliseconds.
 */
public class RetryConfig {
    /** 最大重试次数，默认 10 / Max retries, default 10 */
    private int maxRetries;
    /** 初始退避时间(ms)，默认 1000 / Initial backoff in ms, default 1000 */
    private long initialBackoff;
    /** 最大退避时间(ms)，默认 30000 / Max backoff in ms, default 30000 */
    private long maxBackoff;
    /** 退避系数，默认 2.0 / Backoff factor, default 2.0 */
    private double backoffFactor;
    /** 空闲超时(ms)：超过此时长未收到任何消息视为连接中断，默认 60000 / Idle timeout in ms, default 60000 */
    private long idleTimeout;

    public RetryConfig() {
        this.maxRetries = 10;
        this.initialBackoff = 1000L;
        this.maxBackoff = 30000L;
        this.backoffFactor = 2.0;
        this.idleTimeout = 60000L;
    }

    /**
     * 返回默认重试配置
     * Return default retry config
     */
    public static RetryConfig defaultConfig() {
        return new RetryConfig();
    }

    /**
     * 从环境变量加载重试配置
     * Load retry config from environment variables
     * - VIBEOPS_MAX_RETRIES：最大重试次数
     * - VIBEOPS_IDLE_TIMEOUT：空闲超时（秒），内部统一转换为毫秒
     */
    public static RetryConfig loadFromEnv() {
        RetryConfig cfg = defaultConfig();

        String maxRetries = System.getenv("VIBEOPS_MAX_RETRIES");
        if (maxRetries != null && !maxRetries.trim().isEmpty()) {
            try {
                int n = Integer.parseInt(maxRetries.trim());
                if (n > 0) {
                    cfg.maxRetries = n;
                }
            } catch (NumberFormatException ignored) {
                // 忽略非法值，使用默认值 / ignore invalid value, keep default
            }
        }

        // 环境变量按秒配置，内部统一转换为毫秒
        String idleTimeout = System.getenv("VIBEOPS_IDLE_TIMEOUT");
        if (idleTimeout != null && !idleTimeout.trim().isEmpty()) {
            try {
                int n = Integer.parseInt(idleTimeout.trim());
                if (n > 0) {
                    cfg.idleTimeout = (long) n * 1000L;
                }
            } catch (NumberFormatException ignored) {
                // 忽略非法值，使用默认值 / ignore invalid value, keep default
            }
        }

        return cfg;
    }

    /**
     * 计算退避时间(ms)
     * min(initialBackoff * backoffFactor^(retryCount-1), maxBackoff)
     */
    public long calculateBackoff(int retryCount) {
        double backoff = initialBackoff * Math.pow(backoffFactor, retryCount - 1);
        if (backoff > maxBackoff) {
            return maxBackoff;
        }
        return (long) backoff;
    }

    // Getters and Setters
    public int getMaxRetries() {
        return maxRetries;
    }

    public void setMaxRetries(int maxRetries) {
        this.maxRetries = maxRetries;
    }

    public long getInitialBackoff() {
        return initialBackoff;
    }

    public void setInitialBackoff(long initialBackoff) {
        this.initialBackoff = initialBackoff;
    }

    public long getMaxBackoff() {
        return maxBackoff;
    }

    public void setMaxBackoff(long maxBackoff) {
        this.maxBackoff = maxBackoff;
    }

    public double getBackoffFactor() {
        return backoffFactor;
    }

    public void setBackoffFactor(double backoffFactor) {
        this.backoffFactor = backoffFactor;
    }

    public long getIdleTimeout() {
        return idleTimeout;
    }

    public void setIdleTimeout(long idleTimeout) {
        this.idleTimeout = idleTimeout;
    }
}
