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
)

func main() {
	tempDir, err := os.MkdirTemp("", "go-win-webview2-mapped-*")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mappedPath := filepath.Join(tempDir, "index.html")
	if err := os.WriteFile(mappedPath, []byte(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>mapped-ready</title></head>
<body><h1>mapped page</h1></body>
</html>`), 0o644); err != nil {
		log.Fatalf("写入映射页面失败: %v", err)
	}

	var domLoadedCount atomic.Int32
	var sawIntercept atomic.Bool
	var sawMapped atomic.Bool
	var once atomic.Bool

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:       "resource smoke",
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

	if err := w.AddWebResourceRequestedFilter("https://app.local/*", webview2.WebResourceContextAll); err != nil {
		log.Fatalf("添加请求过滤器失败: %v", err)
	}

	w.OnWebResourceRequested(func(event *webview2.WebResourceRequestedEvent) {
		switch event.URI {
		case "https://app.local/index.html":
			log.Printf("拦截到主文档请求: %s", event.URI)
			if err := event.SetResponse(webview2.WebResourceResponse{
				Content: []byte(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>intercept-start</title></head>
<body>
  <script>
    fetch("https://app.local/api/data")
      .then(resp => resp.json())
      .then(data => {
        document.title = data.title;
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
		case "https://app.local/api/data":
			log.Printf("拦截到 API 请求: %s", event.URI)
			if err := event.SetResponse(webview2.WebResourceResponse{
				Content:      []byte(`{"title":"intercept-done"}`),
				StatusCode:   200,
				ReasonPhrase: "OK",
				Headers:      "Content-Type: application/json; charset=utf-8\r\n",
			}); err != nil {
				log.Printf("设置 API 响应失败: %v", err)
			}
		}
	})

	w.OnDOMContentLoaded(func(event webview2.DOMContentLoadedEvent) {
		log.Printf("DOMContentLoaded: %d", event.NavigationID)
		domLoadedCount.Add(1)
	})

	w.OnTitleChanged(func(title string) {
		log.Printf("标题变化: %s", title)
		if title == "intercept-done" && !once.Load() {
			once.Store(true)
			sawIntercept.Store(true)

			time.AfterFunc(100*time.Millisecond, func() {
				if err := w.SetVirtualHostNameToFolderMapping("mapped.example", tempDir, webview2.HostResourceAccessKindAllow); err != nil {
					log.Printf("设置虚拟主机映射失败: %v", err)
					return
				}
				w.Navigate("https://mapped.example/index.html")
			})
		}

		if title == "mapped-ready" {
			sawMapped.Store(true)
			w.Terminate()
		}
	})

	w.Navigate("https://app.local/index.html")

	time.AfterFunc(5000*time.Millisecond, func() {
		if !sawIntercept.Load() {
			log.Fatal("未观察到请求拦截后的标题变化")
		}
		if !sawMapped.Load() {
			log.Fatal("未观察到虚拟主机映射页面加载")
		}
		if domLoadedCount.Load() < 2 {
			log.Fatal("DOMContentLoaded 次数不足，未覆盖两段导航")
		}
		log.Printf("resource smoke 验证通过")
		os.Exit(0)
	})

	w.Run()
}
