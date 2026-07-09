package com.alibaba.cloud.starops.samples.types;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

/**
 * 工具调用详情
 * Tool call details
 */
public class ItemTool {
    /** 工具调用的唯一标识 / Unique identifier for the tool call */
    @JsonProperty("id")
    private String id;

    /** 工具名称 / Tool name */
    @JsonProperty("name")
    private String name;

    /** 本次调用的上下文 ID / Context ID for this call */
    @JsonProperty("toolCallId")
    private String toolCallId;

    /** 工具调用参数增量 / Tool call arguments delta */
    @JsonProperty("argumentsDelta")
    private String argumentsDelta;

    /** 工具调用参数 / Tool call arguments */
    @JsonProperty("arguments")
    private Object arguments;

    /** 执行阶段状态 / Execution status */
    @JsonProperty("status")
    private ItemStatus status;

    /** 工具执行结果输出 / Tool execution result output */
    @JsonProperty("contents")
    private List<ItemContent> contents;

    public ItemTool() {}

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getToolCallId() {
        return toolCallId;
    }

    public void setToolCallId(String toolCallId) {
        this.toolCallId = toolCallId;
    }

    public String getArgumentsDelta() {
        return argumentsDelta;
    }

    public void setArgumentsDelta(String argumentsDelta) {
        this.argumentsDelta = argumentsDelta;
    }

    public Object getArguments() {
        return arguments;
    }

    public void setArguments(Object arguments) {
        this.arguments = arguments;
    }

    public ItemStatus getStatus() {
        return status;
    }

    public void setStatus(ItemStatus status) {
        this.status = status;
    }

    public List<ItemContent> getContents() {
        return contents;
    }

    public void setContents(List<ItemContent> contents) {
        this.contents = contents;
    }
}
