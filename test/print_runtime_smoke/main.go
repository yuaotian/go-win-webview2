//go:build windows
// +build windows

package main

import (
	"log"
	"os"
	"path/filepath"
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
	pdfPath := filepath.Join(os.TempDir(), "go-win-webview2-print-smoke.pdf")
	_ = os.Remove(pdfPath)

	var printRequested atomic.Bool

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:       "print smoke",
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

	w.OnLoadingStateChanged(func(isLoading bool) {
		log.Printf("加载状态: %v", isLoading)
		if isLoading || printRequested.Load() {
			return
		}
		printRequested.Store(true)

		time.AfterFunc(300*time.Millisecond, func() {
			w.PrintToPDF(pdfPath)
			log.Printf("打印到 PDF 请求已发出: %s", pdfPath)
		})
	})

	w.SetHtml(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>print-smoke</title>
</head>
<body>
  <h1>print smoke</h1>
  <p>verify print to pdf binding</p>
</body>
</html>`)

	time.AfterFunc(3000*time.Millisecond, func() {
		info, err := os.Stat(pdfPath)
		if err != nil {
			log.Fatalf("未生成 PDF 文件: %v", err)
		}
		if info.Size() == 0 {
			log.Fatal("生成的 PDF 文件大小为 0")
		}
		log.Printf("print smoke 验证通过，PDF 文件：%s", pdfPath)
		requestClose(w)
	})

	w.Run()
}
