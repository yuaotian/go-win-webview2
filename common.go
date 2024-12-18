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
	Terminate()
	// 调度函数
	Dispatch(f func())
	// 调度函数
	DispatchAsync(f func())
	Destroy()
	Window() unsafe.Pointer
	SetTitle(title string)
	SetSize(w int, h int, hint Hint)
	Navigate(url string)
	SetHtml(html string)
	Init(js string)
	Eval(js string)
	Bind(name string, f interface{}) error

	// 热键相关
	RegisterHotKey(modifiers int, keyCode int, handler HotKeyHandler) error
	UnregisterHotKey(modifiers int, keyCode int)
	RegisterHotKeyString(hotkey string, handler HotKeyHandler) error

	// 窗口状态
	SetFullscreen(enable bool)
	SetAlwaysOnTop(enable bool)

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
}

// HotKey 表示一个热键组合
type HotKey struct {
	Modifiers int
	KeyCode   int
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
