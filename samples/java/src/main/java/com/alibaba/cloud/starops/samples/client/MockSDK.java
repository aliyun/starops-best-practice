package com.alibaba.cloud.starops.samples.client;

import java.io.*;
import java.nio.charset.StandardCharsets;
import java.util.*;
import com.alibaba.cloud.starops.samples.client.Config;
import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;

/**
 * Mock/Record SDK 层 — 在 openSSEStream 层面拦截
 */
public class MockSDK {
    private static final String STREAM_SEPARATOR = "---";
    private static long mockOffset = 0;
    private static final Gson gson = new Gson();

    public static class MockSSEEvent {
        public String id;
        public String event;
        public Object data;
    }

    /**
     * 从 JSONL 文件读取事件并回放
     */
    public static List<Map<String, Object>> mockStream(Config config) throws IOException {
        List<Map<String, Object>> events = new ArrayList<>();
        if (config.mockFile == null || config.mockFile.isEmpty()) {
            throw new IOException("mockFile is required");
        }

        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(new FileInputStream(config.mockFile), StandardCharsets.UTF_8))) {
            if (mockOffset > 0) {
                reader.skip(mockOffset);
            }
            long bytesRead = 0;
            String line;
            while ((line = reader.readLine()) != null) {
                bytesRead += line.getBytes(StandardCharsets.UTF_8).length + 1;
                line = line.trim();
                if (line.equals(STREAM_SEPARATOR)) {
                    mockOffset += bytesRead;
                    break;
                }
                if (line.isEmpty()) continue;

                try {
                    MockSSEEvent evt = gson.fromJson(line, MockSSEEvent.class);
                    Map<String, Object> resp = new HashMap<>();
                    resp.put("id", evt.id != null ? evt.id : "");
                    resp.put("event", evt.event != null ? evt.event : "message");
                    resp.put("body", evt.data);
                    resp.put("statusCode", 200);
                    events.add(resp);
                } catch (Exception e) {
                    // skip invalid lines
                }
            }
        }
        return events;
    }

    /**
     * 代理真实 SDK 并旁路写入 JSONL 文件
     */
    public static void recordEvent(Config config, Map<String, Object> resp) throws IOException {
        if (config.mockFile == null || config.mockFile.isEmpty()) return;

        MockSSEEvent evt = new MockSSEEvent();
        evt.id = (String) resp.getOrDefault("id", "");
        evt.event = (String) resp.getOrDefault("event", "message");
        evt.data = resp.getOrDefault("body", resp);

        try (FileWriter fw = new FileWriter(config.mockFile, true);
             BufferedWriter bw = new BufferedWriter(fw)) {
            bw.write(gson.toJson(evt));
            bw.newLine();
        }
    }

    public static void writeSeparator(Config config) throws IOException {
        if (config.mockFile == null || config.mockFile.isEmpty()) return;
        try (FileWriter fw = new FileWriter(config.mockFile, true);
             BufferedWriter bw = new BufferedWriter(fw)) {
            bw.write(STREAM_SEPARATOR);
            bw.newLine();
        }
    }
}