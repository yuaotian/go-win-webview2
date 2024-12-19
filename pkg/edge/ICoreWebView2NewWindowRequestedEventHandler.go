//go:build windows
// +build windows
package edge

type iCoreWebView2NewWindowRequestedEventHandlerVtbl struct {
    _IUnknownVtbl
    Invoke ComProc
}

type iCoreWebView2NewWindowRequestedEventHandler struct {
    vtbl *iCoreWebView2NewWindowRequestedEventHandlerVtbl
    impl iCoreWebView2NewWindowRequestedEventHandlerImpl
}

type iCoreWebView2NewWindowRequestedEventHandlerImpl interface {
    _IUnknownImpl
    NewWindowRequested(sender *ICoreWebView2, args *iCoreWebView2NewWindowRequestedEventArgs) uintptr
}

type iCoreWebView2NewWindowRequestedEventArgs struct {
    vtbl *iCoreWebView2NewWindowRequestedEventArgsVtbl
}

func _ICoreWebView2NewWindowRequestedEventHandlerIUnknownQueryInterface(this *iCoreWebView2NewWindowRequestedEventHandler, refiid, object uintptr) uintptr {
    return this.impl.QueryInterface(refiid, object)
}

func _ICoreWebView2NewWindowRequestedEventHandlerIUnknownAddRef(this *iCoreWebView2NewWindowRequestedEventHandler) uintptr {
    return this.impl.AddRef()
}

func _ICoreWebView2NewWindowRequestedEventHandlerIUnknownRelease(this *iCoreWebView2NewWindowRequestedEventHandler) uintptr {
    return this.impl.Release()
}

func _ICoreWebView2NewWindowRequestedEventHandlerInvoke(this *iCoreWebView2NewWindowRequestedEventHandler, sender *ICoreWebView2, args *iCoreWebView2NewWindowRequestedEventArgs) uintptr {
    return this.impl.NewWindowRequested(sender, args)
}

var iCoreWebView2NewWindowRequestedEventHandlerFn = iCoreWebView2NewWindowRequestedEventHandlerVtbl{
    _IUnknownVtbl{
        NewComProc(_ICoreWebView2NewWindowRequestedEventHandlerIUnknownQueryInterface),
        NewComProc(_ICoreWebView2NewWindowRequestedEventHandlerIUnknownAddRef),
        NewComProc(_ICoreWebView2NewWindowRequestedEventHandlerIUnknownRelease),
    },
    NewComProc(_ICoreWebView2NewWindowRequestedEventHandlerInvoke),
}

func newICoreWebView2NewWindowRequestedEventHandler(impl iCoreWebView2NewWindowRequestedEventHandlerImpl) *iCoreWebView2NewWindowRequestedEventHandler {
    return &iCoreWebView2NewWindowRequestedEventHandler{
        vtbl: &iCoreWebView2NewWindowRequestedEventHandlerFn,
        impl: impl,
    }
}

// ... 实现其他必要的接口和方法 