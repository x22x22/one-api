# Coze API 适配器

本适配器支持Coze API的V2和V3两个版本。

## 环境变量配置

使用环境变量`COZE_API_VERSION`来控制使用的API版本：

- 设置为`v2`：使用V2版本API
- 不设置或设置为其他值：使用V3版本API (默认)

示例：

```bash
# 使用V3版本API
export COZE_API_VERSION=v3

# 使用V2版本API (默认)
export COZE_API_VERSION=v2
# 或不设置此环境变量
```

## 版本区别

- V2 API：原始的Coze API
- V3 API：新版本提供更丰富的功能，包括多模态输入、会话管理等

## 配置项说明

在使用V3 API时，如果需要指定会话ID，可以在配置中添加ConversationID字段。 
