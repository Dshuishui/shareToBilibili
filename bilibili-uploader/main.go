// main.go - B站OAuth一键投稿服务
package main

import (
    "bilibili-uploader/config"
    "bilibili-uploader/handlers"
    "bilibili-uploader/services"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    // 加载配置
    config.Load()
    
    // 创建必要的目录
    createDirectories()
    
    // 检查B站配置
    if !config.IsBilibiliConfigured() {
        log.Println("================================================")
        log.Println("⚠️  B站OAuth未配置!")
        log.Println("请按以下步骤操作:")
        log.Println("1. 访问 https://open.bilibili.com 申请开发者")
        log.Println("2. 创建应用，获取Client ID和Secret")
        log.Println("3. 在.env文件中配置:")
        log.Println("   BILIBILI_CLIENT_ID=你的ClientID")
        log.Println("   BILIBILI_CLIENT_SECRET=你的Secret")
        log.Println("================================================")
        log.Println("📌 现在以模拟模式运行...")
    }
    
    // 创建Gin路由器
    router := gin.Default()
    
    // 配置CORS
    corsConfig := cors.DefaultConfig()
    corsConfig.AllowOrigins = []string{"*"}
    corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Cookie"}
    corsConfig.AllowCredentials = true
    router.Use(cors.New(corsConfig))
    
    // 设置文件上传大小限制 (4GB)
    router.MaxMultipartMemory = 4 << 30
    
    // 静态文件服务
    router.Static("/static", "./static")
    router.StaticFS("/downloads", http.Dir("./processed"))
    
    // 主页
    router.GET("/", func(c *gin.Context) {
        c.Redirect(http.StatusMovedPermanently, "/static/index.html")
    })
    
    // API路由组
    api := router.Group("/api")
    {
        // 健康检查
        api.GET("/health", healthCheck)
        
        // OAuth认证相关
        authHandler := handlers.NewAuthHandler()
        auth := api.Group("/auth")
        {
            auth.GET("/url", authHandler.GetAuthURL)           // 获取授权URL
            auth.GET("/callback", authHandler.HandleCallback)  // OAuth回调
            auth.POST("/token", authHandler.ExchangeToken)     // 交换token
            auth.GET("/verify", authHandler.VerifyToken)       // 验证token
        }
        
        // 视频上传相关（需要认证）
        upload := api.Group("/upload")
        upload.Use(authMiddleware())
        {
            upload.POST("/bilibili", handleBilibiliUpload)  // B站上传
            upload.POST("/process", processVideo)           // 视频处理
        }
        
        // 公开的上传接口（用于测试）
        api.POST("/upload/test", handleTestUpload)
    }
    
    // 启动服务器
    port := ":" + config.GlobalConfig.Port
    log.Printf("🚀 B站OAuth一键投稿服务启动")
    log.Printf("📍 访问地址: http://localhost%s", port)
    log.Printf("🔐 OAuth回调: %s", config.GlobalConfig.BilibiliRedirectURI)
    
    if config.IsBilibiliConfigured() {
        log.Printf("✅ B站OAuth已配置")
    } else {
        log.Printf("⚠️  B站OAuth未配置，运行在模拟模式")
    }
    
    if err := router.Run(port); err != nil {
        log.Fatalf("服务器启动失败: %v", err)
    }
}

// createDirectories 创建必要的目录
func createDirectories() {
    dirs := []string{
        config.GlobalConfig.UploadDir,
        config.GlobalConfig.ProcessedDir,
        config.GlobalConfig.TempDir,
    }
    
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            log.Printf("创建目录失败 %s: %v", dir, err)
        }
    }
}

// healthCheck 健康检查
func healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "message": "B站OAuth一键投稿服务运行正常",
        "features": gin.H{
            "oauth_configured": config.IsBilibiliConfigured(),
            "upload_enabled": true,
            "version": "1.0.0",
        },
        "endpoints": []string{
            "/api/auth/url - 获取OAuth授权URL",
            "/api/auth/callback - OAuth回调",
            "/api/auth/verify - 验证token",
            "/api/upload/bilibili - 上传到B站",
        },
    })
}

// authMiddleware 认证中间件
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 如果是模拟模式，跳过认证
        if !config.IsBilibiliConfigured() {
            log.Println("模拟模式：跳过认证")
            c.Next()
            return
        }
        
        // 验证Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "message": "请先登录B站账号",
            })
            c.Abort()
            return
        }
        
        // 验证token（这里简化处理，实际应该调用handlers.ValidateJWT）
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "message": "无效的认证格式",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// handleBilibiliUpload 处理B站上传
func handleBilibiliUpload(c *gin.Context) {
    // 获取上传的文件
    file, header, err := c.Request.FormFile("video")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "获取视频文件失败",
        })
        return
    }
    defer file.Close()
    
    // 获取视频信息
    title := c.PostForm("title")
    desc := c.PostForm("desc")
    tags := c.PostForm("tags")
    category := c.PostForm("category")
    
    if title == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "视频标题不能为空",
        })
        return
    }
    
    log.Printf("📤 收到上传请求:")
    log.Printf("   文件: %s (%.2f MB)", header.Filename, float64(header.Size)/(1024*1024))
    log.Printf("   标题: %s", title)
    log.Printf("   简介: %s", desc)
    log.Printf("   标签: %s", tags)
    log.Printf("   分区: %s", category)
    
    // 保存文件到临时目录
    tempFile := filepath.Join(config.GlobalConfig.TempDir, fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename))
    dst, err := os.Create(tempFile)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "保存文件失败",
        })
        return
    }
    defer dst.Close()
    
    written, err := io.Copy(dst, file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "写入文件失败",
        })
        return
    }
    
    log.Printf("✅ 文件已保存: %s (%d bytes)", tempFile, written)
    
    // 如果配置了B站OAuth，执行真实上传
    if config.IsBilibiliConfigured() {
        // 从JWT中获取B站token
        bilibiliToken, err := handlers.GetBilibiliToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "message": "获取B站认证失败",
            })
            return
        }
        
        // 准备上传参数
        uploadParams := services.VideoUploadParams{
            Title:       title,
            Description: desc,
            Tags:        strings.Split(tags, " "),
            Category:    21, // 默认分区，实际应该从category参数解析
            Copyright:   1,  // 自制
        }
        
        // 执行上传
        log.Printf("🚀 开始上传到B站...")
        bvid, err := services.AutoUploadWithOAuth(bilibiliToken, tempFile, uploadParams)
        if err != nil {
            log.Printf("❌ 上传失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "success": false,
                "message": fmt.Sprintf("上传失败: %v", err),
            })
            return
        }
        
        log.Printf("✅ 上传成功! BV号: %s", bvid)
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "视频上传成功",
            "bvid":    bvid,
            "url":     fmt.Sprintf("https://www.bilibili.com/video/%s", bvid),
        })
        
    } else {
        // 模拟模式
        log.Printf("📝 模拟模式：生成模拟结果")
        
        // 模拟上传延迟
        time.Sleep(2 * time.Second)
        
        // 生成模拟的BV号
        mockBVID := fmt.Sprintf("BV1mock%d", time.Now().Unix()%100000)
        
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "视频上传成功（模拟）",
            "bvid":    mockBVID,
            "url":     fmt.Sprintf("https://www.bilibili.com/video/%s", mockBVID),
            "mode":    "simulation",
            "note":    "这是模拟结果，实际上传需要配置B站OAuth",
        })
    }
    
    // 清理临时文件
    go func() {
        time.Sleep(5 * time.Minute)
        os.Remove(tempFile)
        log.Printf("🗑️ 清理临时文件: %s", tempFile)
    }()
}

// handleTestUpload 测试上传（不需要认证）
func handleTestUpload(c *gin.Context) {
    file, header, err := c.Request.FormFile("video")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "获取文件失败",
        })
        return
    }
    defer file.Close()
    
    log.Printf("📤 测试上传: %s (%.2f MB)", header.Filename, float64(header.Size)/(1024*1024))
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "测试上传成功",
        "filename": header.Filename,
        "size": header.Size,
    })
}

// processVideo 处理视频
func processVideo(c *gin.Context) {
    var req struct {
        Filename string `json:"filename"`
        Quality  string `json:"quality"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "参数错误",
        })
        return
    }
    
    log.Printf("⚙️ 处理视频: %s (质量: %s)", req.Filename, req.Quality)
    
    // 检查是否安装了FFmpeg
    if !services.CheckFFmpeg() {
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "视频处理完成（模拟）",
            "note": "未安装FFmpeg，返回原始文件",
        })
        return
    }
    
    // 执行视频处理
    processor := services.NewVideoProcessor(
        config.GlobalConfig.UploadDir,
        config.GlobalConfig.ProcessedDir,
    )
    
    processOptions := services.ProcessOptions{
        Quality: req.Quality,
    }
    
    outputFile, err := processor.ProcessVideo(req.Filename, processOptions)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": fmt.Sprintf("处理失败: %v", err),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "视频处理完成",
        "output": outputFile,
        "download_url": fmt.Sprintf("/downloads/%s", outputFile),
    })
}