package com.alibaba.cloud.starops.samples.client;

import com.aliyun.credentials.Client;
import com.aliyun.credentials.models.CredentialModel;

/**
 * 默认凭据链支持 - 与 Go 实现保持一致
 * 凭据链优先级：环境变量 > OIDC > CLI配置文件 > 配置文件(~/.alibabacloud/credentials) > IAM角色
 */
public class Credentials {

    /**
     * 通过阿里云默认凭据链获取凭证
     * Load credentials via Alibaba Cloud default credential chain
     *
     * @return String[]{accessKeyId, accessKeySecret}
     * @throws Exception 凭据链初始化或获取失败时抛出
     */
    public static String[] loadFromChain() throws Exception {
        // null config triggers default credential chain
        Client client = new Client();
        CredentialModel credential = client.getCredential();
        String accessKeyId = credential.getAccessKeyId();
        String accessKeySecret = credential.getAccessKeySecret();
        if (accessKeyId == null || accessKeyId.isEmpty()
                || accessKeySecret == null || accessKeySecret.isEmpty()) {
            throw new Exception("凭据链返回的凭证为空");
        }
        return new String[]{accessKeyId, accessKeySecret};
    }
}
