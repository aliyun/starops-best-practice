package com.alibaba.cloud.starops.samples.client;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.BufferedReader;
import java.nio.charset.Charset;
import java.util.ArrayList;
import java.util.List;

/**
 * 模型配置管理
 * Model configuration management
 *
 * 从 config/models.json 加载可用模型列表，提供解析、校验、菜单展示等辅助方法。
 */
public class ModelsConfig {

    private final List<Provider> providers;

    public ModelsConfig(List<Provider> providers) {
        this.providers = providers;
    }

    public List<Provider> getProviders() {
        return providers;
    }

    // ======================== Data classes ========================

    public static class Model {
        private final String modelId;
        private final String displayName;
        private final String shortName;

        public Model(String modelId, String displayName, String shortName) {
            this.modelId = modelId;
            this.displayName = displayName;
            this.shortName = shortName;
        }

        public String getModelId() { return modelId; }
        public String getDisplayName() { return displayName; }
        public String getShortName() { return shortName; }
    }

    public static class Provider {
        private final String provider;
        private final String displayName;
        private final List<Model> models;

        public Provider(String provider, String displayName, List<Model> models) {
            this.provider = provider;
            this.displayName = displayName;
            this.models = models;
        }

        public String getProvider() { return provider; }
        public String getDisplayName() { return displayName; }
        public List<Model> getModels() { return models; }
    }

    // ======================== Static methods ========================

    /**
     * 从文件加载模型配置
     * Load models config from JSON file
     */
    public static ModelsConfig loadModels(String configPath) {
        File file = new File(configPath);
        if (!file.exists() || !file.isFile()) {
            return null;
        }
        try {
            String content = readFileToString(file);
            ObjectMapper mapper = new ObjectMapper();
            JsonNode root = mapper.readTree(content);
            JsonNode providersNode = root.get("providers");
            if (providersNode == null || !providersNode.isArray()) {
                return null;
            }

            List<Provider> providers = new ArrayList<Provider>();
            for (JsonNode pNode : providersNode) {
                String providerName = pNode.has("provider") ? pNode.get("provider").asText() : "";
                String displayName = pNode.has("displayName") ? pNode.get("displayName").asText() : "";
                List<Model> models = new ArrayList<Model>();
                JsonNode modelsNode = pNode.get("models");
                if (modelsNode != null && modelsNode.isArray()) {
                    for (JsonNode mNode : modelsNode) {
                        String modelId = mNode.has("modelId") ? mNode.get("modelId").asText() : "";
                        String mDisplayName = mNode.has("displayName") ? mNode.get("displayName").asText() : "";
                        String shortName = mNode.has("shortName") ? mNode.get("shortName").asText() : "";
                        models.add(new Model(modelId, mDisplayName, shortName));
                    }
                }
                providers.add(new Provider(providerName, displayName, models));
            }
            return new ModelsConfig(providers);
        } catch (Exception e) {
            return null;
        }
    }

    /**
     * 查找 config/models.json 路径
     * Priority: STAROPS_MODELS_CONFIG env > walk up from cwd > default
     */
    public static String findConfigPath() {
        String envPath = System.getenv("STAROPS_MODELS_CONFIG");
        if (envPath != null && !envPath.isEmpty()) {
            return envPath;
        }

        try {
            File dir = new File(System.getProperty("user.dir"));
            while (dir != null) {
                File candidate = new File(dir, "config" + File.separator + "models.json");
                if (candidate.exists() && candidate.isFile()) {
                    return candidate.getAbsolutePath();
                }
                dir = dir.getParentFile();
            }
        } catch (Exception ignored) {
        }

        return "config" + File.separator + "models.json";
    }

    /**
     * 解析 "provider:modelId" 格式
     * Parse model flag in "provider:modelId" format
     *
     * @return String[2]: {provider, modelId}
     * @throws IllegalArgumentException if format is invalid
     */
    public static String[] parseModelFlag(String flag) {
        if (flag == null || flag.isEmpty()) {
            throw new IllegalArgumentException(
                    "invalid model flag \"" + flag + "\": format should be provider:modelId");
        }
        int idx = flag.indexOf(':');
        if (idx <= 0 || idx >= flag.length() - 1) {
            throw new IllegalArgumentException(
                    "invalid model flag \"" + flag + "\": format should be provider:modelId");
        }
        String provider = flag.substring(0, idx);
        String modelId = flag.substring(idx + 1);
        return new String[]{provider, modelId};
    }

    /**
     * 校验 provider+modelId 是否存在于配置中
     * Validate if provider+modelId combination exists in config
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
     * 构建模型 JSON 字符串: {"provider":"x","modelID":"y"}
     * Build model JSON string
     */
    public static String buildModelJson(String provider, String modelId) {
        // 手动拼接以确保字段顺序和 modelID 大写 ID
        return "{\"provider\":\"" + escapeJson(provider) + "\",\"modelID\":\"" + escapeJson(modelId) + "\"}";
    }

    /**
     * 构建完整配置 JSON（含 model 和 disableThreadData）
     * Build full config JSON with model and disableThreadData
     */
    public static String buildConfigWithModel(String provider, String modelId) {
        return "{\"model\":{\"provider\":\"" + escapeJson(provider)
                + "\",\"modelID\":\"" + escapeJson(modelId)
                + "\"},\"disableThreadData\":false}";
    }

    /**
     * 返回 1-based 编号的模型菜单字符串
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
     * 列出所有 --model 可用值及描述
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
     * 通过 0-based 索引获取模型的 provider 和 modelId
     * Get provider and modelId by 0-based flat index (same order as displayMenu)
     *
     * @return String[2]: {provider, modelId}
     * @throws IllegalArgumentException if index out of range
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
        throw new IllegalArgumentException(
                "model index " + index + " out of range (total " + idx + " models)");
    }

    // ======================== Helpers ========================

    private static String readFileToString(File file) throws IOException {
        StringBuilder sb = new StringBuilder();
        FileInputStream fis = new FileInputStream(file);
        try {
            BufferedReader br = new BufferedReader(new InputStreamReader(fis, Charset.forName("UTF-8")));
            char[] buf = new char[4096];
            int n;
            while ((n = br.read(buf)) != -1) {
                sb.append(buf, 0, n);
            }
        } finally {
            fis.close();
        }
        return sb.toString();
    }

    private static String escapeJson(String value) {
        if (value == null) return "";
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < value.length(); i++) {
            char c = value.charAt(i);
            switch (c) {
                case '"': sb.append("\\\""); break;
                case '\\': sb.append("\\\\"); break;
                case '\n': sb.append("\\n"); break;
                case '\r': sb.append("\\r"); break;
                case '\t': sb.append("\\t"); break;
                default: sb.append(c);
            }
        }
        return sb.toString();
    }
}
