# M365 Copilot2API + Personal Provider

基于 HEXUXIU/M365-Copilot2API 的 Go 主项目，接入 Ciallo 的个人 Copilot 通信核心。不是“Python 主站换皮”：管理、账号存储、调度、API Key、会话、用量、Chat Completions / Responses / Messages 协议转换都由原 Go 服务承担。

## 界面

- 根路径保留原版完整控制台；/webapp/ 保留上游 React 控制台。
- 不更换配色、字体、侧栏、表格、卡片与交互风格。
- 在账号页增加个人凭据导入；原企业授权保留为可选入口。
- 默认个人版模型目录：copilot、copilot-reasoning、copilot-thinking、copilot-chat、copilot-search、copilot-research、copilot-study、copilot-coco。它们是个人版模式，不是对应企业 GPT / Claude 模型的保证。

## 架构

用户 → OpenResty HTTPS → Go 主服务 → 私有 Consumer 通信组件 → 微软个人版 Copilot。

Consumer 组件仅执行 Ciallo 的 curl_cffi HTTP/WebSocket 通信；没有独立面板、账号数据库或公网端口。Go 将当前轮的必要凭据通过私有容器网络传入，通信组件不持久化账号。此方案保留原通信实现所需的 TLS 行为，而非假称已将全部 Python 协议重写成 Go。

## 部署

1. 复制 .env.example 为 .env，生成不同的高强度 M365_MASTER_KEY 和 M365_CONSUMER_TRANSPORT_KEY。
2. 创建 secrets/admin-password，写入管理员密码。仅让部署用户和网关容器读到它，禁止提交 GitHub。
3. docker compose up -d --build。Go 主服务只映射 127.0.0.1:4143，通信组件不映射宿主机端口。
4. OpenResty 反代到该端口，保持 WebSocket、SSE 无缓冲和 HTTPS。
5. 默认 M365_CONSUMER_ONLY=1 仅展示个人模式；企业 OAuth 核心代码保留。容器内存上限分别为 384MB 与 256MB，个人请求串行处理。

## 接入个人账号

### copilot.com 新入口（脚本 1.1.0）

上游 Issue #7（https://github.com/MurasameCyan/Ciallo-Ms-365-OpenAI-Proxy-Docker/issues/7）报告同样的个人账号跳转问题；2026-09-28 检查时仍开放且无维护者回复，默认 multi 与 fox 分支脚本均未提供新版域名适配。报告中的新站点使用 Substrate ChatHub，不能只把旧版 Consumer URL 换成 copilot.com 就认为兼容。

本项目脚本 1.1.0 增加新域名匹配、document-start 捕获、右下角入口、无 innerHTML 的个人面板和安全诊断，并消除快捷键重复监听。个人面板不再要求代理地址或 API Key，也不自动上传凭据。老版本需要手动更新并刷新 Copilot 标签页。

检测到旧版 ChatAI 时允许导出；检测到新版 ChatHub 或未验证端点时明确提示不兼容。安全诊断只包含脚本版本、页面域名、布尔能力状态和协议类别，不含完整 WebSocket URL、账号标识或凭据。当前未实现已验证的新版 ChatHub 个人账号认证；不要将显示修复描述为新协议聊天已通过。

1. 在自己的浏览器安装 Tampermonkey BETA；控制台账号页提供导出脚本。
2. 登录 copilot.microsoft.com 并发送消息，点击脚本的“导出个人版凭据”。无需把管理员密码或 API Key 填入脚本。
3. 在控制台账号页导入 JSON。凭据含 Cookie / Token，等同登录材料；不要公开，用完妥善移除下载文件。
4. 去模型测试做真实验证，然后用原版 API Key 管理创建调用密钥。

## 保留与限制

- 保留原 Go 会话、用量、工具路由和协议处理。个人版工具调用仍是提示词/文本解析，依赖模型服从性，不是微软原生工具接口。
- 保留个人版文本流、模式选择、图片输入和上游返回的图片文本片段；图片型 /v1/images 端点没有被伪装成已迁移的个人出图 API。
- 支持带身份校验、并发合并和落盘加密的个人刷新令牌续期；按请求临期触发。Cookie 失效、无可用刷新令牌或微软拒绝时需手动重新导入。未安装 Camoufox 后备浏览器，不承诺永久登录。
- 个人历史由 Go 服务维护，每轮向个人上游发送当前上下文；不声称同步个人版网页全部云端历史。会话页会显示这一限制。
- 企业云端记忆、租户数据和 Studio 能力不适用于个人账号，不会通过改模型名获得权限。
- 原简化 Python 主站已由此次重构替代，旧密码保留。历史提交可用于回滚。

## 验证

Go: go test ./...；React: cd webapp && npm ci && npm run build；通信组件: pytest consumer/test_service.py。

模拟上游的集成测试覆盖三种协议的流式与非流式转换。生产中未经本人微软授权，不得把这些测试与健康检查等同真实模型验收。详见 ACCEPTANCE.md。

## 来源与许可证

基座 HEXUXIU/M365-Copilot2API（完整 AGPL 文本及附加非商业 API 转售限制见 LICENSE）；个人协议 MurasameCyan/Ciallo-Ms-365-OpenAI-Proxy-Docker（Apache 通知见 licenses/Ciallo-LICENSE）。详见 NOTICE.md。仅使用你有权使用的账号，不提供微软账号或权限绕过。
