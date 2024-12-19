//go:build windows
// +build windows

package webview2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/yuaotian/go-win-webview2/internal/w32"
	"github.com/yuaotian/go-win-webview2/pkg/edge"

	"golang.org/x/sys/windows"
)

var (
	windowContext = sync.Map{}
)

func getWindowContext(wnd uintptr) interface{} {
	if v, ok := windowContext.Load(wnd); ok {
		return v
	}
	return nil
}

func setWindowContext(wnd uintptr, data interface{}) {
	windowContext.Store(wnd, data)
}

type browser interface {
	Embed(hwnd uintptr) bool
	Resize()
	Navigate(url string)
	NavigateToString(htmlContent string)
	Init(script string)
	Eval(script string)
	NotifyParentWindowPositionChanged() error
	Focus()
	PrintToPDF(path string) error
	DisableContextMenu() error
	EnableContextMenu() error
	GetSettings() (*edge.ICoreWebViewSettings, error)
}

type webview struct {
	hwnd       uintptr
	mainthread uintptr
	browser    browser
	autofocus  bool
	maxsz      w32.Point
	minsz      w32.Point
	m          sync.Mutex
	bindings   map[string]interface{}
	dispatchq  []func()
	ctx        context.Context
	hotkeys    map[int]HotKeyHandler
	jsHooks    []JSHook // JavaScript hooks

	// 状态听回调
	onLoadingStateChanged func(bool)
	onURLChanged          func(string)
	onTitleChanged        func(string)
	onFullscreenChanged   func(bool)

	wsServer      *http.Server
	wsUpgrader    websocket.Upgrader
	wsHandler     WebSocketHandler
	wsConnections sync.Map

	// 用于处理导航的通道
	navigationChan chan string

	// 消息回调
	messageCallback func(string)
}

type WindowOptions struct {
	Title              string  // 窗口标题
	Width              uint    // 窗口宽度
	Height             uint    // 窗口高度
	IconId             uint    // 图标ID
	Center             bool    // 是否居中
	Frameless          bool    // 是否无边框
	Fullscreen         bool    // 是否全屏
	AlwaysOnTop        bool    // 是否置顶
	Resizable          bool    // 是否可调整大小
	Minimizable        bool    // 是否可最小化
	Maximizable        bool    // 是否可最大化
	Minimized          bool    // 初始是否最小化
	Maximized          bool    // 初始是否最大化
	DisableContextMenu bool    // 是否禁用右键���单
	EnableDragAndDrop  bool    // 是否启用拖放
	HideWindowOnClose  bool    // 关闭时是否隐藏窗口而不是退出
	DefaultBackground  string  // 默认背景色 (CSS 格式，如 "#FFFFFF")
	Opacity            float64 // 初始透明度 (0.0-1.0)
	IconPath           string  // 图标文件路径
	IconData           []byte  // 图标二进制数据
}

// 添加默认配置
func DefaultWindowOptions() WindowOptions {
	return WindowOptions{
		Title:              "WebView2",
		Width:              800,
		Height:             600,
		Center:             true,
		Resizable:          true,
		Minimizable:        true,
		Maximizable:        true,
		Minimized:          false,
		Maximized:          false,
		DisableContextMenu: false,
		EnableDragAndDrop:  true,
		HideWindowOnClose:  false,
		DefaultBackground:  "#FFFFFF",
		Opacity:            1.0,
	}
}

type WebViewOptions struct {
	Window unsafe.Pointer
	Debug  bool

	//DataPath 指定 WebView2 运行时使用的数据路径
	//浏览器实例。
	DataPath string

	//窗口打开时，AutoFocus 将尝试保持 WebView2 小部件聚焦
	//已聚焦。
	AutoFocus bool

	//WindowOptions 自定义创建的窗口以嵌入
	//WebView2 小部件。
	WindowOptions WindowOptions
}

// New 在新窗口中创建个新的 webview。
func New(debug bool) WebView {
	return NewWithOptions(WebViewOptions{Debug: debug})

}

// NewWindow 使用现有窗创一个新的 webview。
//
// 已弃用：使用 NewWithOptions。
func NewWindow(debug bool, window unsafe.Pointer) WebView {
	return NewWithOptions(WebViewOptions{Debug: debug, Window: window})
}

// NewWithOptions 使用提供的选项创建一个的 webview。
func NewWithOptions(options WebViewOptions) WebView {
	// 合并默认选项
	defaultOpts := DefaultWindowOptions()
	if options.WindowOptions.Title == "" {
		options.WindowOptions.Title = defaultOpts.Title
	}
	if options.WindowOptions.Width == 0 {
		options.WindowOptions.Width = defaultOpts.Width
	}
	if options.WindowOptions.Height == 0 {
		options.WindowOptions.Height = defaultOpts.Height
	}
	// 默认居中显示
	if !options.WindowOptions.Center {
		options.WindowOptions.Center = defaultOpts.Center
	}
	// 合并其他默认选项
	if !options.WindowOptions.Resizable {
		options.WindowOptions.Resizable = defaultOpts.Resizable
	}
	if !options.WindowOptions.Minimizable {
		options.WindowOptions.Minimizable = defaultOpts.Minimizable
	}
	if !options.WindowOptions.Maximizable {
		options.WindowOptions.Maximizable = defaultOpts.Maximizable
	}
	if options.WindowOptions.DefaultBackground == "" {
		options.WindowOptions.DefaultBackground = defaultOpts.DefaultBackground
	}

	w := &webview{
		ctx:     context.Background(),
		hotkeys: make(map[int]HotKeyHandler),
	}
	w.bindings = map[string]interface{}{}
	w.autofocus = options.AutoFocus

	chromium := edge.NewChromium()
	chromium.MessageCallback = w.msgcb
	chromium.DataPath = options.DataPath
	chromium.SetPermission(edge.CoreWebView2PermissionKindClipboardRead, edge.CoreWebView2PermissionStateAllow)

	w.browser = chromium
	w.mainthread, _, _ = w32.Kernel32GetCurrentThreadID.Call()
	if !w.CreateWithOptions(options.WindowOptions) {
		return nil
	}

	settings, err := chromium.GetSettings()
	if err != nil {
		log.Printf("Warning: Failed to get settings: %v", err)
		return nil
	}

	
	//禁用上下文菜单
	err = settings.PutAreDefaultContextMenusEnabled(options.Debug)
	if err != nil {
		log.Fatal(err)
	}
	//禁用开者工具
	err = settings.PutAreDevToolsEnabled(options.Debug)
	if err != nil {
		log.Fatal(err)
	}

	// 设置默认消息处理
	w.SetMessageCallback(func(msg string) {
		if chromium, ok := w.browser.(*edge.Chromium); ok {
			chromium.HandleWebMessage(msg)
		}
	})

	return w
}

type rpcMessage struct {
	ID     int               `json:"id"`
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

func jsString(v interface{}) string { b, _ := json.Marshal(v); return string(b) }

func (w *webview) msgcb(msg string) {
	d := rpcMessage{}
	if err := json.Unmarshal([]byte(msg), &d); err != nil {
		log.Printf("Error unmarshaling RPC message: %v", err)
		return
	}

	id := strconv.Itoa(d.ID)
	rejectScript := fmt.Sprintf("window._rpc[%s].reject", id)
	resolveScript := fmt.Sprintf("window._rpc[%s].resolve", id)
	cleanupScript := fmt.Sprintf("window._rpc[%s] = undefined", id)

	if res, err := w.callbinding(d); err != nil {
		w.Dispatch(func() {
			w.Eval(fmt.Sprintf("%s(%s); %s", rejectScript, jsString(err.Error()), cleanupScript))
		})
	} else if b, err := json.Marshal(res); err != nil {
		w.Dispatch(func() {
			w.Eval(fmt.Sprintf("%s(%s); %s", rejectScript, jsString(err.Error()), cleanupScript))
		})
	} else {
		w.Dispatch(func() {
			w.Eval(fmt.Sprintf("%s(%s); %s", resolveScript, string(b), cleanupScript))
		})
	}
}

func (w *webview) callbinding(d rpcMessage) (interface{}, error) {
	w.m.Lock()
	f, ok := w.bindings[d.Method]
	w.m.Unlock()
	if !ok {
		return nil, nil
	}

	v := reflect.ValueOf(f)
	isVariadic := v.Type().IsVariadic()
	numIn := v.Type().NumIn()
	if (isVariadic && len(d.Params) < numIn-1) || (!isVariadic && len(d.Params) != numIn) {
		return nil, errors.New("function arguments mismatch")
	}
	args := []reflect.Value{}
	for i := range d.Params {
		var arg reflect.Value
		if isVariadic && i >= numIn-1 {
			arg = reflect.New(v.Type().In(numIn - 1).Elem())
		} else {
			arg = reflect.New(v.Type().In(i))
		}
		if err := json.Unmarshal(d.Params[i], arg.Interface()); err != nil {
			return nil, err
		}
		args = append(args, arg.Elem())
	}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	res := v.Call(args)
	switch len(res) {
	case 0:
		// No results from the function, just return nil
		return nil, nil

	case 1:
		// One result may be a value, or an error
		if res[0].Type().Implements(errorType) {
			if res[0].Interface() != nil {
				return nil, res[0].Interface().(error)
			}
			return nil, nil
		}
		return res[0].Interface(), nil

	case 2:
		// Two results: first one is value, second is error
		if !res[1].Type().Implements(errorType) {
			return nil, errors.New("second return value must be an error")
		}
		if res[1].Interface() == nil {
			return res[0].Interface(), nil
		}
		return res[0].Interface(), res[1].Interface().(error)

	default:
		return nil, errors.New("unexpected number of return values")
	}
}

func wndproc(hwnd, msg, wp, lp uintptr) uintptr {
	if w, ok := getWindowContext(hwnd).(*webview); ok {
		switch msg {
		case w32.WMMove, w32.WMMoving:
			_ = w.browser.NotifyParentWindowPositionChanged()
		case w32.WMNCLButtonDown:
			if wp == w32.HTCaption {
				// 直接调用 DefWindowProc 处理拖动
				r, _, _ := w32.User32DefWindowProcW.Call(hwnd, msg, wp, lp)
				return r
			}
			_, _, _ = w32.User32SetFocus.Call(w.hwnd)
			r, _, _ := w32.User32DefWindowProcW.Call(hwnd, msg, wp, lp)
			return r
		case w32.WMSize:
			w.browser.Resize()
		case w32.WMActivate:
			if wp == w32.WAInactive {
				break
			}
			if w.autofocus {
				w.browser.Focus()
			}
		case w32.WMClose:
			_, _, _ = w32.User32DestroyWindow.Call(hwnd)
		case w32.WMDestroy:
			w.Terminate()
		case w32.WMGetMinMaxInfo:
			lpmmi := (*w32.MinMaxInfo)(unsafe.Pointer(lp))
			if w.maxsz.X > 0 && w.maxsz.Y > 0 {
				lpmmi.PtMaxSize = w.maxsz
				lpmmi.PtMaxTrackSize = w.maxsz
			}
			if w.minsz.X > 0 && w.minsz.Y > 0 {
				lpmmi.PtMinTrackSize = w.minsz
			}
		case w32.WMNCHitTest:
			// 获取鼠标位置
			x := int32(lp & 0xffff)
			y := int32((lp >> 16) & 0xffff)

			// 获取窗口位置
			var rect w32.Rect
			_, _, _ = w32.User32GetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))

			// 定义边框拖拽区宽度
			const borderWidth = 5

			// 检查是否在拖拽域内
			if y >= rect.Top && y <= rect.Top+borderWidth {
				if x >= rect.Left && x <= rect.Left+borderWidth {
					return w32.HTTopLeft
				}
				if x >= rect.Right-borderWidth && x <= rect.Right {
					return w32.HTTopRight
				}
				return w32.HTTop
			}
			if y >= rect.Bottom-borderWidth && y <= rect.Bottom {
				if x >= rect.Left && x <= rect.Left+borderWidth {
					return w32.HTBottomLeft
				}
				if x >= rect.Right-borderWidth && x <= rect.Right {
					return w32.HTBottomRight
				}
				return w32.HTBottom
			}
			if x >= rect.Left && x <= rect.Left+borderWidth {
				return w32.HTLeft
			}
			if x >= rect.Right-borderWidth && x <= rect.Right {
				return w32.HTRight
			}

			// 允许窗口
			return w32.HTCaption

		case w32.WMLButtonDown:
			if wp == w32.HTCaption {
				_, _, _ = w32.User32SendMessageW.Call(hwnd, w32.WMNCLButtonDown, wp, lp)
				return 0
			}
		case w32.WMHotKey:
			w.m.Lock()
			if handler, ok := w.hotkeys[int(wp)]; ok {
				w.m.Unlock()
				handler()
			} else {
				w.m.Unlock()
			}
			return 0
		default:
			r, _, _ := w32.User32DefWindowProcW.Call(hwnd, msg, wp, lp)
			return r
		}
		return 0
	}
	r, _, _ := w32.User32DefWindowProcW.Call(hwnd, msg, wp, lp)
	return r
}

func (w *webview) Create(debug bool, window unsafe.Pointer) bool {
	// This function signature stopped making sense a long time ago.
	// It is but legacy cruft at this point.
	return w.CreateWithOptions(WindowOptions{})
}

func (w *webview) CreateWithOptions(opts WindowOptions) bool {
	var hinstance windows.Handle
	if err := windows.GetModuleHandleEx(0, nil, &hinstance); err != nil {
		log.Printf("Error getting module handle: %v", err)
		return false
	}

	icon := w.loadWindowIcon(hinstance, opts.IconId, opts)
	if icon == 0 {
		log.Printf("Warning: Failed to load window icon")
	}

	className, _ := windows.UTF16PtrFromString("webview")
	wc := w32.WndClassExW{
		CbSize:        uint32(unsafe.Sizeof(w32.WndClassExW{})),
		HInstance:     hinstance,
		LpszClassName: className,
		HIcon:         windows.Handle(icon),
		HIconSm:       windows.Handle(icon),
		LpfnWndProc:   windows.NewCallback(wndproc),
	}
	_, _, _ = w32.User32RegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	windowName, _ := windows.UTF16PtrFromString(opts.Title)

	windowWidth := opts.Width
	if windowWidth == 0 {
		windowWidth = 640
	}
	windowHeight := opts.Height
	if windowHeight == 0 {
		windowHeight = 480
	}

	var posX, posY uint
	if opts.Center {
		// get screen size
		screenWidth, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CXSCREEN)
		screenHeight, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CYSCREEN)
		// calculate window position
		posX = (uint(screenWidth) - windowWidth) / 2
		posY = (uint(screenHeight) - windowHeight) / 2
	} else {
		// use default position
		posX = w32.CW_USEDEFAULT
		posY = w32.CW_USEDEFAULT
	}

	// 修改窗口样式设置
	var style uint32 = w32.WSOverlappedWindow
	if opts.Frameless {
		style = w32.WSPopup | w32.WSVisible
	} else {
		// 根据选项调整窗样式
		if !opts.Maximizable {
			style &^= w32.WSMaximizeBox
		}
		if !opts.Minimizable {
			style &^= w32.WSMinimizeBox
		}
		if !opts.Resizable {
			style &^= w32.WSThickFrame
		}
	}

	// 添加分层窗口扩展样式
	var exStyle uint32 = w32.WS_EX_LAYERED

	w.hwnd, _, _ = w32.User32CreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style),
		uintptr(posX),
		uintptr(posY),
		uintptr(windowWidth),
		uintptr(windowHeight),
		0,
		0,
		uintptr(hinstance),
		0,
	)

	// 设置初始透明度(默认完全不透明)
	_, _, _ = w32.User32SetLayeredWindowAttributes.Call(
		w.hwnd,
		0,
		255, // 完全不透明
		w32.LWA_ALPHA,
	)

	setWindowContext(w.hwnd, w)

	_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_SHOW)
	_, _, _ = w32.User32UpdateWindow.Call(w.hwnd)
	_, _, _ = w32.User32SetFocus.Call(w.hwnd)

	if !w.browser.Embed(w.hwnd) {
		return false
	}
	w.browser.Resize()

	// 创建窗口后应用全屏和置顶设置
	if opts.Fullscreen {
		w.SetFullscreen(true)
	}

	if opts.AlwaysOnTop {
		w.SetAlwaysOnTop(true)
	}

	// 应用其他选项
	if opts.DisableContextMenu {
		w.DisableContextMenu()
	}

	// 设置默认背色
	if opts.DefaultBackground != "" {
		w.Eval(fmt.Sprintf(`
			document.documentElement.style.background = '%s';
			document.body.style.background = '%s';
		`, opts.DefaultBackground, opts.DefaultBackground))
	}

	// 处理窗口关闭行为
	if opts.HideWindowOnClose {
		w.Bind("__handleWindowClose", func() {
			w.Dispatch(func() {
				_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_HIDE)
			})
		})
		w.Init(`
			window.onbeforeunload = function(e) {
				window.__handleWindowClose();
				e.preventDefault();
				return false;
			};
		`)
	}

	// 设置初始窗口状态
	if opts.Maximizable && opts.Maximized {
		w.Maximize()
	} else if opts.Minimizable && opts.Minimized {
		w.Minimize()
	}

	return true
}

func (w *webview) Destroy() {
	// 注所有热键
	w.m.Lock()
	for id := range w.hotkeys {
		_, _, _ = w32.User32UnregisterHotKey.Call(w.hwnd, uintptr(id))
	}
	w.hotkeys = nil
	w.m.Unlock()

	// 清理资源
	w.m.Lock()
	w.bindings = nil
	w.dispatchq = nil
	w.m.Unlock()

	// 从 windowContext 中移除
	windowContext.Delete(w.hwnd)

	// 发送关闭消息
	_, _, _ = w32.User32PostMessageW.Call(w.hwnd, w32.WMClose, 0, 0)
}

func (w *webview) Run() {
	var msg w32.Msg
	for {
		_, _, _ = w32.User32GetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0,
			0,
			0,
		)
		if msg.Message == w32.WMApp {
			w.m.Lock()
			q := append([]func(){}, w.dispatchq...)
			w.dispatchq = []func(){}
			w.m.Unlock()
			for _, v := range q {
				v()
			}
		} else if msg.Message == w32.WMQuit {
			return
		}
		r, _, _ := w32.User32GetAncestor.Call(uintptr(msg.Hwnd), w32.GARoot)
		r, _, _ = w32.User32IsDialogMessage.Call(r, uintptr(unsafe.Pointer(&msg)))
		if r != 0 {
			continue
		}
		_, _, _ = w32.User32TranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		_, _, _ = w32.User32DispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (w *webview) Terminate() {
	_, _, _ = w32.User32PostQuitMessage.Call(0)
}

func (w *webview) Window() unsafe.Pointer {
	return unsafe.Pointer(uintptr(w.hwnd))
}

func (w *webview) Navigate(url string) {
	w.browser.Navigate(url)
}

func (w *webview) SetHtml(html string) {
	w.browser.NavigateToString(html)
}

func (w *webview) SetTitle(title string) {
	_title, err := windows.UTF16FromString(title)
	if err != nil {
		_title, _ = windows.UTF16FromString("")
	}
	_, _, _ = w32.User32SetWindowTextW.Call(w.hwnd, uintptr(unsafe.Pointer(&_title[0])))
}

func (w *webview) SetSize(width int, height int, hints Hint) {
	index := w32.GWLStyle
	style, _, _ := w32.User32GetWindowLongPtrW.Call(w.hwnd, uintptr(index))
	if hints == HintFixed {
		styleUint32 := uint32(style)
		styleUint32 &^= (w32.WSThickFrame | w32.WSMaximizeBox)
		style = uintptr(styleUint32)
	} else {
		styleUint32 := uint32(style)
		styleUint32 |= (w32.WSThickFrame | w32.WSMaximizeBox)
		style = uintptr(styleUint32)
	}
	_, _, _ = w32.User32SetWindowLongPtrW.Call(w.hwnd, uintptr(index), style)

	if hints == HintMax {
		w.maxsz.X = int32(width)
		w.maxsz.Y = int32(height)
	} else if hints == HintMin {
		w.minsz.X = int32(width)
		w.minsz.Y = int32(height)
	} else {
		r := w32.Rect{}
		r.Left = 0
		r.Top = 0
		r.Right = int32(width)
		r.Bottom = int32(height)
		_, _, _ = w32.User32AdjustWindowRect.Call(
			uintptr(unsafe.Pointer(&r)),
			uintptr(w32.WSOverlappedWindow),
			0,
		)
		_, _, _ = w32.User32SetWindowPos.Call(
			w.hwnd, 0, uintptr(r.Left), uintptr(r.Top),
			uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top),
			w32.SWP_NOZORDER|w32.SWP_NOACTIVATE|w32.SWP_NOMOVE|w32.SWP_FRAMECHANGED)
		w.browser.Resize()
	}
}

// 初始化(加载之前注入js，永久注入)
func (w *webview) Init(js string) {
	// 添加 webview2 导航功能
	baseScript := `
		window.webview2 = {
			navigate: function(url) {
				window.chrome.webview.postMessage(JSON.stringify({
					type: 'navigate',
					url: url
				}));
			}
		};
	`

	// 合并脚本
	fullScript := baseScript
	if js != "" {
		fullScript += ";" + js
	}

	// 初始化浏览器
	w.browser.Init(fullScript)
}

// 执行JS(加载之后注入js，临时注入)
func (w *webview) Eval(js string) {
	// 应用前置钩子
	js = w.processScript(js, JSHookBefore)

	w.Dispatch(func() {
		w.browser.Eval(js)
	})

	// 应用后置钩子
	js = w.processScript(js, JSHookAfter)
}

func (w *webview) Dispatch(f func()) {
	w.m.Lock()
	w.dispatchq = append(w.dispatchq, f)
	w.m.Unlock()
	_, _, _ = w32.User32PostThreadMessageW.Call(w.mainthread, w32.WMApp, 0, 0)
}

func (w *webview) Bind(name string, f interface{}) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return errors.New("only functions can be bound")
	}
	if n := v.Type().NumOut(); n > 2 {
		return errors.New("function may only return a value or a value+error")
	}
	w.m.Lock()
	w.bindings[name] = f
	w.m.Unlock()

	w.Init("(function() { var name = " + jsString(name) + ";" + `
		var RPC = window._rpc = (window._rpc || {nextSeq: 1});
		window[name] = function() {
		  var seq = RPC.nextSeq++;
		  var promise = new Promise(function(resolve, reject) {
			RPC[seq] = {
			  resolve: resolve,
			  reject: reject,
			};
		  });
		  window.external.invoke(JSON.stringify({
			id: seq,
			method: name,
			params: Array.prototype.slice.call(arguments),
		  }));
		  return promise;
		}
	})()`)

	return nil
}

func (w *webview) loadWindowIcon(hinstance windows.Handle, iconId uint, opts WindowOptions) uintptr {
	// 1. 优先使用 IconData
	if len(opts.IconData) > 0 {
		icon, _, _ := w32.User32CreateIconFromResourceEx.Call(
			uintptr(unsafe.Pointer(&opts.IconData[0])),
			uintptr(len(opts.IconData)),
			1,          // IMAGE_ICON
			0x00030000, // 版本
			0, 0,
			w32.LR_DEFAULTSIZE,
		)
		if icon != 0 {
			return icon
		}
	}

	// 2. 其次使用 IconPath
	if opts.IconPath != "" {
		iconPath, _ := windows.UTF16PtrFromString(opts.IconPath)
		icon, _, _ := w32.User32LoadImageW.Call(
			0,
			uintptr(unsafe.Pointer(iconPath)),
			1, // IMAGE_ICON
			0, 0,
			w32.LR_LOADFROMFILE|w32.LR_DEFAULTSIZE,
		)
		if icon != 0 {
			return icon
		}
	}

	// 3. 再次使用 IconId
	if iconId != 0 {
		icon, _, _ := w32.User32LoadImageW.Call(
			uintptr(hinstance),
			uintptr(iconId),
			1,
			0,
			0,
			w32.LR_DEFAULTSIZE|w32.LR_SHARED,
		)
		if icon != 0 {
			return icon
		}
	}

	// 4. 最后使用默认标
	icow, _, _ := w32.User32GetSystemMetrics.Call(w32.SystemMetricsCxIcon)
	icoh, _, _ := w32.User32GetSystemMetrics.Call(w32.SystemMetricsCyIcon)
	icon, _, _ := w32.User32LoadImageW.Call(
		0,
		32512, // IDI_APPLICATION
		1,     // IMAGE_ICON
		icow,
		icoh,
		w32.LR_SHARED,
	)
	return icon
}

func (w *webview) Context() context.Context {
	if w.ctx == nil {
		w.ctx = context.Background()
	}
	return w.ctx
}

func (w *webview) WithContext(ctx context.Context) WebView {
	if ctx == nil {
		ctx = context.Background()
	}
	w.ctx = ctx
	return w
}

func (w *webview) RegisterHotKey(modifiers int, keyCode int, handler HotKeyHandler) error {
	w.m.Lock()
	defer w.m.Unlock()

	// 化键映
	if w.hotkeys == nil {
		w.hotkeys = make(map[int]HotKeyHandler)
	}

	// 生成唯一的热键ID
	hotkeyID := len(w.hotkeys) + 1

	// 注册热键
	ret, _, err := w32.User32RegisterHotKey.Call(
		w.hwnd,
		uintptr(hotkeyID),
		uintptr(modifiers),
		uintptr(keyCode),
	)

	if ret == 0 {
		return fmt.Errorf("failed to register hotkey: %v", err)
	}

	// 保存处理函数
	w.hotkeys[hotkeyID] = handler
	return nil
}

func (w *webview) UnregisterHotKey(modifiers int, keyCode int) {
	w.m.Lock()
	defer w.m.Unlock()

	// 查找对应的热键ID
	var hotkeyID int
	for id := range w.hotkeys {
		// 里简化处理，实际应该保modifiers和keyCode来确保全配
		hotkeyID = id
		break
	}

	if hotkeyID > 0 {
		_, _, _ = w32.User32UnregisterHotKey.Call(w.hwnd, uintptr(hotkeyID))
		delete(w.hotkeys, hotkeyID)
	}
}

// RegisterHotKeyString 通过字符串注册热键
// 例如: "Ctrl+Alt+Q"
func (w *webview) RegisterHotKeyString(hotkey string, handler HotKeyHandler) error {
	hk, err := ParseHotKey(hotkey)
	if err != nil {
		return fmt.Errorf("failed to parse hotkey: %v", err)
	}
	return w.RegisterHotKey(hk.Modifiers, hk.KeyCode, handler)
}

// SetFullscreen 设置窗口全屏状态
func (w *webview) SetFullscreen(enable bool) {
	if enable {
		var rect w32.Rect
		_, _, _ = w32.User32GetWindowRect.Call(w.hwnd, uintptr(unsafe.Pointer(&rect)))

		screenWidth, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CXSCREEN)
		screenHeight, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CYSCREEN)

		style, _, _ := w32.User32GetWindowLongPtrW.Call(w.hwnd, uintptr(w32.GWLStyle&0xFFFFFFFF))
		styleUint32 := uint32(style)
		styleUint32 &^= w32.WSOverlappedWindow
		style = uintptr(styleUint32)
		_, _, _ = w32.User32SetWindowLongPtrW.Call(w.hwnd, uintptr(w32.GWLStyle&0xFFFFFFFF), style)

		_, _, _ = w32.User32SetWindowPos.Call(
			w.hwnd,
			uintptr(w32.HWND_TOP),
			0,
			0,
			screenWidth,
			screenHeight,
			w32.SWP_FRAMECHANGED,
		)
	} else {
		style, _, _ := w32.User32GetWindowLongPtrW.Call(w.hwnd, uintptr(w32.GWLStyle&0xFFFFFFFF))
		styleUint32 := uint32(style)

		// 检查是否为无边框窗口
		isFrameless := (styleUint32 & w32.WSPopup) != 0

		if !isFrameless {
			styleUint32 |= w32.WSOverlappedWindow
		} else {
			styleUint32 = w32.WSPopup | w32.WSVisible
		}

		style = uintptr(styleUint32)
		_, _, _ = w32.User32SetWindowLongPtrW.Call(w.hwnd, uintptr(w32.GWLStyle&0xFFFFFFFF), style)

		// 获取屏幕尺寸
		screenWidth, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CXSCREEN)
		screenHeight, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CYSCREEN)

		// 设置默认窗口大小
		width := uintptr(1024)
		height := uintptr(768)

		// 计算居中置
		x := (screenWidth - width) / 2
		y := (screenHeight - height) / 2

		// 恢复窗口
		_, _, _ = w32.User32SetWindowPos.Call(
			w.hwnd,
			uintptr(w32.HWND_TOP),
			x,
			y,
			width,
			height,
			w32.SWP_FRAMECHANGED,
		)
	}
	w.browser.Resize()

	// 触发屏状态改变回调
	if w.onFullscreenChanged != nil {
		w.onFullscreenChanged(enable)
	}
}

// SetAlwaysOnTop 设置窗口置顶状
func (w *webview) SetAlwaysOnTop(enable bool) {
	flag := w32.HWND_NOTOPMOST
	if enable {
		flag = w32.HWND_TOPMOST
	}
	_, _, _ = w32.User32SetWindowPos.Call(
		w.hwnd,
		uintptr(flag),
		0, 0, 0, 0,
		w32.SWP_NOMOVE|w32.SWP_NOSIZE,
	)
}

// 最小化窗口
func (w *webview) Minimize() {
	w.Dispatch(func() {
		_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_MINIMIZE)
	})
}

// 最大化窗口
func (w *webview) Maximize() {
	w.Dispatch(func() {
		_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_MAXIMIZE)
	})
}

// 还原窗口
func (w *webview) Restore() {
	w.Dispatch(func() {
		_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_RESTORE)
	})
}

// 居中窗口
func (w *webview) Center() {
	w.Dispatch(func() {
		var rect w32.Rect
		_, _, _ = w32.User32GetWindowRect.Call(w.hwnd, uintptr(unsafe.Pointer(&rect)))
		width := int32(rect.Right - rect.Left)
		height := int32(rect.Bottom - rect.Top)

		screenWidth, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CXSCREEN)
		screenHeight, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CYSCREEN)

		x := int32(screenWidth-uintptr(width)) / 2
		y := int32(screenHeight-uintptr(height)) / 2

		_, _, _ = w32.User32SetWindowPos.Call(
			w.hwnd,
			0,
			uintptr(x),
			uintptr(y),
			uintptr(width),
			uintptr(height),
			w32.SWP_NOZORDER|w32.SWP_NOSIZE,
		)
	})
}

// SetOpacity 设置窗口透明度
func (w *webview) SetOpacity(opacity float64) {
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}

	w.Dispatch(func() {
		// 确保窗口有分层属性
		style, _, _ := w32.User32GetWindowLongPtrW.Call(w.hwnd, ^uintptr(15)) // 使用 ^uintptr(15) 替代 GWL_EXSTYLE
		style |= w32.WS_EX_LAYERED

		_, _, _ = w32.User32SetWindowLongPtrW.Call(w.hwnd, ^uintptr(15), style)

		// 设置透明度
		_, _, _ = w32.User32SetLayeredWindowAttributes.Call(
			w.hwnd,
			0,
			uintptr(opacity*255),
			w32.LWA_ALPHA,
		)
	})
}

// 浏览器相关能
func (w *webview) Reload() {
	w.Eval("window.location.reload();")
}

func (w *webview) Back() {
	w.Eval("window.history.back();")
}

func (w *webview) Forward() {
	w.Eval("window.history.forward();")
}

func (w *webview) Stop() {
	w.Eval("window.stop();")
}

// 开发者工具
func (w *webview) OpenDevTools() {
	if w.browser != nil {
		// 需要实现 OpenDevToolsWindow 方法
		//w.browser.OpenDevToolsWindow()
	}
}

func (w *webview) CloseDevTools() {
	// 通过 JavaScript 关闭开发者工具
	w.Eval(`if(window.devtools && window.devtools.isOpen()) window.devtools.close();`)
}

func (w *webview) OnLoadingStateChanged(callback func(bool)) {
	w.onLoadingStateChanged = callback
}

func (w *webview) OnURLChanged(callback func(string)) {
	w.onURLChanged = callback
}

func (w *webview) OnTitleChanged(callback func(string)) {
	w.onTitleChanged = callback
}

func (w *webview) OnFullscreenChanged(callback func(bool)) {
	w.onFullscreenChanged = callback
}

// ClearCache 清除浏览器缓存
func (w *webview) ClearCache() {
	// 通过 JavaScript 清除缓存
	w.Eval(`
		if (window.caches) {
			caches.keys().then(function(keyList) {
				return Promise.all(keyList.map(function(key) {
					return caches.delete(key);
				}));
			});
		}
		localStorage.clear();
		sessionStorage.clear();
	`)
}

// ClearCookies 清除浏览器 cookies
func (w *webview) ClearCookies() {
	// 通过 JavaScript 清除 cookies
	w.Eval(`
		document.cookie.split(";").forEach(function(c) { 
			document.cookie = c.replace(/^ +/, "")
				.replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/"); 
		});
	`)
}

// DispatchAsync 异步分发任务到主线程
func (w *webview) DispatchAsync(f func()) {
	// 使用 channel 来实现异步分发
	go func() {
		w.m.Lock()
		w.dispatchq = append(w.dispatchq, f)
		w.m.Unlock()
		_, _, _ = w32.User32PostThreadMessageW.Call(w.mainthread, w32.WMApp, 0, 0)
	}()
}

// Print 直接使用默认打印机打印前页面
func (w *webview) Print() {
	w.Eval(`window.print()`)
}

// PrintToPDF 将当前页面打印为 PDF 文件
func (w *webview) PrintToPDF(path string) {
	// 使用 WebView2 的 PrintToPdf 方法
	if w.browser != nil {
		if err := w.browser.PrintToPDF(path); err != nil {
			log.Printf("Failed to print to PDF: %v", err)
		}
	}
}

// ShowPrintDialog 显示打印对话框
func (w *webview) ShowPrintDialog() {
	w.Eval(`window.print()`)
}

// DisableContextMenu 禁用右键菜单
func (w *webview) DisableContextMenu() error {
	if w.browser != nil {
		return w.browser.DisableContextMenu()
	}
	return nil
}

// EnableContextMenu 启用右键菜单
func (w *webview) EnableContextMenu() error {
	// 移除 JavaScript 的右键菜单禁用
	w.Eval(`
		document.removeEventListener('contextmenu', function(e) {
			e.preventDefault();
			return false;
		}, true);
	`)

	// 同时恢复 WebView2 的默认上下文菜单
	if w.browser != nil {
		return w.browser.EnableContextMenu()
	}
	return nil
}

// RunAsync 异步运行 webview
func (w *webview) RunAsync() {
	// 启动一个新的 goroutine 来运行消息循环
	go func() {
		// 运行消息循环直到收到退出消息
		var msg w32.Msg
		for {
			r, _, _ := w32.User32GetMessageW.Call(
				uintptr(unsafe.Pointer(&msg)),
				0,
				0,
				0,
			)
			if r == 0 {
				break
			}

			// 处理热键消息
			if msg.Message == w32.WMHotKey {
				if handler, ok := w.hotkeys[int(msg.WParam)]; ok && handler != nil {
					handler()
					continue
				}
			}

			// 处理分发的息
			if msg.Message == w32.WMApp {
				w.m.Lock()
				if len(w.dispatchq) > 0 {
					f := w.dispatchq[0]
					w.dispatchq = w.dispatchq[1:]
					w.m.Unlock()
					f()
					continue
				}
				w.m.Unlock()
			}

			// 理常规窗口消息
			_, _, _ = w32.User32TranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			_, _, _ = w32.User32DispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}()
}

// SetMaximized 设置窗口最大化状态
func (w *webview) SetMaximized(enable bool) {
	w.Dispatch(func() {
		if enable {
			_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_MAXIMIZE)
		} else {
			_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_RESTORE)
		}
	})
}

// SetMinimized 设置窗口最小化状态
func (w *webview) SetMinimized(enable bool) {
	w.Dispatch(func() {
		if enable {
			_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_MINIMIZE)
		} else {
			_, _, _ = w32.User32ShowWindow.Call(w.hwnd, w32.SW_RESTORE)
		}
	})
}

// AddJSHook 添加 JavaScript 钩子
func (w *webview) AddJSHook(hook JSHook) {
	w.m.Lock()
	defer w.m.Unlock()

	// 按优先级插入
	inserted := false
	for i, h := range w.jsHooks {
		if hook.Priority() < h.Priority() {
			// 在此位置插入
			w.jsHooks = append(w.jsHooks[:i], append([]JSHook{hook}, w.jsHooks[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		w.jsHooks = append(w.jsHooks, hook)
	}
}

// RemoveJSHook 移除 JavaScript 钩子
func (w *webview) RemoveJSHook(hook JSHook) {
	w.m.Lock()
	defer w.m.Unlock()

	for i, h := range w.jsHooks {
		if h == hook {
			w.jsHooks = append(w.jsHooks[:i], w.jsHooks[i+1:]...)
			break
		}
	}
}

// ClearJSHooks 清除所有 JavaScript 钩子
func (w *webview) ClearJSHooks() {
	w.m.Lock()
	defer w.m.Unlock()
	w.jsHooks = nil
}

// processScript 处理脚本，用所有钩子
func (w *webview) processScript(script string, hookType JSHookType) string {
	w.m.Lock()
	defer w.m.Unlock()

	result := script
	for _, hook := range w.jsHooks {
		if hook.Type() == hookType {
			result = hook.Handle(result)
		}
	}
	return result
}

// EnableWebSocket 启用 WebSocket 服务
func (w *webview) EnableWebSocket(port int) error {
	w.wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
	}

	// 创建 WebSocket 处理器
	http.HandleFunc("/ws", func(writer http.ResponseWriter, r *http.Request) {
		conn, err := w.wsUpgrader.Upgrade(writer, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}

		// 保存连接
		connID := fmt.Sprintf("%p", conn)
		w.wsConnections.Store(connID, conn)

		// 创建并添加 WebSocket Hook
		wsHook := NewWebSocketHook(conn)
		w.AddJSHook(wsHook)

		// 处理消息
		go func() {
			defer func() {
				conn.Close()
				w.wsConnections.Delete(connID)
				w.RemoveJSHook(wsHook)
			}()

			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						log.Printf("WebSocket read error: %v", err)
					}
					return
				}

				if w.wsHandler != nil {
					w.wsHandler(string(message))
				}
			}
		}()
	})

	// 启动 WebSocket 服务器
	w.wsServer = &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	go func() {
		if err := w.wsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	// 注入 WebSocket 客户端代码
	w.Eval(`
		if (!window._webSocket) {
			window._webSocket = new WebSocket('ws://localhost:` + fmt.Sprint(port) + `/ws');
			window._webSocket.onmessage = function(event) {
				try {
					const data = JSON.parse(event.data);
					if (data.type === 'eval') {
						eval(data.script);
					}
				} catch(e) {
					console.error('WebSocket message error:', e);
				}
			};
		}
	`)

	return nil
}

// DisableWebSocket 禁用 WebSocket 服务
func (w *webview) DisableWebSocket() {
	if w.wsServer != nil {
		// 关闭所有连接
		w.wsConnections.Range(func(key, value interface{}) bool {
			if conn, ok := value.(*websocket.Conn); ok {
				conn.Close()
			}
			return true
		})

		// 关闭服务器
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		w.wsServer.Shutdown(ctx)
		w.wsServer = nil
	}

	// 清理客户端 WebSocket
	w.Eval(`
		if (window._webSocket) {
			window._webSocket.close();
			window._webSocket = null;
		}
	`)
}

// OnWebSocketMessage 设置 WebSocket 消息处理器
func (w *webview) OnWebSocketMessage(handler WebSocketHandler) {
	w.wsHandler = handler
}

// SendWebSocketMessage 发送 WebSocket 消息
func (w *webview) SendWebSocketMessage(message string) {
	w.wsConnections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(*websocket.Conn); ok {
			err := conn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
			}
		}
		return true
	})
}

func (w *webview) OnNavigationStarting(handler func()) {
	if w.browser != nil {
		if chromium, ok := w.browser.(*edge.Chromium); ok {
			chromium.NavigationStartingCallback = handler
		}
	}
}

func (w *webview) Browser() interface{} {
	return w.browser
}

// 设置消息回调
func (w *webview) SetMessageCallback(callback func(string)) {
	w.messageCallback = callback
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		chromium.SetMessageCallback(callback)
	}
}
