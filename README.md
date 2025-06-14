# 钉钉回调处理服务

这个服务用于处理钉钉回调请求，支持解密回调消息并可选转发到其他服务。

## 功能

1. 接收钉钉回调请求
2. 解码URL中的base64参数获取加密密钥等信息
3. 解密钉钉回调消息
4. 可选转发解密后的消息到其他URL

## 项目结构

```
ding_call_back
├── cmd
│   └── main.go               # 启动入口
├── config
│   ├── config.go             # 配置管理
│   └── db_config.go          # 数据库配置
├── controller
│   ├── handler.go            # 回调处理逻辑
│   └── config_handler.go     # 配置注册处理逻辑
├── model
│   └── model.go              # 数据结构定义
├── service
│   ├── logic.go              # 核心业务逻辑
│   └── db_store.go           # 数据库存储服务
├── router
│   └── router.go             # 路由注册
├── utils
│   └── crypto.go             # 工具函数（钉钉加解密）
├── go.mod
├── Dockerfile
├── .dockerignore
└── docker-compose-ding_call_back.yml
```

## 使用方法

### 本地运行

```bash
go run cmd/main.go
```

服务将在3014端口启动（可通过环境变量PORT修改）。

### Docker部署

构建Docker镜像：

```bash
docker build -t ding-callback-service .
```

使用docker-compose运行：

```bash
docker-compose -f docker-compose-ding_call_back.yml up -d
```

### API接口

支持两种方式配置和使用回调：

#### 1. 配置注册方式（推荐）

首先，注册配置信息：

```bash
curl --location 'http://localhost:3014/ding/config' \
--header 'Content-Type: application/json' \
--data '{
  "url": "http://example.com/api/callback",
  "aes_key": "your_aes_key",
  "token": "your_token",
  "corpId": "your_corpId"
}'
```

注册成功后，会返回一个UUID和回调URL：

```json
{
  "success": true,
  "message": "配置注册成功",
  "data": {
    "uuid": "a1b2c3d4-e5f6-7890-abcd-1234567890ab",
    "callback_url": "/ding/callback/a1b2c3d4-e5f6-7890-abcd-1234567890ab"
  }
}
```

然后，将此回调URL配置到钉钉开发者后台的回调地址中。

#### 2. URL路径方式（兼容模式）

接口路由：`/ding/callback/:jsonbase64`

其中`jsonbase64`是经过Base64编码的JSON字符串，包含以下字段：

```json
{
  "url": "转发的具体接口地址(可选)",
  "aes_key": "加密的aes_key",
  "token": "签名token",
  "corpId": "企业自建应用-事件订阅使用appKey;企业自建应用-注册回调地址使用corpId;第三方企业应用使用suiteKey"
}
```

### 处理流程

1. 服务接收到回调请求后，首先根据UUID从数据库获取配置
2. 使用获取到的配置初始化钉钉加密处理器
3. 解密请求中的加密数据
4. 如果提供了url参数，则将解密后的数据转发到指定URL
5. 返回加密后的"success"响应，符合钉钉回调验证要求

### 示例

#### 配置注册

```bash
curl --location 'http://localhost:3014/ding/config' \
--header 'Content-Type: application/json' \
--data '{
  "url": "http://example.com/api/callback",
  "aes_key": "your_aes_key",
  "token": "your_token",
  "corpId": "your_corpId"
}'
```

响应：

```json
{
  "success": true,
  "message": "配置注册成功",
  "data": {
    "uuid": "a1b2c3d4-e5f6-7890-abcd-1234567890ab",
    "callback_url": "/ding/callback/a1b2c3d4-e5f6-7890-abcd-1234567890ab"
  }
}
```

## 数据库配置

服务使用PostgreSQL数据库存储配置信息。数据库连接信息：

```
Host: 120.46.147.53
Port: 5432
Database: pro_db
User: renoelis
Password: renoelis02@gmail.com
```

数据表结构：

```sql
CREATE TABLE IF NOT EXISTS ding_callback_configs (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    url TEXT,
    aes_key VARCHAR(100) NOT NULL,
    token VARCHAR(100) NOT NULL,
    corp_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 注意事项

- AES密钥长度必须为43个字符
- 如果提供了URL，解密后的数据将以POST方式转发到该URL
- 服务默认监听3014端口，可通过环境变量PORT修改
- 配置永久存储在PostgreSQL数据库中，不会自动过期

## 环境变量配置

服务支持通过环境变量进行配置：

| 环境变量 | 说明 | 默认值 |
|---------|------|-------|
| PORT | 服务监听端口 | 3014 |
| DB_HOST | 数据库主机地址 | 120.46.147.53 |
| DB_PORT | 数据库端口 | 5432 |
| DB_NAME | 数据库名称 | pro_db |
| DB_USER | 数据库用户名 | renoelis |
| DB_PASSWORD | 数据库密码 | renoelis02@gmail.com |

在`docker-compose-ding_call_back.yml`文件中已经配置了这些环境变量，您可以根据需要修改它们。 