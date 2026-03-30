//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	webview2 "github.com/yuaotian/go-win-webview2"
)

func main() {
	tempDir, err := os.MkdirTemp("", "go-win-webview2-desktop-demo-*")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	streamFile := filepath.Join(tempDir, "stream.txt")
	if err := os.WriteFile(streamFile, []byte("这是通过文件型 IStream 返回给页面的响应内容。"), 0o644); err != nil {
		log.Fatalf("写入流式响应文件失败: %v", err)
	}
	pdfPath := filepath.Join(tempDir, "desktop-control-demo.pdf")

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: true,
		WindowOptions: webview2.WindowOptions{
			Title:       "B/D/G 桌面控制演示",
			Width:       1120,
			Height:      760,
			Center:      true,
			Resizable:   true,
			Minimizable: true,
			Maximizable: true,
		},
	})
	if w == nil {
		log.Fatal("创建演示窗口失败")
	}
	defer w.Destroy()

	logToPage := func(message string) {
		w.Dispatch(func() {
			w.Eval(fmt.Sprintf("window.appendLog(%s);", strconv.Quote(message)))
		})
	}

	if err := w.Bind("showBrowserPrint", func() {
		if err := w.ShowPrintDialogWithKind(webview2.PrintDialogKindBrowser); err != nil {
			logToPage("打开浏览器打印对话框失败: " + err.Error())
			return
		}
		logToPage("已请求打开浏览器打印对话框")
	}); err != nil {
		log.Fatalf("绑定 showBrowserPrint 失败: %v", err)
	}

	if err := w.Bind("showSystemPrint", func() {
		if err := w.ShowPrintDialogWithKind(webview2.PrintDialogKindSystem); err != nil {
			logToPage("打开系统打印对话框失败: " + err.Error())
			return
		}
		logToPage("已请求打开系统打印对话框")
	}); err != nil {
		log.Fatalf("绑定 showSystemPrint 失败: %v", err)
	}

	if err := w.Bind("printPdf", func() {
		settings, err := w.NewPrintSettings()
		if err != nil {
			logToPage("创建打印设置失败: " + err.Error())
			return
		}
		if err := settings.SetOrientation(webview2.PrintOrientationLandscape); err != nil {
			logToPage("设置打印方向失败: " + err.Error())
			return
		}
		if err := settings.SetShouldPrintBackgrounds(true); err != nil {
			logToPage("设置背景打印失败: " + err.Error())
			return
		}
		if err := settings.SetShouldPrintHeaderAndFooter(true); err != nil {
			logToPage("设置页眉页脚失败: " + err.Error())
			return
		}
		if err := settings.SetHeaderTitle("desktop-control-demo"); err != nil {
			logToPage("设置页眉标题失败: " + err.Error())
			return
		}
		if err := w.PrintToPDFWithSettings(pdfPath, settings, func(result webview2.PrintToPDFResult) {
			if result.Error != nil {
				logToPage("打印到 PDF 失败: " + result.Error.Error())
				return
			}
			if result.Successful {
				logToPage("打印到 PDF 成功: " + pdfPath)
			}
		}); err != nil {
			logToPage("发起打印到 PDF 失败: " + err.Error())
			return
		}
		logToPage("已发起带设置的 PDF 导出")
	}); err != nil {
		log.Fatalf("绑定 printPdf 失败: %v", err)
	}

	if err := w.Bind("printPdfStream", func() {
		settings, err := w.NewPrintSettings()
		if err != nil {
			logToPage("创建打印设置失败: " + err.Error())
			return
		}
		if err := settings.SetShouldPrintBackgrounds(true); err != nil {
			logToPage("设置背景打印失败: " + err.Error())
			return
		}
		if err := w.PrintToPDFStreamWithSettings(settings, func(result webview2.PrintToPDFStreamResult) {
			if result.Error != nil {
				logToPage("导出 PDF 流失败: " + result.Error.Error())
				return
			}
			defer result.Stream.Release()
			if err := result.Stream.SaveToFile(pdfPath); err != nil {
				logToPage("保存 PDF 流失败: " + err.Error())
				return
			}
			logToPage("PrintToPDFStream 成功并已保存到: " + pdfPath)
		}); err != nil {
			logToPage("发起 PrintToPDFStream 失败: " + err.Error())
			return
		}
		logToPage("已发起 PrintToPDFStream")
	}); err != nil {
		log.Fatalf("绑定 printPdfStream 失败: %v", err)
	}

	if err := w.AddWebResourceRequestedFilter("https://app.local/*", webview2.WebResourceContextAll); err != nil {
		log.Fatalf("添加请求拦截过滤器失败: %v", err)
	}

	w.OnWebResourceRequested(func(event *webview2.WebResourceRequestedEvent) {
		switch event.URI {
		case "https://app.local/index.html":
			if err := event.SetResponse(webview2.WebResourceResponse{
				Content: []byte(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>desktop-control-demo</title>
  <style>
    :root {
      --bg: #f5f2ea;
      --panel: #fffdf8;
      --ink: #1f2a1f;
      --accent: #0f766e;
      --accent-soft: rgba(15, 118, 110, 0.12);
      --border: rgba(31, 42, 31, 0.12);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Segoe UI", "Microsoft YaHei", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(circle at top right, rgba(15, 118, 110, 0.12), transparent 35%),
        linear-gradient(135deg, #faf7f1 0%, #f1efe7 100%);
      min-height: 100vh;
    }
    .shell {
      width: min(980px, calc(100% - 32px));
      margin: 24px auto;
      display: grid;
      gap: 18px;
    }
    .hero, .panel {
      background: var(--panel);
      border: 1px solid var(--border);
      border-radius: 22px;
      box-shadow: 0 12px 40px rgba(31, 42, 31, 0.08);
    }
    .hero {
      padding: 28px;
    }
    .hero h1 {
      margin: 0 0 10px;
      font-size: clamp(28px, 4vw, 44px);
    }
    .hero p {
      margin: 0;
      line-height: 1.7;
      color: rgba(31, 42, 31, 0.72);
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
      gap: 18px;
    }
    .panel {
      padding: 20px;
    }
    .panel h2 {
      margin: 0 0 12px;
      font-size: 20px;
    }
    .panel p {
      margin: 0 0 16px;
      line-height: 1.7;
      color: rgba(31, 42, 31, 0.72);
    }
    button {
      appearance: none;
      border: 0;
      border-radius: 999px;
      padding: 10px 16px;
      background: var(--accent);
      color: white;
      cursor: pointer;
      margin-right: 10px;
      margin-bottom: 10px;
      font-size: 14px;
    }
    button.secondary {
      background: var(--accent-soft);
      color: var(--accent);
    }
    .target {
      margin-top: 12px;
      padding: 18px;
      border-radius: 16px;
      border: 1px dashed rgba(15, 118, 110, 0.4);
      background: rgba(15, 118, 110, 0.05);
    }
    .log {
      min-height: 180px;
      padding: 16px;
      border-radius: 16px;
      background: #14221c;
      color: #d8f7ef;
      font-family: Consolas, monospace;
      white-space: pre-wrap;
      line-height: 1.6;
    }
  </style>
</head>
<body>
  <div class="shell">
    <section class="hero">
      <h1>B / D / G 桌面控制演示</h1>
      <p>这个 demo 同时演示三件事：通过文件型 IStream 返回资源、拦截原生右键菜单并插入 Go 自定义项、以及使用打印设置对象驱动打印控制。</p>
    </section>

    <section class="grid">
      <article class="panel">
        <h2>B：文件型 IStream 响应</h2>
        <p id="stream-status">正在读取 https://app.local/stream.txt ...</p>
      </article>

	      <article class="panel">
	        <h2>D：原生右键菜单</h2>
	        <p>在下面区域点击右键，查看 Go 宿主插入的原生菜单项，以及 checkbox / radio / submenu。</p>
	        <div class="target">右键点击这里。点击 “Go 宿主菜单项” 后，右侧日志会追加记录。</div>
	      </article>

      <article class="panel">
        <h2>G：打印控制</h2>
        <p>按钮分别演示浏览器打印 UI、系统打印 UI、带设置 PDF 导出、以及 PDF 流导出。</p>
        <button onclick="window.showBrowserPrint()">浏览器打印 UI</button>
        <button class="secondary" onclick="window.showSystemPrint()">系统打印 UI</button>
        <button onclick="window.printPdf()">导出 PDF</button>
        <button class="secondary" onclick="window.printPdfStream()">导出 PDF 流</button>
      </article>
    </section>

    <section class="panel">
      <h2>日志</h2>
      <div class="log" id="log"></div>
    </section>
  </div>

  <script>
    window.appendLog = function (message) {
      const log = document.getElementById("log");
      log.textContent += message + "\n";
    };

    window.onGoMenu = function () {
      window.appendLog("点击了 Go 宿主菜单项");
    };

    fetch("https://app.local/stream.txt")
      .then(resp => resp.text())
      .then(text => {
        document.getElementById("stream-status").textContent = "文件型 IStream 响应成功，长度：" + text.length + "，内容：" + text;
        window.appendLog("文件型 IStream 资源读取成功");
      })
      .catch(err => {
        document.getElementById("stream-status").textContent = "读取失败：" + err;
        window.appendLog("文件型 IStream 资源读取失败：" + err);
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
		case "https://app.local/stream.txt":
			stream, err := webview2.NewStreamFromFile(streamFile)
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

	w.OnContextMenuRequested(func(event *webview2.ContextMenuRequestedEvent) {
		items, err := event.MenuItems()
		if err != nil {
			logToPage("读取右键菜单项失败: " + err.Error())
			return
		}
		if err := items.AppendSeparator(); err != nil {
			logToPage("插入菜单分隔符失败: " + err.Error())
			return
		}
		if err := items.AppendCommand("Go 宿主菜单项", func() {
			logToPage("Go 宿主菜单项回调已触发")
			w.Dispatch(func() {
				w.Eval(`window.onGoMenu();`)
			})
		}); err != nil {
			logToPage("插入 Go 宿主菜单项失败: " + err.Error())
			return
		}
		if err := items.AppendCheckBox("显示更多日志", true, func() {
			logToPage("checkbox 菜单项回调已触发")
		}); err != nil {
			logToPage("插入 checkbox 菜单项失败: " + err.Error())
			return
		}
		if err := items.AppendRadio("单选项示例", true, func() {
			logToPage("radio 菜单项回调已触发")
		}); err != nil {
			logToPage("插入 radio 菜单项失败: " + err.Error())
			return
		}
		if err := items.AppendSubMenu("Go 子菜单", nil, func(children *webview2.ContextMenuItemCollection) error {
			return children.AppendCommand("子菜单命令", func() {
				logToPage("submenu 菜单项回调已触发")
			})
		}); err != nil {
			logToPage("插入 submenu 菜单项失败: " + err.Error())
			return
		}
		logToPage(fmt.Sprintf("收到右键菜单请求：target=%d (%d,%d)", event.Target.Kind, event.Location.X, event.Location.Y))
	})

	w.Navigate("https://app.local/index.html")
	w.Run()
}
