package main

import (
	"embed"
	"log"
	"os"
	"sync"
	"unsafe"

	webview2 "github.com/yuaotian/go-win-webview2"
	"github.com/yuaotian/go-win-webview2/internal/w32"
)

//go:embed favicon.ico
var iconData embed.FS

// WindowState 管理窗口状态
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

// 窗口边缘调整大小的区域大小
const (
	resizeAreaSize = 8
)

func getIconBytes() []byte {
	data, err := iconData.ReadFile("favicon.ico")
	if err != nil {
		log.Printf("读取图标失败: %v", err)
		return nil
	}
	if len(data) == 0 {
		log.Printf("警告: 图标文件为空")
		return nil
	}
	return data
}

func main() {
	// 初始化窗口状态
	state := &WindowState{
		opacity: 1.0,
	}

	// 创建自定义样式的窗口
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: true,
		WindowOptions: webview2.WindowOptions{
			Title:              "现代化窗口示例",
			//IconData:           getIconBytes(),
			IconPath:           "favicon.ico",
			Width:              1024,
			Height:             768,
			Center:             true,
			Frameless:          true,
			AlwaysOnTop:        false,
			DisableContextMenu: false,
			DefaultBackground:  "#ffffff",
			Opacity:            state.opacity,
			Resizable:          false,
		},
	})

	// 绑定窗口控制函数
	bindWindowControls(w, state)

	// 设置HTML内容
	w.SetHtml(getHTML())

	// 注册快捷键
	registerHotkeys(w, state)

	// 运行窗口
	w.Run()
}

func bindWindowControls(w webview2.WebView, state *WindowState) {
	// 最小化
	w.Bind("minimize", func() {
		log.Println("窗口最小化")
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

			// 获取工作区大小
			var workArea w32.Rect
			w32.SystemParametersInfo(w32.SPI_GETWORKAREA, 0, unsafe.Pointer(&workArea), 0)

			log.Println("窗口最大化")
			w.Maximize()
		} else {
			log.Println("窗口还原")
			w.Restore()
		}
	})

	// 关闭窗口
	w.Bind("closeWindow", func() {
		log.Println("关闭窗口")
		w.Terminate()
		w.Destroy()
		os.Exit(0)
	})

	// 设置透明度
	w.Bind("setOpacity", func(value float64) {
		log.Printf("设置透明度: %.2f", value)
		state.Lock()
		state.opacity = value
		state.Unlock()
		w.SetOpacity(value)
	})

	// 全屏切换
	w.Bind("toggleFullscreen", func() {
		state.Lock()
		state.isFullscreen = !state.isFullscreen
		state.Unlock()
		log.Printf("切换全屏: %v", state.isFullscreen)
		w.SetFullscreen(state.isFullscreen)
	})

	// 窗口居中
	w.Bind("centerWindow", func() {
		log.Println("窗口居中")
		w.Center()
	})

	// 窗口拖动
	w.Bind("startDragging", func() {
		log.Println("开始拖动窗口")
		hwnd := w.Window()
		w32.ReleaseCapture()
		w32.SendMessage(w32.Handle(uintptr(hwnd)), w32.WMNCLButtonDown, w32.HTCaption, 0)
	})

	// 窗口大小调整
	w.Bind("startResizing", func(edge string) {
		log.Printf("开始调整窗口大小: %s", edge)
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

func registerHotkeys(w webview2.WebView, state *WindowState) {
	// Ctrl+Q 退出
	w.RegisterHotKeyString("Ctrl+Q", func() {
		log.Println("触发快捷键: Ctrl+Q")
		w.Terminate()
		w.Destroy()
		os.Exit(0)
	})

	// Ctrl+M 最小化
	w.RegisterHotKeyString("Ctrl+M", func() {
		log.Println("触发快捷键: Ctrl+M")
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
		log.Println("触发快捷键: F11")
		state.Lock()
		state.isFullscreen = !state.isFullscreen
		state.Unlock()
		w.SetFullscreen(state.isFullscreen)
	})
}

func getHTML() string {
	return `
    <!DOCTYPE html>
    <html>
    <head>
        <title>现代化窗口示例</title>
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
                font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
                background: var(--bg-color);
                color: var(--text-color);
                overflow: hidden;
                user-select: none;
            }

            .resize-handle {
                position: fixed;
                z-index: 9999;
            }

            .resize-handle.top {
                top: 0;
                left: var(--resize-area);
                right: var(--resize-area);
                height: var(--resize-area);
                cursor: n-resize;
            }

            .resize-handle.right {
                top: var(--resize-area);
                right: 0;
                bottom: var(--resize-area);
                width: var(--resize-area);
                cursor: e-resize;
            }

            .resize-handle.bottom {
                bottom: 0;
                left: var(--resize-area);
                right: var(--resize-area);
                height: var(--resize-area);
                cursor: s-resize;
            }

            .resize-handle.left {
                top: var(--resize-area);
                left: 0;
                bottom: var(--resize-area);
                width: var(--resize-area);
                cursor: w-resize;
            }

            .resize-handle.top-left {
                top: 0;
                left: 0;
                width: var(--resize-area);
                height: var(--resize-area);
                cursor: nw-resize;
            }

            .resize-handle.top-right {
                top: 0;
                right: 0;
                width: var(--resize-area);
                height: var(--resize-area);
                cursor: ne-resize;
            }

            .resize-handle.bottom-left {
                bottom: 0;
                left: 0;
                width: var(--resize-area);
                height: var(--resize-area);
                cursor: sw-resize;
            }

            .resize-handle.bottom-right {
                bottom: 0;
                right: 0;
                width: var(--resize-area);
                height: var(--resize-area);
                cursor: se-resize;
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
                cursor: move;
                box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            }

            .controls {
                -webkit-app-region: no-drag;
                display: flex;
                align-items: center;
                gap: 4px;
                cursor: default;
            }

            .title {
                flex: 1;
                font-size: 14px;
                font-weight: 500;
                margin-right: 16px;
            }

            .ctrl-btn {
                border: none;
                background: none;
                color: white;
                width: 46px;
                height: var(--title-bar-height);
                cursor: pointer;
                font-family: "Segoe MDL2 Assets", sans-serif;
                font-size: 14px;
                transition: all 0.2s ease;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .ctrl-btn:hover {
                background: var(--hover-color);
            }

            .close-btn:hover {
                background: #e81123 !important;
            }

            .main-content {
                margin-top: var(--title-bar-height);
                padding: 24px;
                height: calc(100vh - var(--title-bar-height));
                overflow-y: auto;
            }

            .card {
                background: white;
                border-radius: 8px;
                padding: 24px;
                box-shadow: 0 2px 8px rgba(0,0,0,0.1);
                margin-bottom: 24px;
                animation: fadeIn 0.3s ease;
            }

            @keyframes fadeIn {
                from { opacity: 0; transform: translateY(10px); }
                to { opacity: 1; transform: translateY(0); }
            }

            h1 {
                font-size: 24px;
                font-weight: 500;
                margin-bottom: 16px;
                color: var(--primary-color);
            }

            .feature-list {
                list-style: none;
                margin-top: 16px;
            }

            .feature-list li {
                padding: 8px 0;
                display: flex;
                align-items: center;
                gap: 8px;
                transition: all 0.2s ease;
            }

            .feature-list li:hover {
                transform: translateX(4px);
            }

            .feature-list li::before {
                content: "•";
                color: var(--primary-color);
                font-weight: bold;
            }

            .shortcut {
                background: #f5f5f5;
                padding: 2px 6px;
                border-radius: 4px;
                font-family: monospace;
                font-size: 12px;
                transition: all 0.2s ease;
            }

            .shortcut:hover {
                background: var(--primary-color);
                color: white;
            }
        </style>
        <script>
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

                for (var i = 0; i < resizeAreas.length; i++) {
                    var area = resizeAreas[i];
                    var handle = document.createElement('div');
                    handle.className = 'resize-handle ' + area.class;
                    handle.addEventListener('mousedown', function(e) {
                        e.preventDefault();
                        window.startResizing(this.getAttribute('data-edge'));
                    });
                    handle.setAttribute('data-edge', area.edge);
                    document.body.appendChild(handle);
                }

                // 窗口拖动
                titleBar.addEventListener('mousedown', function(e) {
                    if (!e.target.closest('.controls')) {
                        window.startDragging();
                    }
                });

                // 添加键盘快捷键提示
                var shortcuts = document.querySelectorAll('.shortcut');
                for (var i = 0; i < shortcuts.length; i++) {
                    shortcuts[i].setAttribute('title', '点击使用此快捷键');
                }
            });
        </script>
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
        <div class="main-content">
            <div class="card">
                <h1>功能介绍</h1>
                <div class="content">
                    <ul class="feature-list">
                        <li>现代化无边框设计，支持窗口拖动</li>
                        <li>窗口控制：最小化、最大化、关闭</li>
                        <li>快捷键支持：
                            <span class="shortcut">Ctrl+Q</span> 退出
                            <span class="shortcut">Ctrl+M</span> 最小化
                            <span class="shortcut">F11</span> 全屏
                        </li>
                        <li>完善的窗口状态管理</li>
                        <li>支持窗口边缘拖动调整大小</li>
                        <li>优雅的动画过渡效果</li>
                        <li>响应式布局设计</li>
                    </ul>
                </div>
            </div>
        </div>
    </body>
    </html>
    `
}
