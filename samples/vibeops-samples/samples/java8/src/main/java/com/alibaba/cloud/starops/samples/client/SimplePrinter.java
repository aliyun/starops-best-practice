package com.alibaba.cloud.starops.samples.client;

import com.fasterxml.jackson.databind.JsonNode;
import java.util.HashSet;
import java.util.Set;

/**
 * 简洁模式打印器
 * Simple mode printer - only outputs final assistant text content
 */
public class SimplePrinter {
    private final StringBuilder buffer;
    private final Set<String> seenArtifacts;

    public SimplePrinter() {
        this.buffer = new StringBuilder();
        this.seenArtifacts = new HashSet<>();
    }

    /**
     * 处理事件，提取文本内容
     * Process event and extract text content
     */
    public String processEvent(ChatEvent event) {
        if (event == null || event.getBody() == null) {
            return "";
        }

        // 利用 event 字段快速跳过非文本事件
        String eventType = event.getEvent();
        if (eventType != null && !"text".equals(eventType) && !"task_finished".equals(eventType)) {
            return "";
        }

        StringBuilder extracted = new StringBuilder();
        JsonNode body = event.getBody();

        if (body.has("messages") && body.get("messages").isArray()) {
            for (JsonNode msg : body.get("messages")) {
                // Only process system role messages (contains final result artifacts)
                if (msg.has("role") && "system".equals(msg.get("role").asText())) {
                    String text = extractTextFromArtifacts(msg.get("artifacts"));
                    if (!text.isEmpty() && !seenArtifacts.contains(text)) {
                        seenArtifacts.add(text);
                        extracted.append(text);
                        buffer.append(text);
                    }
                }
            }
        }

        return extracted.toString();
    }

    /**
     * 从 artifacts 中提取文本
     * Extract text from artifacts
     */
    private String extractTextFromArtifacts(JsonNode artifacts) {
        if (artifacts == null || !artifacts.isArray()) {
            return "";
        }

        StringBuilder result = new StringBuilder();
        for (JsonNode artifact : artifacts) {
            if (artifact.has("parts") && artifact.get("parts").isArray()) {
                for (JsonNode part : artifact.get("parts")) {
                    if (part.has("kind") && "text".equals(part.get("kind").asText())) {
                        if (part.has("text")) {
                            result.append(part.get("text").asText());
                        }
                    }
                }
            }
        }
        return result.toString();
    }

    /**
     * 获取最终文本
     * Get final text
     */
    public String getFinalText() {
        return buffer.toString();
    }

    /**
     * 重置缓冲区
     * Reset buffer
     */
    public void reset() {
        buffer.setLength(0);
        seenArtifacts.clear();
    }
}
