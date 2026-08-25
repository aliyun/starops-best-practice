package com.alibaba.cloud.starops.samples.client;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * 模型配置加载与辅助方法
 * Model configuration loading and helper methods
 */
public class ModelsConfig {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private List<Provider> providers;

    public ModelsConfig() {
        this.providers = new ArrayList<>();
    }

    public List<Provider> getProviders() {
        return providers;
    }

    public void setProviders(List<Provider> providers) {
        this.providers = providers;
    }

    // --- POJO ---

    @JsonIgnoreProperties(ignoreUnknown = true)
    public static class Provider {
        private String provider;
        private String displayName;
        private List<Model> models;

        public Provider() {
            this.models = new ArrayList<>();
        }

        public String getProvider() { return provider; }
        public void setProvider(String provider) { this.provider = provider; }

        public String getDisplayName() { return displayName; }
        public void setDisplayName(String displayName) { this.displayName = displayName; }

        public List<Model> getModels() { return models; }
        public void setModels(List<Model> models) { this.models = models; }
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    public static class Model {
        @JsonProperty("modelId")
        private String modelId;
        private String displayName;
        private String shortName;

        public String getModelId() { return modelId; }
        public void setModelId(String modelId) { this.modelId = modelId; }

        public String getDisplayName() { return displayName; }
        public void setDisplayName(String displayName) { this.displayName = displayName; }

        public String getShortName() { return shortName; }
        public void setShortName(String shortName) { this.shortName = shortName; }
    }

    // --- Static methods ---

    /**
     * 加载模型配置
     * Load models config from the given path
     */
    public static ModelsConfig loadModels(String configPath) {
        try {
            byte[] data = Files.readAllBytes(Paths.get(configPath));
            return MAPPER.readValue(data, ModelsConfig.class);
        } catch (IOException e) {
            return null;
        }
    }

    /**
     * 定位 config/models.json 文件
     * Priority: STAROPS_MODELS_CONFIG env > walk up from cwd > default
     */
    public static String findConfigPath() {
        String envPath = System.getenv("STAROPS_MODELS_CONFIG");
        if (envPath != null && !envPath.isEmpty()) {
            return envPath;
        }

        try {
            Path dir = Paths.get("").toAbsolutePath();
            while (dir != null) {
                Path candidate = dir.resolve("config").resolve("models.json");
                if (Files.exists(candidate)) {
                    return candidate.toString();
                }
                dir = dir.getParent();
            }
        } catch (Exception ignored) {
        }

        return "config/models.json";
    }

    /**
     * 解析 "provider:modelId" 格式
     * Parse model flag in "provider:modelId" format
     * @return String[2] = {provider, modelId}
     */
    public static String[] parseModelFlag(String flag) {
        if (flag == null || flag.isEmpty()) {
            return null;
        }
        int idx = flag.indexOf(':');
        if (idx <= 0 || idx >= flag.length() - 1) {
            return null;
        }
        String provider = flag.substring(0, idx);
        String modelId = flag.substring(idx + 1);
        if (provider.isEmpty() || modelId.isEmpty()) {
            return null;
        }
        return new String[]{provider, modelId};
    }

    /**
     * 校验 provider + modelId 组合是否存在于配置中
     * Validate that the provider:modelId combination exists in config
     */
    public static boolean validateModel(ModelsConfig cfg, String provider, String modelId) {
        if (cfg == null) return false;
        for (Provider p : cfg.getProviders()) {
            if (p.getProvider().equals(provider)) {
                for (Model m : p.getModels()) {
                    if (m.getModelId().equals(modelId)) {
                        return true;
                    }
                }
            }
        }
        return false;
    }

    /**
     * 生成模型 JSON 字符串: {"provider":"x","modelID":"y"}
     * Build model JSON string
     */
    public static String buildModelJson(String provider, String modelId) {
        // 注意: modelID 大写 ID
        return "{\"provider\":\"" + escapeJson(provider) + "\",\"modelID\":\"" + escapeJson(modelId) + "\"}";
    }

    /**
     * 生成完整 config JSON: {"model":{...},"disableThreadData":false}
     * Build full config JSON with model and disableThreadData
     */
    public static String buildConfigWithModel(String provider, String modelId) {
        return "{\"model\":{\"provider\":\"" + escapeJson(provider)
                + "\",\"modelID\":\"" + escapeJson(modelId)
                + "\"},\"disableThreadData\":false}";
    }

    /**
     * 返回编号菜单字符串（1-based）
     * Display numbered menu of all models (1-based)
     */
    public static String displayMenu(ModelsConfig cfg) {
        StringBuilder sb = new StringBuilder();
        sb.append("可选模型列表：\n");
        int idx = 1;
        for (Provider p : cfg.getProviders()) {
            for (Model m : p.getModels()) {
                sb.append(String.format("  %d. %s: %s\n", idx, p.getDisplayName(), m.getDisplayName()));
                idx++;
            }
        }
        return sb.toString();
    }

    /**
     * 返回所有有效 --model 取值及描述
     * List all valid --model flag values with descriptions
     */
    public static String listModelFlags(ModelsConfig cfg) {
        StringBuilder sb = new StringBuilder();
        sb.append("可传入 --model 的模型取值（格式: provider:modelId）：\n");
        for (Provider p : cfg.getProviders()) {
            for (Model m : p.getModels()) {
                String flag = p.getProvider() + ":" + m.getModelId();
                sb.append(String.format("  %-30s %s: %s\n", flag, p.getDisplayName(), m.getDisplayName()));
            }
        }
        return sb.toString();
    }

    /**
     * 按 0-based 索引获取 provider 和 modelId
     * Get model by 0-based flat index (same order as displayMenu)
     * @return String[2] = {provider, modelId}, or null if out of range
     */
    public static String[] getModelByIndex(ModelsConfig cfg, int index) {
        int idx = 0;
        for (Provider p : cfg.getProviders()) {
            for (Model m : p.getModels()) {
                if (idx == index) {
                    return new String[]{p.getProvider(), m.getModelId()};
                }
                idx++;
            }
        }
        return null;
    }

    private static String escapeJson(String s) {
        if (s == null) return "";
        return s.replace("\\", "\\\\").replace("\"", "\\\"");
    }
}
