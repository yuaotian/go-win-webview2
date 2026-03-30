//go:build windows
// +build windows

package main

import (
	"log"
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
	var once atomic.Bool
	var sawContextMenu atomic.Bool
	var sawCustomInsert atomic.Bool
	var sawPageTarget atomic.Bool
	var sawExtendedItems atomic.Bool

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:       "context menu smoke",
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

	w.OnContextMenuRequested(func(event *webview2.ContextMenuRequestedEvent) {
		log.Printf("收到右键菜单请求: target=%d location=(%d,%d)", event.Target.Kind, event.Location.X, event.Location.Y)
		sawContextMenu.Store(true)
		if event.Target.Kind == webview2.ContextMenuTargetKindPage {
			sawPageTarget.Store(true)
		}

		items, err := event.MenuItems()
		if err != nil {
			log.Printf("读取菜单项失败: %v", err)
			return
		}
		if err := items.AppendCommand("smoke-command", func() {
			log.Printf("smoke-command 被点击")
		}); err != nil {
			log.Printf("插入自定义菜单项失败: %v", err)
		} else {
			sawCustomInsert.Store(true)
		}
		if err := items.AppendCheckBox("smoke-checkbox", true, func() {
			log.Printf("smoke-checkbox 被点击")
		}); err != nil {
			log.Printf("插入 checkbox 菜单项失败: %v", err)
		} else if err := items.AppendRadio("smoke-radio", true, func() {
			log.Printf("smoke-radio 被点击")
		}); err != nil {
			log.Printf("插入 radio 菜单项失败: %v", err)
		} else if err := items.AppendSubMenu("smoke-submenu", nil, func(children *webview2.ContextMenuItemCollection) error {
			return children.AppendCommand("submenu-command", func() {
				log.Printf("submenu-command 被点击")
			})
		}); err != nil {
			log.Printf("插入 submenu 菜单项失败: %v", err)
		} else {
			sawExtendedItems.Store(true)
		}

		// 自动化 smoke 只验证事件和自定义项插入链路，菜单本身不展示，避免阻塞。
		if err := event.SetHandled(true); err != nil {
			log.Printf("设置 handled 失败: %v", err)
		}
	})

	w.OnLoadingStateChanged(func(isLoading bool) {
		if isLoading || once.Load() {
			return
		}
		once.Store(true)

		time.AfterFunc(400*time.Millisecond, func() {
			if err := w.CallDevToolsProtocolMethod("Input.dispatchMouseEvent", `{"type":"mousePressed","x":60,"y":60,"button":"right","buttons":2,"clickCount":1}`, nil); err != nil {
				log.Printf("发送鼠标按下失败: %v", err)
			}
			if err := w.CallDevToolsProtocolMethod("Input.dispatchMouseEvent", `{"type":"mouseReleased","x":60,"y":60,"button":"right","buttons":0,"clickCount":1}`, nil); err != nil {
				log.Printf("发送鼠标抬起失败: %v", err)
			}
		})
	})

	w.SetHtml(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>context-menu-smoke</title></head>
<body>
  <div style="margin:40px;width:200px;height:120px;border:1px solid #999;">right click target</div>
</body>
</html>`)

	time.AfterFunc(4000*time.Millisecond, func() {
		if !sawContextMenu.Load() {
			log.Fatal("未观察到 ContextMenuRequested 事件")
		}
		if !sawCustomInsert.Load() {
			log.Fatal("未观察到自定义菜单项插入成功")
		}
		if !sawExtendedItems.Load() {
			log.Fatal("未观察到 checkbox/radio/submenu 插入成功")
		}
		if !sawPageTarget.Load() {
			log.Fatal("未观察到页面目标类型")
		}
		log.Printf("context menu smoke 验证通过")
		requestClose(w)
	})

	w.Run()
}
