//go:build windows
// +build windows
package webview2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/yuaotian/go-win-webview2/internal/w32"
	"github.com/gorilla/websocket"
)

// 错误定义
var (
	ErrInvalidFunction   = errors.New("only functions can be bound")
	ErrTooManyReturns    = errors.New("function may only return a value or a value+error")
	ErrInvalidReturnType = errors.New("second return value must be an error")
)

// 在错误定义之后添加
// HotKeyHandler 是热键处理函数的类型
type HotKeyHandler func()

// 提示用于配置窗口大小和调整大小行为。
type Hint int

const (
	//HintNone 指定宽度和高度为默认尺寸
	HintNone Hint = iota

	//HintFixed 指定用户无法更改窗口大小
	HintFixed

	//HintMin 指定宽度和高度为最小边界
	HintMin

	//HintMax 指定宽度和高度为最大边界
	HintMax
)

// WebView是webview的接口。
type WebView interface {
	// 上下文管理相关方法
	Context() context.Context
	WithContext(ctx context.Context) WebView

	// 窗口控制
	Run()
	// 运行
	RunAsync()
	// 终止
	Terminate()
	// 调度函数
	Dispatch(f func())
	// 调度函数
	DispatchAsync(f func())
	// 销毁
	Destroy()
	// 获取窗口
	Window() unsafe.Pointer
	// 设置标题
	SetTitle(title string)
	// 设置大小
	SetSize(w int, h int, hint Hint)
	// 导航
	Navigate(url string)
	// 设置HTML
	SetHtml(html string)
	//初始化(加载之前注入js，永久注入)
	Init(js string)
	// 执行JS（加载之后注入js，临时注入）
	Eval(js string)
	// 绑定函数
	Bind(name string, f interface{}) error

	// 热键相关
	RegisterHotKey(modifiers int, keyCode int, handler HotKeyHandler) error
	// 注销热键
	UnregisterHotKey(modifiers int, keyCode int)
	// 注册热键字符串
	RegisterHotKeyString(hotkey string, handler HotKeyHandler) error

	// 窗口状态
	// 设置全屏
	SetFullscreen(enable bool)
	// 设置置顶
	SetAlwaysOnTop(enable bool)
	// 设置最小化
	SetMinimized(enable bool)
	// 设置最大化
	SetMaximized(enable bool)

	// 新增方法
	Minimize()                  // 最小化窗口
	Maximize()                  // 最大化窗口
	Restore()                   // 还原窗口
	Center()                    // 居中窗口
	SetOpacity(opacity float64) // 设置窗口透明度 (0.0-1.0)

	// 浏览器功能
	Reload()       // 刷新页面
	Back()         // 后退
	Forward()      // 前进
	Stop()         // 停止加载
	ClearCache()   // 清除缓存
	ClearCookies() // 清除 cookies

	// 开发工具
	OpenDevTools()  // 打开开发者工具
	CloseDevTools() // 关闭开发者工具

	// 状态监听
	OnLoadingStateChanged(func(isLoading bool))  // 加载状态变化
	OnURLChanged(func(url string))               // URL 变化
	OnTitleChanged(func(title string))           // 标题变化
	OnFullscreenChanged(func(isFullscreen bool)) // 全屏状态变化

	// 打印相关方法
	Print()                    // 直接打印
	PrintToPDF(path string)    // 打印到 PDF 文件
	ShowPrintDialog()          // 显示打印对话框

	// 右键菜单控制
	DisableContextMenu()    // 禁用右键菜单
	EnableContextMenu()     // 启用右键菜单

	// JavaScript Hook 相关方法
	AddJSHook(hook JSHook)           // 添加 JS Hook
	RemoveJSHook(hook JSHook)        // 移除 JS Hook
	ClearJSHooks()                   // 清除所有 JS Hook

	// WebSocket 相关方法
	EnableWebSocket(port int) error              // 启用 WebSocket 服务
	DisableWebSocket()                           // 禁用 WebSocket 服务
	OnWebSocketMessage(handler WebSocketHandler) // 设置 WebSocket 消息处理器
	SendWebSocketMessage(message string)         // 发送 WebSocket 消息
}

// HotKey 表示一个键组合
type HotKey struct {
	Modifiers int // 修饰符
	KeyCode   int // 键码
}

// ParseHotKey 将热键字符串解析为 HotKey 结构
// 例如: "Ctrl+Alt+Q" -> HotKey{MOD_CONTROL|MOD_ALT, 'Q'}
func ParseHotKey(s string) (HotKey, error) {
	parts := strings.Split(s, "+")
	if len(parts) < 2 {
		return HotKey{}, errors.New("invalid hotkey format")
	}

	var modifiers int
	key := strings.ToUpper(parts[len(parts)-1])

	// 解析修饰符
	for _, mod := range parts[:len(parts)-1] {
		switch strings.ToLower(strings.TrimSpace(mod)) {
		case "ctrl":
			modifiers |= w32.MOD_CONTROL
		case "alt":
			modifiers |= w32.MOD_ALT
		case "shift":
			modifiers |= w32.MOD_SHIFT
		case "win":
			modifiers |= w32.MOD_WIN
		default:
			return HotKey{}, fmt.Errorf("unknown modifier: %s", mod)
		}
	}

	// 解析键码
	var keyCode int
	if len(key) == 1 {
		// 字母和数字键
		keyCode = int(key[0])
	} else {
		// 特殊键
		switch key {
		case "f1":
			keyCode = w32.VK_F1
		case "f2":
			keyCode = w32.VK_F2
		case "f3":
			keyCode = w32.VK_F3
		case "f4":
			keyCode = w32.VK_F4
		case "f5":
			keyCode = w32.VK_F5
		case "f6":
			keyCode = w32.VK_F6
		case "f7":
			keyCode = w32.VK_F7
		case "f8":
			keyCode = w32.VK_F8
		case "f9":
			keyCode = w32.VK_F9
		case "f10":
			keyCode = w32.VK_F10
		case "f11":
			keyCode = w32.VK_F11
		case "f12":
			keyCode = w32.VK_F12
		case "esc":
			keyCode = w32.VK_ESCAPE
		case "tab":
			keyCode = w32.VK_TAB
		case "space":
			keyCode = w32.VK_SPACE
		default:
			return HotKey{}, fmt.Errorf("unknown key: %s", key)
		}
	}

	return HotKey{Modifiers: modifiers, KeyCode: keyCode}, nil
}

// JSHookType 定义 Hook 的类型
type JSHookType int

const (
	JSHookBefore JSHookType = iota  // JS 执行前
	JSHookAfter                     // JS 执行后
)

// JSHook 定义 JavaScript 钩子接口
type JSHook interface {
	Type() JSHookType              // 获取 Hook 类型
	Handle(script string) string   // 处理脚本
	Priority() int                 // Hook 优先级，数字越小优先级越高
}

// BaseJSHook 提供基本的 JSHook 实现
type BaseJSHook struct {
	HookType     JSHookType
	Handler      func(script string) string
	HookPriority int
}

func (h *BaseJSHook) Type() JSHookType {
	return h.HookType
}

func (h *BaseJSHook) Handle(script string) string {
	if h.Handler != nil {
		return h.Handler(script)
	}
	return script
}

func (h *BaseJSHook) Priority() int {
	return h.HookPriority
}

// WebSocketHandler 定义 WebSocket 消息处理函数
type WebSocketHandler func(message string)

// WebSocketHook 定义 WebSocket 相关的 JSHook
type WebSocketHook struct {
	BaseJSHook
	wsConn *websocket.Conn
}

// NewWebSocketHook 创建新的 WebSocket Hook
func NewWebSocketHook(conn *websocket.Conn) *WebSocketHook {
	return &WebSocketHook{
		BaseJSHook: BaseJSHook{
			HookType:     JSHookBefore,
			HookPriority: 0,
		},
		wsConn: conn,
	}
}
