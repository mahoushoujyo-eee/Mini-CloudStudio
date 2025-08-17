# Mini-CloudStudio

一个基于 Kubernetes 的云端开发环境平台，旨在复刻 Cloud Studio 的核心功能，为用户提供可定制规格的 code-server 虚拟实践环境。

## 项目概述

Mini-CloudStudio 是一个云原生的开发环境即服务（Development Environment as a Service）平台，允许用户快速创建和管理个人的云端开发环境。通过 Web 界面，用户可以启动自定义规格的 code-server 实例，在浏览器中进行代码开发、调试和部署。

## 核心功能

### 🚀 用户管理
- 用户注册、登录、认证
- 基于 JWT 的身份验证
- 邮件验证和密码重置

### 💻 开发环境管理
- 创建自定义规格的 code-server 实例
- 支持 CPU、内存资源配置
- 实时监控环境状态
- 环境的启动、停止、删除操作

### ☸️ Kubernetes 集成
- 自动化 Pod、Service、PVC 创建
- 命名空间隔离
- 资源配额管理
- 持久化存储支持

### 🔧 基础设施
- Redis 缓存支持
- MySQL 数据持久化
- 邮件服务集成
- RESTful API 设计

## 技术架构

### 后端技术栈
- **框架**: CloudWeGo Hertz - 高性能 Go HTTP 框架
- **数据库**: MySQL + GORM ORM
- **缓存**: Redis
- **容器编排**: Kubernetes Client-Go
- **认证**: JWT + 中间件
- **邮件**: go-mail

### 项目结构
```
Mini-CloudStudio/
├── biz/                    # 业务逻辑层
│   ├── config/            # 配置管理
│   ├── handler/           # HTTP 处理器
│   ├── middleware/        # 中间件
│   ├── model/            # 数据模型
│   ├── router/           # 路由定义
│   ├── service/          # 业务服务
│   └── util/             # 工具类
├── main.go               # 应用入口
└── router.go            # 路由注册
```

## 设计理念

### 云原生架构
- 基于 Kubernetes 的容器化部署
- 微服务架构设计
- 资源弹性伸缩

### 用户体验优先
- 简洁直观的 Web 界面
- 快速环境启动（秒级）
- 资源使用透明化

### 安全性保障
- 命名空间级别的资源隔离
- JWT 令牌认证
- 数据加密存储

### 可扩展性
- 模块化代码组织
- 插件化架构设计
- 支持多种开发环境模板

## 快速开始

### 环境要求
- Go 1.24+
- Kubernetes 集群
- MySQL 5.7+
- Redis 6.0+

### 安装部署
```bash
# 克隆项目
git clone <repository-url>
cd LearningHertz

# 安装依赖
go mod tidy

# 配置环境变量
cp .env.example .env

# 启动服务
go run main.go
```

### 配置说明
- 数据库连接配置
- Kubernetes 集群配置
- Redis 连接配置
- 邮件服务配置

## 开发状态

🚧 **项目正在积极开发中**

当前已完成的功能：
- [x] 用户认证系统
- [x] Kubernetes 客户端集成
- [x] 基础的环境创建流程
- [x] 数据模型设计

正在开发的功能：
- [ ] 前端用户界面
- [ ] 环境监控面板
- [ ] 资源使用统计
- [ ] 多环境模板支持

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进项目。

## 许可证

本项目采用 MIT 许可证。

## 联系方式

如有问题或建议，请通过 Issue 与我们联系。

---

*让云端开发变得简单而强大* ☁️✨