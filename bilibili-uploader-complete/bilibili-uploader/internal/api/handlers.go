package api

import (
	"bilibili-uploader/internal/bilibili"
	"bilibili-uploader/internal/models"
	"bilibili-uploader/internal/utils"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler API处理器
type Handler struct {
	bilibiliClient *bilibili.Client
	uploadSessions map[string]*models.UploadSession
	sessionMutex   sync.RWMutex
	uploadDir      string
}

// NewHandler 创建新的API处理器
func NewHandler(sessdata, biliJct, uploadDir string) *Handler {
	return &Handler{
		bilibiliClient: bilibili.NewClient(sessdata, biliJct),
		uploadSessions: make(map[string]*models.UploadSession),
		uploadDir:      uploadDir,
	}
}

// InitUpload 初始化上传
func (h *Handler) InitUpload(c *gin.Context) {
	var req models.UploadInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证文件类型
	if !utils.IsVideoFile(req.Filename) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "不支持的视频文件格式",
		})
		return
	}

	// 调用Bilibili预上传接口
	preuploadResp, err := h.bilibiliClient.Preupload(req.Filename, req.Filesize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "预上传失败: " + err.Error(),
		})
		return
	}

	// 保存上传会话
	session := &models.UploadSession{
		UploadID:  preuploadResp.Data.UploadID,
		BizID:     preuploadResp.Data.BizID,
		Endpoint:  preuploadResp.Data.Endpoint,
		ChunkSize: preuploadResp.Data.ChunkSize,
		Chunks:    preuploadResp.Data.Chunks,
		Auth:      preuploadResp.Data.Auth,
		Filename:  req.Filename,
		Filesize:  req.Filesize,
		CreatedAt: time.Now(),
	}

	h.sessionMutex.Lock()
	h.uploadSessions[session.UploadID] = session
	h.sessionMutex.Unlock()

	// 返回响应
	resp := models.UploadInitResponse{
		Code:    0,
		Message: "初始化成功",
	}
	resp.Data.UploadID = session.UploadID
	resp.Data.BizID = session.BizID
	resp.Data.Endpoint = session.Endpoint
	resp.Data.ChunkSize = session.ChunkSize
	resp.Data.Chunks = session.Chunks
	resp.Data.Auth = session.Auth

	c.JSON(http.StatusOK, resp)
}

// UploadChunk 上传分片
func (h *Handler) UploadChunk(c *gin.Context) {
	var req models.ChunkUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取上传会话
	h.sessionMutex.RLock()
	session, exists := h.uploadSessions[req.UploadID]
	h.sessionMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Code:    -1,
			Message: "上传会话不存在",
		})
		return
	}

	// 获取上传的文件分片
	fileHeader, err := c.FormFile("file_chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "获取文件分片失败: " + err.Error(),
		})
		return
	}

	// 读取分片数据
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "打开文件分片失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	chunkData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "读取文件分片失败: " + err.Error(),
		})
		return
	}

	// 上传分片到Bilibili
	_, err = h.bilibiliClient.UploadChunk(
		session.Endpoint,
		req.UploadID,
		req.BizID,
		req.ChunkIndex,
		req.TotalChunks,
		chunkData,
		req.MD5,
		req.CRC32,
		req.Hash,
		req.Auth,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "上传分片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "分片上传成功",
	})
}

// CompleteUpload 完成上传
func (h *Handler) CompleteUpload(c *gin.Context) {
	var req models.CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取上传会话
	h.sessionMutex.RLock()
	session, exists := h.uploadSessions[req.UploadID]
	h.sessionMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Code:    -1,
			Message: "上传会话不存在",
		})
		return
	}

	// 调用Bilibili完成上传接口
	completeResp, err := h.bilibiliClient.CompleteUpload(
		session.Endpoint,
		req.UploadID,
		req.BizID,
		req.Filename,
		req.Filesize,
		req.MD5,
		req.CRC32,
		req.Hash,
		req.Auth,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "完成上传失败: " + err.Error(),
		})
		return
	}

	// 清理上传会话
	h.sessionMutex.Lock()
	delete(h.uploadSessions, req.UploadID)
	h.sessionMutex.Unlock()

	// 返回响应
	resp := models.CompleteUploadResponse{
		Code:    0,
		Message: "上传完成",
	}
	resp.Data.UposURI = completeResp.Data.UposURI

	c.JSON(http.StatusOK, resp)
}

// SubmitVideo 提交视频稿件
func (h *Handler) SubmitVideo(c *gin.Context) {
	var req models.SubmitVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 调用Bilibili提交稿件接口
	submitResp, err := h.bilibiliClient.SubmitVideo(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "提交稿件失败: " + err.Error(),
		})
		return
	}

	// 返回响应
	resp := models.SubmitVideoResponse{
		Code:    0,
		Message: "稿件提交成功",
	}
	resp.Data.AID = submitResp.Data.AID
	resp.Data.BVID = submitResp.Data.BVID

	c.JSON(http.StatusOK, resp)
}

// UploadCover 上传封面
func (h *Handler) UploadCover(c *gin.Context) {
	// 获取上传的封面文件
	fileHeader, err := c.FormFile("cover_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "获取封面文件失败: " + err.Error(),
		})
		return
	}

	// 验证文件类型
	if !utils.IsImageFile(fileHeader.Filename) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "不支持的图片文件格式",
		})
		return
	}

	// 读取文件数据
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "打开封面文件失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "读取封面文件失败: " + err.Error(),
		})
		return
	}

	// 获取MIME类型
	mimeType := utils.GetMimeType(fileHeader.Filename)

	// 上传封面到Bilibili
	coverResp, err := h.bilibiliClient.UploadCover(imageData, mimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    -1,
			Message: "上传封面失败: " + err.Error(),
		})
		return
	}

	// 返回响应
	resp := models.CoverUploadResponse{
		Code:    0,
		Message: "封面上传成功",
	}
	resp.Data.URL = coverResp.Data.URL

	c.JSON(http.StatusOK, resp)
}

// GetUploadStatus 获取上传状态
func (h *Handler) GetUploadStatus(c *gin.Context) {
	uploadID := c.Param("upload_id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    -1,
			Message: "上传ID不能为空",
		})
		return
	}

	h.sessionMutex.RLock()
	session, exists := h.uploadSessions[uploadID]
	h.sessionMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Code:    -1,
			Message: "上传会话不存在",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取状态成功",
		Data:    session,
	})
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "服务正常运行",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		},
	})
}

// CleanupSessions 清理过期的上传会话
func (h *Handler) CleanupSessions() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			h.sessionMutex.Lock()
			now := time.Now()
			for uploadID, session := range h.uploadSessions {
				// 清理超过1小时的会话
				if now.Sub(session.CreatedAt) > time.Hour {
					delete(h.uploadSessions, uploadID)
				}
			}
			h.sessionMutex.Unlock()
		}
	}()
}

