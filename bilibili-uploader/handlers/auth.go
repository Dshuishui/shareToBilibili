// handlers/auth.go - OAuth认证处理
package handlers

import (
    "bilibili-uploader/config"
    "bilibili-uploader/services"
    // "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v4"
)

// AuthHandler OAuth认证处理器
type AuthHandler struct {
    oauthService *services.BilibiliOAuth
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
    return &AuthHandler{
        oauthService: services.NewBilibiliOAuth(
            config.GlobalConfig.BilibiliClientID,
            config.GlobalConfig.BilibiliClientSecret,
            config.GlobalConfig.BilibiliRedirectURI,
        ),
    }
}

// GetAuthURL 获取授权URL
func (h *AuthHandler) GetAuthURL(c *gin.Context) {
    // 生成state参数（防止CSRF攻击）
    state := generateState()
    
    // 保存state到session或缓存（这里简化处理，实际应该用Redis等）
    c.SetCookie("oauth_state", state, 3600, "/", "", false, true)
    
    // 构建授权URL
    authURL := fmt.Sprintf(
        "https://passport.bilibili.com/register/pc_oauth2.html?client_id=%s&response_type=code&redirect_uri=%s&scope=video-upload&state=%s",
        config.GlobalConfig.BilibiliClientID,
        config.GlobalConfig.BilibiliRedirectURI,
        state,
    )
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "auth_url": authURL,
    })
}

// HandleCallback OAuth回调处理
func (h *AuthHandler) HandleCallback(c *gin.Context) {
    // 获取授权码和state
    code := c.Query("code")
    state := c.Query("state")
    
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "授权码不能为空",
        })
        return
    }
    
    // 验证state（防止CSRF）
    savedState, err := c.Cookie("oauth_state")
    if err != nil || savedState != state {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "state验证失败",
        })
        return
    }
    
    // 用授权码换取access token
    tokenResp, err := h.oauthService.ExchangeCode(code)
    if err != nil {
        log.Printf("换取token失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "获取访问令牌失败",
        })
        return
    }
    
    // 获取用户信息
    userInfo, err := h.oauthService.GetUserInfo(tokenResp.AccessToken)
    if err != nil {
        log.Printf("获取用户信息失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "获取用户信息失败",
        })
        return
    }
    
    // 生成内部JWT token（包含B站access token）
    jwtToken, err := generateJWT(tokenResp.AccessToken, userInfo)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "生成令牌失败",
        })
        return
    }
    
    // 返回成功页面（带token）
    html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>授权成功</title>
        <style>
            body { 
                font-family: Arial; 
                display: flex; 
                justify-content: center; 
                align-items: center; 
                height: 100vh;
                background: linear-gradient(135deg, #00a1d6, #0081c6);
            }
            .container {
                background: white;
                padding: 40px;
                border-radius: 12px;
                text-align: center;
                box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            }
            h1 { color: #4caf50; }
            p { color: #666; margin: 20px 0; }
            .btn {
                display: inline-block;
                padding: 12px 24px;
                background: #00a1d6;
                color: white;
                text-decoration: none;
                border-radius: 6px;
                margin-top: 20px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>✅ 授权成功！</h1>
            <p>您已成功授权B站账号: <strong>%s</strong></p>
            <p>现在可以使用一键投稿功能了</p>
            <a href="/" class="btn">返回主页</a>
        </div>
        <script>
            // 保存token到localStorage
            localStorage.setItem('bilibili_token', '%s');
            localStorage.setItem('bilibili_user', '%s');
            
            // 3秒后自动跳转
            setTimeout(() => {
                window.location.href = '/';
            }, 3000);
        </script>
    </body>
    </html>
    `, userInfo.Username, jwtToken, userInfo.Username)
    
    c.Header("Content-Type", "text/html; charset=utf-8")
    c.String(http.StatusOK, html)
}

// ExchangeToken 交换令牌（前端AJAX调用）
func (h *AuthHandler) ExchangeToken(c *gin.Context) {
    var req struct {
        Code        string `json:"code"`
        RedirectURI string `json:"redirect_uri"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "参数错误",
        })
        return
    }
    
    // 交换token
    tokenResp, err := h.oauthService.ExchangeCode(req.Code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "获取访问令牌失败",
        })
        return
    }
    
    // 获取用户信息
    userInfo, err := h.oauthService.GetUserInfo(tokenResp.AccessToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "获取用户信息失败",
        })
        return
    }
    
    // 生成JWT
    jwtToken, err := generateJWT(tokenResp.AccessToken, userInfo)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "生成令牌失败",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success":      true,
        "access_token": jwtToken,
        "username":     userInfo.Username,
        "uid":          userInfo.UID,
    })
}

// VerifyToken 验证令牌
func (h *AuthHandler) VerifyToken(c *gin.Context) {
    // 从Authorization header获取token
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "未提供认证令牌",
        })
        return
    }
    
    // 解析Bearer token
    tokenString := ""
    if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
        tokenString = authHeader[7:]
    }
    
    // 验证JWT
    claims, err := validateJWT(tokenString)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "令牌无效或已过期",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success":  true,
        "username": claims.Username,
        "uid":      claims.UID,
    })
}

// JWTClaims JWT声明
type JWTClaims struct {
    BilibiliToken string `json:"bilibili_token"`
    Username      string `json:"username"`
    UID           int64  `json:"uid"`
    jwt.RegisteredClaims
}

// generateJWT 生成JWT令牌
func generateJWT(bilibiliToken string, userInfo *services.UserInfo) (string, error) {
    claims := JWTClaims{
        BilibiliToken: bilibiliToken,
        Username:      userInfo.Username,
        UID:           userInfo.UID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7天有效期
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.GlobalConfig.JWTSecret))
}

// validateJWT 验证JWT令牌
func validateJWT(tokenString string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.GlobalConfig.JWTSecret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid token")
}

// GetBilibiliToken 从JWT中提取B站token
func GetBilibiliToken(c *gin.Context) (string, error) {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        return "", fmt.Errorf("no authorization header")
    }
    
    tokenString := ""
    if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
        tokenString = authHeader[7:]
    }
    
    claims, err := validateJWT(tokenString)
    if err != nil {
        return "", err
    }
    
    return claims.BilibiliToken, nil
}

// generateState 生成随机state
func generateState() string {
    b := make([]byte, 16)
    for i := range b {
        b[i] = byte(65 + (time.Now().UnixNano() % 26))
    }
    return string(b)
}