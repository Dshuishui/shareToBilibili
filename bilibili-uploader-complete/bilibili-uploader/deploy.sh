#!/bin/bash

# Bilibili视频上传工具部署脚本

set -e

echo "=== Bilibili视频上传工具部署脚本 ==="
echo ""

# 检查参数
if [ $# -lt 2 ]; then
    echo "使用方法: $0 <SESSDATA> <BILI_JCT> [PORT]"
    echo ""
    echo "参数说明:"
    echo "  SESSDATA  - Bilibili SESSDATA Cookie (必需)"
    echo "  BILI_JCT  - Bilibili bili_jct Cookie (必需)"
    echo "  PORT      - 服务端口 (可选，默认8080)"
    echo ""
    echo "获取Cookie的方法:"
    echo "1. 在浏览器中登录 https://www.bilibili.com"
    echo "2. 打开开发者工具 (F12)"
    echo "3. 在 Application/Storage -> Cookies 中找到 SESSDATA 和 bili_jct"
    exit 1
fi

SESSDATA=$1
BILI_JCT=$2
PORT=${3:-8080}

echo "配置信息:"
echo "  端口: $PORT"
echo "  SESSDATA: ${SESSDATA:0:20}..."
echo "  BILI_JCT: ${BILI_JCT:0:20}..."
echo ""

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go语言环境"
    echo "请先安装Go语言: https://golang.org/dl/"
    exit 1
fi

# 检查Node.js是否安装
if ! command -v node &> /dev/null; then
    echo "错误: 未找到Node.js环境"
    echo "请先安装Node.js: https://nodejs.org/"
    exit 1
fi

# 编译后端
echo "1. 编译Go后端服务..."
cd "$(dirname "$0")"
go mod tidy
go build -o bilibili-uploader cmd/main.go
echo "   ✓ 后端编译完成"

# 构建前端
echo ""
echo "2. 构建React前端..."
cd ../bilibili-uploader-frontend
npm install --legacy-peer-deps
npm run build
echo "   ✓ 前端构建完成"

# 复制前端文件到后端静态目录
echo ""
echo "3. 部署前端文件..."
cd ../bilibili-uploader
rm -rf web/static/*
rm -rf web/templates/*
cp -r ../bilibili-uploader-frontend/dist/* web/static/
cp web/static/index.html web/templates/
echo "   ✓ 前端文件部署完成"

# 创建上传目录
mkdir -p uploads
echo "   ✓ 创建上传目录"

# 启动服务
echo ""
echo "4. 启动服务..."
echo "正在启动Bilibili视频上传服务..."

# 停止已有服务
pkill -f "bilibili-uploader" || true
sleep 2

# 启动新服务
nohup ./bilibili-uploader \
    -sessdata="$SESSDATA" \
    -bili_jct="$BILI_JCT" \
    -port="$PORT" \
    > service.log 2>&1 &

sleep 3

# 检查服务状态
if curl -s "http://localhost:$PORT/api/health" > /dev/null; then
    echo "   ✓ 服务启动成功"
    echo ""
    echo "=== 部署完成 ==="
    echo ""
    echo "服务信息:"
    echo "  访问地址: http://localhost:$PORT"
    echo "  API健康检查: http://localhost:$PORT/api/health"
    echo "  日志文件: $(pwd)/service.log"
    echo ""
    echo "使用说明:"
    echo "1. 打开浏览器访问 http://localhost:$PORT"
    echo "2. 选择视频文件并填写相关信息"
    echo "3. 点击'开始上传'进行视频投稿"
    echo ""
    echo "停止服务: pkill -f bilibili-uploader"
    echo "查看日志: tail -f $(pwd)/service.log"
else
    echo "   ✗ 服务启动失败"
    echo ""
    echo "请检查日志文件: $(pwd)/service.log"
    exit 1
fi

