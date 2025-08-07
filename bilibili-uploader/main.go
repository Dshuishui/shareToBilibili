// main.go - 程序入口文件
package main

import (
    // "fmt"
    "log"
    "net/http"
    
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    // 创建Gin路由器
    router := gin.Default()
    
    // 配置CORS中间件，允许前端跨域访问
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Cookie"}
    config.AllowCredentials = true
    router.Use(cors.New(config))
    
    // 设置文件上传大小限制 (4GB)
    router.MaxMultipartMemory = 4 << 30
    
    // API路由组
    api := router.Group("/api")
    {
        // 健康检查
        api.GET("/health", healthCheck)
        
        // 检查登录状态
        api.POST("/check-login", checkLogin)
        
        // 获取上传配置
        api.POST("/pre-upload", preUpload)
        
        // 上传视频文件
        api.POST("/upload", uploadVideo)
        
        // 提交视频信息
        api.POST("/submit", submitVideo)
    }
    
    // 静态文件服务（如果需要提供前端页面）
    router.Static("/static", "./static")
    
    // 启动服务器
    port := ":8080"
    log.Printf("服务器启动在 http://localhost%s", port)
    log.Printf("API地址: http://localhost%s/api", port)
    
    if err := router.Run(port); err != nil {
        log.Fatalf("服务器启动失败: %v", err)
    }
}

// healthCheck 健康检查接口
func healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "message": "B站上传服务运行正常",
    })
}

// checkLogin 检查B站登录状态
func checkLogin(c *gin.Context) {
    // 接收前端传来的cookie
    var req struct {
        Cookie string `json:"cookie" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "缺少cookie参数",
        })
        return
    }
    
    // TODO: 这里将调用B站API检查登录状态
    // 暂时返回模拟数据
    c.JSON(http.StatusOK, gin.H{
        "logged_in": true,
        "username": "测试用户",
        "message": "登录状态检查完成",
    })
}

// preUpload 预上传接口，获取上传地址
func preUpload(c *gin.Context) {
    var req struct {
        Filename string `json:"filename" binding:"required"`
        Filesize int64  `json:"filesize" binding:"required"`
        Cookie   string `json:"cookie" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "参数错误",
        })
        return
    }
    
    log.Printf("预上传请求: 文件名=%s, 大小=%d", req.Filename, req.Filesize)
    
    // TODO: 调用B站预上传API
    // 暂时返回模拟数据
    c.JSON(http.StatusOK, gin.H{
        "upload_url": "https://example.com/upload",
        "biz_id": "123456",
        "chunk_size": 5242880, // 5MB
        "message": "获取上传地址成功",
    })
}

// uploadVideo 处理视频上传
func uploadVideo(c *gin.Context) {
    // 获取上传的文件
    file, header, err := c.Request.FormFile("video")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "获取文件失败",
        })
        return
    }
    defer file.Close()
    
    // 获取其他表单参数
    // cookie := c.PostForm("cookie")
    title := c.PostForm("title")
    
    log.Printf("收到文件上传: %s (%.2f MB)", header.Filename, float64(header.Size)/(1024*1024))
    log.Printf("视频标题: %s", title)
    
    // TODO: 实现分片上传到B站
    // 这里需要：
    // 1. 将文件分片
    // 2. 逐片上传到B站
    // 3. 返回上传进度
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "filename": header.Filename,
        "size": header.Size,
        "message": "文件上传成功",
    })
}

// submitVideo 提交视频信息到B站
func submitVideo(c *gin.Context) {
    var req struct {
        Cookie      string   `json:"cookie" binding:"required"`
        Title       string   `json:"title" binding:"required"`
        Description string   `json:"description"`
        Tags        []string `json:"tags"`
        Category    int      `json:"category"`
        Filename    string   `json:"filename" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "参数错误",
        })
        return
    }
    
    log.Printf("提交视频: %s", req.Title)
    
    // TODO: 调用B站提交API
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "bvid": "BV1xx411c7mD", // 模拟的BV号
        "message": "视频投稿成功",
    })
}