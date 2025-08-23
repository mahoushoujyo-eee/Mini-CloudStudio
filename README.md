# Mini-CloudStudio

一个基于 Kubernetes 的云端开发环境平台，复刻 Cloud Studio 核心功能，支持自定义 code-server 实践环境。

## 项目简介
Mini-CloudStudio 是面向开发者的云原生开发环境即服务（DEaaS）平台。用户可通过 Web 界面一键创建、管理、监控个人云端开发环境，支持资源自定义、环境隔离与持久化。

## 技术架构
- **后端框架**：CloudWeGo Hertz
- **容器编排**：Kubernetes（Pod/Service/PVC 自动化）
- **数据库**：MySQL（GORM 驱动）
- **缓存**：Redis
- **认证**：JWT
- **邮件服务**：go-mail
- **其他依赖**：uuid、x/crypto 等

## 主要功能
### 用户管理
- 用户注册、登录、认证
- JWT 身份验证
- 邮件验证、密码重置
- 角色权限管理

### 云开发环境管理
- 创建/启动/停止/删除 code-server 实例
- 支持 CPU/内存资源自定义
- 实时监控环境状态
- Kubernetes 命名空间隔离、资源配额
- 持久化存储

### 基础设施集成
- Redis 缓存
- MySQL 持久化
- 邮件服务
- RESTful API 设计

## 快速启动
1. 安装 Go 1.24+，并配置好 Kubernetes 集群、MySQL、Redis。
2. 克隆项目并安装依赖：
   ```bash
   git clone <repo-url>
   cd Mini-CloudStudio
   go mod tidy
   ```
3. 配置相关连接信息（见 biz/config/config_file/）。
4. 启动服务：
   ```bash
   go run main.go
   ```

## 目录结构
- biz/config/        配置与初始化
- biz/handler/       路由处理
- biz/middleware/    中间件（JWT等）
- biz/model/         数据模型
- biz/router/        路由注册
- biz/service/       业务逻辑
- biz/util/          工具类
- script/            部署脚本
- main.go            项目入口

## API 简要说明
- 用户注册/登录：POST /api/register, /api/login
- 环境管理：POST/GET /api/clouddev
- 监控与操作：GET/POST /api/clouddev/status, /api/clouddev/action

详细接口文档请见 router/ 及 handler/ 目录。

## 贡献与许可证
欢迎提交 Issue 和 PR。项目采用 MIT 许可证。

---
如需定制或集成更多功能，欢迎联系维护者。
