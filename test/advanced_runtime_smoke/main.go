//go:build windows
// +build windows

package main

import (
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	webview2 "github.com/yuaotian/go-win-webview2"
)

func main() {
	var sawNewWindow atomic.Bool
	var sawWindowClose atomic.Bool
	var sawDevToolsMethod atomic.Bool
	var sawDevToolsEvent atomic.Bool
	var once atomic.Bool
	var unsubscribe func()

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:       "advanced smoke",
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

	w.OnNewWindowRequested(func(event *webview2.NewWindowRequestedEvent) {
		log.Printf("收到新窗口请求: %s", event.URI)
		sawNewWindow.Store(true)
		if err := event.Block(); err != nil {
			log.Printf("阻止新窗口失败: %v", err)
		}
	})

	w.OnWindowCloseRequested(func() {
		log.Printf("收到窗口关闭请求")
		sawWindowClose.Store(true)
		w.Terminate()
	})

	w.OnLoadingStateChanged(func(isLoading bool) {
		log.Printf("加载状态: %v", isLoading)
		if isLoading || once.Load() {
			return
		}
		once.Store(true)

		if err := w.CallDevToolsProtocolMethod("Runtime.enable", `{}`, func(err error, result string) {
			if err != nil {
				log.Printf("启用 Runtime 域失败: %v", err)
				return
			}
			log.Printf("Runtime.enable 返回: %s", result)

			if err := w.CallDevToolsProtocolMethod("Runtime.evaluate", `{"expression":"document.title"}`, func(err error, result string) {
				if err != nil {
					log.Printf("Runtime.evaluate 失败: %v", err)
					return
				}
				log.Printf("Runtime.evaluate 返回: %s", result)
				if strings.Contains(result, "runtime-advanced") {
					sawDevToolsMethod.Store(true)
				}
			}); err != nil {
				log.Printf("调用 Runtime.evaluate 失败: %v", err)
			}

			if err := w.CallDevToolsProtocolMethod("Security.enable", `{}`, func(err error, result string) {
				if err != nil {
					log.Printf("启用 Security 域失败: %v", err)
					return
				}
				log.Printf("Security.enable 返回: %s", result)

				var subscribeErr error
				unsubscribe, subscribeErr = w.SubscribeDevToolsProtocolEvent("Security.securityStateChanged", func(json string) {
					log.Printf("收到 DevTools 事件: %s", json)
					sawDevToolsEvent.Store(true)
				})
				if subscribeErr != nil {
					log.Printf("订阅 DevTools 事件失败: %v", subscribeErr)
				}

				w.Eval(`
					setTimeout(function () {
						window.open("about:blank#popup", "_blank");
					}, 80);
					setTimeout(function () {
						window.close();
					}, 220);
				`)
			}); err != nil {
				log.Printf("调用 Security.enable 失败: %v", err)
			}
		}); err != nil {
			log.Printf("调用 Runtime.enable 失败: %v", err)
		}
	})

	w.SetHtml(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>runtime-advanced</title>
</head>
<body>
  <h1>advanced smoke</h1>
</body>
</html>`)

	time.AfterFunc(4000*time.Millisecond, func() {
		if unsubscribe != nil {
			unsubscribe()
		}
		if !sawNewWindow.Load() {
			log.Fatal("未观察到新窗口请求")
		}
		if !sawWindowClose.Load() {
			log.Fatal("未观察到窗口关闭请求")
		}
		if !sawDevToolsMethod.Load() {
			log.Fatal("未观察到 DevTools method 回调")
		}
		if !sawDevToolsEvent.Load() {
			log.Fatal("未观察到 DevTools event 回调")
		}

		log.Printf("advanced smoke 验证通过")
		os.Exit(0)
	})

	w.Run()
}
