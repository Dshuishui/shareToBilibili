package main

import (
	"bilibili-uploader/internal/api"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 命令行参数
	var (
		port     = flag.String("port", "8080", "服务端口")
		sessdata = flag.String("sessdata", "", "Bilibili SESSDATA Cookie")
		biliJct  = flag.String("bili_jct", "", "Bilibili bili_jct Cookie")
		uploadDir = flag.String("upload_dir", "./uploads", "上传文件存储目录")
	)
	flag.Parse()

	// 检查必需的参数
	if *sessdata == "" {
		fmt.Println("错误: 必须提供 SESSDATA Cookie")
		fmt.Println("使用方法: ./bilibili-uploader -sessdata=YOUR_SESSDATA -bili_jct=YOUR_BILI_JCT")
		fmt.Println("")
		fmt.Println("获取Cookie的方法:")
		fmt.Println("1. 在浏览器中登录 https://www.bilibili.com")
		fmt.Println("2. 打开开发者工具 (F12)")
		fmt.Println("3. 在 Application/Storage -> Cookies 中找到:")
		fmt.Println("   - SESSDATA: 用于身份认证")
		fmt.Println("   - bili_jct: 用于CSRF保护")
		os.Exit(1)
	}

	if *biliJct == "" {
		fmt.Println("错误: 必须提供 bili_jct Cookie")
		fmt.Println("使用方法: ./bilibili-uploader -sessdata=YOUR_SESSDATA -bili_jct=YOUR_BILI_JCT")
		os.Exit(1)
	}

	// 创建上传目录
	if err := os.MkdirAll(*uploadDir, 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// 创建绝对路径
	absUploadDir, err := filepath.Abs(*uploadDir)
	if err != nil {
		log.Fatalf("获取上传目录绝对路径失败: %v", err)
	}

	// 创建API处理器
	handler := api.NewHandler(*sessdata, *biliJct, absUploadDir)
	
	// 启动会话清理
	handler.CleanupSessions()

	// 设置路由
	router := api.SetupRoutes(handler)

	// 启动服务器
	fmt.Printf("Bilibili视频上传服务启动成功!\n")
	fmt.Printf("服务地址: http://0.0.0.0:%s\n", *port)
	fmt.Printf("上传目录: %s\n", absUploadDir)
	fmt.Printf("API文档: http://0.0.0.0:%s/api/health\n", *port)
	fmt.Println("")
	fmt.Println("按 Ctrl+C 停止服务")

	if err := router.Run("0.0.0.0:" + *port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

