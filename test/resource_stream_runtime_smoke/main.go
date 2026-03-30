//go:build windows
// +build windows

package main

import (
	"bytes"
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
	tempDir, err := os.MkdirTemp("", "go-win-webview2-stream-*")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	const payloadSize = 2 * 1024 * 1024
	largeFilePath := filepath.Join(tempDir, "large.txt")
	if err := os.WriteFile(largeFilePath, bytes.Repeat([]byte("A"), payloadSize), 0o644); err != nil {
		log.Fatalf("写入大文件失败: %v", err)
	}

	var sawFileResponse atomic.Bool
	var sawTitle atomic.Bool

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:       "resource stream smoke",
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

	if err := w.AddWebResourceRequestedFilter("https://stream.local/*", webview2.WebResourceContextAll); err != nil {
		log.Fatalf("添加请求过滤器失败: %v", err)
	}

	w.OnWebResourceRequested(func(event *webview2.WebResourceRequestedEvent) {
		switch event.URI {
		case "https://stream.local/index.html":
			if err := event.SetResponse(webview2.WebResourceResponse{
				Content: []byte(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>stream-start</title></head>
<body>
  <script>
    fetch("https://stream.local/large.txt")
      .then(resp => resp.text())
      .then(text => {
        document.title = "stream-" + text.length;
      });
  </script>
</body>
</html>`),
				StatusCode:   200,
				ReasonPhrase: "OK",
				Headers:      "Content-Type: text/html; charset=utf-8\r\n",
			}); err != nil {
				log.Printf("设置主文档响应失败: %v", err)
			}
		case "https://stream.local/large.txt":
			sawFileResponse.Store(true)
			stream, err := webview2.NewStreamFromFile(largeFilePath)
			if err != nil {
				log.Printf("创建 raw stream 失败: %v", err)
				return
			}
			defer stream.Release()
			if err := event.SetStreamResponse(stream, 200, "OK", "Content-Type: text/plain; charset=utf-8\r\n"); err != nil {
				log.Printf("设置文件流响应失败: %v", err)
			}
		}
	})

	w.OnTitleChanged(func(title string) {
		log.Printf("标题变化: %s", title)
		if title == "stream-2097152" {
			sawTitle.Store(true)
		}
	})

	w.Navigate("https://stream.local/index.html")

	time.AfterFunc(5000*time.Millisecond, func() {
		if !sawFileResponse.Load() {
			log.Fatal("未观察到文件流响应")
		}
		if !sawTitle.Load() {
			log.Fatal("未观察到文件流响应后的标题变化")
		}
		log.Printf("resource stream smoke 验证通过")
		requestClose(w)
	})

	w.Run()
}
