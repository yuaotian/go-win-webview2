//go:build windows
// +build windows

package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	webview2 "github.com/yuaotian/go-win-webview2"
	"github.com/yuaotian/go-win-webview2/internal/w32"
)

// 通过主线程发送 WM_CLOSE，确保 smoke 按正常窗口关闭链路收尾。
func requestClose(w webview2.WebView) {
	w.Dispatch(func() {
		hwnd := uintptr(w.Window())
		if hwnd == 0 {
			return
		}
		_, _, _ = w32.User32PostMessageW.Call(hwnd, w32.WMClose, 0, 0)
	})
}

func main() {
	previewPath := filepath.Join(os.TempDir(), "go-win-webview2-runtime-smoke.png")
	_ = os.Remove(previewPath)

	var sawURLChange atomic.Bool
	var sawTitleChange atomic.Bool
	var captureRequested atomic.Bool

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:       "runtime smoke",
			Width:       960,
			Height:      640,
			Center:      true,
			Resizable:   true,
			Minimizable: true,
			Maximizable: true,
		},
	})
	if w == nil {
		log.Fatal("创建 WebView 失败")
	}

	w.OnURLChanged(func(url string) {
		log.Printf("URL 变化: %s", url)
		if strings.Contains(url, "#verified") {
			sawURLChange.Store(true)
		}
	})

	w.OnTitleChanged(func(title string) {
		log.Printf("标题变化: %s", title)
		if title == "runtime-finished" {
			sawTitleChange.Store(true)
		}
	})

	w.OnLoadingStateChanged(func(isLoading bool) {
		log.Printf("加载状态: %v", isLoading)
		if isLoading || captureRequested.Load() {
			return
		}

		captureRequested.Store(true)
		time.AfterFunc(300*time.Millisecond, func() {
			if err := w.CapturePreview(previewPath); err != nil {
				log.Printf("截图请求失败: %v", err)
			} else {
				log.Printf("截图请求已发出: %s", previewPath)
			}

			if os.Getenv("GO_WIN_WEBVIEW2_SMOKE_OPEN_DEVTOOLS") == "1" {
				log.Printf("尝试打开 DevTools")
				w.OpenDevTools()
			} else {
				log.Printf("跳过 DevTools 自动验证，避免阻塞 smoke 流程")
			}
		})
	})

	w.SetHtml(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>runtime-start</title>
  <script>
    window.addEventListener('DOMContentLoaded', function () {
      setTimeout(function () {
        document.title = 'runtime-finished';
        location.hash = 'verified';
      }, 120);
    });
  </script>
</head>
<body>
  <h1>runtime smoke</h1>
</body>
</html>`)

	time.AfterFunc(2500*time.Millisecond, func() {
		if !sawURLChange.Load() {
			log.Fatal("未观察到 hash 触发的 URL 变化")
		}
		if !sawTitleChange.Load() {
			log.Fatal("未观察到标题变化")
		}
		if _, err := os.Stat(previewPath); err != nil {
			log.Fatalf("未生成截图文件: %v", err)
		}

		log.Printf("运行态 smoke 验证通过，截图文件：%s", previewPath)
		requestClose(w)
	})

	w.Run()
}
