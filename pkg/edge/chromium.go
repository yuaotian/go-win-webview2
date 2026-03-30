//go:build windows
// +build windows

package edge

import (
	"encoding/json"
	"errors"
	"fmt"

	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/yuaotian/go-win-webview2/internal/w32"
	"golang.org/x/sys/windows"
)

var (
	shlwapi                    = windows.NewLazySystemDLL("shlwapi.dll")
	shCreateStreamOnFileExProc = shlwapi.NewProc("SHCreateStreamOnFileEx")
)

const (
	stgmWrite          = 0x00000001
	stgmShareExclusive = 0x00000010
	stgmCreate         = 0x00001000
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
	newWindowRequested    *ICoreWebView2NewWindowRequestedEventHandler
	windowCloseRequested  *ICoreWebView2WindowCloseRequestedEventHandler
	domContentLoaded      *ICoreWebView2DOMContentLoadedEventHandler
	navigationStarting    *ICoreWebView2NavigationStartingEventHandler
	sourceChanged         *ICoreWebView2SourceChangedEventHandler
	documentTitleChanged  *ICoreWebView2DocumentTitleChangedEventHandler
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
	NewWindowRequestedCallback   func(sender *ICoreWebView2, args *ICoreWebView2NewWindowRequestedEventArgs)
	WindowCloseRequestedCallback func()
	DOMContentLoadedCallback     func(sender *ICoreWebView2, args *ICoreWebView2DOMContentLoadedEventArgs)
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

	dispatch struct {
		queue []func()
		sync.Mutex
	}

	previewRequests  sync.Map
	devToolsRequests sync.Map
	devToolsEvents   sync.Map
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
	e.newWindowRequested = newICoreWebView2NewWindowRequestedEventHandler(e)
	e.windowCloseRequested = newICoreWebView2WindowCloseRequestedEventHandler(e)
	e.domContentLoaded = newICoreWebView2DOMContentLoadedEventHandler(e)
	e.navigationStarting = newICoreWebView2NavigationStartingEventHandler(e)
	e.sourceChanged = newICoreWebView2SourceChangedEventHandler(e)
	e.documentTitleChanged = newICoreWebView2DocumentTitleChangedEventHandler(e)
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
	if err := e.webview.AddNewWindowRequested(e.newWindowRequested, &token); err != nil {
		log.Fatal(err)
	}
	if err := e.webview.AddWindowCloseRequested(e.windowCloseRequested, &token); err != nil {
		log.Fatal(err)
	}
	if webview2_2 := e.GetICoreWebView2_2(); webview2_2 != nil {
		if err := webview2_2.AddDomContentLoaded(e.domContentLoaded, &token); err != nil {
			log.Fatal(err)
		}
	}

	// 补齐原生导航、URL 和标题事件，驱动外层状态监听。
	if err := e.webview.AddNavigationStarting(e.navigationStarting, &token); err != nil {
		log.Fatal(err)
	}
	if err := e.webview.AddSourceChanged(e.sourceChanged, &token); err != nil {
		log.Fatal(err)
	}
	if err := e.webview.AddDocumentTitleChanged(e.documentTitleChanged, &token); err != nil {
		log.Fatal(err)
	}
	if err := e.webview.AddNavigationCompleted(e.navigationCompleted, &token); err != nil {
		log.Fatal(err)
	}

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

func (e *Chromium) NewWindowRequested(sender *ICoreWebView2, args *ICoreWebView2NewWindowRequestedEventArgs) uintptr {
	if e.NewWindowRequestedCallback != nil {
		e.NewWindowRequestedCallback(sender, args)
	}
	return 0
}

func (e *Chromium) WindowCloseRequested(_ *ICoreWebView2, _ *iUnknown) uintptr {
	if e.WindowCloseRequestedCallback != nil {
		e.WindowCloseRequestedCallback()
		return 0
	}

	// 默认沿用现有窗口关闭语义，避免页面 window.close() 变成空操作。
	_, _, _ = w32.User32PostMessageW.Call(e.hwnd, w32.WMClose, 0, 0)
	return 0
}

func (e *Chromium) DOMContentLoaded(sender *ICoreWebView2, args *ICoreWebView2DOMContentLoadedEventArgs) uintptr {
	if e.DOMContentLoadedCallback != nil {
		e.DOMContentLoadedCallback(sender, args)
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

func (e *Chromium) NavigationStarting(sender *ICoreWebView2, _ *ICoreWebView2NavigationStartingEventArgs) uintptr {
	e.updateState(func(c *Chromium) {
		c.state.isLoading = true
	})

	if e.NavigationStartingCallback != nil {
		e.NavigationStartingCallback()
	}
	return 0
}

func (e *Chromium) SourceChanged(sender *ICoreWebView2, _ *ICoreWebView2SourceChangedEventArgs) uintptr {
	source, err := sender.GetSource()
	if err != nil {
		log.Printf("读取当前 URL 失败: %v", err)
		return 0
	}

	e.updateState(func(c *Chromium) {
		c.state.currentURL = source
	})
	return 0
}

func (e *Chromium) DocumentTitleChanged(sender *ICoreWebView2, _ *iUnknown) uintptr {
	title, err := sender.GetDocumentTitle()
	if err != nil {
		log.Printf("读取当前标题失败: %v", err)
		return 0
	}

	e.updateState(func(c *Chromium) {
		c.state.currentTitle = title
	})
	return 0
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

func (e *Chromium) OpenDevToolsWindow() error {
	if e.webview == nil {
		return errors.New("webview not initialized")
	}
	return e.webview.OpenDevToolsWindow()
}

func (e *Chromium) CoreWebView2() *ICoreWebView2 {
	return e.webview
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

func (e *Chromium) ShowPrintUI(kind COREWEBVIEW2_PRINT_DIALOG_KIND) error {
	if e.webview == nil {
		return errors.New("webview not initialized")
	}
	return e.webview.ShowPrintUI(kind)
}

type capturePreviewRequest struct {
	stream   *w32.IUnknown
	handler  *ICoreWebView2CapturePreviewCompletedHandler
	chromium *Chromium
}

func (r *capturePreviewRequest) QueryInterface(_, _ uintptr) uintptr {
	return 0
}

func (r *capturePreviewRequest) AddRef() uintptr {
	return 1
}

func (r *capturePreviewRequest) Release() uintptr {
	return 1
}

func (r *capturePreviewRequest) CapturePreviewCompleted(result uintptr) uintptr {
	if r.stream != nil {
		r.stream.Release()
		r.stream = nil
	}
	if r.chromium != nil && r.handler != nil {
		r.chromium.previewRequests.Delete(uintptr(unsafe.Pointer(r.handler)))
	}
	if int32(result) < 0 {
		log.Printf("截图预览失败: %08x", result)
	}
	return 0
}

func createPreviewFileStream(path string) (*w32.IUnknown, error) {
	var stream *w32.IUnknown
	hr, _, _ := shCreateStreamOnFileExProc.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(path))),
		uintptr(stgmCreate|stgmWrite|stgmShareExclusive),
		uintptr(windows.FILE_ATTRIBUTE_NORMAL),
		1,
		0,
		uintptr(unsafe.Pointer(&stream)),
	)
	if int32(hr) < 0 {
		return nil, fmt.Errorf("create preview stream failed: %08x", hr)
	}
	return stream, nil
}

func (e *Chromium) CapturePreview(path string) error {
	return e.CapturePreviewWithFormat(path, COREWEBVIEW2_CAPTURE_PREVIEW_IMAGE_FORMAT_PNG)
}

func (e *Chromium) CapturePreviewWithFormat(path string, format COREWEBVIEW2_CAPTURE_PREVIEW_IMAGE_FORMAT) error {
	if e.webview == nil {
		return errors.New("webview not initialized")
	}

	stream, err := createPreviewFileStream(path)
	if err != nil {
		return err
	}

	request := &capturePreviewRequest{
		stream:   stream,
		chromium: e,
	}
	request.handler = newICoreWebView2CapturePreviewCompletedHandler(request)

	key := uintptr(unsafe.Pointer(request.handler))
	e.previewRequests.Store(key, request)

	hr, _, _ := e.webview.vtbl.CapturePreview.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(format),
		uintptr(unsafe.Pointer(stream)),
		uintptr(unsafe.Pointer(request.handler)),
	)
	if int32(hr) < 0 {
		e.previewRequests.Delete(key)
		stream.Release()
		return fmt.Errorf("capture preview request failed: %08x", hr)
	}

	return nil
}

type devToolsMethodRequest struct {
	handler  *ICoreWebView2CallDevToolsProtocolMethodCompletedHandler
	chromium *Chromium
	callback func(error, string)
}

func (r *devToolsMethodRequest) QueryInterface(_, _ uintptr) uintptr {
	return 0
}

func (r *devToolsMethodRequest) AddRef() uintptr {
	return 1
}

func (r *devToolsMethodRequest) Release() uintptr {
	return 1
}

func (r *devToolsMethodRequest) CallDevToolsProtocolMethodCompleted(errorCode uintptr, result *uint16) uintptr {
	if r.chromium != nil && r.handler != nil {
		r.chromium.devToolsRequests.Delete(uintptr(unsafe.Pointer(r.handler)))
	}
	if r.callback != nil {
		if int32(errorCode) < 0 {
			r.callback(fmt.Errorf("devtools 协议调用失败: %08x", errorCode), "")
		} else {
			r.callback(nil, w32.Utf16PtrToString(result))
		}
	}
	return 0
}

func (e *Chromium) CallDevToolsProtocolMethod(method string, paramsJSON string, callback func(error, string)) error {
	if e.webview == nil {
		return errors.New("webview not initialized")
	}

	if callback == nil {
		return e.webview.CallDevToolsProtocolMethod(method, paramsJSON, nil)
	}

	request := &devToolsMethodRequest{
		chromium: e,
		callback: callback,
	}
	request.handler = newICoreWebView2CallDevToolsProtocolMethodCompletedHandler(request)

	key := uintptr(unsafe.Pointer(request.handler))
	e.devToolsRequests.Store(key, request)
	if err := e.webview.CallDevToolsProtocolMethod(method, paramsJSON, request.handler); err != nil {
		e.devToolsRequests.Delete(key)
		return err
	}
	return nil
}

type devToolsEventSubscription struct {
	handler  *ICoreWebView2DevToolsProtocolEventReceivedEventHandler
	receiver *ICoreWebView2DevToolsProtocolEventReceiver
	token    _EventRegistrationToken
	chromium *Chromium
	callback func(string)
}

func (s *devToolsEventSubscription) QueryInterface(_, _ uintptr) uintptr {
	return 0
}

func (s *devToolsEventSubscription) AddRef() uintptr {
	return 1
}

func (s *devToolsEventSubscription) Release() uintptr {
	return 1
}

func (s *devToolsEventSubscription) DevToolsProtocolEventReceived(_ *ICoreWebView2, args *ICoreWebView2DevToolsProtocolEventReceivedEventArgs) uintptr {
	if s.callback == nil {
		return 0
	}
	jsonText, err := args.GetParameterObjectAsJson()
	if err != nil {
		log.Printf("读取 DevTools 事件参数失败: %v", err)
		return 0
	}
	s.callback(jsonText)
	return 0
}

func (e *Chromium) SubscribeDevToolsProtocolEvent(eventName string, handler func(string)) (func(), error) {
	if e.webview == nil {
		return nil, errors.New("webview not initialized")
	}
	receiver, err := e.webview.GetDevToolsProtocolEventReceiver(eventName)
	if err != nil {
		return nil, err
	}

	subscription := &devToolsEventSubscription{
		receiver: receiver,
		chromium: e,
		callback: handler,
	}
	subscription.handler = newICoreWebView2DevToolsProtocolEventReceivedEventHandler(subscription)

	var token _EventRegistrationToken
	if err := receiver.AddDevToolsProtocolEventReceived(subscription.handler, &token); err != nil {
		receiver.Release()
		return nil, err
	}
	subscription.token = token

	key := uintptr(unsafe.Pointer(subscription.handler))
	e.devToolsEvents.Store(key, subscription)

	unsubscribe := func() {
		if subscription.receiver != nil {
			_ = subscription.receiver.RemoveDevToolsProtocolEventReceived(subscription.token)
			subscription.receiver.Release()
			subscription.receiver = nil
		}
		e.devToolsEvents.Delete(key)
	}

	return unsubscribe, nil
}

func (e *Chromium) SetDOMContentLoadedCallback(callback func(sender *ICoreWebView2, args *ICoreWebView2DOMContentLoadedEventArgs)) {
	e.DOMContentLoadedCallback = callback
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
	if e.webview == nil || f == nil {
		return
	}

	e.dispatch.Lock()
	e.dispatch.queue = append(e.dispatch.queue, f)
	e.dispatch.Unlock()

	// 通过固定的内部消息触发主线程执行，避免覆盖外部消息回调。
	_, _, _ = e.webview.vtbl.ExecuteScript.Call(
		uintptr(unsafe.Pointer(e.webview)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(
			`window.setTimeout(() => { window.chrome.webview.postMessage('__dispatch__'); }, 0);`,
		))),
		0,
	)
}

func (e *Chromium) HandleDispatchMessage(msg string) bool {
	if msg != "__dispatch__" {
		return false
	}

	e.dispatch.Lock()
	if len(e.dispatch.queue) == 0 {
		e.dispatch.Unlock()
		return true
	}
	f := e.dispatch.queue[0]
	e.dispatch.queue = e.dispatch.queue[1:]
	e.dispatch.Unlock()

	f()
	return true
}

// HandleWebMessage 处理来自 WebView 的消息
func (e *Chromium) HandleWebMessage(message string) bool {
	var msg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
		URL     string          `json:"url"`
		Loading bool            `json:"loading"`
		Title   string          `json:"title"`
	}

	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		return false
	}
	if msg.Type == "" {
		return false
	}

	switch msg.Type {
	case "navigate":
		var payload struct {
			URL string `json:"url"`
		}
		if len(msg.Payload) > 0 {
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("解析 navigate 消息失败: %v", err)
				return true
			}
		} else {
			payload.URL = msg.URL
		}
		if payload.URL == "" {
			return true
		}
		e.Navigate(payload.URL)
		return true

	case "stateChange":
		var payload struct {
			Loading bool   `json:"loading"`
			URL     string `json:"url"`
			Title   string `json:"title"`
		}
		if len(msg.Payload) > 0 {
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("解析 stateChange 消息失败: %v", err)
				return true
			}
		} else {
			payload.Loading = msg.Loading
			payload.URL = msg.URL
			payload.Title = msg.Title
		}
		e.updateState(func(c *Chromium) {
			c.state.isLoading = payload.Loading
			c.state.currentURL = payload.URL
			c.state.currentTitle = payload.Title
		})
		return true
	}
	return false
}

// 添加消息处理回调设置方法
func (e *Chromium) SetMessageCallback(callback func(string)) {
	e.MessageCallback = callback
}

// 供 webview 层桥接外部状态监听到 Chromium 内部事件。
func (e *Chromium) SetLoadingStateChangedCallback(callback func(bool)) {
	e.callbacks.Lock()
	defer e.callbacks.Unlock()
	e.callbacks.onLoadingStateChanged = callback
}

func (e *Chromium) SetURLChangedCallback(callback func(string)) {
	e.callbacks.Lock()
	defer e.callbacks.Unlock()
	e.callbacks.onURLChanged = callback
}

func (e *Chromium) SetTitleChangedCallback(callback func(string)) {
	e.callbacks.Lock()
	defer e.callbacks.Unlock()
	e.callbacks.onTitleChanged = callback
}

func (e *Chromium) SetNewWindowRequestedCallback(callback func(sender *ICoreWebView2, args *ICoreWebView2NewWindowRequestedEventArgs)) {
	e.NewWindowRequestedCallback = callback
}

func (e *Chromium) SetWindowCloseRequestedCallback(callback func()) {
	e.WindowCloseRequestedCallback = callback
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
