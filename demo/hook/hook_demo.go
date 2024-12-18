package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	webview2 "github.com/yuaotian/go-win-webview2"
)

// 全局配置
const (
	defaultPort   = 8080
	defaultTitle  = "HTML5 测试 - JSHook 和 WebSocket 演示"
	defaultWidth  = 1024
	defaultHeight = 768
	wsWindowTitle = "WebSocket 测试控制台"
	wsWindowWidth = 800
	wsWindowHeight = 600
)

// 应用状态
type AppState struct {
	startTime     time.Time
	wsConnections sync.Map
	perfStats     map[string]interface{}
	mu            sync.RWMutex
}

// WSTestWindow WebSocket测试窗口
type WSTestWindow struct {
	w       webview2.WebView
	mainWin webview2.WebView
}

func main() {
	// 初始化应用状态
	state := &AppState{
		startTime: time.Now(),
		perfStats: make(map[string]interface{}),
	}

	// 创建带调试功能的 webview
	w := createWebView(true)
	if w == nil {
		log.Fatalln("Failed to create webview")
	}
	defer w.Destroy()

	// 初始化基本配置
	setupWebView(w, state)

	// 添加钩子
	setupHooks(w)

	// 启用 WebSocket
	if err := setupWebSocket(w, state); err != nil {
		log.Printf("WebSocket 启动失败: %v", err)
	}

	// 注入控制面板和测试脚本
	injectUIComponents(w)

	// 导航到测试页面
	w.Navigate("https://html5test.com")

	// 启动状态监控
	go monitorServerStatus(w, state)

	// 运行
	w.Run()
}

// createWebView 创建并配置WebView实例
func createWebView(debug bool) webview2.WebView {
	w := webview2.New(debug)
	if w == nil {
		return nil
	}

	w.SetTitle(defaultTitle)
	w.SetSize(defaultWidth, defaultHeight, webview2.HintNone)

	return w
}

// setupWebView 设置WebView基本配置
func setupWebView(w webview2.WebView, state *AppState) {
	// 注册状态变化回调
	w.OnLoadingStateChanged(func(isLoading bool) {
		if isLoading {
			log.Println("页面加载中...")
		} else {
			log.Println("页面加载完成")
			// 页面加载完成后重新注入所有组件
			injectCustomStyles(w)
			injectUIComponents(w)
			w.Eval("window._initWebSocket();")
		}
	})

	// 注册URL变化回调
	w.OnURLChanged(func(url string) {
		log.Printf("URL已变更: %s", url)
		// URL变更后也需要重新注入
		w.Eval(`
			setTimeout(() => {
				if (!document.getElementById('controls')) {
					injectControlPanel();
				}
			}, 1000);
		`)
	})
}

// setupHooks 设置JS钩子
func setupHooks(w webview2.WebView) {
	// 性能监控钩子
	perfHook := &webview2.BaseJSHook{
		HookType: webview2.JSHookBefore,
		Handler: func(script string) string {
			if strings.Contains(script, "_webSocket") {
				return script
			}
			return wrapWithPerformanceMonitoring(script)
		},
		HookPriority: 0,
	}

	// 安全检查钩子
	securityHook := &webview2.BaseJSHook{
		HookType: webview2.JSHookBefore,
		Handler: func(script string) string {
			checkSecurityIssues(script)
			return script
		},
		HookPriority: 1,
	}

	w.AddJSHook(perfHook)
	w.AddJSHook(securityHook)
}

// wrapWithPerformanceMonitoring 包装脚本以添加性能监控
func wrapWithPerformanceMonitoring(script string) string {
	return fmt.Sprintf(`
		console.time('执行时间');
		try {
			const startMemory = performance.memory ? performance.memory.usedJSHeapSize : 0;
			%s
			if (performance.memory) {
				const endMemory = performance.memory.usedJSHeapSize;
				console.log('内存使用:', (endMemory - startMemory) / 1024 / 1024, 'MB');
			}
		} finally {
			console.timeEnd('执行时间');
			console.log('脚本大小:', %d, '字节');
		}
	`, script, len(script))
}

// checkSecurityIssues 检查安全问题
func checkSecurityIssues(script string) {
	sensitivePatterns := map[string]string{
		"eval\\(":           "使用eval",
		"document\\.cookie": "访问cookie",
		"localStorage":      "使用localStorage",
		"sessionStorage":    "使用sessionStorage",
		"new Function":      "动态创建函数",
		"innerHTML":         "使用innerHTML",
		"document\\.write":  "使用document.write",
	}

	for pattern, desc := range sensitivePatterns {
		if matched, _ := regexp.MatchString(pattern, script); matched {
			log.Printf("安全警告: 检测到%s", desc)
		}
	}
}

// setupWebSocket 设置WebSocket服务
func setupWebSocket(w webview2.WebView, state *AppState) error {
	err := w.EnableWebSocket(defaultPort)
	if err != nil {
		return fmt.Errorf("启动WebSocket失败: %v", err)
	}

	// 创建测试窗口
	wsWindow := createWSTestWindow(w)
	if wsWindow != nil {
		go wsWindow.w.Run()
	}

	// 确保WebSocket连接在页面加载完成后初始化
	w.Eval(fmt.Sprintf(`
		window._initWebSocket = function() {
			if (!window._webSocket || window._webSocket.readyState !== 1) {
				try {
					window._webSocket = new WebSocket('ws://localhost:%d/ws');
					window._webSocket.onopen = () => {
						console.log('WebSocket已连接');
						window._webSocket.send(JSON.stringify({
							type: 'status',
							data: '主窗口WebSocket已连接'
						}));
					};
					window._webSocket.onclose = () => console.log('WebSocket已断开');
					window._webSocket.onerror = (e) => console.error('WebSocket错误:', e);
					window._webSocket.onmessage = (event) => {
						try {
							const data = JSON.parse(event.data);
							console.log('收到消息:', data);
							if (data.type === 'console') {
								eval(data.data);
							}
						} catch(e) {
							console.error('解析消息失败:', e);
						}
					};
				} catch(e) {
					console.error('WebSocket连接失败:', e);
				}
			}
			return window._webSocket;
		};
		
		// 定期检查并重连
		setInterval(() => {
			if (!window._webSocket || window._webSocket.readyState !== 1) {
				window._initWebSocket();
			}
		}, 3000);
	`, defaultPort))

	w.OnLoadingStateChanged(func(isLoading bool) {
		if !isLoading {
			w.Eval("window._initWebSocket();")
		}
	})

	w.OnWebSocketMessage(func(message string) {
		handleWebSocketMessage(w, message, state)
	})

	return nil
}

// handleWebSocketMessage 处理WebSocket消息
func handleWebSocketMessage(w webview2.WebView, message string, state *AppState) {
	var msg struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Printf("解析WebSocket消息失败: %v", err)
		return
	}

	switch msg.Type {
	case "performance":
		handlePerformanceData(msg.Data, state)
	case "security":
		handleSecurityData(msg.Data)
	case "html5test":
		handleHTML5TestData(msg.Data)
	default:
		log.Printf("收到未知类型消息: %s", msg.Type)
	}

	// 发送确认
	w.SendWebSocketMessage(fmt.Sprintf(`{"type":"ack","data":"已收到%s类型消息"}`, msg.Type))
}

// monitorServerStatus 监控服务器状态
func monitorServerStatus(w webview2.WebView, state *AppState) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := getServerStats(state)
		message, err := json.Marshal(map[string]interface{}{
			"type": "serverStatus",
			"data": stats,
		})
		if err != nil {
			log.Printf("JSON序列化错误: %v", err)
			continue
		}
		w.SendWebSocketMessage(string(message))
	}
}

// getServerStats 获取服务器状态
func getServerStats(state *AppState) map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(state.startTime).String(),
		"memory": map[string]uint64{
			"allocated": m.Alloc,
			"total":     m.TotalAlloc,
			"system":    m.Sys,
			"gc_cycles": uint64(m.NumGC),
		},
		"goroutines": runtime.NumGoroutine(),
	}
}

// injectCustomStyles 注入自定义样式
func injectCustomStyles(w webview2.WebView) {
	w.Eval(`
		if (!document.getElementById('custom-controls-style')) {
			const style = document.createElement('style');
			style.id = 'custom-controls-style';
			style.textContent = ` + "`" + `
				#controls { 
					position: fixed !important;
					top: 20px !important;
					right: 20px !important;
					background: rgba(240, 240, 240, 0.95) !important;
					padding: 15px !important;
					border-radius: 8px !important;
					box-shadow: 0 2px 10px rgba(0,0,0,0.2) !important;
					z-index: 9999 !important;
					backdrop-filter: blur(5px) !important;
					transition: all 0.3s ease !important;
				}
				#controls:hover {
					transform: translateY(-2px) !important;
					box-shadow: 0 4px 15px rgba(0,0,0,0.3) !important;
				}
				#controls button {
					margin: 5px 0 !important;
					padding: 8px 12px !important;
					border: none !important;
					border-radius: 4px !important;
					background: #007bff !important;
					color: white !important;
					cursor: pointer !important;
					transition: background 0.2s !important;
					display: block !important;
					width: 100% !important;
				}
				#controls button:hover {
					background: #0056b3 !important;
				}
			` + "`" + `;
			document.head.appendChild(style);
		}
	`)
}

// injectUIComponents 注入控制面板和测试脚本
func injectUIComponents(w webview2.WebView) {
	// 先注入样式
	injectCustomStyles(w)

	// 注入控制面板和功能
	w.Eval(`
		// 拖拽功能设置
		function setupDragAndDrop() {
			const controls = document.getElementById('controls');
			let isDragging = false;
			let currentX;
			let currentY;
			let initialX;
			let initialY;
			let xOffset = 0;
			let yOffset = 0;

			controls.addEventListener('mousedown', dragStart);
			document.addEventListener('mousemove', drag);
			document.addEventListener('mouseup', dragEnd);

			function dragStart(e) {
				if (e.target.tagName === 'BUTTON') return;
				initialX = e.clientX - xOffset;
				initialY = e.clientY - yOffset;
				if (e.target === controls) {
					isDragging = true;
				}
			}

			function drag(e) {
				if (isDragging) {
					e.preventDefault();
					currentX = e.clientX - initialX;
					currentY = e.clientY - initialY;
					xOffset = currentX;
					yOffset = currentY;
					controls.style.transform = 'translate3d(' + currentX + 'px, ' + currentY + 'px, 0)';
				}
			}

			function dragEnd() {
				isDragging = false;
			}
		}

		// HTML5特性测试
		function testFeatures() {
			const features = {
				webSocket: 'WebSocket' in window,
				webGL: (() => {
					try {
						return !!window.WebGLRenderingContext && !!document.createElement('canvas').getContext('experimental-webgl');
					} catch(e) {
						return false;
					}
				})(),
				localStorage: (() => {
					try {
						return 'localStorage' in window && window.localStorage !== null;
					} catch(e) {
						return false;
					}
				})(),
				webWorkers: 'Worker' in window,
				webRTC: 'RTCPeerConnection' in window,
				geolocation: 'geolocation' in navigator
			};

			console.info('HTML5特性测试结果:', features);
			if (window._webSocket) {
				window._webSocket.send(JSON.stringify({
					type: 'html5test',
					data: features
				}));
			}
		}

		// 性能分析
		function capturePerformance() {
			const timing = performance.timing;
			const performanceData = {
				loadTime: timing.loadEventEnd - timing.navigationStart,
				domReadyTime: timing.domContentLoadedEventEnd - timing.navigationStart,
				firstPaintTime: performance.getEntriesByType('paint')[0]?.startTime || 0,
				resourceCount: performance.getEntriesByType('resource').length,
				memoryInfo: performance.memory ? {
					usedJSHeapSize: performance.memory.usedJSHeapSize,
					totalJSHeapSize: performance.memory.totalJSHeapSize
				} : null
			};

			console.info('性能数据:', performanceData);
			if (window._webSocket) {
				window._webSocket.send(JSON.stringify({
					type: 'performance',
					data: performanceData
				}));
			}
		}

		// 安全检查
		function checkSecurity() {
			const securityInfo = {
				https: location.protocol === 'https:',
				contentSecurityPolicy: !!document.querySelector('meta[http-equiv="Content-Security-Policy"]'),
				xssProtection: !!document.querySelector('meta[http-equiv="X-XSS-Protection"]'),
				frameOptions: !!document.querySelector('meta[http-equiv="X-Frame-Options"]'),
				cookies: document.cookie ? document.cookie.split(';').length : 0,
				localStorage: Object.keys(localStorage).length,
				thirdPartyScripts: Array.from(document.scripts).filter(s => {
					if (!s.src) return false;
					try {
						const url = new URL(s.src);
						return url.host !== window.location.host;
					} catch(e) {
						return false;
					}
				}).length
			};

			console.info('安全检查结果:', securityInfo);
			if (window._webSocket) {
				window._webSocket.send(JSON.stringify({
					type: 'security',
					data: securityInfo
				}));
			}
		}

		// WebSocket测试
		function toggleWebSocket() {
			if (!window._webSocket || window._webSocket.readyState !== 1) {
				window._initWebSocket();
				console.info('正在重新连接WebSocket...');
			} else {
				console.info('WebSocket连接状态:', {
					readyState: window._webSocket.readyState,
					bufferedAmount: window._webSocket.bufferedAmount,
					protocol: window._webSocket.protocol
				});
			}
		}

		function injectControlPanel() {
			if (!document.getElementById('controls')) {
				const div = document.createElement('div');
				div.id = 'controls';
				div.innerHTML = ` + "`" + `
					<div style="margin-bottom:10px;font-weight:bold;text-align:center;cursor:move;">
						测试控制面板
						<span style="float:right;cursor:pointer;padding:0 5px" onclick="toggleConsole()">🔽</span>
					</div>
					<div class="button-group">
						<button onclick="testFeatures()">测试 HTML5 特性</button>
						<button onclick="capturePerformance()">性能分析</button>
						<button onclick="checkSecurity()">安全检查</button>
						<button onclick="toggleWebSocket()">WebSocket 测试</button>
					</div>
					<div id="controlConsole" style="display:none">
						<div class="console-container">
							<div id="consoleOutput" class="console-output"></div>
							<div class="console-input-group">
								<textarea id="codeInput" placeholder="输入JavaScript代码..." rows="3"></textarea>
								<button onclick="executeCode()">执行</button>
								<button onclick="clearConsole()">清除</button>
							</div>
						</div>
					</div>
				` + "`" + `;
				document.body.appendChild(div);
				setupDragAndDrop();
				setupConsole();
			}
		}

		// 添加新���样式
		const extraStyles = document.createElement('style');
		extraStyles.textContent = ` + "`" + `
			.button-group {
				display: grid;
				grid-template-columns: 1fr 1fr;
				gap: 5px;
				margin-bottom: 10px;
			}
			.console-container {
				background: #1e1e1e;
				border-radius: 4px;
				margin-top: 10px;
				overflow: hidden;
			}
			.console-output {
				height: 150px;
				overflow-y: auto;
				padding: 8px;
				font-family: monospace;
				font-size: 12px;
				color: #fff;
			}
			.console-output .log { color: #fff; }
			.console-output .error { color: #ff6b6b; }
			.console-output .warn { color: #ffd93d; }
			.console-output .info { color: #4dabf7; }
			.console-input-group {
				padding: 8px;
				background: #2d2d2d;
				display: flex;
				flex-direction: column;
				gap: 5px;
			}
			#codeInput {
				background: #1e1e1e;
				color: #fff;
				border: 1px solid #3d3d3d;
				padding: 8px;
				font-family: monospace;
				resize: vertical;
			}
			.console-input-group button {
				background: #0d6efd;
				margin: 2px;
			}
		` + "`" + `;
		document.head.appendChild(extraStyles);

		// 控制台功能设置
		function setupConsole() {
			// 重写控制台方法
			const consoleOutput = document.getElementById('consoleOutput');
			const originalConsole = {
				log: console.log,
				error: console.error,
				warn: console.warn,
				info: console.info
			};

			function logToPanel(type, ...args) {
				const entry = document.createElement('div');
				entry.className = type;
				entry.textContent = args.map(arg => 
					typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
				).join(' ');
				consoleOutput.appendChild(entry);
				consoleOutput.scrollTop = consoleOutput.scrollHeight;
				
				// 同时调用原始方法
				originalConsole[type].apply(console, args);
			}

			console.log = (...args) => logToPanel('log', ...args);
			console.error = (...args) => logToPanel('error', ...args);
			console.warn = (...args) => logToPanel('warn', ...args);
			console.info = (...args) => logToPanel('info', ...args);
		}

		// 切换控制台显示
		function toggleConsole() {
			const console = document.getElementById('controlConsole');
			const arrow = event.target;
			if (console.style.display === 'none') {
				console.style.display = 'block';
				arrow.textContent = '🔼';
			} else {
				console.style.display = 'none';
				arrow.textContent = '🔽';
			}
		}

		// 执行代码
		function executeCode() {
			const input = document.getElementById('codeInput');
			const code = input.value.trim();
			if (!code) return;

			console.info('执行代码:', code);
			try {
				const result = eval(code);
				if (result !== undefined) {
					console.log('返回值:', result);
				}
			} catch (err) {
				console.error('执行错误:', err.message);
			}
		}

		// 清除控制台
		function clearConsole() {
			document.getElementById('consoleOutput').innerHTML = '';
		}

		// 添加快捷键
		document.addEventListener('keydown', function(e) {
			if (e.ctrlKey && e.key === 'Enter' && document.activeElement === document.getElementById('codeInput')) {
				executeCode();
				e.preventDefault();
			}
		});

		// 初始化
		if (document.readyState === 'loading') {
			document.addEventListener('DOMContentLoaded', injectControlPanel);
		} else {
			injectControlPanel();
		}
	`)
}

// 其他辅助函数...
func handlePerformanceData(data json.RawMessage, state *AppState) {
	var perfData map[string]interface{}
	if err := json.Unmarshal(data, &perfData); err != nil {
		log.Printf("解析性能数据失败: %v", err)
		return
	}

	state.mu.Lock()
	state.perfStats = perfData
	state.mu.Unlock()

	log.Printf("性能数据: %+v", perfData)
}

func handleSecurityData(data json.RawMessage) {
	var securityData map[string]interface{}
	if err := json.Unmarshal(data, &securityData); err != nil {
		log.Printf("解析安全数据失败: %v", err)
		return
	}

	log.Printf("安全检查结果: %+v", securityData)
}

func handleHTML5TestData(data json.RawMessage) {
	var testData map[string]interface{}
	if err := json.Unmarshal(data, &testData); err != nil {
		log.Printf("解析HTML5测试数据失败: %v", err)
		return
	}

	log.Printf("HTML5测试结果: %+v", testData)
}

// createWSTestWindow 创建WebSocket测试窗口
func createWSTestWindow(mainWin webview2.WebView) *WSTestWindow {
	w := webview2.New(true)
	if w == nil {
		log.Println("创建WebSocket测试窗口失败")
		return nil
	}

	w.SetTitle(wsWindowTitle)
	w.SetSize(wsWindowWidth, wsWindowHeight, webview2.HintNone)

	// 直接设置完整的HTML内容
	w.SetHtml(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>WebSocket测试控制台</title>
			<style>
				body { 
					margin: 0; 
					padding: 20px; 
					font-family: Arial, sans-serif;
					background: #f5f5f5;
				}
				.container {
					display: grid;
					grid-template-columns: 1fr 1fr;
					gap: 20px;
					height: calc(100vh - 40px);
				}
				.panel {
					background: white;
					border-radius: 8px;
					padding: 15px;
					box-shadow: 0 2px 4px rgba(0,0,0,0.1);
					display: flex;
					flex-direction: column;
				}
				.panel-header {
					font-size: 16px;
					font-weight: bold;
					margin-bottom: 10px;
					padding-bottom: 10px;
					border-bottom: 1px solid #eee;
				}
				.message-list {
					flex: 1;
					overflow-y: auto;
					background: #f8f9fa;
					border-radius: 4px;
					padding: 10px;
					margin-bottom: 10px;
				}
				.message {
					margin: 5px 0;
					padding: 8px;
					border-radius: 4px;
					background: white;
					border-left: 4px solid #007bff;
					box-shadow: 0 1px 2px rgba(0,0,0,0.05);
				}
				.message.sent { border-left-color: #28a745; }
				.message.error { border-left-color: #dc3545; }
				.controls {
					display: flex;
					flex-direction: column;
					gap: 10px;
				}
				button {
					padding: 8px 16px;
					border: none;
					border-radius: 4px;
					background: #007bff;
					color: white;
					cursor: pointer;
					transition: background 0.2s;
				}
				button:hover { background: #0056b3; }
				textarea {
					width: 100%;
					height: 100px;
					padding: 8px;
					border: 1px solid #ddd;
					border-radius: 4px;
					resize: vertical;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="panel">
					<div class="panel-header">消息记录</div>
					<div id="messageList" class="message-list"></div>
					<div class="controls">
						<textarea id="messageInput" placeholder="输入要发送的消息..."></textarea>
						<button onclick="sendMessage()">发送消息</button>
						<button onclick="clearMessages()">清除记录</button>
					</div>
				</div>
				<div class="panel">
					<div class="panel-header">脚本执行</div>
					<div id="scriptOutput" class="message-list"></div>
					<div class="controls">
						<textarea id="scriptInput" placeholder="输入要执行的JavaScript代码..."></textarea>
						<button onclick="executeScript()">执行脚本</button>
						<button onclick="clearScriptOutput()">清除输出</button>
					</div>
				</div>
			</div>
			<script>
				// WebSocket初始化
				window._webSocket = new WebSocket('ws://localhost:8080/ws');
				window._webSocket.onopen = () => addMessage('WebSocket已连接', 'sent');
				window._webSocket.onclose = () => addMessage('WebSocket已断开', 'error');
				window._webSocket.onerror = (e) => console.error('WebSocket错误:', e);
				window._webSocket.onmessage = (event) => {
					try {
						const data = JSON.parse(event.data);
						addMessage('收到: ' + JSON.stringify(data, null, 2));
					} catch(e) {
						addMessage('收到: ' + event.data);
					}
				};

				// 工具函数
				function addMessage(text, type = '') {
					const messageList = document.getElementById('messageList');
					const message = document.createElement('div');
					message.className = 'message ' + type;
					message.textContent = text;
					messageList.appendChild(message);
					messageList.scrollTop = messageList.scrollHeight;
				}

				function addScriptOutput(text, type = '') {
					const outputList = document.getElementById('scriptOutput');
					const output = document.createElement('div');
					output.className = 'message ' + type;
					output.textContent = text;
					outputList.appendChild(output);
					outputList.scrollTop = outputList.scrollHeight;
				}

				// 交互功能
				function sendMessage() {
					const input = document.getElementById('messageInput');
					const message = input.value.trim();
					if (message && window._webSocket && window._webSocket.readyState === 1) {
						try {
							window._webSocket.send(JSON.stringify({
								type: 'console',
								data: message
							}));
							addMessage('发送: ' + message, 'sent');
							input.value = '';
						} catch(e) {
							addMessage('发送失败: ' + e.message, 'error');
						}
					} else {
						addMessage('WebSocket未连接或消息为空', 'error');
					}
				}

				function executeScript() {
					const input = document.getElementById('scriptInput');
					const script = input.value.trim();
					if (script) {
						try {
							window.executeInMain(script);
							addScriptOutput('执行脚本: ' + script);
							input.value = '';
						} catch (e) {
							addScriptOutput('执行错误: ' + e.message, 'error');
						}
					}
				}

				function clearMessages() {
					document.getElementById('messageList').innerHTML = '';
				}

				function clearScriptOutput() {
					document.getElementById('scriptOutput').innerHTML = '';
				}

				// 添加快捷键
				document.addEventListener('keydown', function(e) {
					if (e.ctrlKey && e.key === 'Enter') {
						if (document.activeElement === document.getElementById('messageInput')) {
							sendMessage();
						} else if (document.activeElement === document.getElementById('scriptInput')) {
							executeScript();
						}
					}
				});
			</script>
		</body>
		</html>
	`)

	// 绑定脚本执行函数
	w.Bind("executeInMain", func(script string) {
		mainWin.Dispatch(func() {
			log.Println("执行脚本:", script)
			mainWin.Eval(script)
		})
	})

	return &WSTestWindow{w: w, mainWin: mainWin}
}
