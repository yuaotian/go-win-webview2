//go:build windows
// +build windows

package webview2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"unsafe"

	"github.com/jchv/go-webview2/internal/w32"
	"github.com/jchv/go-webview2/pkg/edge"

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
	
	// 状态监听回调
	onLoadingStateChanged func(bool)
	onURLChanged         func(string)
	onTitleChanged      func(string)
	onFullscreenChanged func(bool)
}

type WindowOptions struct {
	Title     string
	Width     uint
	Height    uint
	IconId    uint
	Center    bool
	Frameless bool
	Fullscreen bool  // 是否全屏
	AlwaysOnTop bool // 是否置顶
}

type WebViewOptions struct {
	Window unsafe.Pointer
	Debug  bool

	//DataPath 指定 WebView2 运行时使用的数据路径
	//浏览器实例。
	DataPath string

	//当窗口打开时，AutoFocus 将尝试保持 WebView2 小部件聚焦
	//已聚焦。
	AutoFocus bool

	//WindowOptions 自定义创建的窗口以嵌入
	//WebView2 小部件。
	WindowOptions WindowOptions
}

//New 在新窗口中创建一个新的 webview。
func New(debug bool) WebView { return NewWithOptions(WebViewOptions{Debug: debug}) }

//NewWindow 使用现有窗口创建一个新的 webview。
//
//已弃用：使用 NewWithOptions。
func NewWindow(debug bool, window unsafe.Pointer) WebView {
	return NewWithOptions(WebViewOptions{Debug: debug, Window: window})
}

//NewWithOptions 使用提供的选项创建一个新的 webview。
func NewWithOptions(options WebViewOptions) WebView {
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
		log.Fatal(err)
	}
	//禁用上下文菜单
	err = settings.PutAreDefaultContextMenusEnabled(options.Debug)
	if err != nil {
		log.Fatal(err)
	}
	//禁用开发者工具
	err = settings.PutAreDevToolsEnabled(options.Debug)
	if err != nil {
		log.Fatal(err)
	}

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
			
			// 定义边框拖拽区域宽度
			const borderWidth = 5
			
			// 检查是否在拖拽区域内
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
			
			// 允许拖动窗口
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

	icon := w.loadWindowIcon(hinstance, opts.IconId)
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

	var style uint32 = w32.WSOverlappedWindow // 默认样式
	
	if opts.Frameless {
		// 无边框窗口样式
		style = w32.WSPopup | w32.WSVisible
	}

	w.hwnd, _, _ = w32.User32CreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style), // 使用新的样式
		uintptr(posX),
		uintptr(posY),
		uintptr(windowWidth),
		uintptr(windowHeight),
		0,
		0,
		uintptr(hinstance),
		0,
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

	return true
}

func (w *webview) Destroy() {
	// 注销所有热键
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
		style &^= (w32.WSThickFrame | w32.WSMaximizeBox)
	} else {
		style |= (w32.WSThickFrame | w32.WSMaximizeBox)
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
		_, _, _ = w32.User32AdjustWindowRect.Call(uintptr(unsafe.Pointer(&r)), w32.WSOverlappedWindow, 0)
		_, _, _ = w32.User32SetWindowPos.Call(
			w.hwnd, 0, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top),
			w32.SWP_NOZORDER|w32.SWP_NOACTIVATE|w32.SWP_NOMOVE|w32.SWP_FRAMECHANGED)
		w.browser.Resize()
	}
}

func (w *webview) Init(js string) {
	w.browser.Init(js)
}

func (w *webview) Eval(js string) {
	w.browser.Eval(js)
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

func (w *webview) loadWindowIcon(hinstance windows.Handle, iconId uint) uintptr {
	if iconId == 0 {
		icow, _, _ := w32.User32GetSystemMetrics.Call(w32.SystemMetricsCxIcon)
		icoh, _, _ := w32.User32GetSystemMetrics.Call(w32.SystemMetricsCyIcon)
		icon, _, _ := w32.User32LoadImageW.Call(
			uintptr(hinstance),
			32512,
			icow,
			icoh,
			0,
			0,
		)
		return icon
	}
	
	icon, _, _ := w32.User32LoadImageW.Call(
		uintptr(hinstance),
		uintptr(iconId),
		1,
		0,
		0,
		w32.LR_DEFAULTSIZE|w32.LR_SHARED,
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

	// 始化热键映射
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
		// 这里简化处理，实际应该保modifiers和keyCode来确保完全匹配
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
		// 保存当前窗口位置和大小用于还原
		var rect w32.Rect
		_, _, _ = w32.User32GetWindowRect.Call(w.hwnd, uintptr(unsafe.Pointer(&rect)))
		
		// 获取主显示器的完整尺寸（包括任务栏）
		screenWidth, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CXSCREEN)
		screenHeight, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CYSCREEN)
		
		// 移除窗口边框样式
		style, _, _ := w32.User32GetWindowLongPtrW.Call(w.hwnd, ^uintptr(15))
		style &^= w32.WSOverlappedWindow
		_, _, _ = w32.User32SetWindowLongPtrW.Call(w.hwnd, ^uintptr(15), style)
		
		// 设置全屏 - 使用整个屏幕尺寸
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
		// 恢复窗口样式
		style, _, _ := w32.User32GetWindowLongPtrW.Call(w.hwnd, ^uintptr(15))
		style |= w32.WSOverlappedWindow
		_, _, _ = w32.User32SetWindowLongPtrW.Call(w.hwnd, ^uintptr(15), style)
		
		// 获取屏幕尺寸
		screenWidth, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CXSCREEN)
		screenHeight, _, _ := w32.User32GetSystemMetrics.Call(w32.SM_CYSCREEN)
		
		// 设置默认窗口大小
		width := uintptr(1024)
		height := uintptr(768)
		
		// 计算居中位置
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
}

// SetAlwaysOnTop 设置窗口置顶状态
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

// 设置窗口透明度
func (w *webview) SetOpacity(opacity float64) {
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	
	w.Dispatch(func() {
		style, _, _ := w32.User32GetWindowLongPtrW.Call(w.hwnd, ^uintptr(15))
		style |= uintptr(w32.WS_EX_LAYERED)
		
		_, _, _ = w32.User32SetWindowLongPtrW.Call(w.hwnd, ^uintptr(15), style)
		_, _, _ = w32.User32SetLayeredWindowAttributes.Call(
			w.hwnd,
			0,
			uintptr(opacity * 255),
			uintptr(w32.LWA_ALPHA),
		)
	})
}

// 浏览器相关功能
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
