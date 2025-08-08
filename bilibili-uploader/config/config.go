// config/config.go - 配置管理
package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

// Config 全局配置
type Config struct {
    // 服务器配置
    Port string
    
    // B站OAuth配置
    BilibiliClientID     string
    BilibiliClientSecret string
    BilibiliRedirectURI  string
    
    // 文件存储配置
    UploadDir    string
    ProcessedDir string
    TempDir      string
    
    // JWT密钥（用于生成自己的token）
    JWTSecret string
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Load 加载配置
func Load() {
    // 加载.env文件
    if err := godotenv.Load(); err != nil {
        log.Println("未找到.env文件，使用环境变量")
    }
    
    GlobalConfig = &Config{
        // 服务器配置
        Port: getEnv("PORT", "8080"),
        
        // B站OAuth配置
        BilibiliClientID:     getEnv("BILIBILI_CLIENT_ID", ""),
        BilibiliClientSecret: getEnv("BILIBILI_CLIENT_SECRET", ""),
        BilibiliRedirectURI:  getEnv("BILIBILI_REDIRECT_URI", "http://localhost:8080/api/auth/callback"),
        
        // 文件存储配置  
        UploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
        ProcessedDir: getEnv("PROCESSED_DIR", "./processed"),
        TempDir:      getEnv("TEMP_DIR", "./temp"),
        
        // JWT密钥
        JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-this"),
    }
    
    // 验证必要配置
    validateConfig()
}

// validateConfig 验证配置
func validateConfig() {
    if GlobalConfig.BilibiliClientID == "" {
        log.Println("⚠️  警告: BILIBILI_CLIENT_ID 未设置")
        log.Println("   请在.env文件中设置B站OAuth客户端ID")
        log.Println("   申请地址: https://open.bilibili.com")
    }
    
    if GlobalConfig.BilibiliClientSecret == "" {
        log.Println("⚠️  警告: BILIBILI_CLIENT_SECRET 未设置")
    }
    
    if GlobalConfig.JWTSecret == "your-secret-key-change-this" {
        log.Println("⚠️  警告: 使用默认JWT密钥，请修改JWT_SECRET")
    }
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// IsBilibiliConfigured 检查B站配置是否完整
func IsBilibiliConfigured() bool {
    return GlobalConfig.BilibiliClientID != "" && 
           GlobalConfig.BilibiliClientSecret != ""
}