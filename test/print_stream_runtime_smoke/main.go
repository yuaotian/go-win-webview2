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
	pdfPath := filepath.Join(os.TempDir(), "go-win-webview2-print-stream-smoke.pdf")
	_ = os.Remove(pdfPath)

	var requested atomic.Bool
	var saved atomic.Bool

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:       "print stream smoke",
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
		if isLoading || requested.Load() {
			return
		}
		requested.Store(true)

		time.AfterFunc(300*time.Millisecond, func() {
			settings, err := w.NewPrintSettings()
			if err != nil {
				log.Fatalf("创建打印设置失败: %v", err)
			}
			if err := settings.SetShouldPrintBackgrounds(true); err != nil {
				log.Fatalf("设置背景打印失败: %v", err)
			}
			if err := w.PrintToPDFStreamWithSettings(settings, func(result webview2.PrintToPDFStreamResult) {
				if result.Error != nil {
					log.Printf("PrintToPDFStream 回调失败: %v", result.Error)
					return
				}
				defer result.Stream.Release()
				if err := result.Stream.SaveToFile(pdfPath); err != nil {
					log.Printf("保存 PDF 流失败: %v", err)
					return
				}
				saved.Store(true)
			}); err != nil {
				log.Fatalf("发起 PrintToPDFStream 失败: %v", err)
			}
		})
	})

	w.SetHtml(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>print-stream-smoke</title>
</head>
<body>
  <h1>print stream smoke</h1>
</body>
</html>`)

	time.AfterFunc(3000*time.Millisecond, func() {
		if !saved.Load() {
			log.Fatal("未观察到 PrintToPDFStream 成功保存结果")
		}
		info, err := os.Stat(pdfPath)
		if err != nil {
			log.Fatalf("未生成 PDF 文件: %v", err)
		}
		if info.Size() == 0 {
			log.Fatal("生成的 PDF 文件大小为 0")
		}
		log.Printf("print stream smoke 验证通过，PDF 文件：%s", pdfPath)
		requestClose(w)
	})

	w.Run()
}
