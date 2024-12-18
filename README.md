# 🚀 go-win-webview2

<p align="center">
  <img src="assets/logo.png" alt="Logo" width="200" height="200">
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

- ⚡️ 原生窗口控制
  - 支持无边框窗口
  - 窗口大小调整
  - 全屏切换
  - 窗口置顶
  - 透明度控制
  
- 🌐 完整的Web功能
  - HTML/CSS/JavaScript支持
  - 双向通信机制
  - Cookie管理
  - 缓存控制
  
- ⌨️ 快捷键支持
  - 全局热键注册
  - 自定义快捷键
  - 多组合键支持

- 🎮 窗口操作
  - 最大化/最小化
  - 窗口居中
  - 自定义图标
  - 窗口样式定制

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
    w := webview2.New(true)
    defer w.Destroy()
    w.SetTitle("示例应用")
    w.SetSize(800, 600, webview2.HintNone)
    w.Navigate("https://github.com")
    w.Run()
}
```

### 高级示例
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
            Frameless:   true,
            Fullscreen:  false,
            AlwaysOnTop: false,
        },
    })
    defer w.Destroy()

    // 注册全局热键
    w.RegisterHotKeyString("Ctrl+Alt+Q", func() {
        log.Println("退出应用...")
        w.Terminate()
    })

    // 注册窗口状态热键
    w.RegisterHotKeyString("F11", func() {
        w.SetFullscreen(true)
    })

    // 设置窗口透明度
    w.SetOpacity(0.95)

    // 注入CSS样式
    w.Init(`
        body { 
            background: #f0f0f0;
            font-family: 'Segoe UI', sans-serif;
        }
    `)

    // 绑定Go函数到JavaScript
    w.Bind("greet", func(name string) string {
        return "Hello, " + name + "!"
    })

    // 监听页面加载状态
    w.OnLoadingStateChanged(func(isLoading bool) {
        if isLoading {
            log.Println("页面加载中...")
        } else {
            log.Println("页面加载完成!")
        }
    })

    w.Navigate("https://example.com")
    w.Run()
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

### 热键管理
| API | 描述 |
|-----|------|
| `RegisterHotKeyString(string, func())` | 注册热键 |
| `UnregisterHotKey(int, int)` | 注销热键 |

### 浏览器控制
| API | 描述 |
|-----|------|
| `Navigate(string)` | 导航到URL |
| `Reload()` | 刷新页面 |
| `Back()` | 后退 |
| `Forward()` | 前进 |
| `ClearCache()` | 清除缓存 |
| `ClearCookies()` | 清除Cookies |

### 事件监听
| API | 描述 |
|-----|------|
| `OnLoadingStateChanged(func(bool))` | 加载状态变化 |
| `OnURLChanged(func(string))` | URL变化 |
| `OnTitleChanged(func(string))` | 标题变化 |
| `OnFullscreenChanged(func(bool))` | 全屏状态变化 |

## 📝 常见问题

### Q: 如何处理窗口关闭事件?
```go
w.Bind("onClose", func() {
    // 执行清理操作
    w.Terminate()
})
```

### Q: 如何注入自定义CSS?
```go
w.Init(`
    document.head.insertAdjacentHTML('beforeend', '
        <style>
            body { background: #f0f0f0; }
        </style>
    ');
`)
```

### Q: 如何实现拖拽功能?
```go
w.Init(`
    document.body.style.webkitAppRegion = 'drag';
    document.querySelector('.no-drag').style.webkitAppRegion = 'no-drag';
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
