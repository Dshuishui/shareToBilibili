const express = require('express');
const multer = require('multer');
const path = require('path');
const axios = require('axios');
const puppeteer = require('puppeteer');

const app = express();
const port = 3000;

// Python后端地址
const PYTHON_BACKEND = 'http://localhost:5001';

// 分区ID映射
const CATEGORY_MAP = {
    'douga': 1,      // 动画
    'game': 4,       // 游戏  
    'kichiku': 119,  // 鬼畜
    'music': 3,      // 音乐
    'dance': 129,    // 舞蹈
    'cinephile': 181, // 影视
    'ent': 5,        // 娱乐
    'knowledge': 36,  // 知识
    'tech': 188,     // 科技
    'information': 202, // 资讯
    'food': 76,      // 美食
    'life': 160,     // 生活
    'car': 223,      // 汽车
    'fashion': 155,  // 时尚
    'sports': 234,   // 运动
    'animal': 217    // 动物圈
};

// 配置文件上传
const storage = multer.diskStorage({
    destination: function (req, file, cb) {
        cb(null, 'uploads/')
    },
    filename: function (req, file, cb) {
        cb(null, Date.now() + '-' + file.originalname)
    }
});
const upload = multer({ storage: storage });

// 用于存储浏览器实例
let browser = null;

// 静态文件服务
app.use(express.static('public'));
app.use(express.json());

// 健康检查
app.get('/health', (req, res) => {
    res.json({ 
        status: 'ok', 
        message: 'Node.js前端服务正常',
        pythonBackend: PYTHON_BACKEND
    });
});

// 登录接口 - 用Puppeteer获取Cookie并发送给Python后端
app.get('/login', async (req, res) => {
    try {
        console.log('开始启动浏览器进行B站登录...');
        
        // 启动浏览器（显示界面让用户登录）
        browser = await puppeteer.launch({
            headless: false,
            defaultViewport: null,
            args: ['--start-maximized'],
            executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
        });

        const page = await browser.newPage();
        
        // 跳转到B站登录页面
        await page.goto('https://passport.bilibili.com/login', {
            waitUntil: 'networkidle2'
        });

        console.log('B站登录页面已打开，等待用户完成登录...');

        // 等待用户完成登录（检测URL变化）
        await page.waitForFunction(() => {
            return window.location.href.includes('bilibili.com') && 
                   !window.location.href.includes('passport.bilibili.com');
        }, { timeout: 300000 }); // 5分钟超时

        console.log('检测到用户已登录，正在获取Cookie...');

        // 获取所有cookies
        const cookies = await page.cookies();
        
        console.log(`Cookie获取成功，数量: ${cookies.length}`);

        // 关闭浏览器
        await browser.close();
        browser = null;

        // 转换Cookie格式并发送给Python后端
        const cookieData = {
            cookies: cookies.map(cookie => ({
                name: cookie.name,
                value: cookie.value,
                domain: cookie.domain
            }))
        };

        console.log('正在将Cookie发送给Python后端...');

        // 发送Cookie给Python后端
        const response = await axios.post(`${PYTHON_BACKEND}/login`, cookieData);
        
        console.log('Python后端响应:', response.data);

        res.json({
            success: true,
            message: '登录成功！Cookie已保存到后端，可以开始上传视频。',
            cookieCount: cookies.length,
            backendResponse: response.data
        });

    } catch (error) {
        console.error('登录过程出错:', error);
        
        if (browser) {
            await browser.close();
            browser = null;
        }

        res.json({
            success: false,
            message: '登录失败: ' + error.message
        });
    }
});

// 检查登录状态
app.get('/check-login', async (req, res) => {
    try {
        const response = await axios.get(`${PYTHON_BACKEND}/check-login`);
        res.json(response.data);
    } catch (error) {
        res.json({ 
            isLoggedIn: false, 
            message: '无法连接到Python后端' 
        });
    }
});

// 视频上传接口
app.post('/upload-video', upload.single('video'), async (req, res) => {
    try {
        const { title, description, tags, category } = req.body;
        const videoFile = req.file;
        
        if (!videoFile) {
            return res.json({
                success: false,
                message: '未收到视频文件'
            });
        }
        
        // 转换分区ID
        const categoryId = CATEGORY_MAP[category] || 160; // 默认生活分区
        
        console.log('收到视频上传请求:', {
            title,
            description,
            tags,
            category: `${category} (${categoryId})`,
            filename: videoFile.filename
        });
        
        // 创建FormData转发给Python后端
        const FormData = require('form-data');
        const fs = require('fs');
        
        const formData = new FormData();
        formData.append('video', fs.createReadStream(videoFile.path), videoFile.originalname);
        formData.append('title', title);
        formData.append('description', description || '');
        formData.append('tags', tags || '');
        formData.append('category', categoryId.toString()); // 发送数字ID
        
        const response = await axios.post(`${PYTHON_BACKEND}/upload-video`, formData, {
            headers: {
                ...formData.getHeaders(),
            },
            timeout: 300000 // 5分钟超时
        });
        
        // 删除临时文件
        fs.unlinkSync(videoFile.path);
        
        res.json(response.data);
        
    } catch (error) {
        console.error('上传视频失败:', error.message);
        res.json({
            success: false,
            message: '上传失败: ' + error.message
        });
    }
});

// 旧的投稿页面接口（保留兼容性）
app.get('/open-upload-page', (req, res) => {
    res.json({
        success: true,
        message: '新版本将直接上传到B站，无需手动操作！'
    });
});

// 启动服务器
app.listen(port, () => {
    console.log(`Node.js前端服务运行在 http://localhost:${port}`);
    console.log(`Python后端地址: ${PYTHON_BACKEND}`);
    console.log('请确保Python后端服务也在运行！');
});

// 优雅关闭
process.on('SIGINT', async () => {
    console.log('\n正在关闭服务器...');
    if (browser) {
        await browser.close();
    }
    process.exit(0);
});