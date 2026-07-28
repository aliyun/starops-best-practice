package com.alibaba.cloud.starops.samples.logger;

import com.fasterxml.jackson.databind.ObjectMapper;
import java.io.PrintStream;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
import java.util.HashMap;
import java.util.Map;

/**
 * 结构化日志器
 * Structured logger
 */
public class Logger {
    private LogLevel level;
    private PrintStream output;
    private final ObjectMapper objectMapper;

    public enum LogLevel {
        DEBUG(0), INFO(1), WARN(2), ERROR(3);

        private final int value;

        LogLevel(int value) {
            this.value = value;
        }

        public int getValue() {
            return value;
        }

        public static LogLevel fromString(String s) {
            if (s == null) return INFO;
            switch (s.toLowerCase().trim()) {
                case "debug": return DEBUG;
                case "warn":
                case "warning": return WARN;
                case "error": return ERROR;
                default: return INFO;
            }
        }
    }

    public Logger(LogLevel level, PrintStream output) {
        this.level = level;
        this.output = output != null ? output : System.out;
        this.objectMapper = new ObjectMapper();
    }

    public static Logger fromEnv() {
        String levelStr = System.getenv("STAROPS_LOG_LEVEL");
        LogLevel level = LogLevel.fromString(levelStr);
        return new Logger(level, System.out);
    }

    public void setLevel(LogLevel level) {
        this.level = level;
    }

    public LogLevel getLevel() {
        return level;
    }

    public void setOutput(PrintStream output) {
        this.output = output;
    }

    private void log(LogLevel level, String message, Map<String, Object> context, Throwable error, boolean includeStack) {
        if (level.getValue() < this.level.getValue()) {
            return;
        }

        Map<String, Object> entry = new HashMap<>();
        entry.put("timestamp", DateTimeFormatter.ISO_INSTANT.format(Instant.now()));
        entry.put("level", level.name().toLowerCase());
        entry.put("message", message);

        if (context != null && !context.isEmpty()) {
            entry.put("context", context);
        }

        if (error != null) {
            entry.put("error", error.getMessage());
            if (includeStack) {
                entry.put("stack", getStackTrace(error));
            }
        }

        try {
            output.println(objectMapper.writeValueAsString(entry));
        } catch (Exception e) {
            output.printf("%s [%s] %s%n", entry.get("timestamp"), level.name(), message);
        }
    }

    public void debug(String message, Map<String, Object> context) {
        log(LogLevel.DEBUG, message, context, null, false);
    }

    public void info(String message, Map<String, Object> context) {
        log(LogLevel.INFO, message, context, null, false);
    }

    public void warn(String message, Map<String, Object> context) {
        log(LogLevel.WARN, message, context, null, false);
    }

    public void error(String message, Throwable error, Map<String, Object> context) {
        log(LogLevel.ERROR, message, context, error, true);
    }

    private String getStackTrace(Throwable error) {
        StringBuilder sb = new StringBuilder();
        for (StackTraceElement element : error.getStackTrace()) {
            sb.append(element.toString()).append("\n");
        }
        return sb.toString();
    }
}
