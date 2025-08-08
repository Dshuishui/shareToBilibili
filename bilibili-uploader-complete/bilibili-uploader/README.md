# Bilibili 视频上传工具

基于 Go 语言开发的 B站视频一键投稿服务，支持大文件分片上传和完整的视频稿件提交流程。

## 功能特性

- ✅ **视频分片上传**: 支持大文件分片上传，提高上传成功率
- ✅ **封面上传**: 支持自定义视频封面图片
- ✅ **稿件提交**: 完整的视频稿件信息提交
- ✅ **RESTful API**: 标准的 REST API 接口设计
- ✅ **跨域支持**: 支持前端跨域访问
- ✅ **会话管理**: 自动管理上传会话和清理过期数据
- ✅ **错误处理**: 完善的错误处理和日志记录

## 快速开始

### 1. 获取 Bilibili Cookie

在使用本工具之前，需要获取 Bilibili 的认证 Cookie：

1. 在浏览器中登录 [https://www.bilibili.com](https://www.bilibili.com)
2. 打开开发者工具 (F12)
3. 在 Application/Storage -> Cookies 中找到：
   - `SESSDATA`: 用于身份认证
   - `bili_jct`: 用于CSRF保护

### 2. 编译和运行

```bash
# 克隆项目
git clone <repository-url>
cd bilibili-uploader

# 安装依赖
go mod tidy

# 编译
go build -o bilibili-uploader cmd/main.go

# 运行服务
./bilibili-uploader -sessdata=YOUR_SESSDATA -bili_jct=YOUR_BILI_JCT
```

### 3. 命令行参数

```bash
./bilibili-uploader [选项]

选项:
  -port string
        服务端口 (默认 "8080")
  -sessdata string
        Bilibili SESSDATA Cookie (必需)
  -bili_jct string
        Bilibili bili_jct Cookie (必需)
  -upload_dir string
        上传文件存储目录 (默认 "./uploads")
```

### 4. 访问服务

服务启动后，可以通过以下地址访问：

- **Web界面**: http://localhost:8080
- **健康检查**: http://localhost:8080/api/health
- **API文档**: 查看 Web 界面中的接口说明

## API 接口文档

### 1. 初始化上传

**接口**: `POST /api/upload/init`

**请求体**:
```json
{
    "filename": "video.mp4",
    "filesize": 1048576
}
```

**响应**:
```json
{
    "code": 0,
    "message": "初始化成功",
    "data": {
        "upload_id": "xxx",
        "biz_id": 123,
        "endpoint": "https://...",
        "chunk_size": 10485760,
        "chunks": 1,
        "auth": "xxx"
    }
}
```

### 2. 上传分片

**接口**: `POST /api/upload/chunk`

**请求体** (multipart/form-data):
- `upload_id`: 上传ID
- `biz_id`: 业务ID
- `chunk_index`: 分片索引
- `total_chunks`: 总分片数
- `file_chunk`: 文件分片
- `md5`: 分片MD5
- `crc32`: 分片CRC32
- `hash`: 分片哈希
- `auth`: 认证信息

### 3. 完成上传

**接口**: `POST /api/upload/complete`

**请求体**:
```json
{
    "upload_id": "xxx",
    "biz_id": 123,
    "filename": "video.mp4",
    "filesize": 1048576,
    "md5": "xxx",
    "crc32": "xxx",
    "hash": "xxx",
    "auth": "xxx"
}
```

### 4. 提交视频稿件

**接口**: `POST /api/submit/video`

**请求体**:
```json
{
    "upos_uri": "xxx",
    "title": "视频标题",
    "desc": "视频简介",
    "tags": "标签1,标签2,标签3",
    "tid": 122,
    "cover": "https://...",
    "copyright": 1,
    "source": ""
}
```

### 5. 上传封面

**接口**: `POST /api/upload/cover`

**请求体** (multipart/form-data):
- `cover_file`: 封面图片文件

## 项目结构

```
bilibili-uploader/
├── cmd/
│   └── main.go              # 主程序入口
├── internal/
│   ├── api/
│   │   ├── handlers.go      # API处理器
│   │   └── routes.go        # 路由配置
│   ├── bilibili/
│   │   └── client.go        # Bilibili API客户端
│   ├── models/
│   │   └── types.go         # 数据模型定义
│   └── utils/
│       ├── crypto.go        # 加密和哈希工具
│       └── file.go          # 文件处理工具
├── web/
│   ├── static/              # 静态文件
│   └── templates/
│       └── index.html       # 首页模板
├── configs/                 # 配置文件
├── docs/                    # 文档
├── go.mod                   # Go模块文件
├── go.sum                   # 依赖校验文件
└── README.md               # 项目说明
```

## 技术栈

- **后端框架**: Gin (Go Web框架)
- **HTTP客户端**: Go标准库 net/http
- **跨域处理**: gin-contrib/cors
- **文件处理**: Go标准库
- **JSON处理**: Go标准库 encoding/json

## 注意事项

1. **Cookie安全**: 请妥善保管 SESSDATA 和 bili_jct，不要泄露给他人
2. **文件格式**: 目前支持常见的视频格式 (mp4, avi, mov, wmv, flv, mkv, webm, m4v)
3. **文件大小**: 支持大文件分片上传，理论上无大小限制
4. **网络环境**: 建议在稳定的网络环境下使用
5. **合规使用**: 请遵守 Bilibili 的使用条款和相关法律法规

## 开发和贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

### 开发环境要求

- Go 1.21 或更高版本
- Git

### 本地开发

```bash
# 克隆项目
git clone <repository-url>
cd bilibili-uploader

# 安装依赖
go mod tidy

# 运行开发服务器
go run cmd/main.go -sessdata=YOUR_SESSDATA -bili_jct=YOUR_BILI_JCT
```

## 许可证

本项目采用 MIT 许可证，详情请查看 LICENSE 文件。

## 免责声明

本工具仅供学习和研究使用，使用者需要遵守 Bilibili 的服务条款。作者不对使用本工具产生的任何后果承担责任。

