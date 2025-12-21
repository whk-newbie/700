# Line账号管理系统

> Line账号分组管理与进线统计系统

## 📋 项目简介

本系统是一个基于Go + Vue3，通过使用cursor配合下开发的Line账号管理系统，主要功能包括：

- **分组管理**: 激活码分组，权限控制
- **Line账号管理**: 账号监控，二维码生成
- **进线统计**: 实时统计，去重逻辑
- **客户管理**: 客户信息，跟进记录
- **底库管理**: 联系人导入，数据管理
- **大模型集成**: AI客服，智能回复

## 🛠️ 技术栈

### 后端
- **Go 1.21+**: 主力开发语言
- **Gin**: Web框架
- **GORM**: ORM框架
- **PostgreSQL**: 主数据库
- **Redis**: 缓存和会话存储
- **WebSocket**: 实时通信
- **JWT**: 身份认证

### 前端
- **Vue 3**: 渐进式前端框架
- **Element Plus**: UI组件库
- **Less**: CSS预处理器
- **Vite**: 构建工具
- **Pinia**: 状态管理
- **Axios**: HTTP客户端

### 部署
- **Docker**: 容器化部署
- **Nginx**: 反向代理
- **Docker Compose**: 编排工具

## 📁 项目结构

```
line007/
├── API文档规划.md              # API文档规范
├── README.md                   # 项目说明
├── docker-compose.yml          # Docker编排
├── backend/                    # Go后端
│   ├── cmd/server/main.go      # 应用入口
│   ├── internal/               # 内部包
│   │   ├── config/            # 配置管理
│   │   ├── models/            # 数据模型
│   │   ├── schemas/           # 请求响应结构
│   │   ├── handlers/          # HTTP处理器
│   │   ├── services/          # 业务逻辑
│   │   ├── websocket/         # WebSocket服务
│   │   └── middleware/        # 中间件
│   ├── pkg/                   # 公共包
│   │   ├── database/          # 数据库连接
│   │   ├── redis/             # Redis连接
│   │   └── logger/            # 日志系统
│   ├── migrations/            # 数据库迁移
│   ├── scripts/               # 脚本文件
│   └── static/                # 静态文件
├── frontend/                   # Vue3前端
│   ├── src/
│   │   ├── api/               # API封装
│   │   ├── components/        # 公共组件
│   │   ├── views/             # 页面组件
│   │   ├── router/            # 路由配置
│   │   ├── store/             # 状态管理
│   │   ├── styles/            # 样式文件
│   │   │   ├── variables.less # 样式变量
│   │   │   ├── mixins.less    # 样式混合
│   │   │   └── index.less     # 全局样式
│   │   └── utils/             # 工具函数
│   ├── public/                # 公共资源
│   └── package.json           # 项目配置
└── docs/                      # 项目文档
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 16+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose

### 后端启动

1. **进入后端目录**
   ```bash
   cd backend
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置环境变量**
   ```bash
   cp env.example.txt .env
   # 编辑 .env 文件，配置数据库等信息
   ```

4. **启动服务**
   ```bash
   go run cmd/server/main.go
   ```

### 前端启动

1. **进入前端目录**
   ```bash
   cd frontend
   ```

2. **安装依赖**
   ```bash
   npm install
   ```

3. **启动开发服务器**
   ```bash
   npm run dev
   ```

4. **访问应用**
   ```
   http://localhost:3000
   ```

### Docker部署

1. **构建并启动服务**
   ```bash
   docker-compose up -d
   ```

2. **查看服务状态**
   ```bash
   docker-compose ps
   ```

## 📚 开发指南

### 代码规范

#### Go代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 遵循官方Go命名规范
- 使用有意义的变量和函数名

#### Vue代码规范
- 使用ES6+语法
- 使用Vue 3 Composition API
- 遵循Vue风格指南
- 使用TypeScript（可选）

#### 样式规范
- 使用Less变量定义颜色和尺寸
- 使用BEM命名规范
- 移动端优先的响应式设计
- 避免深度选择器

### 提交规范

```
<type>(<scope>): <subject>

type: feat, fix, docs, style, refactor, test, chore
scope: 影响的模块名
subject: 简短的描述
```

### 分支管理

```
main        # 主分支
develop     # 开发分支
feature/*   # 功能分支
hotfix/*    # 热修复分支
release/*   # 发布分支
```

## 📖 文档

- [API文档规划](API文档规划.md) - 接口文档规范
- [项目实施规划](项目实施规划.md) - 开发计划和进度
- [状态码定义](状态码定义.md) - 统一的错误状态码
- [数据库表设计](数据库表设计-完整版.md) - 数据库设计文档

## 🤝 贡献

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 👥 联系我们

项目维护者：whk-newbie

邮箱：whk-newbie@example.com

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

---

**注意**: 这是一个正在开发中的项目，功能和API可能会发生变化。
