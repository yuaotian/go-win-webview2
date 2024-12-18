# 🚀 go-win-webview2

<p align="center">
  <img src="assets/logo.svg" alt="WebView2 Logo" width="200" height="200">
</p>

<p align="center">
  <a href="https://github.com/yuaotian/go-win-webview2/stargazers"><img src="https://img.shields.io/github/stars/yuaotian/go-win-webview2" alt="Stars"></a>
  <a href="https://github.com/yuaotian/go-win-webview2/network/members"><img src="https://img.shields.io/github/forks/yuaotian/go-win-webview2" alt="Forks"></a>
  <a href="https://github.com/yuaotian/go-win-webview2/issues"><img src="https://img.shields.io/github/issues/yuaotian/go-win-webview2" alt="Issues"></a>
  <a href="https://github.com/yuaotian/go-win-webview2/blob/master/LICENSE"><img src="https://img.shields.io/github/license/yuaotian/go-win-webview2" alt="License"></a>
  <img src="https://img.shields.io/badge/platform-windows-blue" alt="Platform">
  <img src="https://img.shields.io/badge/Go-%3E%3D%201.16-blue" alt="Go Version">
</p>

<p align="center">
  <a href="#快速开始">快速开始</a> •
  <a href="#特性">特性</a> •
  <a href="#安装">安装</a> •
  <a href="#使用示例">使用示例</a> •
  <a href="#文档">文档</a>
</p>

> 🌟 基于Microsoft Edge WebView2的Go语言界面开发包,提供简单易用的API接口。本项目基于[webview/webview](https://github.com/webview/webview) | [jchv/go-webview2](https://github.com/jchv/go-webview2)改进,专注于Windows平台的WebView2功能增强。

## ✨ 特性

- 🎯 完全兼容webview/webview的API
- 💪 专注于Windows平台WebView2的增强功能
- 🔌 简单易用的Go语言接口
- 🛡️ 稳定可靠的性能表现
- 🎨 丰富的界面定制选项
- 🔒 内置安全机制
- 🚀 快速的启动速度
- 📦 零依赖部署

## 🎯 主要功能

### 窗口控制
- 🎨 丰富的窗口操作
  - 无边框窗口
  - 窗口大小调整
  - 全屏切换
  - 窗口置顶
  - 透明度控制
  - 窗口最大化/最小化/还原
  - 窗口居中
  - 自定义图标
  - 窗口样式定制

### 浏览器功能
- 🌐 完整的Web功能
  - HTML/CSS/JavaScript支持
  - 双向通信机制
  - Cookie管理
  - 缓存控制
  - 页面导航(前进/后退/刷新)
  - 开发者工具
  - 打印功能(直接打印/PDF导出)

### 事件监听
- 📡 丰富的事件回调
  - 页面加载状态
  - URL变化
  - 标题变化
  - 全屏状态变化

### 扩展功能
- ⚡ WebSocket支持
  - 内置WebSocket服务器
  - ��向实时通信
  - 消息处理回调
- 🔌 JavaScript Hook机制
  - 前置/后置处理钩子
  - 优先级控制
  - 灵活的脚本注入

### 热键支持
- ⌨️ 全局热键系统
  - 支持组合键
  - 字符串格式配置
  - 动态注册/注销

## 📦 安装

### 前置要求

- Windows 10+ 操作系统
- Go 1.16+
- Microsoft Edge WebView2 Runtime

> 💡 Windows 10+系统通常已预装WebView2 runtime。如果没有,可以从[Microsoft官网](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)下载安装。

### 通过go get安装

```bash
go get github.com/yuaotian/go-win-webview2
```

## 🎮 使用示例

### 基础示例
```go
package main

import "github.com/yuaotian/go-win-webview2"

func main() {
    w := webview2.NewWithOptions(webview2.WebViewOptions{
        Debug: true,
        WindowOptions: webview2.WindowOptions{
            Title:  "基础示例",
            Width:  800,
            Height: 600,
            Center: true,
        },
    })
    defer w.Destroy()
    
    w.Navigate("https://example.com")
    w.Run()
}
```

### 高级功能示例
```go
package main

import (
    "log"
    "github.com/yuaotian/go-win-webview2"
)

func main() {
    // 创建带选项的窗口
    w := webview2.NewWithOptions(webview2.WebViewOptions{
        Debug: true,
        AutoFocus: true,
        WindowOptions: webview2.WindowOptions{
            Title:       "高级示例",
            Width:       1024,
            Height:      768,
            Center:      true,
            Frameless:   false,
            Fullscreen:  false,
            AlwaysOnTop: false,
        },
    })
    defer w.Destroy()

    // 注册热键
    w.RegisterHotKeyString("Ctrl+Alt+Q", func() {
        log.Println("退出应用...")
        w.Terminate()
    })

    // 设置事件监听
    w.OnLoadingStateChanged(func(isLoading bool) {
        if isLoading {
            log.Println("页面加载中...")
        } else {
            log.Println("页面加载完成!")
        }
    })

    // 启用WebSocket
    if err := w.EnableWebSocket(8080); err != nil {
        log.Printf("WebSocket启动失败: %v", err)
    }

    // 添加JavaScript钩子
    w.AddJSHook(&webview2.BaseJSHook{
        HookType: webview2.JSHookBefore,
        Handler: func(script string) string {
            log.Printf("执行脚本: %s", script)
            return script
        },
    })

    // 绑定Go函数到JavaScript
    w.Bind("greet", func(name string) string {
        return "Hello, " + name + "!"
    })

    w.Navigate("https://example.com")
    w.Run()
}
```

### WebSocket通信示例
```go
// 设置WebSocket消息处理器
w.OnWebSocketMessage(func(message string) {
    log.Printf("收到WebSocket消息: %s", message)
    // 发送响应
    w.SendWebSocketMessage(`{"type":"response","data":"消息已收到"}`)
})

// 在JavaScript中使用WebSocket
w.Eval(`
    window._webSocket.send(JSON.stringify({
        type: 'message',
        data: 'Hello from JavaScript!'
    }));
`)
```

### 事件监听示例
```go
// 监听页面加载状态
w.OnLoadingStateChanged(func(isLoading bool) {
    if isLoading {
        log.Println("页面加载中...")
    } else {
        log.Println("页面加载完成!")
    }
})

// 监听URL变化
w.OnURLChanged(func(url string) {
    log.Printf("页面URL已变更: %s", url)
})

// 监听标题变化
w.OnTitleChanged(func(title string) {
    log.Printf("页面标题已变更: %s", title)
    w.SetTitle(title) // 自动更新窗口标题
})

// 监听全屏状态变化
w.OnFullscreenChanged(func(isFullscreen bool) {
    log.Printf("全屏状态已变更: %v", isFullscreen)
})
```

### 热键绑定示例
```go
// 注册基本热键
w.RegisterHotKeyString("Ctrl+Q", func() {
    log.Println("退出应用...")
    w.Terminate()
})

// 注册功能热键
w.RegisterHotKeyString("F11", func() {
    log.Println("切换全屏...")
    // 在这里保存当前状态
    isFullscreen := false // 实际应用中需要跟踪此状态
    isFullscreen = !isFullscreen
    w.SetFullscreen(isFullscreen)
})

// 注册组合键
w.RegisterHotKeyString("Ctrl+Shift+D", func() {
    log.Println("打开开发者工具...")
    w.OpenDevTools()
})

// 注册窗口控制热键
w.RegisterHotKeyString("Ctrl+M", func() {
    log.Println("最小化窗口...")
    w.Minimize()
})
```

### JavaScript交互示例
```go
// 绑定Go函数到JavaScript
w.Bind("sayHello", func(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
})

// 绑定带错误处理的函数
w.Bind("divide", func(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("除数不能为零")
    }
    return a / b, nil
})

// 绑定异步操作
w.Bind("fetchData", func() interface{} {
    // 模拟异步操作
    time.Sleep(1 * time.Second)
    return map[string]interface{}{
        "status": "success",
        "data": []string{"item1", "item2", "item3"},
    }
})

// 在JavaScript中调用
w.Eval(`
    // 调用简单函数
    sayHello("World").then(result => {
        console.log(result); // 输出: Hello, World!
    });

    // 调用带错误处理的函数
    divide(10, 2).then(result => {
        console.log("10 ÷ 2 =", result);
    }).catch(err => {
        console.error("计算错误:", err);
    });

    // 调用异步函数
    fetchData().then(result => {
        console.log("获取的数据:", result);
    });
`)
```

### 窗口样式定制示例
```go
// 创建自定义样式的窗口
w := webview2.NewWithOptions(webview2.WebViewOptions{
    Debug: true,
    WindowOptions: webview2.WindowOptions{
        Title:       "自定义样式示例",
        Width:       800,
        Height:      600,
        Center:      true,
        Frameless:   true,  // 无边框模式
        AlwaysOnTop: true,  // 窗口置顶
    },
})

// 注入自定义样式
w.Init(`
    // 添加自定义标题栏
    const titleBar = document.createElement('div');
    titleBar.innerHTML = `+"`"+`
        <div style="display:flex;justify-content:space-between;align-items:center;padding:0 10px;">
            <div class="title">自定义标题栏</div>
            <div class="controls">
                <button onclick="minimize()">-</button>
                <button onclick="maximize()">□</button>
                <button onclick="closeWindow()">×</button>
            </div>
        </div>
    `+"`"+`;
    titleBar.style.cssText = 'position:fixed;top:0;left:0;right:0;height:30px;background:#f0f0f0;-webkit-app-region:drag;';
    document.body.appendChild(titleBar);

    // 添加控制按钮样式
    const style = document.createElement('style');
    style.textContent = `+"`"+`
        .controls button {
            border: none;
            background: none;
            padding: 5px 10px;
            cursor: pointer;
            -webkit-app-region: no-drag;
        }
        .controls button:hover {
            background: #e0e0e0;
        }
    `+"`"+`;
    document.head.appendChild(style);
`)

// 绑定窗口控制函数
w.Bind("minimize", func() {
    w.Minimize()
})

w.Bind("maximize", func() {
    // 这里可以添加最大化/还原的切换逻辑
    w.Maximize()
})

w.Bind("closeWindow", func() {
    w.Terminate()
})
```

### WebSocket高级示例
```go
// 启用WebSocket并处理不同类型的消息
w.EnableWebSocket(8080)

// 定义消息结构
type WSMessage struct {
    Type string      `json:"type"`
    Data interface{} `json:"data"`
}

// 设置消息处理器
w.OnWebSocketMessage(func(message string) {
    var msg WSMessage
    if err := json.Unmarshal([]byte(message), &msg); err != nil {
        log.Printf("解析消息失败: %v", err)
        return
    }

    // 根据消息类型处理
    switch msg.Type {
    case "ping":
        w.SendWebSocketMessage(`{"type":"pong"}`)
    
    case "eval":
        if script, ok := msg.Data.(string); ok {
            w.Eval(script)
        }
    
    case "notification":
        // 处理通知消息
        if data, ok := msg.Data.(map[string]interface{}); ok {
            log.Printf("收到通知: %v", data)
        }
    
    default:
        log.Printf("未知消息类型: %s", msg.Type)
    }
})

// 注入WebSocket客户端增强代码
w.Init(`
    // WebSocket 重连机制
    class WSClient {
        constructor(url, options = {}) {
            this.url = url;
            this.options = {
                reconnectInterval: 1000,
                maxReconnects: 5,
                ...options
            };
            this.reconnectCount = 0;
            this.handlers = new Map();
            this.connect();
        }

        connect() {
            this.ws = new WebSocket(this.url);
            this.ws.onopen = () => {
                console.log('WebSocket已连接');
                this.reconnectCount = 0;
                this.handlers.get('open')?.forEach(fn => fn());
            };
            
            this.ws.onclose = () => {
                console.log('WebSocket已断开');
                this.reconnect();
                this.handlers.get('close')?.forEach(fn => fn());
            };
            
            this.ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                this.handlers.get('message')?.forEach(fn => fn(data));
            };
        }

        reconnect() {
            if (this.reconnectCount < this.options.maxReconnects) {
                this.reconnectCount++;
                setTimeout(() => this.connect(), this.options.reconnectInterval);
            }
        }

        on(event, handler) {
            if (!this.handlers.has(event)) {
                this.handlers.set(event, new Set());
            }
            this.handlers.get(event).add(handler);
        }

        send(data) {
            if (this.ws.readyState === WebSocket.OPEN) {
                this.ws.send(JSON.stringify(data));
            }
        }
    }

    // 创建WebSocket客户端实例
    window._ws = new WSClient('ws://localhost:8080/ws', {
        reconnectInterval: 2000,
        maxReconnects: 10
    });

    // 添加事件监听
    window._ws.on('message', data => {
        console.log('收到消息:', data);
    });
`)
```

## 🛠 API参考

### 窗口控制
| API | 描述 |
|-----|------|
| `SetFullscreen(bool)` | 设置全屏模式 |
| `SetAlwaysOnTop(bool)` | 设置窗口置顶 |
| `SetOpacity(float64)` | 设置窗口透明度 |
| `Minimize()` | 最小化窗口 |
| `Maximize()` | 最大化窗口 |
| `Restore()` | 还原窗口 |
| `Center()` | 居中窗口 |

### 浏览器控制
| API | 描述 |
|-----|------|
| `Navigate(string)` | 导航到URL |
| `SetHtml(string)` | 设置HTML内容 |
| `Reload()` | 刷新页面 |
| `Back()` | 后退 |
| `Forward()` | 前进 |
| `Stop()` | 停止加载 |
| `ClearCache()` | 清除缓存 |
| `ClearCookies()` | 清除Cookies |

### 开发工具
| API | 描述 |
|-----|------|
| `OpenDevTools()` | 打开开发者工具 |
| `CloseDevTools()` | 关闭开发者工具 |
| `DisableContextMenu()` | 禁用右键菜单 |
| `EnableContextMenu()` | 启用右键菜单 |

### WebSocket相关
| API | 描述 |
|-----|------|
| `EnableWebSocket(port)` | 启用WebSocket服务 |
| `DisableWebSocket()` | 禁用WebSocket服务 |
| `OnWebSocketMessage(handler)` | 设置消息处理器 |
| `SendWebSocketMessage(message)` | 发送WebSocket消息 |

### JavaScript Hook
| API | 描述 |
|-----|------|
| `AddJSHook(hook)` | 添加JS钩子 |
| `RemoveJSHook(hook)` | 移除JS钩子 |
| `ClearJSHooks()` | 清除所有钩子 |

## 📝 常见问题

### Q: 如何处理窗口关闭事件?
```go
w.Bind("onClose", func() {
    // 执行清理操作
    w.Terminate()
})
```

### Q: 如何实现自定义标题栏?
```go
// 设置无边框窗口
w := webview2.NewWithOptions(webview2.WebViewOptions{
    WindowOptions: webview2.WindowOptions{
        Frameless: true,
    },
})

// 注入自定义标题栏HTML和CSS
w.Init(`
    const titleBar = document.createElement('div');
    titleBar.style.cssText = 'position:fixed;top:0;left:0;right:0;height:30px;-webkit-app-region:drag;background:#f0f0f0;';
    document.body.appendChild(titleBar);
`)
```

### Q: 如何优化WebSocket连接?
```go
// 启用带自动重连的WebSocket
w.Init(`
    function connectWebSocket() {
        if (!window._webSocket || window._webSocket.readyState !== 1) {
            window._webSocket = new WebSocket('ws://localhost:8080/ws');
            window._webSocket.onclose = () => {
                setTimeout(connectWebSocket, 1000);
            };
        }
    }
    connectWebSocket();
`)
```

## 🤝 贡献指南

欢迎提交问题和改进建议! 请查看我们的[贡献指南](CONTRIBUTING.md)了解更多信息。

1. Fork 项目
2. 创建新分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交Pull Request

## 📄 版权说明

该项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件

## 🙏 鸣谢
- [jchv/go-webview2](https://github.com/jchv/go-webview2)
- [webview/webview](https://github.com/webview/webview)
- [Microsoft Edge WebView2](https://docs.microsoft.com/microsoft-edge/webview2/)
- [Wails](https://wails.io/)

## 📊 项目状态

![Alt](https://repobeats.axiom.co/api/embed/your-analytics-hash.svg "Repobeats analytics image")

---

<p align="center">
  <sub>Built with ❤️ by 煎饼果子卷鲨鱼辣椒</sub>
</p>
