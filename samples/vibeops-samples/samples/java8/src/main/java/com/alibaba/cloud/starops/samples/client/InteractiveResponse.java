package com.alibaba.cloud.starops.samples.client;

import com.alibaba.cloud.starops.samples.types.InteractionType;
import java.util.Map;

/**
 * 交互响应
 * Interactive response
 */
public class InteractiveResponse {
    private final String callId;
    private final InteractionType type;
    private final Map<String, Object> response;
    private Map<String, Object> source;
    private Map<String, Object> modifiedData;
    private Map<String, Object> formData;
    private String decision;

    public InteractiveResponse(String callId, InteractionType type, Map<String, Object> response) {
        this.callId = callId;
        this.type = type;
        this.response = response;
    }

    public String getCallId() {
        return callId;
    }

    public InteractionType getType() {
        return type;
    }

    public Map<String, Object> getResponse() {
        return response;
    }

    public Map<String, Object> getSource() {
        return source;
    }

    public void setSource(Map<String, Object> source) {
        this.source = source;
    }

    public Map<String, Object> getModifiedData() {
        return modifiedData;
    }

    public void setModifiedData(Map<String, Object> modifiedData) {
        this.modifiedData = modifiedData;
    }

    public Map<String, Object> getFormData() {
        return formData;
    }

    public void setFormData(Map<String, Object> formData) {
        this.formData = formData;
    }

    public String getDecision() {
        return decision;
    }

    public void setDecision(String decision) {
        this.decision = decision;
    }

    @Override
    public String toString() {
        return "InteractiveResponse{" +
                "callId='" + callId + '\'' +
                ", type=" + type +
                ", response=" + response +
                ", decision='" + decision + '\'' +
                '}';
    }
}