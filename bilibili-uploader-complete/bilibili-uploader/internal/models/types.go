package models

import "time"

// UploadInitRequest 上传初始化请求
type UploadInitRequest struct {
	Filename string `json:"filename" binding:"required"`
	Filesize int64  `json:"filesize" binding:"required"`
}

// UploadInitResponse 上传初始化响应
type UploadInitResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		UploadID  string `json:"upload_id"`
		BizID     int64  `json:"biz_id"`
		Endpoint  string `json:"endpoint"`
		ChunkSize int64  `json:"chunk_size"`
		Chunks    int    `json:"chunks"`
		Auth      string `json:"auth"`
	} `json:"data"`
}

// ChunkUploadRequest 分片上传请求
type ChunkUploadRequest struct {
	UploadID    string `form:"upload_id" binding:"required"`
	BizID       int64  `form:"biz_id" binding:"required"`
	ChunkIndex  int    `form:"chunk_index" binding:"required"`
	TotalChunks int    `form:"total_chunks" binding:"required"`
	MD5         string `form:"md5" binding:"required"`
	CRC32       string `form:"crc32" binding:"required"`
	Hash        string `form:"hash" binding:"required"`
	Auth        string `form:"auth" binding:"required"`
}

// CompleteUploadRequest 完成上传请求
type CompleteUploadRequest struct {
	UploadID string `json:"upload_id" binding:"required"`
	BizID    int64  `json:"biz_id" binding:"required"`
	Filename string `json:"filename" binding:"required"`
	Filesize int64  `json:"filesize" binding:"required"`
	MD5      string `json:"md5" binding:"required"`
	CRC32    string `json:"crc32" binding:"required"`
	Hash     string `json:"hash" binding:"required"`
	Auth     string `json:"auth" binding:"required"`
}

// CompleteUploadResponse 完成上传响应
type CompleteUploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		UposURI string `json:"upos_uri"`
	} `json:"data"`
}

// SubmitVideoRequest 提交视频稿件请求
type SubmitVideoRequest struct {
	UposURI   string `json:"upos_uri" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Desc      string `json:"desc" binding:"required"`
	Tags      string `json:"tags" binding:"required"`
	TID       int    `json:"tid" binding:"required"`
	Cover     string `json:"cover"`
	Copyright int    `json:"copyright" binding:"required"` // 1:原创, 2:转载
	Source    string `json:"source"`                      // 转载来源
}

// SubmitVideoResponse 提交视频稿件响应
type SubmitVideoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AID  int64  `json:"aid"`
		BVID string `json:"bvid"`
	} `json:"data"`
}

// CoverUploadResponse 封面上传响应
type CoverUploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		URL string `json:"url"`
	} `json:"data"`
}

// BilibiliPreuploadResponse Bilibili预上传响应
type BilibiliPreuploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Endpoint  string `json:"endpoint"`
		BizID     int64  `json:"biz_id"`
		UposURI   string `json:"upos_uri"`
		ChunkSize int64  `json:"chunk_size"`
		Chunks    int    `json:"chunks"`
		TotalSize int64  `json:"total_size"`
		UploadID  string `json:"upload_id"`
		Auth      string `json:"auth"`
		OK        bool   `json:"ok"`
	} `json:"data"`
}

// BilibiliUploadResponse Bilibili上传响应
type BilibiliUploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		UploadID  string `json:"upload_id"`
		BizID     int64  `json:"biz_id"`
		UposURI   string `json:"upos_uri"`
		ChunkSize int64  `json:"chunk_size"`
		Chunks    int    `json:"chunks"`
		TotalSize int64  `json:"total_size"`
		Auth      string `json:"auth"`
		OK        bool   `json:"ok"`
	} `json:"data"`
}

// BilibiliSubmitResponse Bilibili提交稿件响应
type BilibiliSubmitResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		AID  int64  `json:"aid"`
		BVID string `json:"bvid"`
	} `json:"data"`
}

// BilibiliCoverResponse Bilibili封面上传响应
type BilibiliCoverResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		URL string `json:"url"`
	} `json:"data"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// UploadSession 上传会话
type UploadSession struct {
	UploadID  string    `json:"upload_id"`
	BizID     int64     `json:"biz_id"`
	Endpoint  string    `json:"endpoint"`
	ChunkSize int64     `json:"chunk_size"`
	Chunks    int       `json:"chunks"`
	Auth      string    `json:"auth"`
	Filename  string    `json:"filename"`
	Filesize  int64     `json:"filesize"`
	CreatedAt time.Time `json:"created_at"`
}

