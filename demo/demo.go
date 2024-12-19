//go:build windows
// +build windows

package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	webview2 "github.com/yuaotian/go-win-webview2"
)

//go:embed tmp.html
var htmlTest embed.FS

func main() {
	var isFullscreen, isTopmost bool

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     true,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:              "增强版 Webview 演示",
			Width:              1024,
			Height:             768,
			IconId:             1,
			IconPath:           "logo.ico",
			Center:             true,      // 居中
			Frameless:          false,     // 无边框
			Fullscreen:         false,     // 全屏
			AlwaysOnTop:        false,     // 置顶
			Resizable:          true,      // 可调整大小
			Minimizable:        true,      // 可最小化
			Maximizable:        true,      // 可最大化
			DisableContextMenu: false,      // 禁用右键菜单
			EnableDragAndDrop:  true,      // 启用拖放
			HideWindowOnClose:  false,     // 关闭时是否隐藏窗口不是退出
			DefaultBackground:  "#FFFFFF", // 默认背景色 (CSS 格式，如 "#FFFFFF")
		},
	})
	if w == nil {
		log.Fatalln("加载 webview 失败。")
	}
	defer w.Destroy()
	miniWindow := false
	// 基本控制热键
	hotkeys := map[string]struct {
		desc    string
		handler webview2.HotKeyHandler
	}{
		"Ctrl+Alt+Q": {"退出程序", func() {
			log.Println("Ctrl+Alt+Q 按下，退出...")
			w.Terminate()
			os.Exit(0)
		}},
		"Ctrl+Alt+M": {"最小化窗口", func() {
			if miniWindow {
				log.Println("还原窗口...")
				w.Restore()
				miniWindow = false
			} else {
				log.Println("最小化窗口...")
				w.Minimize()
				miniWindow = true
			}
		}},
		"Ctrl+Alt+F": {"切换全屏", func() {
			isFullscreen = !isFullscreen
			log.Printf("切换全屏: %v", isFullscreen)
			w.SetFullscreen(isFullscreen)
		}},
		"Ctrl+Alt+T": {"切换置顶", func() {
			isTopmost = !isTopmost
			log.Printf("切换置顶: %v", isTopmost)
			w.SetAlwaysOnTop(isTopmost)
		}},
		// // 浏览器控制热键
		// "Ctrl+R": {"刷新页面", func() {
		// 	log.Println("刷新页面...")
		// 	w.Reload()
		// }},
		// "Alt+Left": {"后退", func() {
		// 	log.Println("后退...")
		// 	w.Back()
		// }},
		// "Alt+Right": {"前进", func() {
		// 	log.Println("前进...")
		// 	w.Forward()
		// }},
		// "Esc": {"停止加载", func() {
		// 	log.Println("停止加载...")
		// 	w.Stop()
		// }},

		// // 窗口控制热键
		// "Ctrl+M": {"最小化", func() {
		// 	log.Println("最小化窗口...")
		// 	w.Minimize()
		// }},
		// "Ctrl+Up": {"最大化", func() {
		// 	log.Println("最大化窗口...")
		// 	w.Maximize()
		// }},
		// "Ctrl+Down": {"还原窗口", func() {
		// 	log.Println("还原窗口...")
		// 	w.Restore()
		// }},
		// "Ctrl+C": {"居中窗口", func() {
		// 	log.Println("居中窗口...")
		// 	w.Center()
		// }},

		// // 开发工具热键
		// "F12": {"开发者工具", func() {
		// 	log.Println("打开开发者工具...")
		// 	w.OpenDevTools()
		// }},

		// // 清理热键
		// "Ctrl+Shift+Delete": {"清除缓存", func() {
		// 	log.Println("清除缓存...")
		// 	w.ClearCache()
		// }},
		// "Ctrl+Shift+C": {"清除Cookies", func() {
		// 	log.Println("清除Cookies...")
		// 	w.ClearCookies()
		// }},
	}

	// 注册所有热键
	for key, item := range hotkeys {
		if err := w.RegisterHotKeyString(key, item.handler); err != nil {
			log.Printf("警告: 注册热键 %s (%s) 失败: %v", key, item.desc, err)
		} else {
			log.Printf("成功注册热键: %s (%s)", key, item.desc)
		}
	}

	// 读取HTML内容
	htmlContent, err := fs.ReadFile(htmlTest, "tmp.html")
	if err != nil {
		log.Fatalln("无法读取 HTML:", err)
	}

	// 注入HTML内容
	w.Init(`
		document.addEventListener('DOMContentLoaded', function() {
			document.body.insertAdjacentHTML('beforeend', ` + "`" + string(htmlContent) + "`" + `);
		});
	`)

	// 关闭开发者工具
	w.CloseDevTools()

	// 禁用右键菜单
	w.DisableContextMenu()

	// 加载网页
	w.Navigate("https://html5test.com/")

	// 设置半透明
	w.SetOpacity(0.95)

	// 居中显示
	w.Center()

	// 状态监听
	w.OnLoadingStateChanged(func(isLoading bool) {
		if isLoading {
			log.Println("页面正在加载...")
		} else {
			log.Println("页面加载完成!")
		}
	})

	w.OnURLChanged(func(url string) {
		log.Printf("URL已更改: %s", url)
	})

	w.OnTitleChanged(func(title string) {
		log.Printf("标题已更改: %s", title)
		w.SetTitle(title) // 自动更新窗口标题
	})

	w.OnFullscreenChanged(func(isFullscreen bool) {
		log.Printf("全屏状态: %v", isFullscreen)
	})

	// 定时任务例：每60秒刷新一次页面
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			w.Dispatch(func() {
				log.Println("执行定时刷新...")
				w.Reload()
			})
		}
	}()

	// 新窗口请求处理
	w.OnNewWindowRequested(func(url string) bool {
		log.Printf("检测到新窗口请求: %s", url)

		// 根据 URL 决定打开方式
		if strings.Contains(url, "html5test.com") {
			// html5test.com 的链接在当前窗口打开
			log.Printf("在当前窗口打开链接: %s", url)
			return true
		} else {
			// 其他链接允许在新窗口打开
			log.Printf("允许在新窗口打开链接: %s", url)
			return false
		}
	})

	// 添加一些测试链接到页面
	w.Eval(`
		(function() {
			// 等待 DOM 完全加载
			if (document.readyState === 'loading') {
				document.addEventListener('DOMContentLoaded', addLinks);
			} else {
				addLinks();
			}

			function addLinks() {
				// 检查是否已经添加过链接
				if (document.getElementById('test-links-container')) {
					return;
				}

				// 创建测试链接列表
				const links = [
					{ url: 'https://html5test.com/results.html', text: 'HTML5 Test Results (当前窗口打开)' },
					{ url: 'https://www.google.com', text: 'Google (新窗口打开)' },
					{ url: 'https://example.com', text: '示例链接 (当前窗口打开)' }
				];

				// 创建链接容器
				const container = document.createElement('div');
				container.id = 'test-links-container';
				container.style.cssText = 'position: fixed; top: 10px; right: 10px; background: white; padding: 10px; border: 1px solid #ccc; border-radius: 5px; z-index: 9999;';
				
				// 添加标题
				const title = document.createElement('h3');
				title.textContent = '新窗口测试链接';
				container.appendChild(title);

				// 创建链接列表
				const ul = document.createElement('ul');
				links.forEach(link => {
					const li = document.createElement('li');
					const a = document.createElement('a');
					a.href = link.url;
					a.target = '_blank';
					a.textContent = link.text;
					// 移除点击事件处理，让 WebView2 直接处理新窗口请求
					li.appendChild(a);
					ul.appendChild(li);
				});
				container.appendChild(ul);

				// 添加到页面
				document.body.appendChild(container);
			}
		})();
	`)

	// 运行
	w.Run()
}
