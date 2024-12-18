//go:build windows
// +build windows

package edge

import (
	"encoding/json"
	"errors"

	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/yuaotian/go-win-webview2/internal/w32"
	"golang.org/x/sys/windows"
)

// Chromium 结构体定义
type Chromium struct {
	hwnd                  uintptr
	focusOnInit           bool
	controller            *ICoreWebView2Controller
	webview               *ICoreWebView2
	inited                uintptr
	envCompleted          *iCoreWebView2CreateCoreWebView2EnvironmentCompletedHandler
	controllerCompleted   *iCoreWebView2CreateCoreWebView2ControllerCompletedHandler
	webMessageReceived    *iCoreWebView2WebMessageReceivedEventHandler
	permissionRequested   *iCoreWebView2PermissionRequestedEventHandler
	webResourceRequested  *iCoreWebView2WebResourceRequestedEventHandler
	acceleratorKeyPressed *ICoreWebView2AcceleratorKeyPressedEventHandler
	navigationCompleted   *ICoreWebView2NavigationCompletedEventHandler

	environment *ICoreWebView2Environment

	// Settings
	DataPath string

	// permissions
	permissions      map[CoreWebView2PermissionKind]CoreWebView2PermissionState
	globalPermission *CoreWebView2PermissionState

	// Callbacks
	MessageCallback              func(string)
	WebResourceRequestedCallback func(request *ICoreWebView2WebResourceRequest, args *ICoreWebView2WebResourceRequestedEventArgs)
	NavigationCompletedCallback  func(sender *ICoreWebView2, args *ICoreWebView2NavigationCompletedEventArgs)
	AcceleratorKeyCallback       func(uint) bool
	NavigationStartingCallback   func()

	// 状态管理
	state struct {
		isLoading    bool
		currentURL   string
		currentTitle string
		isFullscreen bool
		sync.RWMutex
	}

	// 状态变化回调
	callbacks struct {
		onLoadingStateChanged func(bool)
		onURLChanged          func(string)
		onTitleChanged        func(string)
		onFullscreenChanged   func(bool)
		sync.RWMutex
	}
}

func NewChromium() *Chromium {
	e := &Chromium{}

	/*
	   	 所有这处理程序都通过带有"uintptr(unsafe.Pointer(handler))"的系统调用传递给本机代码，我们知道
	   	 指向这些的指针将保留在本机代码中。此外，这些处理程序还包含指向其他 Go 的指针
	   	 vtable 这样的结构。
	   	 这违反了 unsafe.Pointer 规则"(4) 在调用 syscall.Syscall 时将指针转换为 uintptr"。因为
	   	 无法保证 Go 不会���动这些对象。
	   据我所前 Go 运行时不会移动 HEAP 对象，因此我们使用这些处理程序应该是安全的。但他们不
	   	 保证它，因为Go 可能使用压缩 GC。
	   	 有人建议添加一个runtime.Pin函数，以防止移动固定对象，这将允许轻松修复
	   	 只需固定处理程序即可解决此问题。 https://go-review.googlesource.com/c/go/+/367296/应该登陆 Go 1.19。
	*/
	e.envCompleted = newICoreWebView2CreateCoreWebView2EnvironmentCompletedHandler(e)
	e.controllerCompleted = newICoreWebView2CreateCoreWebView2ControllerCompletedHandler(e)
	e.webMessageReceived = newICoreWebView2WebMessageReceivedEventHandler(e)
	e.permissionRequested = newICoreWebView2PermissionRequestedEventHandler(e)
	e.webResourceRequested = newICoreWebView2WebResourceRequestedEventHandler(e)
	e.acceleratorKeyPressed = newICoreWebView2AcceleratorKeyPressedEventHandler(e)
	e.navigationCompleted = newICoreWebView2NavigationCompletedEventHandler(e)
	e.permissions = make(map[CoreWebView2PermissionKind]CoreWebView2PermissionState)

	return e
}

func (e *Chromium) Embed(hwnd uintptr) bool {
	e.hwnd = hwnd

	dataPath := e.DataPath
	if dataPath == "" {
		currentExePath := make([]uint16, windows.MAX_PATH)
		_, err := windows.GetModuleFileName(windows.Handle(0), &currentExePath[0], windows.MAX_PATH)
		if err != nil {
			// What to do here?
			return false
		}
		currentExeName := filepath.Base(windows.UTF16ToString(currentExePath))
		dataPath = filepath.Join(os.Getenv("AppData"), currentExeName)
	}

	res, err := createCoreWebView2EnvironmentWithOptions(nil, windows.StringToUTF16Ptr(dataPath), 0, e.envCompleted)
	if err != nil {
		log.Printf("Error calling Webview2Loader: %v", err)
		return false
	} else if res != 0 {
		log.Printf("Result: %08x", res)
		return false
	}
	var msg w32.Msg
	for {
		if atomic.LoadUintptr(&e.inited) != 0 {
			break
		}
		r, _, _ := w32.User32GetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0,
			0,
			0,
		)
		if r == 0 {
			break
		}
		_, _, _ = w32.User32TranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		_, _, _ = w32.User32DispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
	e.Init("window.external={invoke:s=>window.chrome.webview.postMessage(s)}")
	return true
}

// 导航URL
func (e *Chromium) Navigate(url string) {
	_, _, _ = e.webview.vtbl.Navigate.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(url))),
	)
}

// NavigateToString 将 HTML 内容注入到 WebView 中
func (e *Chromium) NavigateToString(htmlContent string) {
	_, _, _ = e.webview.vtbl.NavigateToString.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(htmlContent))),
	)
}

func (e *Chromium) Init(script string) {
	baseScript := `console.log('hello world')`

	// 合并脚本
	fullScript := baseScript
	if script != "" {
		fullScript += ";" + script
	}

	_, _, _ = e.webview.vtbl.AddScriptToExecuteOnDocumentCreated.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(fullScript))),
		0,
	)
}

func (e *Chromium) Eval(script string) {
	_script, err := windows.UTF16PtrFromString(script)
	if err != nil {
		log.Fatal(err)
	}

	_, _, _ = e.webview.vtbl.ExecuteScript.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(_script)),
		0,
	)
}

func (e *Chromium) Show() error {
	return e.controller.PutIsVisible(true)
}

func (e *Chromium) Hide() error {
	return e.controller.PutIsVisible(false)
}

func (e *Chromium) QueryInterface(_, _ uintptr) uintptr {
	return 0
}

func (e *Chromium) AddRef() uintptr {
	return 1
}

func (e *Chromium) Release() uintptr {
	return 1
}

func (e *Chromium) EnvironmentCompleted(res uintptr, env *ICoreWebView2Environment) uintptr {
	if int64(res) < 0 {
		log.Fatalf("Creating environment failed with %08x", res)
	}
	_, _, _ = env.vtbl.AddRef.Call(uintptr(unsafe.Pointer(env)))
	e.environment = env

	_, _, _ = env.vtbl.CreateCoreWebView2Controller.Call(
		uintptr(unsafe.Pointer(env)),
		e.hwnd,
		uintptr(unsafe.Pointer(e.controllerCompleted)),
	)
	return 0
}

func (e *Chromium) CreateCoreWebView2ControllerCompleted(res uintptr, controller *ICoreWebView2Controller) uintptr {
	if int64(res) < 0 {
		log.Fatalf("Creating controller failed with %08x", res)
	}
	_, _, _ = controller.vtbl.AddRef.Call(uintptr(unsafe.Pointer(controller)))
	e.controller = controller

	var token _EventRegistrationToken
	_, _, _ = controller.vtbl.GetCoreWebView2.Call(
		uintptr(unsafe.Pointer(controller)),
		uintptr(unsafe.Pointer(&e.webview)),
	)
	_, _, _ = e.webview.vtbl.AddRef.Call(
		uintptr(unsafe.Pointer(e.webview)),
	)
	_, _, _ = e.webview.vtbl.AddWebMessageReceived.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(e.webMessageReceived)),
		uintptr(unsafe.Pointer(&token)),
	)
	_, _, _ = e.webview.vtbl.AddPermissionRequested.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(e.permissionRequested)),
		uintptr(unsafe.Pointer(&token)),
	)
	_, _, _ = e.webview.vtbl.AddWebResourceRequested.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(e.webResourceRequested)),
		uintptr(unsafe.Pointer(&token)),
	)
	_, _, _ = e.webview.vtbl.AddNavigationCompleted.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(e.navigationCompleted)),
		uintptr(unsafe.Pointer(&token)),
	)

	_ = e.controller.AddAcceleratorKeyPressed(e.acceleratorKeyPressed, &token)

	atomic.StoreUintptr(&e.inited, 1)

	if e.focusOnInit {
		e.Focus()
	}

	return 0
}

func (e *Chromium) MessageReceived(sender *ICoreWebView2, args *iCoreWebView2WebMessageReceivedEventArgs) uintptr {
	var message *uint16
	_, _, _ = args.vtbl.TryGetWebMessageAsString.Call(
		uintptr(unsafe.Pointer(args)),
		uintptr(unsafe.Pointer(&message)),
	)
	if e.MessageCallback != nil {
		e.MessageCallback(w32.Utf16PtrToString(message))
	}
	_, _, _ = sender.vtbl.PostWebMessageAsString.Call(
		uintptr(unsafe.Pointer(sender)),
		uintptr(unsafe.Pointer(message)),
	)
	windows.CoTaskMemFree(unsafe.Pointer(message))
	return 0
}

func (e *Chromium) SetPermission(kind CoreWebView2PermissionKind, state CoreWebView2PermissionState) {
	e.permissions[kind] = state
}

func (e *Chromium) SetGlobalPermission(state CoreWebView2PermissionState) {
	e.globalPermission = &state
}

func (e *Chromium) PermissionRequested(_ *ICoreWebView2, args *iCoreWebView2PermissionRequestedEventArgs) uintptr {
	var kind CoreWebView2PermissionKind
	_, _, _ = args.vtbl.GetPermissionKind.Call(
		uintptr(unsafe.Pointer(args)),
		uintptr(kind),
	)
	var result CoreWebView2PermissionState
	if e.globalPermission != nil {
		result = *e.globalPermission
	} else {
		var ok bool
		result, ok = e.permissions[kind]
		if !ok {
			result = CoreWebView2PermissionStateDefault
		}
	}
	_, _, _ = args.vtbl.PutState.Call(
		uintptr(unsafe.Pointer(args)),
		uintptr(result),
	)
	return 0
}

func (e *Chromium) WebResourceRequested(sender *ICoreWebView2, args *ICoreWebView2WebResourceRequestedEventArgs) uintptr {
	req, err := args.GetRequest()
	if err != nil {
		log.Fatal(err)
	}
	if e.WebResourceRequestedCallback != nil {
		e.WebResourceRequestedCallback(req, args)
	}
	return 0
}

func (e *Chromium) AddWebResourceRequestedFilter(filter string, ctx COREWEBVIEW2_WEB_RESOURCE_CONTEXT) {
	err := e.webview.AddWebResourceRequestedFilter(filter, ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Chromium) Environment() *ICoreWebView2Environment {
	return e.environment
}

// AcceleratorKeyPressed is called when an accelerator key is pressed.
// If the AcceleratorKeyCallback method has been set, it will defer handling of the keypress
// to the callback. That callback returns a bool indicating if the event was handled.
func (e *Chromium) AcceleratorKeyPressed(sender *ICoreWebView2Controller, args *ICoreWebView2AcceleratorKeyPressedEventArgs) uintptr {
	if e.AcceleratorKeyCallback == nil {
		return 0
	}
	eventKind, _ := args.GetKeyEventKind()
	if eventKind == COREWEBVIEW2_KEY_EVENT_KIND_KEY_DOWN ||
		eventKind == COREWEBVIEW2_KEY_EVENT_KIND_SYSTEM_KEY_DOWN {
		virtualKey, _ := args.GetVirtualKey()
		status, _ := args.GetPhysicalKeyStatus()
		if !status.WasKeyDown {
			_ = args.PutHandled(e.AcceleratorKeyCallback(virtualKey))
			return 0
		}
	}
	_ = args.PutHandled(false)
	return 0
}

func (e *Chromium) GetSettings() (*ICoreWebViewSettings, error) {
	return e.webview.GetSettings()
}

func (e *Chromium) GetController() *ICoreWebView2Controller {
	return e.controller
}

func (e *Chromium) NavigationCompleted(sender *ICoreWebView2, args *ICoreWebView2NavigationCompletedEventArgs) uintptr {
	e.updateState(func(c *Chromium) {
		c.state.isLoading = false
	})

	if e.NavigationCompletedCallback != nil {
		e.NavigationCompletedCallback(sender, args)
	}
	return 0
}

func (e *Chromium) NotifyParentWindowPositionChanged() error {
	//看起来控制器初始化完成之前就调用了wndproc函。
	//此控制器为零
	if e.controller == nil {
		return nil
	}
	return e.controller.NotifyParentWindowPositionChanged()
}

func (e *Chromium) Focus() {
	if e.controller == nil {
		e.focusOnInit = true
		return
	}
	_ = e.controller.MoveFocus(COREWEBVIEW2_MOVE_FOCUS_REASON_PROGRAMMATIC)
}

// PrintToPDF 将当前页面打印到 PDF 文件
func (e *Chromium) PrintToPDF(path string) error {
	if e.webview == nil {
		return errors.New("webview not initialized")
	}

	// 转换文件路径为 UTF16
	_path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	// 调用 WebView2 的 PrintToPdf 方法
	_, _, err = e.webview.vtbl.PrintToPdf.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(_path)),
		0, // 使用默认打印设置
	)

	if err != windows.ERROR_SUCCESS {
		return err
	}

	return nil
}

// 添加打印相关的回调处理
func (e *Chromium) handlePrintCompleted(success bool, errorCode int) {
	// 处理打印成件
	if !success {
		log.Printf("Print failed with error code: %d", errorCode)
	}
}

// DisableContextMenu 禁用上下文菜单
func (e *Chromium) DisableContextMenu() error {
	if settings, err := e.GetSettings(); err != nil {
		return err
	} else {
		return settings.PutAreDefaultContextMenusEnabled(false)
	}
}

// EnableContextMenu 启用上下文菜单
func (e *Chromium) EnableContextMenu() error {
	if settings, err := e.GetSettings(); err != nil {
		return err
	} else {
		return settings.PutAreDefaultContextMenusEnabled(true)
	}
}

// Dispatch 在主线程中执行函数
func (e *Chromium) Dispatch(f func()) {
	if e.webview == nil {
		return
	}

	// 创建个通道来同步执行
	done := make(chan struct{})

	go func() {
		// 在主线程中执行函数
		_, _, _ = e.webview.vtbl.ExecuteScript.Call(
			uintptr(unsafe.Pointer(e.webview)),
			uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(
				`window.setTimeout(() => { window.chrome.webview.postMessage('__dispatch__'); }, 0);`,
			))),
			0,
		)

		// 等待执行完成
		<-done
	}()

	// 设置一个临时的消息处理器
	oldCallback := e.MessageCallback
	e.MessageCallback = func(msg string) {
		if msg == "__dispatch__" {
			f()
			e.MessageCallback = oldCallback
			close(done)
		} else if oldCallback != nil {
			oldCallback(msg)
		}
	}
}

// HandleWebMessage 处理来自 WebView 的消息
func (e *Chromium) HandleWebMessage(message string) {
	var msg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	switch msg.Type {
	case "navigate":
		var payload struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Failed to parse navigate payload: %v", err)
			return
		}
		e.Navigate(payload.URL)

	case "stateChange":
		var payload struct {
			Loading bool   `json:"loading"`
			URL     string `json:"url"`
			Title   string `json:"title"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Failed to parse state change payload: %v", err)
			return
		}
		e.updateState(func(c *Chromium) {
			c.state.isLoading = payload.Loading
			c.state.currentURL = payload.URL
			c.state.currentTitle = payload.Title
		})
	}
}

// 添加消息处理回调设置方法
func (e *Chromium) SetMessageCallback(callback func(string)) {
	e.MessageCallback = callback
}

// 加状态管理方法
func (e *Chromium) updateState(update func(*Chromium)) {
	e.state.Lock()
	defer e.state.Unlock()

	// 记录更新前的状态
	oldURL := e.state.currentURL
	oldTitle := e.state.currentTitle
	oldFullscreen := e.state.isFullscreen

	// 执行更新
	update(e)

	// 触发相应调
	e.callbacks.RLock()
	defer e.callbacks.RUnlock()

	// 触发加载状态变化回调
	if e.callbacks.onLoadingStateChanged != nil {
		e.callbacks.onLoadingStateChanged(e.state.isLoading)
	}

	// 触发 URL 变化回调
	if e.callbacks.onURLChanged != nil && oldURL != e.state.currentURL {
		e.callbacks.onURLChanged(e.state.currentURL)
	}

	// 触发标题变化回调
	if e.callbacks.onTitleChanged != nil && oldTitle != e.state.currentTitle {
		e.callbacks.onTitleChanged(e.state.currentTitle)
	}

	// 触发全屏状态变化回调
	if e.callbacks.onFullscreenChanged != nil && oldFullscreen != e.state.isFullscreen {
		e.callbacks.onFullscreenChanged(e.state.isFullscreen)
	}
}

// 设置导航开始回调
func (e *Chromium) SetNavigationStartingCallback(callback func()) {
	e.NavigationStartingCallback = callback
}

// boolToInt 将 bool 转换为 uintptr
func boolToInt(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
