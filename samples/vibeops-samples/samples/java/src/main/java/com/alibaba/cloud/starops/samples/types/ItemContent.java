package com.alibaba.cloud.starops.samples.types;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * 消息内容
 * Message content
 */
public class ItemContent {
    /** 内容类型 / Content type */
    @JsonProperty("type")
    private ContentType type;

    /** 内容值 / Content value */
    @JsonProperty("value")
    private String value;

    /** 是否是追加内容 / Whether this is appended content */
    @JsonProperty("append")
    private boolean append;

    /** 是否是最后一个 chunk / Whether this is the last chunk */
    @JsonProperty("lastChunk")
    private boolean lastChunk;

    public ItemContent() {}

    public ItemContent(ContentType type, String value) {
        this.type = type;
        this.value = value;
    }

    public ContentType getType() {
        return type;
    }

    public void setType(ContentType type) {
        this.type = type;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public boolean isAppend() {
        return append;
    }

    public void setAppend(boolean append) {
        this.append = append;
    }

    public boolean isLastChunk() {
        return lastChunk;
    }

    public void setLastChunk(boolean lastChunk) {
        this.lastChunk = lastChunk;
    }
}
