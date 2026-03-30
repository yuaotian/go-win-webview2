//go:build windows
// +build windows

package main

import (
	"log"
	"time"

	webview2 "github.com/yuaotian/go-win-webview2"
)

func main() {
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: true,
		WindowOptions: webview2.WindowOptions{
			Title:       "高级能力演示",
			Width:       1100,
			Height:      760,
			Center:      true,
			Resizable:   true,
			Minimizable: true,
			Maximizable: true,
		},
	})
	if w == nil {
		log.Fatal("创建高级能力演示窗口失败")
	}
	defer w.Destroy()

	w.OnNewWindowRequested(func(event *webview2.NewWindowRequestedEvent) {
		log.Printf("收到新窗口请求: uri=%s userInitiated=%v", event.URI, event.IsUserInitiated)
		if err := event.Block(); err != nil {
			log.Printf("阻止新窗口失败: %v", err)
		}
	})

	w.OnWindowCloseRequested(func() {
		log.Printf("收到页面关闭请求")
		w.Terminate()
	})

	w.OnDOMContentLoaded(func(event webview2.DOMContentLoadedEvent) {
		log.Printf("DOMContentLoaded: navigationId=%d", event.NavigationID)
	})

	if err := w.AddWebResourceRequestedFilter("https://app.local/*", webview2.WebResourceContextAll); err != nil {
		log.Printf("添加请求拦截过滤器失败: %v", err)
	}

	w.OnWebResourceRequested(func(event *webview2.WebResourceRequestedEvent) {
		switch event.URI {
		case "https://app.local/index.html":
			log.Printf("拦截主文档请求: %s", event.URI)
			if err := event.SetResponse(webview2.WebResourceResponse{
				Content: []byte(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>advanced-demo</title></head>
<body>
  <h1>高级能力演示</h1>
  <p>当前稳定可演示：请求拦截、新窗口请求、窗口关闭请求、截图、DevTools method。</p>
  <p>当前仍在继续深挖：DevTools event receiver、虚拟主机映射运行态。</p>
  <button onclick="window.open('about:blank#popup', '_blank')">测试新窗口请求</button>
  <button onclick="window.close()">测试窗口关闭请求</button>
  <script>
    fetch("https://app.local/api/data")
      .then(resp => resp.json())
      .then(data => {
        document.body.insertAdjacentHTML("beforeend", "<pre>" + JSON.stringify(data, null, 2) + "</pre>");
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
			log.Printf("拦截 API 请求: %s", event.URI)
			if err := event.SetResponse(webview2.WebResourceResponse{
				Content:      []byte(`{"feature":"web resource requested","status":"ok"}`),
				StatusCode:   200,
				ReasonPhrase: "OK",
				Headers:      "Content-Type: application/json; charset=utf-8\r\n",
			}); err != nil {
				log.Printf("设置 API 响应失败: %v", err)
			}
		}
	})

	if err := w.CallDevToolsProtocolMethod("Runtime.enable", `{}`, func(err error, result string) {
		if err != nil {
			log.Printf("发起 Runtime.enable 失败: %v", err)
			return
		}
		log.Printf("Runtime.enable 返回: %s", result)
	}); err != nil {
		log.Printf("调用 Runtime.enable 失败: %v", err)
	}

	if err := w.CallDevToolsProtocolMethod("Runtime.evaluate", `{"expression":"document.title"}`, func(err error, result string) {
		if err != nil {
			log.Printf("发起 Runtime.evaluate 失败: %v", err)
			return
		}
		log.Printf("Runtime.evaluate 返回: %s", result)
	}); err != nil {
		log.Printf("调用 Runtime.evaluate 失败: %v", err)
	}

	// 当前本机运行态里，SubscribeDevToolsProtocolEvent 仍在继续深挖，故这里保留示例调用并记录结果。
	if _, err := w.SubscribeDevToolsProtocolEvent("Security.securityStateChanged", func(payload string) {
		log.Printf("收到 DevTools 事件: %s", payload)
	}); err != nil {
		log.Printf("订阅 DevTools 事件失败（当前仍在深挖）: %v", err)
	}

	// 当前本机运行态里，虚拟主机映射仍在继续深挖，故这里仅保留目标调用方式，不默认执行。
	_ = func() error {
		return w.SetVirtualHostNameToFolderMapping("demo.example", "C:/replace/with/real/path", webview2.HostResourceAccessKindAllow)
	}

	w.Navigate("https://app.local/index.html")

	time.AfterFunc(2*time.Second, func() {
		if err := w.CapturePreviewWithOptions("advanced-demo-preview.png", webview2.CapturePreviewOptions{
			Format: webview2.CapturePreviewFormatPNG,
		}); err != nil {
			log.Printf("截图失败: %v", err)
		} else {
			log.Printf("截图请求已发出: advanced-demo-preview.png")
		}
	})

	w.Run()
}
