# B站一键投稿工具

这是一个自动化投稿工具，可以帮助你将视频一键上传到B站，支持两种投稿方式：全自动浏览器投稿和API投稿。

## 🌟 功能特点

- **自动登录**：通过Puppeteer自动打开浏览器登录B站获取Cookie
- **视频上传**：支持选择本地视频文件并填写标题、描述、标签等信息
- **分区选择**：支持选择B站的各个视频分区
- **双模式投稿**：
  - **浏览器自动化投稿**：自动填写表单，暂停供预览确认，然后自动提交
  - **API投稿**：通过Python后端直接调用B站API上传
- **用户友好**：提供清晰的进度反馈和错误提示

## 📁 项目结构

```
bilibili-auto-upload/
├── public/
│   └── index.html          # 前端界面
├── python-backend/
│   ├── app.py             # Python后端服务
│   ├── test_upload.py     # 上传测试脚本
│   └── uploads/           # 临时文件存储目录
├── app.js                 # Node.js前端服务
├── package.json           # Node.js依赖配置
└── README.md              # 项目说明文档
```

## 🚀 快速开始

### 环境要求

- Node.js (v12或更高版本)
- Python 3.7或更高版本
- Google Chrome浏览器

### 安装步骤

1. **安装Node.js依赖**
```bash
npm install
```

2. **安装Python依赖**
```bash
cd python-backend
pip install bilibili-api-python flask flask-cors
```

### 启动服务

1. **启动Python后端服务**
```bash
cd python-backend
python app.py
```

2. **启动Node.js前端服务**
```bash
# 在项目根目录下
node app.js
```

3. **访问前端页面**
打开浏览器访问: http://localhost:3000

## 📖 使用说明

### 方式一：浏览器自动化投稿（推荐）

1. **登录B站账号**
   点击"🔑 登录B站账号"按钮，系统会自动打开Chrome浏览器，你需要在弹出的窗口中完成B站登录。登录成功后，浏览器会保持打开状态。

2. **选择视频文件**
   点击"选择文件"按钮，选择你要上传的视频文件。

3. **填写视频信息**
   - **视频标题**：必填，最多80个字符
   - **视频简介**：可选，最多2000个字符
   - **视频标签**：可选，多个标签用英文逗号分隔
   - **视频分区**：必选，选择视频所属的分区

4. **自动化投稿**
   点击"🎯 自动化投稿"按钮，系统会：
   - 自动跳转到B站投稿页面
   - 自动上传视频文件
   - 自动填写视频信息
   - 暂停供你预览确认
   - 确认后自动提交投稿

### 方式二：API投稿

1. **登录B站账号**
   点击"🔑 登录B站账号"按钮，系统会自动打开Chrome浏览器，你需要在弹出的窗口中完成B站登录。

2. **选择视频文件和填写信息**
   同方式一的步骤2-3。

3. **API投稿**
   点击"🚀 API投稿"按钮，系统会通过Python后端直接调用B站API完成上传。

## ⚙️ 技术架构

### 前端 (Node.js)
- 使用Express框架提供Web服务
- 使用Puppeteer处理B站登录和浏览器自动化操作
- 使用Multer处理文件上传
- 与Python后端通过HTTP API通信

### 后端 (Python)
- 使用Flask框架提供RESTful API
- 使用bilibili-api-python库与B站API交互
- 处理视频上传逻辑

### 通信流程
```
浏览器 -> Node.js前端 -> (浏览器自动化 | Python后端) -> B站API
```

## 🔧 API接口

### Node.js前端接口
- `GET /health` - 健康检查
- `GET /login` - B站登录（浏览器自动化）
- `GET /check-login` - 检查登录状态
- `POST /auto-upload` - 浏览器自动化投稿
- `POST /api-upload` - API投稿（转发到Python后端）
- `POST /close-browser` - 关闭浏览器

### Python后端接口
- `GET /health` - 健康检查
- `POST /login` - 处理登录凭证
- `GET /check-login` - 检查登录状态
- `POST /upload-video` - 视频上传

## 📋 注意事项

1. **浏览器要求**：需要安装Google Chrome浏览器，程序会自动调用
2. **网络环境**：确保网络环境稳定，大文件上传可能需要较长时间
3. **文件大小**：B站对视频文件大小有限制，请确保文件符合要求
4. **Cookie有效期**：登录获取的Cookie有一定有效期，过期后需要重新登录
5. **浏览器占用**：自动化投稿模式会保持浏览器打开，完成投稿后可手动关闭或调用关闭接口

## 🛠 常见问题

### 1. 启动时报错找不到模块
请确保已正确安装所有依赖：
```bash
npm install
pip install bilibili-api-python flask flask-cors
```

### 2. Chrome浏览器未找到
请检查Chrome浏览器是否安装在默认路径，或修改[app.js](file:///Users/cong/Documents/七牛云/shareToBilibili/app.js)中的executablePath配置。

### 3. 上传失败
- 检查网络连接
- 确认B站账号登录状态
- 检查视频文件格式和大小是否符合B站要求
- 查看控制台日志获取详细错误信息

### 4. 自动化投稿卡住
- 可能是B站页面结构发生变化，需要更新选择器
- 确保浏览器窗口保持可见状态，不要最小化
- 页面加载可能需要较长时间，请耐心等待

## 📝 开发说明

项目采用前后端分离架构，Node.js前端负责用户界面和浏览器自动化操作，Python后端负责与B站API交互。

### 扩展开发
1. 可以根据需要添加更多视频信息字段
2. 可以优化浏览器自动化脚本以适应B站页面更新
3. 可以增强错误处理和重试机制

## 📄 许可证

本项目仅供学习交流使用，请遵守B站相关服务条款。

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个项目。