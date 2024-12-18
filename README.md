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
  <a href="#API参考">API参考</a> •
  <a href="#性能优化">性能优化</a> •
  <a href="#错误处理">错误处理</a> •
  <a href="#最佳实践">最佳实践</a>
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
  - 双向实时通信
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

#### 基础窗口配置
```go
// 创建自定义样式的窗口
w := webview2.NewWithOptions(webview2.WebViewOptions{
    Debug: true,
    WindowOptions: webview2.WindowOptions{
        Title:              "现代化窗口示例",
        Width:              1024,
        Height:             768,
        Center:            true,
        Frameless:         true,  // 无边框模式
        AlwaysOnTop:       false,
        DisableContextMenu: false,
        DefaultBackground: "#ffffff",
        Opacity:           1.0,
        Resizable:         true,
    },
})
```

#### 窗口状态管理
```go
// 定义窗口状态结构
type WindowState struct {
    sync.Mutex
    isFullscreen bool
    isMaximized  bool
    isMinimized  bool
    opacity      float64
    lastWidth    int
    lastHeight   int
    lastX        int
    lastY        int
}

// 初始化窗口状态
state := &WindowState{
    opacity: 1.0,
}
```

#### 自定义标题栏和窗口控制
```go
// 注入HTML和CSS样式
w.SetHtml(`
<!DOCTYPE html>
<html>
<head>
    <style>
        :root {
            --primary-color: #2196F3;
            --hover-color: #1976D2;
            --bg-color: #ffffff;
            --text-color: #333333;
            --title-bar-height: 36px;
            --resize-area: 8px;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body { 
            margin: 0;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto;
            background: var(--bg-color);
            color: var(--text-color);
            overflow: hidden;
            user-select: none;
        }

        .title-bar {
            -webkit-app-region: drag;
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            height: var(--title-bar-height);
            background: var(--primary-color);
            color: white;
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0 16px;
            z-index: 9998;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .controls {
            -webkit-app-region: no-drag;
            display: flex;
            align-items: center;
            gap: 4px;
        }

        .ctrl-btn {
            border: none;
            background: none;
            color: white;
            width: 46px;
            height: var(--title-bar-height);
            cursor: pointer;
            font-size: 14px;
            transition: all 0.2s ease;
        }

        .ctrl-btn:hover {
            background: var(--hover-color);
        }

        .close-btn:hover {
            background: #e81123 !important;
        }

        .resize-handle {
            position: fixed;
            z-index: 9999;
        }

        .resize-handle.top { top: 0; left: var(--resize-area); right: var(--resize-area); height: var(--resize-area); cursor: n-resize; }
        .resize-handle.right { top: var(--resize-area); right: 0; bottom: var(--resize-area); width: var(--resize-area); cursor: e-resize; }
        .resize-handle.bottom { bottom: 0; left: var(--resize-area); right: var(--resize-area); height: var(--resize-area); cursor: s-resize; }
        .resize-handle.left { top: var(--resize-area); left: 0; bottom: var(--resize-area); width: var(--resize-area); cursor: w-resize; }
        .resize-handle.top-left { top: 0; left: 0; width: var(--resize-area); height: var(--resize-area); cursor: nw-resize; }
        .resize-handle.top-right { top: 0; right: 0; width: var(--resize-area); height: var(--resize-area); cursor: ne-resize; }
        .resize-handle.bottom-left { bottom: 0; left: 0; width: var(--resize-area); height: var(--resize-area); cursor: sw-resize; }
        .resize-handle.bottom-right { bottom: 0; right: 0; width: var(--resize-area); height: var(--resize-area); cursor: se-resize; }
    </style>
</head>
<body>
    <div class="title-bar">
        <div class="title">现代化窗口示例</div>
        <div class="controls">
            <button class="ctrl-btn" onclick="window.minimize()" title="最小化">─</button>
            <button class="ctrl-btn" onclick="window.toggleMaximize()" title="最大化">□</button>
            <button class="ctrl-btn close-btn" onclick="window.closeWindow()" title="关闭">×</button>
        </div>
    </div>
    <div id="content">
        <!-- 页面内容 -->
    </div>
</body>
</html>
`)
```

#### 窗口控制函数绑定
```go
// 绑定窗口控制函数
func bindWindowControls(w webview2.WebView, state *WindowState) {
    // 最小化
    w.Bind("minimize", func() {
        state.Lock()
        state.isMinimized = true
        state.Unlock()
        w.Minimize()
    })

    // 最大化切换
    w.Bind("toggleMaximize", func() {
        state.Lock()
        defer state.Unlock()

        state.isMaximized = !state.isMaximized
        if state.isMaximized {
            // 保存当前窗口位置和大小
            var rect w32.Rect
            w32.GetWindowRect(w32.Handle(w.Window()), &rect)
            state.lastX = int(rect.Left)
            state.lastY = int(rect.Top)
            state.lastWidth = int(rect.Right - rect.Left)
            state.lastHeight = int(rect.Bottom - rect.Top)
            w.Maximize()
        } else {
            w.Restore()
        }
    })

    // 关闭窗口
    w.Bind("closeWindow", func() {
        w.Terminate()
    })

    // 窗口拖动
    w.Bind("startDragging", func() {
        hwnd := w.Window()
        w32.ReleaseCapture()
        w32.SendMessage(w32.Handle(uintptr(hwnd)), w32.WMNCLButtonDown, w32.HTCaption, 0)
    })

    // 窗口大小调整
    w.Bind("startResizing", func(edge string) {
        hwnd := w.Window()
        w32.ReleaseCapture()
        var hitTest uintptr
        switch edge {
        case "top":
            hitTest = w32.HTTop
        case "right":
            hitTest = w32.HTRight
        case "bottom":
            hitTest = w32.HTBottom
        case "left":
            hitTest = w32.HTLeft
        case "topLeft":
            hitTest = w32.HTTopLeft
        case "topRight":
            hitTest = w32.HTTopRight
        case "bottomLeft":
            hitTest = w32.HTBottomLeft
        case "bottomRight":
            hitTest = w32.HTBottomRight
        }
        w32.SendMessage(w32.Handle(uintptr(hwnd)), w32.WMNCLButtonDown, hitTest, 0)
    })
}
```

#### 注册快捷键
```go
// 注册窗口控制快捷键
func registerHotkeys(w webview2.WebView, state *WindowState) {
    // Ctrl+Q 退出
    w.RegisterHotKeyString("Ctrl+Q", func() {
        w.Terminate()
    })

    // Ctrl+M 最小化
    w.RegisterHotKeyString("Ctrl+M", func() {
        state.Lock()
        state.isMinimized = !state.isMinimized
        state.Unlock()
        if state.isMinimized {
            w.Minimize()
        } else {
            w.Restore()
        }
    })

    // F11 全屏
    w.RegisterHotKeyString("F11", func() {
        state.Lock()
        state.isFullscreen = !state.isFullscreen
        state.Unlock()
        w.SetFullscreen(state.isFullscreen)
    })
}
```

#### JavaScript事件处理
```javascript
// 添加到HTML中的JavaScript代码
document.addEventListener('DOMContentLoaded', function() {
    var titleBar = document.querySelector('.title-bar');
    
    // 添加窗口大小调整句柄
    var resizeAreas = [
        { class: 'top', edge: 'top' },
        { class: 'right', edge: 'right' },
        { class: 'bottom', edge: 'bottom' },
        { class: 'left', edge: 'left' },
        { class: 'top-left', edge: 'topLeft' },
        { class: 'top-right', edge: 'topRight' },
        { class: 'bottom-left', edge: 'bottomLeft' },
        { class: 'bottom-right', edge: 'bottomRight' }
    ];

    resizeAreas.forEach(area => {
        var handle = document.createElement('div');
        handle.className = 'resize-handle ' + area.class;
        handle.addEventListener('mousedown', function(e) {
            e.preventDefault();
            window.startResizing(area.edge);
        });
        document.body.appendChild(handle);
    });

    // 窗口拖动
    titleBar.addEventListener('mousedown', function(e) {
        if (!e.target.closest('.controls')) {
            window.startDragging();
        }
    });
});
```

这个示例展示了如何创建一个现代化的自定义窗口，包括：

1. 自定义标题栏
2. 窗口拖动
3. 边缘调整大小
4. 最大化/最小化/关闭控制
5. 快捷键支持
6. 窗口状态管理
7. 平滑动画过渡
8. 响应式布局

主要特点：
- 无边框设计
- 现代化UI风格
- 完整的窗口控制
- 状态同步管理
- 用户体验优化

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

## 🛠 性能优化

### 内存管理
```go
// 使用对象池复用WebView实例
var webviewPool = sync.Pool{
    New: func() interface{} {
        return webview2.NewWithOptions(webview2.WebViewOptions{
            Debug: false,
            WindowOptions: webview2.WindowOptions{
                Width:  800,
                Height: 600,
            },
        })
    },
}

// 获取WebView实例
w := webviewPool.Get().(webview2.WebView)
defer webviewPool.Put(w)
```

### 资源释放
```go
// 确保资源正确释放
func cleanup(w webview2.WebView) {
    w.Eval(`
        // 清理DOM事件监听器
        document.querySelectorAll('*').forEach(el => {
            el.replaceWith(el.cloneNode(true));
        });
        // 清理WebSocket连接
        if(window._ws) {
            window._ws.close();
        }
        // 清理定时器
        for(let i = setTimeout(()=>{}, 0); i > 0; i--) {
            clearTimeout(i);
        }
    `)
    w.Destroy()
}
```

### 渲染优化
```go
// 优化渲染性能
w.Init(`
    // 使用CSS containment优化重排
    .optimized-container {
        contain: content;
    }
    
    // 使用transform代替top/left
    .animated-element {
        transform: translate3d(0, 0, 0);
        will-change: transform;
    }
    
    // 避免大量DOM操作
    const fragment = document.createDocumentFragment();
    items.forEach(item => {
        const div = document.createElement('div');
        div.textContent = item;
        fragment.appendChild(div);
    });
    container.appendChild(fragment);
`)
```

## ⚠️ 错误处理

### 全局错误处理
```go
func setupErrorHandling(w webview2.WebView) {
    // JavaScript错误处理
    w.Init(`
        window.onerror = function(msg, url, line, col, error) {
            console.error('JavaScript错误:', {
                message: msg,
                url: url,
                line: line,
                column: col,
                error: error
            });
            return false;
        };
        
        window.onunhandledrejection = function(event) {
            console.error('未处理的Promise拒绝:', event.reason);
        };
    `)
    
    // Go端错误处理
    w.Bind("handleError", func(err string) {
        log.Printf("应用错误: %s", err)
        // 可以添加错误上报逻辑
    })
}
```

### 优雅降级
```go
// 功能检测和降级处理
w.Init(`
    // WebSocket支持检测
    if (!window.WebSocket) {
        console.warn('浏览器不支持WebSocket,使用轮询替代');
        startPolling();
    }
    
    // 存储API检测
    const storage = window.localStorage || {
        _data: {},
        setItem(id, val) { this._data[id] = val; },
        getItem(id) { return this._data[id]; }
    };
`)
```

### 错误恢复
```go
// 实现错误恢复机制
func recoverableOperation(w webview2.WebView, operation func() error) {
    const maxRetries = 3
    var err error
    
    for i := 0; i < maxRetries; i++ {
        err = operation()
        if err == nil {
            return
        }
        log.Printf("操作失败(重试 %d/%d): %v", i+1, maxRetries, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    // 最终失败处理
    w.Eval(`alert('操作失败,请稍后重试')`)
}
```

## 📚 最佳实践

### 代码组织
```go
// 模块化组织代码
type Application struct {
    webview webview2.WebView
    state   *WindowState
    config  *Config
}

func NewApplication() *Application {
    return &Application{
        webview: webview2.NewWithOptions(defaultOptions),
        state:   NewWindowState(),
        config:  LoadConfig(),
    }
}

func (app *Application) Initialize() {
    app.setupErrorHandling()
    app.setupEventListeners()
    app.setupHotkeys()
    app.loadInitialContent()
}
```

### 状态管理
```go
// 使用发布订阅模式管理状态
type StateManager struct {
    state     map[string]interface{}
    listeners map[string][]func(interface{})
    mu        sync.RWMutex
}

func (sm *StateManager) Subscribe(key string, listener func(interface{})) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    sm.listeners[key] = append(sm.listeners[key], listener)
}

func (sm *StateManager) SetState(key string, value interface{}) {
    sm.mu.Lock()
    sm.state[key] = value
    listeners := sm.listeners[key]
    sm.mu.Unlock()
    
    for _, listener := range listeners {
        listener(value)
    }
}
```

### 安全实践
```go
// 实现CSP策略
w.Init(`
    // 添加CSP meta标签
    const meta = document.createElement('meta');
    meta.httpEquiv = 'Content-Security-Policy';
    meta.content = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';";
    document.head.appendChild(meta);
    
    // 防止XSS
    function sanitizeHTML(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
`)

// 实现安全的消息传递
type SecureMessage struct {
    Payload   interface{} `json:"payload"`
    Timestamp int64      `json:"timestamp"`
    Signature string     `json:"signature"`
}

func (app *Application) sendSecureMessage(payload interface{}) {
    msg := SecureMessage{
        Payload:   payload,
        Timestamp: time.Now().Unix(),
        Signature: app.generateSignature(payload),
    }
    app.webview.Eval(fmt.Sprintf("window.handleSecureMessage(%s)", toJSON(msg)))
}
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
