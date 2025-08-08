// services/video_processor.go - 视频处理服务
package services

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// VideoProcessor 视频处理器
type VideoProcessor struct {
    InputDir  string
    OutputDir string
}

// NewVideoProcessor 创建视频处理器
func NewVideoProcessor(inputDir, outputDir string) *VideoProcessor {
    return &VideoProcessor{
        InputDir:  inputDir,
        OutputDir: outputDir,
    }
}

// ProcessVideo 处理视频文件
func (vp *VideoProcessor) ProcessVideo(filename string, options ProcessOptions) (string, error) {
    inputPath := filepath.Join(vp.InputDir, filename)
    outputFilename := fmt.Sprintf("bilibili_%s", filename)
    outputPath := filepath.Join(vp.OutputDir, outputFilename)
    
    // 检查输入文件是否存在
    if _, err := os.Stat(inputPath); os.IsNotExist(err) {
        return "", fmt.Errorf("输入文件不存在: %s", inputPath)
    }
    
    // 构建FFmpeg命令
    args := vp.buildFFmpegArgs(inputPath, outputPath, options)
    
    // 执行FFmpeg命令
    cmd := exec.Command("ffmpeg", args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    log.Printf("执行FFmpeg命令: ffmpeg %s", strings.Join(args, " "))
    
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("FFmpeg处理失败: %v", err)
    }
    
    return outputFilename, nil
}

// ProcessOptions 视频处理选项
type ProcessOptions struct {
    Resolution string // 分辨率: 1080p, 720p, 480p
    Bitrate    string // 码率: 6000k, 4000k, 2000k
    Format     string // 格式: mp4, flv
    Quality    string // 质量: high, medium, low
}

// buildFFmpegArgs 构建FFmpeg参数
func (vp *VideoProcessor) buildFFmpegArgs(input, output string, options ProcessOptions) []string {
    args := []string{
        "-i", input, // 输入文件
        "-y",        // 覆盖输出文件
    }
    
    // 根据B站推荐设置参数
    switch options.Quality {
    case "high":
        // 高质量设置（适合1080P）
        args = append(args,
            "-c:v", "libx264",      // 使用H.264编码
            "-preset", "slow",       // 慢速编码，更好的压缩
            "-crf", "19",           // 质量因子（越低质量越高）
            "-c:a", "aac",          // 音频编码
            "-b:a", "320k",         // 音频码率
            "-ar", "48000",         // 音频采样率
            "-maxrate", "6000k",    // 最大码率
            "-bufsize", "12000k",   // 缓冲区大小
        )
    case "medium":
        // 中等质量（适合720P）
        args = append(args,
            "-c:v", "libx264",
            "-preset", "medium",
            "-crf", "23",
            "-c:a", "aac",
            "-b:a", "192k",
            "-ar", "44100",
            "-maxrate", "4000k",
            "-bufsize", "8000k",
        )
    default:
        // 低质量/快速处理
        args = append(args,
            "-c:v", "libx264",
            "-preset", "fast",
            "-crf", "28",
            "-c:a", "aac",
            "-b:a", "128k",
            "-ar", "44100",
            "-maxrate", "2000k",
            "-bufsize", "4000k",
        )
    }
    
    // 设置分辨率
    if options.Resolution != "" {
        switch options.Resolution {
        case "1080p":
            args = append(args, "-vf", "scale=-2:1080")
        case "720p":
            args = append(args, "-vf", "scale=-2:720")
        case "480p":
            args = append(args, "-vf", "scale=-2:480")
        }
    }
    
    // 添加B站推荐的其他参数
    args = append(args,
        "-pix_fmt", "yuv420p",     // 像素格式
        "-movflags", "+faststart",  // 优化流媒体播放
    )
    
    // 输出文件
    args = append(args, output)
    
    return args
}

// GetVideoInfo 获取视频信息
func (vp *VideoProcessor) GetVideoInfo(filename string) (*VideoInfo, error) {
    inputPath := filepath.Join(vp.InputDir, filename)
    
    // 使用ffprobe获取视频信息
    cmd := exec.Command("ffprobe",
        "-v", "quiet",
        "-print_format", "json",
        "-show_format",
        "-show_streams",
        inputPath,
    )
    
    _, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("获取视频信息失败: %v", err)
    }
    
    // 这里应该解析JSON输出
    // 为了简化，我们返回基本信息
    info := &VideoInfo{
        Filename: filename,
        Duration: "未知",
        Resolution: "未知",
        Bitrate: "未知",
        Size: getFileSize(inputPath),
    }
    
    return info, nil
}

// VideoInfo 视频信息
type VideoInfo struct {
    Filename   string
    Duration   string
    Resolution string
    Bitrate    string
    Size       int64
}

// getFileSize 获取文件大小
func getFileSize(filepath string) int64 {
    info, err := os.Stat(filepath)
    if err != nil {
        return 0
    }
    return info.Size()
}

// CheckFFmpeg 检查FFmpeg是否安装
func CheckFFmpeg() bool {
    cmd := exec.Command("ffmpeg", "-version")
    if err := cmd.Run(); err != nil {
        log.Println("⚠️ FFmpeg未安装，视频处理功能将不可用")
        log.Println("💡 请安装FFmpeg: brew install ffmpeg")
        return false
    }
    return true
}