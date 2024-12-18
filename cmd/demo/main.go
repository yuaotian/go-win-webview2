//go:build windows
// +build windows

package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"github.com/yuaotian/go-win-webview2"
)

//go:embed tmp.html
var content embed.FS

func main() {
	var isFullscreen, isTopmost bool

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     true,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:       "Enhanced Webview Demo",
			Width:       1024,
			Height:      768,
			IconId:      2,    // 图标资源ID
			Center:      true, // 是否居中
			Frameless:   true, // 无边框窗口样式
			Fullscreen:  true, // 启动时全屏
			AlwaysOnTop: true, // 启动时置顶
		},
	})
	if w == nil {
		log.Fatalln("Failed to load webview.")
	}
	defer w.Destroy()

	// 使用字符串注册热键
	err := w.RegisterHotKeyString("Ctrl+Alt+Q", func() {
		log.Println("Ctrl+Alt+Q 按下，退出...")
		w.Terminate()
		os.Exit(0)
	})
	if err != nil {
		log.Printf("Warning: 注册热键失败: %v", err)
	}

	// 可以注册多个热键
	err = w.RegisterHotKeyString("Ctrl+Alt+M", func() {
		log.Println("Minimizing window...")
		w.Dispatch(func() {
			w.Minimize()
		})
	})
	if err != nil {
		log.Printf("Warning: 注册热键失败: %v", err)
	}

	// 添加全屏热键 (F11)
	err = w.RegisterHotKeyString("F11", func() {
		isFullscreen = !isFullscreen
		log.Printf("切换全屏: %v", isFullscreen)
		w.SetFullscreen(isFullscreen)
	})
	if err != nil {
		log.Printf("Warning: 注册热键失败: %v", err)
	}

	// 添加置顶热键 (Ctrl+T)
	err = w.RegisterHotKeyString("Ctrl+T", func() {
		isTopmost = !isTopmost
		log.Printf("始终在顶部切换: %v", isTopmost)
		w.SetAlwaysOnTop(isTopmost)
	})
	if err != nil {
		log.Printf("Warning: 注册热键失败: %v", err)
	}

	// 读取HTML内容
	htmlContent, err := fs.ReadFile(content, "tmp.html")
	if err != nil {
		log.Fatalln("无法读取 HTML:", err)
	}
	//println(string(htmlContent))
	// 注入HTML内容
	w.Init(`
		document.addEventListener('DOMContentLoaded', function() {
			document.body.insertAdjacentHTML('beforeend', ` + "`" + string(htmlContent) + "`" + `);
		});
	`)
	// 加载网页
	w.Navigate("https://html5test.com/")
	// 注入js
	w.Eval("console.log('Hello, World!');")

	// 注册热键示例
	w.RegisterHotKeyString("Ctrl+M", func() {
		w.Minimize()
	})

	w.RegisterHotKeyString("Ctrl+R", func() {
		w.Reload()
	})

	// 设置半透明
	w.SetOpacity(0.9)

	// 居中显示
	w.Center()

	// 加载状态监听
	w.OnLoadingStateChanged(func(isLoading bool) {
		if isLoading {
			log.Println("页面正在加载...")
		} else {
			log.Println("页面已加载!")
		}
	})

	w.Run()
}
