package bilibili

import (
	"bilibili-uploader/internal/models"
	"bilibili-uploader/internal/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client Bilibili API客户端
type Client struct {
	httpClient *http.Client
	sessdata   string
	biliJct    string
	userAgent  string
}

// NewClient 创建新的Bilibili客户端
func NewClient(sessdata, biliJct string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		sessdata:  sessdata,
		biliJct:   biliJct,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}
}

// addAuthHeaders 添加认证头
func (c *Client) addAuthHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", "https://member.bilibili.com/")
	req.Header.Set("Origin", "https://member.bilibili.com")
	
	// 添加Cookie
	cookie := fmt.Sprintf("SESSDATA=%s; bili_jct=%s", c.sessdata, c.biliJct)
	req.Header.Set("Cookie", cookie)
}

// Preupload 预上传，获取上传元数据
func (c *Client) Preupload(filename string, filesize int64) (*models.BilibiliPreuploadResponse, error) {
	baseURL := "https://member.bilibili.com/preupload"
	
	// 构建查询参数
	params := url.Values{}
	params.Set("name", filename)
	params.Set("r", "upos")
	params.Set("profile", "ugcfx/bup")
	params.Set("probe_version", "20221109")
	params.Set("upcdn", "txa")
	params.Set("zone", "sh001")
	
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建预上传请求失败: %w", err)
	}
	
	c.addAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送预上传请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取预上传响应失败: %w", err)
	}
	
	var result models.BilibiliPreuploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析预上传响应失败: %w", err)
	}
	
	if result.Code != 0 {
		return nil, fmt.Errorf("预上传失败: %s", result.Message)
	}
	
	return &result, nil
}

// UploadMetadata 上传视频元数据
func (c *Client) UploadMetadata(endpoint, uploadID string, bizID int64, filename string, filesize int64, fileMD5, fileCRC32, fileHash, auth string, chunkSize int64, chunks int) (*models.BilibiliUploadResponse, error) {
	// 构建查询参数
	params := url.Values{}
	params.Set("output", "json")
	params.Set("profile", "ugcfx/bup")
	params.Set("upload_id", uploadID)
	params.Set("biz_id", strconv.FormatInt(bizID, 10))
	params.Set("upcdn", "txa")
	params.Set("file_size", strconv.FormatInt(filesize, 10))
	params.Set("file_name", filename)
	params.Set("file_md5", fileMD5)
	params.Set("file_crc32", fileCRC32)
	params.Set("file_hash", fileHash)
	params.Set("chunk_size", strconv.FormatInt(chunkSize, 10))
	params.Set("chunks", strconv.Itoa(chunks))
	params.Set("total_size", strconv.FormatInt(filesize, 10))
	params.Set("auth", auth)
	
	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建上传元数据请求失败: %w", err)
	}
	
	c.addAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送上传元数据请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取上传元数据响应失败: %w", err)
	}
	
	var result models.BilibiliUploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析上传元数据响应失败: %w", err)
	}
	
	if result.Code != 0 {
		return nil, fmt.Errorf("上传元数据失败: %s", result.Message)
	}
	
	return &result, nil
}

// UploadChunk 上传视频分片
func (c *Client) UploadChunk(endpoint, uploadID string, bizID int64, chunkIndex, totalChunks int, chunkData []byte, chunkMD5, chunkCRC32, chunkHash, auth string) (*models.BilibiliUploadResponse, error) {
	// 构建查询参数
	params := url.Values{}
	params.Set("output", "json")
	params.Set("profile", "ugcfx/bup")
	params.Set("upload_id", uploadID)
	params.Set("biz_id", strconv.FormatInt(bizID, 10))
	params.Set("upcdn", "txa")
	params.Set("chunk", strconv.Itoa(chunkIndex))
	params.Set("chunks", strconv.Itoa(totalChunks))
	params.Set("size", strconv.Itoa(len(chunkData)))
	params.Set("total", strconv.Itoa(len(chunkData)))
	params.Set("md5", chunkMD5)
	params.Set("crc32", chunkCRC32)
	params.Set("hash", chunkHash)
	params.Set("auth", auth)
	
	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	
	req, err := http.NewRequest("PUT", fullURL, bytes.NewReader(chunkData))
	if err != nil {
		return nil, fmt.Errorf("创建上传分片请求失败: %w", err)
	}
	
	c.addAuthHeaders(req)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送上传分片请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取上传分片响应失败: %w", err)
	}
	
	var result models.BilibiliUploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析上传分片响应失败: %w", err)
	}
	
	if result.Code != 0 {
		return nil, fmt.Errorf("上传分片失败: %s", result.Message)
	}
	
	return &result, nil
}

// CompleteUpload 完成上传
func (c *Client) CompleteUpload(endpoint, uploadID string, bizID int64, filename string, filesize int64, fileMD5, fileCRC32, fileHash, auth string) (*models.BilibiliUploadResponse, error) {
	// 构建查询参数
	params := url.Values{}
	params.Set("output", "json")
	params.Set("profile", "ugcfx/bup")
	params.Set("upload_id", uploadID)
	params.Set("biz_id", strconv.FormatInt(bizID, 10))
	params.Set("upcdn", "txa")
	params.Set("file_size", strconv.FormatInt(filesize, 10))
	params.Set("file_name", filename)
	params.Set("file_md5", fileMD5)
	params.Set("file_crc32", fileCRC32)
	params.Set("file_hash", fileHash)
	params.Set("auth", auth)
	
	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建完成上传请求失败: %w", err)
	}
	
	c.addAuthHeaders(req)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送完成上传请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取完成上传响应失败: %w", err)
	}
	
	var result models.BilibiliUploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析完成上传响应失败: %w", err)
	}
	
	if result.Code != 0 {
		return nil, fmt.Errorf("完成上传失败: %s", result.Message)
	}
	
	return &result, nil
}

// SubmitVideo 提交视频稿件
func (c *Client) SubmitVideo(req *models.SubmitVideoRequest) (*models.BilibiliSubmitResponse, error) {
	submitURL := "https://member.bilibili.com/x/vu/web/add"
	
	// 构建表单数据
	data := url.Values{}
	data.Set("csrf", c.biliJct)
	data.Set("copyright", strconv.Itoa(req.Copyright))
	if req.Source != "" {
		data.Set("source", req.Source)
	}
	data.Set("tid", strconv.Itoa(req.TID))
	data.Set("cover", req.Cover)
	data.Set("title", req.Title)
	data.Set("desc", req.Desc)
	data.Set("desc_format_id", "0")
	data.Set("tag", req.Tags)
	
	// 构建视频信息JSON
	videos := []map[string]interface{}{
		{
			"filename": req.UposURI,
			"title":    req.Title,
			"desc":     req.Desc,
		},
	}
	videosJSON, _ := json.Marshal(videos)
	data.Set("videos", string(videosJSON))
	
	httpReq, err := http.NewRequest("POST", submitURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建提交稿件请求失败: %w", err)
	}
	
	c.addAuthHeaders(httpReq)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送提交稿件请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取提交稿件响应失败: %w", err)
	}
	
	var result models.BilibiliSubmitResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析提交稿件响应失败: %w", err)
	}
	
	if result.Code != 0 {
		return nil, fmt.Errorf("提交稿件失败: %s", result.Message)
	}
	
	return &result, nil
}

// UploadCover 上传封面
func (c *Client) UploadCover(imageData []byte, mimeType string) (*models.BilibiliCoverResponse, error) {
	coverURL := "https://member.bilibili.com/x/vu/web/cover/up"
	
	// 将图片编码为base64
	base64Image := utils.EncodeImageToBase64(imageData, mimeType)
	
	// 构建表单数据
	data := url.Values{}
	data.Set("csrf", c.biliJct)
	data.Set("cover", base64Image)
	
	req, err := http.NewRequest("POST", coverURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建上传封面请求失败: %w", err)
	}
	
	c.addAuthHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送上传封面请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取上传封面响应失败: %w", err)
	}
	
	var result models.BilibiliCoverResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析上传封面响应失败: %w", err)
	}
	
	if result.Code != 0 {
		return nil, fmt.Errorf("上传封面失败: %s", result.Message)
	}
	
	return &result, nil
}

