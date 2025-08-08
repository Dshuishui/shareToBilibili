// services/bilibili_oauth.go - B站OAuth授权和自动投稿服务
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "time"
)

// BilibiliOAuth B站OAuth服务
type BilibiliOAuth struct {
    ClientID     string
    ClientSecret string
    RedirectURI  string
    BaseURL      string
}

// NewBilibiliOAuth 创建B站OAuth服务
func NewBilibiliOAuth(clientID, clientSecret, redirectURI string) *BilibiliOAuth {
    return &BilibiliOAuth{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURI:  redirectURI,
        BaseURL:      "https://api.bilibili.com",
    }
}

// TokenResponse OAuth token响应
type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
}

// UserInfo 用户信息
type UserInfo struct {
    UID      int64  `json:"uid"`
    Username string `json:"uname"`
    Face     string `json:"face"`
}

// ExchangeCode 用授权码换取访问令牌
func (b *BilibiliOAuth) ExchangeCode(code string) (*TokenResponse, error) {
    // 构建请求参数
    data := url.Values{}
    data.Set("client_id", b.ClientID)
    data.Set("client_secret", b.ClientSecret)
    data.Set("grant_type", "authorization_code")
    data.Set("code", code)
    data.Set("redirect_uri", b.RedirectURI)
    
    // 发送请求
    resp, err := http.PostForm("https://passport.bilibili.com/api/oauth2/access_token", data)
    if err != nil {
        return nil, fmt.Errorf("请求token失败: %v", err)
    }
    defer resp.Body.Close()
    
    // 解析响应
    var tokenResp TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return nil, fmt.Errorf("解析token响应失败: %v", err)
    }
    
    return &tokenResp, nil
}

// GetUserInfo 获取用户信息
func (b *BilibiliOAuth) GetUserInfo(accessToken string) (*UserInfo, error) {
    req, err := http.NewRequest("GET", b.BaseURL+"/x/web-interface/nav", nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+accessToken)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Code int      `json:"code"`
        Data UserInfo `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    if result.Code != 0 {
        return nil, fmt.Errorf("获取用户信息失败: code=%d", result.Code)
    }
    
    return &result.Data, nil
}

// BilibiliUploader B站视频上传器
type BilibiliUploader struct {
    AccessToken string
    BaseURL     string
}

// NewBilibiliUploader 创建上传器
func NewBilibiliUploader(accessToken string) *BilibiliUploader {
    return &BilibiliUploader{
        AccessToken: accessToken,
        BaseURL:     "https://member.bilibili.com",
    }
}

// VideoUploadParams 视频上传参数
type VideoUploadParams struct {
    Title       string   `json:"title"`       // 标题
    Description string   `json:"desc"`        // 简介
    Tags        []string `json:"tags"`        // 标签
    Category    int      `json:"tid"`         // 分区ID
    Cover       string   `json:"cover"`       // 封面URL
    Source      string   `json:"source"`      // 来源
    Copyright   int      `json:"copyright"`   // 1:自制 2:转载
}

// UploadVideo 上传视频到B站
func (u *BilibiliUploader) UploadVideo(videoPath string, params VideoUploadParams) (string, error) {
    // Step 1: 预上传，获取上传地址
    uploadInfo, err := u.preUpload(videoPath)
    if err != nil {
        return "", fmt.Errorf("预上传失败: %v", err)
    }
    
    // Step 2: 分片上传视频文件
    if err := u.uploadChunks(videoPath, uploadInfo); err != nil {
        return "", fmt.Errorf("上传视频失败: %v", err)
    }
    
    // Step 3: 提交稿件
    bvid, err := u.submitVideo(uploadInfo.Filename, params)
    if err != nil {
        return "", fmt.Errorf("提交稿件失败: %v", err)
    }
    
    return bvid, nil
}

// UploadInfo 上传信息
type UploadInfo struct {
    URL      string `json:"url"`
    Filename string `json:"filename"`
    BizID    int    `json:"biz_id"`
}

// preUpload 预上传
func (u *BilibiliUploader) preUpload(videoPath string) (*UploadInfo, error) {
    file, err := os.Open(videoPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    fileInfo, _ := file.Stat()
    
    // 构建请求
    params := url.Values{}
    params.Set("name", filepath.Base(videoPath))
    params.Set("size", fmt.Sprintf("%d", fileInfo.Size()))
    params.Set("r", "upos")
    params.Set("profile", "ugcupos/bup")
    
    req, err := http.NewRequest("GET", 
        fmt.Sprintf("%s/preupload?%s", u.BaseURL, params.Encode()), nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+u.AccessToken)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var uploadInfo UploadInfo
    if err := json.NewDecoder(resp.Body).Decode(&uploadInfo); err != nil {
        return nil, err
    }
    
    uploadInfo.Filename = filepath.Base(videoPath)
    return &uploadInfo, nil
}

// uploadChunks 分片上传
func (u *BilibiliUploader) uploadChunks(videoPath string, uploadInfo *UploadInfo) error {
    file, err := os.Open(videoPath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    fileInfo, _ := file.Stat()
    fileSize := fileInfo.Size()
    
    // 分片大小：5MB
    chunkSize := int64(5 * 1024 * 1024)
    chunks := (fileSize + chunkSize - 1) / chunkSize
    
    for i := int64(0); i < chunks; i++ {
        start := i * chunkSize
        end := start + chunkSize
        if end > fileSize {
            end = fileSize
        }
        
        // 读取分片数据
        chunkData := make([]byte, end-start)
        _, err := file.ReadAt(chunkData, start)
        if err != nil && err != io.EOF {
            return fmt.Errorf("读取分片失败: %v", err)
        }
        
        // 上传分片
        if err := u.uploadChunk(uploadInfo.URL, chunkData, i+1, chunks); err != nil {
            return fmt.Errorf("上传分片%d失败: %v", i+1, err)
        }
    }
    
    return nil
}

// uploadChunk 上传单个分片
func (u *BilibiliUploader) uploadChunk(uploadURL string, data []byte, partNum, totalParts int64) error {
    url := fmt.Sprintf("%s?partNumber=%d&parts=%d", uploadURL, partNum, totalParts)
    
    req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/octet-stream")
    req.Header.Set("Authorization", "Bearer "+u.AccessToken)
    
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("上传分片失败: status=%d", resp.StatusCode)
    }
    
    return nil
}

// submitVideo 提交稿件
func (u *BilibiliUploader) submitVideo(filename string, params VideoUploadParams) (string, error) {
    // 构建提交数据
    submitData := map[string]interface{}{
        "copyright": params.Copyright,
        "source":    params.Source,
        "title":     params.Title,
        "tid":       params.Category,
        "tag":       joinTags(params.Tags),
        "desc":      params.Description,
        "cover":     params.Cover,
        "videos": []map[string]string{
            {
                "filename": filename,
                "title":    params.Title,
                "desc":     "",
            },
        },
    }
    
    jsonData, err := json.Marshal(submitData)
    if err != nil {
        return "", err
    }
    
    req, err := http.NewRequest("POST", u.BaseURL+"/web/add", bytes.NewBuffer(jsonData))
    if err != nil {
        return "", err
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+u.AccessToken)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var result struct {
        Code int    `json:"code"`
        BVid string `json:"bvid"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }
    
    if result.Code != 0 {
        return "", fmt.Errorf("提交稿件失败: code=%d", result.Code)
    }
    
    return result.BVid, nil
}

// AutoUploadWithOAuth 使用OAuth自动上传视频
func AutoUploadWithOAuth(accessToken, videoPath string, params VideoUploadParams) (string, error) {
    uploader := NewBilibiliUploader(accessToken)
    return uploader.UploadVideo(videoPath, params)
}

// joinTags 连接标签
func joinTags(tags []string) string {
    if len(tags) == 0 {
        return ""
    }
    result := ""
    for i, tag := range tags {
        if i > 0 {
            result += ","
        }
        result += tag
    }
    return result
}

// ===== 模拟实现（用于测试） =====

// SimulatedUpload 模拟上传（用于开发测试）
func SimulatedUpload(videoPath string, params VideoUploadParams) (string, error) {
    // 模拟上传延迟
    time.Sleep(3 * time.Second)
    
    // 生成模拟的BV号
    bvid := fmt.Sprintf("BV1%s%d", generateRandomString(8), time.Now().Unix()%1000)
    
    fmt.Printf("模拟上传视频:\n")
    fmt.Printf("  文件: %s\n", videoPath)
    fmt.Printf("  标题: %s\n", params.Title)
    fmt.Printf("  简介: %s\n", params.Description)
    fmt.Printf("  分区: %d\n", params.Category)
    fmt.Printf("  标签: %v\n", params.Tags)
    fmt.Printf("  BV号: %s\n", bvid)
    
    return bvid, nil
}

func generateRandomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, length)
    for i := range result {
        result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
    }
    return string(result)
}

// CreateMultipartUpload 创建分片上传请求
func CreateMultipartUpload(videoPath string, params map[string]string) (*bytes.Buffer, string, error) {
    file, err := os.Open(videoPath)
    if err != nil {
        return nil, "", err
    }
    defer file.Close()
    
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    // 添加文件
    part, err := writer.CreateFormFile("video", filepath.Base(videoPath))
    if err != nil {
        return nil, "", err
    }
    
    _, err = io.Copy(part, file)
    if err != nil {
        return nil, "", err
    }
    
    // 添加其他参数
    for key, val := range params {
        _ = writer.WriteField(key, val)
    }
    
    err = writer.Close()
    if err != nil {
        return nil, "", err
    }
    
    return body, writer.FormDataContentType(), nil
}