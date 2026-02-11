<div align="center">

# ClashSub

**一个现代化的 Clash/Mihomo 配置管理与订阅分发系统**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.4+-4FC08D?style=flat&logo=vue.js&logoColor=white)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

## 简介

ClashSub 是一个功能完整的 Web 管理系统，用于管理 Clash/Mihomo 代理配置。它提供了直观的 Web 界面，支持节点管理、规则配置、策略组设置、订阅生成等功能，帮助用户轻松管理和分发代理配置。

## 核心特性

### 代理节点管理
- 支持多种代理协议：Shadowsocks、VMess、VLESS、Trojan、Hysteria2、SOCKS5、HTTP
- 节点批量导入功能
- 节点配置可视化编辑

### 规则与策略
- 灵活的路由规则配置（DOMAIN-SUFFIX、IP-CIDR、GEOIP 等）
- 策略组管理，支持多种选择模式（手动选择、自动测速、故障转移）
- 规则批量导入

### DNS 配置
- 自定义 DNS 服务器设置
- Fake-IP 模式支持
- 域名解析规则配置

### 订阅分发
- 生成专属订阅链接
- 访问日志记录
- 在线用户监控
- 订阅令牌管理

### 用户系统
- JWT 认证
- 多用户管理
- 密码重置工具

## 技术栈

### 后端
- **语言**: Go 1.24+
- **框架**: Gin - 高性能 Web 框架
- **数据库**: SQLite + GORM
- **认证**: JWT (golang-jwt/jwt/v5)
- **配置**: YAML (gopkg.in/yaml.v3)

### 前端
- **框架**: Vue.js 3.4+
- **UI 组件**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **构建工具**: Vite
- **HTTP 客户端**: Axios

## 快速开始

### 环境要求
- Go 1.24 或更高版本
- Node.js 18 或更高版本

### 安装与运行

#### 方式一：使用预编译二进制文件

```bash
# 下载最新版本的预编译文件
# Windows
clash-manager.exe

# Linux/macOS
./clash-manager
```

#### 方式二：从源码运行

**后端启动：**
```bash
# 克隆仓库
git clone https://github.com/yourusername/ClashSub.git
cd ClashSub

# 运行后端服务
go run cmd/server/main.go
```

后端默认运行在 `http://localhost:8090`

**前端开发（可选）：**
```bash
cd web

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建生产版本
npm run build
```

### 首次使用

1. 访问 `http://localhost:8090`
2. 首次访问会进入初始化页面，设置管理员密码
3. 登录后即可开始配置

## 命令行工具

### 密码重置工具

```bash
# 列出所有用户
go run cmd/reset-password/main.go --list

# 重置指定用户密码（交互式输入）
go run cmd/reset-password/main.go --user admin

# 直接指定新密码
go run cmd/reset-password/main.go --user admin --password newpass123

# 编译后使用
go build -o reset-password cmd/reset-password/main.go
./reset-password --list
```

### 服务器启动选项

```bash
# 重置 admin 密码
go run cmd/server/main.go --reset-admin=<新密码>
```

## 功能截图

> 待添加项目截图

## 配置说明

### 默认配置
- 服务器端口: `:8090`
- 数据库路径: `data/clash.db`

修改配置请编辑 `internal/config/config.go`

### 支持的节点类型
- Shadowsocks (ss)
- VMess
- VLESS
- Trojan
- Hysteria2
- SOCKS5
- HTTP

### 支持的规则类型
- DOMAIN (精确域名匹配)
- DOMAIN-SUFFIX (域名后缀匹配)
- DOMAIN-KEYWORD (域名关键词匹配)
- IP-CIDR (IP段匹配)
- GEOIP (地理位置匹配)
- MATCH (兜底规则)

## API 接口

### 认证接口
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/auth/login` | 用户登录 |
| POST | `/api/auth/setup` | 初始化系统 |
| POST | `/api/auth/register` | 创建用户 (需认证) |
| POST | `/api/auth/password` | 修改密码 (需认证) |

### 节点管理
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/nodes` | 获取节点列表 |
| POST | `/api/nodes` | 创建节点 |
| POST | `/api/nodes/import` | 导入节点 |
| PUT | `/api/nodes/:id` | 更新节点 |
| DELETE | `/api/nodes/:id` | 删除节点 |

### 规则管理
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/rules` | 获取规则列表 |
| POST | `/api/rules` | 创建规则 |
| POST | `/api/rules/import` | 导入规则 |
| PUT | `/api/rules/:id` | 更新规则 |
| DELETE | `/api/rules/:id` | 删除规则 |

### 策略组管理
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/groups` | 获取策略组列表 |
| POST | `/api/groups` | 创建策略组 |
| PUT | `/api/groups/:id` | 更新策略组 |
| DELETE | `/api/groups/:id` | 删除策略组 |

### DNS 设置
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/settings/dns` | 获取 DNS 配置 |
| POST | `/api/settings/dns` | 保存 DNS 配置 |

### 订阅管理
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/subscription/token` | 获取订阅令牌 |
| POST | `/api/subscription/token/refresh` | 刷新订阅令牌 |
| GET | `/api/subscription/url` | 获取订阅链接 |
| GET | `/api/subscription/logs` | 获取访问日志 |
| GET | `/api/subscription/stats` | 获取订阅统计 |
| DELETE | `/api/subscription/logs/old` | 删除旧日志 |
| GET | `/api/subscription/online` | 获取在线用户 |

### 订阅获取（公开接口）
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/sub/:token` | 获取 Clash 配置文件 |

## 项目结构

```
ClashSub/
├── cmd/
│   ├── server/              # 主程序入口
│   └── reset-password/      # 密码重置工具
├── internal/
│   ├── api/                 # API 处理器
│   ├── config/              # 配置文件
│   ├── middleware/          # 中间件
│   ├── model/               # 数据模型
│   ├── repository/          # 数据访问层
│   ├── server/              # 服务器初始化
│   └── service/             # 业务逻辑层
├── web/                     # 前端代码
│   └── src/
│       ├── api/             # API 调用
│       ├── views/           # 页面组件
│       └── router/          # 路由配置
├── docs/                    # 文档
├── data/                    # 数据库文件
└── build/                   # 前端构建输出
```

## 开发指南

### 添加新的节点类型

1. 在 `internal/model/clash_yaml.go` 中添加节点配置结构
2. 在 `internal/model/db_model.go` 中更新 Node 模型
3. 在前端 `web/src/views/Nodes.vue` 中添加表单字段
4. 在 `internal/service/generator.go` 中添加生成逻辑

### 添加新的规则类型

1. 在前端 `web/src/views/Rules.vue` 中添加类型选项
2. 在 `internal/service/generator.go` 中添加规则生成逻辑

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 致谢

- [Clash](https://github.com/Dreamacro/clash) - 核心代理引擎
- [Mihomo](https://github.com/MetaCubeX/mihomo) - Clash 内核的持续维护版本
- [Gin](https://github.com/gin-gonic/gin) - Go Web 框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - Vue 3 组件库

---

<div align="center">

**如果这个项目对你有帮助，请给一个 Star**

</div>
