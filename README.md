# Copilot Personal Hub

个人 Copilot 核心与中文管理界面的融合项目。后端基于 Ciallo，界面设计与导航参考 M365-Copilot2API，保留全部上游许可证与来源。不是微软官方 API。

## 功能

- 中文 React 控制台：总览、账号管理、密钥创建/绑定/启停、真实对话测试、个人账号接入指南。
- Ciallo Consumer 核心及原生高级管理、用户自助页保留。
- 管理员会话采用上游 HttpOnly Cookie；密钥仅保存在页面内存，不写入 localStorage。
- 默认仅监听 127.0.0.1:4142，需 HTTPS 反代；640MB 容器上限，不启用共享浏览器。

## 部署

复制 .env.example 为 .env，设置足够强且不同的 ADMIN_PASSWORD 和 API_KEY。切勿留空，否则上游部分接口可能无需认证。

```
docker compose up -d --build
```

反向代理到 127.0.0.1:4142，启用 HTTPS、SSE 无缓冲和 WebSocket 转发。将站点根路径跳转 /console/，将 /self-service 反代到后端 /，其他路径原样转发。示例 compose 的域名与 ALLOWED_ORIGINS 请改成自己的域名。

- /console/：新融合控制台，使用 ADMIN_PASSWORD。
- /admin：原生高级管理。
- /self-service：用户自助页，使用绑定账号的 API Key。
- /v1：兼容接口基础地址。

## 个人账号接入

1. 在账号管理创建空账号，然后创建绑定到该账号的 API Key。
2. 安装 Tampermonkey BETA，并在控制台指南中安装本仓库 get_token.user.js。审阅后只向你信任的服务器推送凭据。
3. 在自己的浏览器打开 copilot.microsoft.com，登录个人账号并发送消息。
4. 脚本中填写自己的站点地址和 API Key，点击“一键推送个人版”。不要使用 M365 Only 授权按钮。
5. 回控制台刷新，确认 Consumer 凭据已导入，再做真实聊天测试。

默认不包含 Camoufox：部分自动续期路径不可用，失效后需重新推送；不能保证永不过期。Cookie/Token 等同账号凭据，不要提交到代码仓库。

## 验证与限制

前端：进入 console 执行 npm ci 和 npm run build。后端：uv sync --extra dev 后执行 uv run pytest。

完整上游资料见 README.upstream.md。真实协议验收状态见 ACCEPTANCE.md。未完成微软账号授权前，不可把健康检查成功说成模型调用成功。消费者协议及工具能力会随微软变更。

## 来源与许可

详见 NOTICE.md、LICENSE 和 licenses/。组合发行保留 M365-Copilot2API 的 AGPL 文本和非商业 API 转售限制，以及 Ciallo 的 Apache 通知。仅供有权使用账号的用户使用，不提供账号、不绕过账号权限、不承诺免费额度。
