package com.alibaba.cloud.starops.samples.types;

/**
 * 执行状态枚举
 * Item status enumeration
 */
public enum ItemStatus {
    /** 初始化 / Initialization */
    INIT("init"),
    /** 开始执行 / Start execution */
    START("start"),
    /** 执行中 / In progress */
    PROGRESS("progress"),
    /** 暂停 / Suspended */
    SUSPENDED("suspended"),
    /** 执行完成(成功) / Completed successfully */
    SUCCESS("success"),
    /** 执行完成(失败) / Completed with failure */
    FAIL("fail");

    private final String value;

    ItemStatus(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }

    public static ItemStatus fromValue(String value) {
        for (ItemStatus status : values()) {
            if (status.value.equals(value)) {
                return status;
            }
        }
        return null;
    }
}
