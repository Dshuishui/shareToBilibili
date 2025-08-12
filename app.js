const express = require('express');
const cors = require('cors'); // 添加这行
const multer = require('multer');
const path = require('path');
const axios = require('axios');
const puppeteer = require('puppeteer');
const fs = require('fs');

const app = express();
const port = 3000;

// Python后端地址
const PYTHON_BACKEND = 'http://localhost:5001';

// 添加CORS配置 - 在其他中间件之前
app.use(cors({
    origin: [
      'http://localhost:5173', // XBuilder开发服务器
      'http://localhost:3000', // 允许同源请求
      // 如果有其他需要的域名可以继续添加
    ],
    credentials: true, // 允许携带cookies
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization']
  }));

// 分区ID映射 - B站真实分区ID
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
        const uploadDir = 'uploads/';
        if (!fs.existsSync(uploadDir)) {
            fs.mkdirSync(uploadDir, { recursive: true });
        }
        cb(null, uploadDir);
    },
    filename: function (req, file, cb) {
        cb(null, Date.now() + '-' + file.originalname);
    }
});
const upload = multer({
    storage: storage,
    limits: {
        fileSize: 8 * 1024 * 1024 * 1024 // 8GB 限制
    }
});

// 用于存储浏览器实例
let browser = null;
let currentPage = null;

// 静态文件服务
app.use(express.static('public'));
app.use(express.json());

// 健康检查
app.get('/health', (req, res) => {
    res.json({
        status: 'ok',
        message: 'Node.js前端服务正常',
        pythonBackend: PYTHON_BACKEND,
        browserActive: !!browser
    });
});

// 登录接口 - 用Puppeteer获取Cookie
app.get('/login', async (req, res) => {
    try {
        console.log('🔑 开始启动浏览器进行B站登录...');

        // 启动浏览器（显示界面让用户登录）
        browser = await puppeteer.launch({
            headless: false,
            defaultViewport: null,
            args: [
                '--start-maximized',
                '--no-sandbox',
                '--disable-setuid-sandbox'
            ],
            executablePath: process.platform === 'darwin'
                ? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
                : undefined
        });

        const page = await browser.newPage();
        currentPage = page;

        // 跳转到B站登录页面
        await page.goto('https://passport.bilibili.com/login', {
            waitUntil: 'networkidle2'
        });

        console.log('🌐 B站登录页面已打开，等待用户完成登录...');

        // 等待用户完成登录（检测URL变化或特定元素）
        await page.waitForFunction(() => {
            return window.location.href.includes('bilibili.com') &&
                !window.location.href.includes('passport.bilibili.com');
        }, { timeout: 300000 }); // 5分钟超时

        console.log('✅ 检测到用户已登录，正在获取Cookie...');

        // 获取所有cookies
        const cookies = await page.cookies();

        console.log(`🍪 Cookie获取成功，数量: ${cookies.length}`);

        // 保持浏览器打开，供后续自动化投稿使用
        // 注意：不关闭浏览器

        res.json({
            success: true,
            message: '登录成功！浏览器保持打开状态，现在可以使用自动化投稿功能。',
            cookieCount: cookies.length,
            browserReady: true
        });

    } catch (error) {
        console.error('❌ 登录过程出错:', error);

        if (browser) {
            await browser.close();
            browser = null;
            currentPage = null;
        }

        res.json({
            success: false,
            message: '登录失败: ' + error.message
        });
    }
});

// 🚀 新增：自动化投稿接口
// 修改路由定义，支持video和cover两个字段
app.post('/auto-upload', upload.fields([
    { name: 'video', maxCount: 1 },
    { name: 'cover', maxCount: 1 }  // 添加封面字段支持
]), async (req, res) => {
    try {
        const { title, description, tags, category } = req.body;
        const videoFile = req.files['video'] ? req.files['video'][0] : null;
        const coverFile = req.files['cover'] ? req.files['cover'][0] : null;

        console.log('🎬 收到自动化投稿请求:', {
            title,
            description: description?.substring(0, 50) + '...',
            tags,
            category,
            hasFile: !!videoFile,
            fileSize: videoFile ? `${(videoFile.size / 1024 / 1024).toFixed(2)} MB` : 'N/A',
            hasCover: !!coverFile, // 添加封面信息
            coverSize: coverFile ? `${(coverFile.size / 1024).toFixed(2)} KB` : 'N/A'
        });

        // 验证必要参数
        if (!videoFile) {
            return res.json({
                success: false,
                message: '未收到视频文件'
            });
        }

        if (!title?.trim()) {
            return res.json({
                success: false,
                message: '视频标题不能为空'
            });
        }

        if (!browser || !currentPage) {
            return res.json({
                success: false,
                message: '浏览器未准备就绪，请先完成B站登录'
            });
        }

        // 执行自动化投稿流程，传递封面文件
        const result = await performAutomatedUpload(videoFile, {
            title: title.trim(),
            description: description?.trim() || '',
            tags: tags?.trim() || '',
            category,
            coverFile: coverFile // 传递封面文件对象
        });

        res.json(result);

    } catch (error) {
        console.error('❌ 自动化投稿失败:', error);
        res.json({
            success: false,
            message: '自动化投稿失败: ' + error.message
        });
    }
});

// 在文件顶部添加辅助函数
const { setTimeout } = require('timers/promises');


// 上传封面文件 - 新增函数
async function uploadCoverFile(coverFile) {
    try {
        console.log('🔍 查找封面上传元素...');

        // B站投稿页面的封面上传选择器（需要根据实际页面调整）
        const coverUploadSelectors = [
            'input[accept*="image"]',              // 通用图片上传
            '.cover-upload input[type="file"]',    // 封面上传区域
            '.upload-cover input[type="file"]',    // 封面上传
            '[class*="cover"] input[type="file"]', // 包含cover的类名
            '.bcc-upload-cover input[type="file"]', // B站封面上传组件
        ];

        let coverInput = null;
        
        // 首先尝试直接查找封面上传input
        for (const selector of coverUploadSelectors) {
            try {
                await currentPage.waitForSelector(selector, { timeout: 3000 });
                coverInput = await currentPage.$(selector);
                if (coverInput) {
                    console.log(`✅ 找到封面上传元素: ${selector}`);
                    break;
                }
            } catch (e) {
                console.log(`⚠️ 封面上传选择器 ${selector} 未找到`);
                continue;
            }
        }

        // 如果找不到直接的input，尝试点击封面上传按钮来激活
        if (!coverInput) {
            console.log('🖱️ 尝试点击封面上传按钮...');
            const coverButtonSelectors = [
                'button:contains("上传封面")',
                '.cover-upload-btn',
                '.upload-cover-btn',
                '[class*="cover"][class*="upload"]',
                '.cover-area',
                '.cover-container'
            ];

            for (const btnSelector of coverButtonSelectors) {
                try {
                    if (btnSelector.includes(':contains')) {
                        const text = btnSelector.match(/contains\("([^"]+)"/)[1];
                        const coverBtn = await currentPage.evaluateHandle((text) => {
                            const buttons = Array.from(document.querySelectorAll('*'));
                            return buttons.find(btn => btn.textContent && btn.textContent.includes(text));
                        }, text);

                        if (coverBtn && await coverBtn.asElement()) {
                            await coverBtn.click();
                            console.log(`🖱️ 点击封面按钮: ${text}`);
                            await setTimeout(2000);
                            break;
                        }
                    } else {
                        const coverBtn = await currentPage.$(btnSelector);
                        if (coverBtn) {
                            await coverBtn.click();
                            console.log(`🖱️ 点击封面按钮: ${btnSelector}`);
                            await setTimeout(2000);
                            break;
                        }
                    }
                } catch (e) {
                    continue;
                }
            }

            // 重新查找input
            coverInput = await currentPage.$('input[type="file"][accept*="image"]');
        }

        if (!coverInput) {
            console.log('⚠️ 无法找到封面上传元素，跳过封面上传');
            return;
        }

        // 上传封面文件
        console.log('📤 开始上传封面文件...');
        await coverInput.uploadFile(coverFile.path);
        
        console.log('⏳ 等待封面上传处理...');
        await setTimeout(3000);

        // 等待封面上传完成
        try {
            await currentPage.waitForFunction(() => {
                // 检查封面上传完成的标识
                const indicators = [
                    document.querySelector('.cover-success'),
                    document.querySelector('[class*="cover"][class*="success"]'),
                    document.querySelector('img[src*="cover"]'), // 封面预览图
                ];
                return indicators.some(el => el !== null);
            }, { timeout: 30000 });

            console.log('✅ 封面上传完成');
        } catch (e) {
            console.log('⚠️ 未检测到明确的封面上传完成标识，继续后续流程...');
        }

    } catch (error) {
        console.error('❌ 封面文件上传失败:', error);
        console.log('⚠️ 封面上传失败，继续视频投稿流程...');
        // 不抛出错误，让投稿流程继续
    }
}

// 替换原来的 performAutomatedUpload 函数
async function performAutomatedUpload(videoFile, metadata) {
    try {
        console.log('🚀 开始自动化投稿流程...');

        // 第1步：导航到B站投稿页面
        console.log('📍 Step 1: 导航到投稿页面');
        await currentPage.goto('https://member.bilibili.com/platform/upload/video/frame', {
            waitUntil: 'networkidle2',
            timeout: 30000
        });

        // 替换 waitForTimeout 为 setTimeout
        console.log('⏳ 等待页面完全加载...');
        // await setTimeout(1000); // 等待1秒

        // 第2步：上传视频文件
        console.log('📁 Step 2: 上传视频文件');
        await uploadVideoFile(videoFile);

        // 第3步：上传封面文件（新增）
        if (metadata.coverFile) {
            console.log('🖼️ Step 3: 上传封面文件');
            await uploadCoverFile(metadata.coverFile);
        }

        // 第4步：填写视频信息
        console.log('✏️ Step 3: 填写视频信息');
        await fillVideoInformation(metadata);

        // 第5步：等待用户预览和确认
        console.log('👀 Step 4: 等待用户预览确认');
        const confirmed = await waitForUserConfirmation();

        if (!confirmed) {
            return {
                success: false,
                message: '用户取消了投稿'
            };
        }

        // 第6步：自动提交
        console.log('🎯 Step 5: 自动提交投稿');
        const submitResult = await submitVideo();

        return submitResult;

    } catch (error) {
        console.error('❌ 自动化投稿过程出错:', error);
        throw error;
    }
}

// 上传视频文件
async function uploadVideoFile(videoFile) {
    try {
        // 等待文件上传区域出现
        await currentPage.waitForSelector('input[type="file"]', { timeout: 10000 });

        // 选择文件
        const fileInput = await currentPage.$('input[type="file"]');
        await fileInput.uploadFile(videoFile.path);

        console.log('📤 视频文件已选择，等待上传完成...');

        // 等待上传完成 - 监听上传进度或完成标识
        // 这里需要根据B站实际页面元素调整
        await currentPage.waitForFunction(() => {
            // 检查是否有上传完成的标识
            const uploadStatus = document.querySelector('.upload-status, .upload-success, .video-info');
            return uploadStatus && uploadStatus.textContent.includes('上传完成') ||
                document.querySelector('.video-title-input'); // 或者等待标题输入框出现
        }, { timeout: 300000 }); // 5分钟超时

        console.log('✅ 视频上传完成');

    } catch (error) {
        console.error('❌ 视频文件上传失败:', error);
        throw new Error('视频文件上传失败: ' + error.message);
    }
}

// 修复 fillVideoInformation 函数
async function fillVideoInformation(metadata) {
    try {
        console.log('📝 开始填写视频信息...');

        // 等待页面稳定
        await setTimeout(1000);

        // 填写标题 - 使用更多选择器
        const titleSelectors = [
            'input[placeholder*="标题"]',
            'input[placeholder*="title"]',
            '.title-input input',
            '.video-title input',
            'input.input[maxlength="80"]',
            '.form-item input[type="text"]'
        ];

        let titleFilled = false;
        for (const selector of titleSelectors) {
            try {
                await currentPage.waitForSelector(selector, { timeout: 3000 });

                // 清空并填写标题
                await currentPage.click(selector, { clickCount: 3 }); // 三击选中全部
                await currentPage.keyboard.press('Backspace');
                await currentPage.type(selector, metadata.title, { delay: 100 });

                console.log('✅ 标题已填写');
                titleFilled = true;
                break;
            } catch (e) {
                console.log(`⚠️ 标题选择器 ${selector} 失败，尝试下一个...`);
                continue;
            }
        }

        if (!titleFilled) {
            console.log('⚠️ 标题填写失败，用户需要手动填写');
        }

        // 填写简介
        if (metadata.description) {
            const descSelectors = [
                'textarea[placeholder*="简介"]',
                'textarea[placeholder*="描述"]',
                '.desc-input textarea',
                '.video-desc textarea',
                'textarea[maxlength="2000"]'
            ];

            for (const selector of descSelectors) {
                try {
                    await currentPage.waitForSelector(selector, { timeout: 3000 });
                    await currentPage.click(selector, { clickCount: 3 });
                    await currentPage.keyboard.press('Backspace');
                    await currentPage.type(selector, metadata.description, { delay: 50 });
                    console.log('✅ 简介已填写');
                    break;
                } catch (e) {
                    continue;
                }
            }
        }

        // 填写标签
        if (metadata.tags) {
            const tagSelectors = [
                'input[placeholder*="标签"]',
                '.tag-input input',
                '.tags-input input'
            ];

            const tagList = metadata.tags.split(',').map(tag => tag.trim()).filter(tag => tag);

            for (const selector of tagSelectors) {
                try {
                    await currentPage.waitForSelector(selector, { timeout: 3000 });

                    for (const tag of tagList) {
                        await currentPage.click(selector);
                        await currentPage.type(selector, tag, { delay: 100 });
                        await currentPage.keyboard.press('Enter');
                        await setTimeout(500);
                    }
                    console.log('✅ 标签已填写');
                    break;
                } catch (e) {
                    continue;
                }
            }
        }

        console.log('✅ 视频信息填写完成');

    } catch (error) {
        console.error('❌ 填写视频信息失败:', error);
        // 不抛出错误，让用户手动填写
        console.log('⚠️ 自动填写失败，用户可以手动调整信息');
    }
}

// 修复 waitForUserConfirmation 函数
async function waitForUserConfirmation() {
    try {
        console.log('⏳ 等待用户预览和确认...');

        // 在页面上注入确认对话框
        await currentPage.evaluate(() => {
            // 移除可能存在的旧对话框
            const oldModal = document.getElementById('auto-upload-confirm');
            if (oldModal) oldModal.remove();

            // 创建确认对话框
            const modal = document.createElement('div');
            modal.id = 'auto-upload-confirm';
            modal.style.cssText = `
                position: fixed;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);
                background: white;
                padding: 30px;
                border-radius: 12px;
                box-shadow: 0 8px 32px rgba(0,0,0,0.3);
                z-index: 999999;
                text-align: center;
                border: 3px solid #00a1d6;
                max-width: 400px;
                font-family: Arial, sans-serif;
            `;

            modal.innerHTML = `
                <h3 style="color: #00a1d6; margin: 0 0 15px 0; font-size: 18px;">🤖 自动化投稿确认</h3>
                <p style="margin: 10px 0; color: #333; line-height: 1.5;">
                    请检查视频信息是否正确：<br>
                    • 标题是否准确<br>
                    • 简介和标签是否合适<br>
                    • 分区选择是否正确<br>
                    • 封面是否满意
                </p>
                <div style="margin: 25px 0;">
                    <button id="confirm-upload" style="background: #52c41a; color: white; border: none; padding: 12px 24px; border-radius: 6px; margin-right: 15px; cursor: pointer; font-size: 16px; font-weight: bold;">✅ 确认投稿</button>
                    <button id="cancel-upload" style="background: #ff4d4f; color: white; border: none; padding: 12px 24px; border-radius: 6px; cursor: pointer; font-size: 16px; font-weight: bold;">❌ 取消</button>
                </div>
                <p style="font-size: 12px; color: #666; margin: 0;">
                    确认后将自动点击"立即投稿"按钮
                </p>
            `;

            document.body.appendChild(modal);

            // 绑定事件
            document.getElementById('confirm-upload').onclick = () => {
                window.autoUploadConfirmed = true;
                document.body.removeChild(modal);
            };

            document.getElementById('cancel-upload').onclick = () => {
                window.autoUploadConfirmed = false;
                document.body.removeChild(modal);
            };

            // 重置确认状态
            window.autoUploadConfirmed = undefined;
        });

        console.log('💡 确认对话框已显示，等待用户选择...');

        // 等待用户做出选择
        await currentPage.waitForFunction(() => {
            return window.autoUploadConfirmed !== undefined;
        }, { timeout: 300000 }); // 5分钟超时

        const confirmed = await currentPage.evaluate(() => window.autoUploadConfirmed);

        console.log(confirmed ? '✅ 用户确认投稿' : '❌ 用户取消投稿');
        return confirmed;

    } catch (error) {
        console.error('❌ 等待用户确认超时或出错:', error);
        return false;
    }
}
// 修复 submitVideo 函数
async function submitVideo() {
    try {
        console.log('🎯 开始提交投稿...');

        // 更新的投稿按钮选择器 - 按优先级排序
        const submitSelectors = [
            'span.submit-add',                           // 最精确的选择器
            'span[data-reporter-id="29"]',               // 基于data属性
            '.submit-add',                               // 基于class名
            'span:contains("立即投稿")',                  // 基于文本内容（需要特殊处理）
            'span:contains("投稿")',                     // 更宽泛的文本匹配
            '.submit-btn',                               // 备用选择器
            '.publish-btn',                              // 备用选择器
            '[class*="submit"]'                          // 包含submit的class
        ];

        let submitSuccess = false;

        for (const selector of submitSelectors) {
            try {
                // 特殊处理文本内容选择器
                if (selector.includes(':contains')) {
                    const text = selector.match(/contains\("([^"]+)"/)[1];
                    console.log(`🔍 尝试按文本查找按钮: "${text}"`);
                    
                    const submitButton = await currentPage.evaluateHandle((text) => {
                        const elements = Array.from(document.querySelectorAll('span, button'));
                        return elements.find(el => el.textContent?.includes(text));
                    }, text);

                    if (submitButton && await submitButton.asElement()) {
                        console.log(`🚀 找到投稿按钮 (文本匹配): ${text}`);
                        await submitButton.click();
                        submitSuccess = true;
                        break;
                    }
                } else {
                    // 常规CSS选择器
                    console.log(`🔍 尝试选择器: ${selector}`);
                    
                    // 等待元素出现，但不要等太久
                    await currentPage.waitForSelector(selector, { timeout: 3000 });
                    
                    // 检查元素是否可见和可点击
                    const isVisible = await currentPage.evaluate((sel) => {
                        const el = document.querySelector(sel);
                        if (!el) return false;
                        
                        const style = window.getComputedStyle(el);
                        return style.display !== 'none' && 
                               style.visibility !== 'hidden' && 
                               style.opacity !== '0';
                    }, selector);
                    
                    if (isVisible) {
                        await currentPage.click(selector);
                        submitSuccess = true;
                        console.log(`🚀 投稿按钮已点击: ${selector}`);
                        break;
                    } else {
                        console.log(`⚠️ 元素存在但不可见: ${selector}`);
                    }
                }
            } catch (e) {
                console.log(`⚠️ 选择器失败 ${selector}: ${e.message}`);
                continue;
            }
        }

        if (!submitSuccess) {
            console.log('⚠️ 所有选择器都失败，尝试最后的备用方案...');
            
            // 最后的备用方案：查找所有可能的提交元素
            try {
                const found = await currentPage.evaluate(() => {
                    const keywords = ['立即投稿', '投稿', '提交', '发布'];
                    const selectors = ['span', 'button', 'div[role="button"]', '[class*="submit"]', '[class*="publish"]'];
                    
                    for (const sel of selectors) {
                        const elements = document.querySelectorAll(sel);
                        for (const el of elements) {
                            const text = el.textContent?.trim();
                            if (keywords.some(keyword => text?.includes(keyword))) {
                                el.click();
                                return { success: true, text, selector: sel };
                            }
                        }
                    }
                    return { success: false };
                });
                
                if (found.success) {
                    submitSuccess = true;
                    console.log(`🚀 备用方案成功: 点击了包含"${found.text}"的${found.selector}元素`);
                }
            } catch (e) {
                console.log('❌ 备用方案也失败了:', e.message);
            }
        }

        if (!submitSuccess) {
            console.log('⚠️ 未找到投稿按钮，请用户手动点击投稿');
            
            // 在页面上显示提示
            await currentPage.evaluate(() => {
                const tip = document.createElement('div');
                tip.style.cssText = `
                    position: fixed;
                    top: 50%;
                    left: 50%;
                    transform: translate(-50%, -50%);
                    background: #ff9800;
                    color: white;
                    padding: 20px;
                    border-radius: 8px;
                    z-index: 999999;
                    font-size: 16px;
                    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
                `;
                tip.textContent = '⚠️ 请手动点击"立即投稿"按钮完成投稿';
                document.body.appendChild(tip);
                
                setTimeout(() => tip.remove(), 10000); // 10秒后自动消失
            });
            
            return {
                success: true,
                message: '视频信息已填写完成，请手动点击"立即投稿"按钮完成投稿',
                url: currentPage.url()
            };
        }

        // 等待投稿完成或跳转
        console.log('⏳ 等待投稿处理完成...');
        try {
            await currentPage.waitForFunction(() => {
                return document.querySelector('.success, .complete, [class*="success"]') ||
                    window.location.href.includes('/video/') ||
                    document.querySelector('[class*="result"]') ||
                    document.querySelector('.upload-result') ||
                    document.querySelector('[class*="complete"]');
            }, { timeout: 60000 });

            console.log('🎉 投稿提交完成！');
        } catch (e) {
            console.log('⚠️ 投稿状态检测超时，但提交操作已完成');
        }

        return {
            success: true,
            message: '视频投稿已提交！请在B站查看投稿状态。',
            url: currentPage.url()
        };

    } catch (error) {
        console.error('❌ 提交投稿失败:', error);
        return {
            success: false,
            message: '提交投稿失败: ' + error.message
        };
    }
}



// 检查登录状态
app.get('/check-login', async (req, res) => {
    try {
        const response = await axios.get(`${PYTHON_BACKEND}/check-login`);
        res.json({
            ...response.data,
            browserReady: !!browser && !!currentPage
        });
    } catch (error) {
        res.json({
            isLoggedIn: false,
            browserReady: !!browser && !!currentPage,
            message: '无法连接到Python后端'
        });
    }
});

// 修复 uploadVideoFile 函数
async function uploadVideoFile(videoFile) {
    try {
        console.log('🔍 查找文件上传元素...');

        // 等待文件上传区域出现 - 使用更精确的选择器
        const uploadSelectors = [
            'input[type="file"]',
            'input[accept*="video"]',
            '.upload-wrapper input[type="file"]',
            '.bcc-upload input[type="file"]'
        ];

        let fileInput = null;
        for (const selector of uploadSelectors) {
            try {
                await currentPage.waitForSelector(selector, { timeout: 5000 });
                fileInput = await currentPage.$(selector);
                if (fileInput) {
                    console.log(`✅ 找到文件输入元素: ${selector}`);
                    break;
                }
            } catch (e) {
                console.log(`⚠️ 选择器 ${selector} 未找到，尝试下一个...`);
                continue;
            }
        }

        if (!fileInput) {
            // 尝试点击上传按钮来激活文件选择
            const uploadBtnSelectors = [
                'button:contains("上传视频")',
                '.upload-btn',
                '.bcc-upload-dragger',
                '[class*="upload"]'
            ];

            for (const btnSelector of uploadBtnSelectors) {
                try {
                    const uploadBtn = await currentPage.$(btnSelector);
                    if (uploadBtn) {
                        console.log(`🖱️ 点击上传按钮: ${btnSelector}`);
                        await uploadBtn.click();
                        await setTimeout(2000);

                        // 再次查找文件输入
                        fileInput = await currentPage.$('input[type="file"]');
                        if (fileInput) break;
                    }
                } catch (e) {
                    continue;
                }
            }
        }

        if (!fileInput) {
            throw new Error('无法找到文件上传元素，请检查页面是否正确加载');
        }

        // 上传文件
        console.log('📤 开始上传视频文件...');
        await fileInput.uploadFile(videoFile.path);

        console.log('⏳ 等待文件上传处理...');
        await setTimeout(5000);

        // 等待上传完成的标识
        console.log('🔍 等待上传完成标识...');
        try {
            await currentPage.waitForFunction(() => {
                // 检查多种可能的上传完成标识
                const indicators = [
                    document.querySelector('.upload-success'),
                    document.querySelector('.video-info'),
                    document.querySelector('[class*="success"]'),
                    document.querySelector('input[placeholder*="标题"]'),
                    document.querySelector('.title-input'),
                    document.querySelector('.video-title')
                ];
                return indicators.some(el => el !== null);
            }, { timeout: 300000 }); // 5分钟超时

            console.log('✅ 视频上传完成');
        } catch (e) {
            console.log('⚠️ 未检测到明确的上传完成标识，继续后续流程...');
        }

    } catch (error) {
        console.error('❌ 视频文件上传失败:', error);
        throw new Error('视频文件上传失败: ' + error.message);
    }
}

// 关闭浏览器接口
app.post('/close-browser', async (req, res) => {
    try {
        if (browser) {
            await browser.close();
            browser = null;
            currentPage = null;
            console.log('🔒 浏览器已关闭');
        }
        res.json({
            success: true,
            message: '浏览器已关闭'
        });
    } catch (error) {
        res.json({
            success: false,
            message: '关闭浏览器失败: ' + error.message
        });
    }
});

// 启动服务器
app.listen(port, () => {
    console.log(`🚀 Node.js服务运行在 http://localhost:${port}`);
    console.log(`🔗 Python后端地址: ${PYTHON_BACKEND}`);
    console.log('=' * 50);
    console.log('🆕 新功能: 浏览器自动化投稿');
    console.log('📋 使用步骤:');
    console.log('1. 点击"登录B站账号" - 浏览器将保持打开');
    console.log('2. 选择视频文件并填写信息');
    console.log('3. 选择投稿方式:');
    console.log('   - 自动化投稿: 浏览器自动填写表单，暂停预览，确认后提交');
    console.log('   - API投稿: 直接通过Python后端API上传');
    console.log('=' * 50);
});

// 优雅关闭
process.on('SIGINT', async () => {
    console.log('\n🔄 正在关闭服务器...');
    if (browser) {
        await browser.close();
        console.log('🔒 浏览器已关闭');
    }
    process.exit(0);
});