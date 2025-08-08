// handlers/upload.go - 视频上传处理
package handlers

import (
    "bilibili-uploader/config"
    "bilibili-uploader/services"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
)

// UploadHandler 上传处理器
type UploadHandler struct {
    uploadService *services.BilibiliUploader
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler() *UploadHandler {
    return &UploadHandler{}
}

// UploadToBilibili 上传视频到B站
func (h *UploadHandler) UploadToBilibili(c *gin.Context) {
    // 从context获取B站token
    bilibiliToken, err := GetBilibiliToken(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "请先授权B站账号",
        })
        return
    }
    
    // 解析multipart form
    err = c.Request.ParseMultipartForm(32 << 20) // 32MB
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "解析表单失败",
        })
        return
    }
    
    // 获取文件
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
    description := c.PostForm("desc")
    tags := c.PostForm("tags")
    categoryStr := c.PostForm("category")
    
    // 验证必填字段
    if title == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "视频标题不能为空",
        })
        return
    }
    
    // 解析分区ID
    category := 21 // 默认日常分区
    if categoryStr != "" {
        if cat, err := strconv.Atoi(categoryStr); err == nil {
            category = cat
        }
    }
    
    // 处理标签
    tagList := []string{}
    if tags != "" {
        tagList = strings.Split(tags, " ")
    }
    
    log.Printf("📤 开始上传到B站:")
    log.Printf("   文件: %s (%.2f MB)", header.Filename, float64(header.Size)/(1024*1024))
    log.Printf("   标题: %s", title)
    log.Printf("   分区: %d", category)
    log.Printf("   标签: %v", tagList)
    
    // 保存到临时文件
    tempDir := config.GlobalConfig.TempDir
    tempFile := filepath.Join(tempDir, fmt.Sprintf("upload_%d_%s", time.Now().Unix(), header.Filename))
    
    dst, err := os.Create(tempFile)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "创建临时文件失败",
        })
        return
    }
    defer dst.Close()
    defer os.Remove(tempFile) // 上传完成后删除临时文件
    
    // 复制文件内容
    written, err := io.Copy(dst, file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "保存文件失败",
        })
        return
    }
    
    log.Printf("✅ 临时文件已保存: %s (%d bytes)", tempFile, written)
    
    // 创建上传参数
    uploadParams := services.VideoUploadParams{
        Title:       title,
        Description: description,
        Tags:        tagList,
        Category:    category,
        Copyright:   1, // 自制
    }
    
    // 使用进度回调上传
    progressChan := make(chan float64, 100)
    go func() {
        for progress := range progressChan {
            log.Printf("📊 上传进度: %.2f%%", progress)
            // 可以通过WebSocket发送给前端
        }
    }()
    
    // 创建上传器并执行上传
    uploader := services.NewBilibiliUploader(bilibiliToken)
    
    log.Println("🚀 开始执行B站API上传...")
    bvid, err := uploader.UploadVideo(tempFile, uploadParams)
    close(progressChan)
    
    if err != nil {
        log.Printf("❌ 上传失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": fmt.Sprintf("上传失败: %v", err),
        })
        return
    }
    
    log.Printf("🎉 上传成功! BV号: %s", bvid)
    
    // 返回成功结果
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "视频投稿成功",
        "data": gin.H{
            "bvid": bvid,
            "url":  fmt.Sprintf("https://www.bilibili.com/video/%s", bvid),
            "title": title,
            "upload_time": time.Now().Format("2006-01-02 15:04:05"),
        },
    })
}

// UploadProgress 上传进度（WebSocket）
func (h *UploadHandler) UploadProgress(c *gin.Context) {
    // 这里可以实现WebSocket来实时推送上传进度
    // 暂时返回模拟进度
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "progress": 50,
        "status": "uploading",
    })
}

// GetUploadHistory 获取上传历史
func (h *UploadHandler) GetUploadHistory(c *gin.Context) {
    // 从数据库或缓存获取用户的上传历史
    // 这里返回模拟数据
    history := []gin.H{
        {
            "bvid": "BV1xx411c7mD",
            "title": "测试视频1",
            "upload_time": "2024-01-01 12:00:00",
            "status": "published",
        },
        {
            "bvid": "BV1xx411c7mE", 
            "title": "测试视频2",
            "upload_time": "2024-01-02 13:00:00",
            "status": "reviewing",
        },
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": history,
    })
}

// CheckUploadStatus 检查上传状态
func (h *UploadHandler) CheckUploadStatus(c *gin.Context) {
    bvid := c.Param("bvid")
    
    if bvid == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "BV号不能为空",
        })
        return
    }
    
    // 这里应该调用B站API检查视频状态
    // 暂时返回模拟数据
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "bvid": bvid,
            "status": "published", // reviewing, published, failed
            "play_count": 1234,
            "like_count": 56,
            "coin_count": 12,
        },
    })
}