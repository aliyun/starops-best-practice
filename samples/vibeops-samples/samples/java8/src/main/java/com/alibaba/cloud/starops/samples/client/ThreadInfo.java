package com.alibaba.cloud.starops.samples.client;

/**
 * 会话信息
 * Thread information
 */
public class ThreadInfo {
    private String threadId;
    private String title;
    private String status;
    private String createTime;
    private String updateTime;

    public ThreadInfo() {}

    public ThreadInfo(String threadId, String title, String status, String createTime, String updateTime) {
        this.threadId = threadId;
        this.title = title;
        this.status = status;
        this.createTime = createTime;
        this.updateTime = updateTime;
    }

    public String getThreadId() {
        return threadId;
    }

    public void setThreadId(String threadId) {
        this.threadId = threadId;
    }

    public String getTitle() {
        return title;
    }

    public void setTitle(String title) {
        this.title = title;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getCreateTime() {
        return createTime;
    }

    public void setCreateTime(String createTime) {
        this.createTime = createTime;
    }

    public String getUpdateTime() {
        return updateTime;
    }

    public void setUpdateTime(String updateTime) {
        this.updateTime = updateTime;
    }
}
