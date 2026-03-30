//go:build windows
// +build windows

package edge

type _ICoreWebView2CapturePreviewCompletedHandlerVtbl struct {
	_IUnknownVtbl
	Invoke ComProc
}

type ICoreWebView2CapturePreviewCompletedHandler struct {
	vtbl *_ICoreWebView2CapturePreviewCompletedHandlerVtbl
	impl _ICoreWebView2CapturePreviewCompletedHandlerImpl
}

func _ICoreWebView2CapturePreviewCompletedHandlerIUnknownQueryInterface(this *ICoreWebView2CapturePreviewCompletedHandler, refiid, object uintptr) uintptr {
	return this.impl.QueryInterface(refiid, object)
}

func _ICoreWebView2CapturePreviewCompletedHandlerIUnknownAddRef(this *ICoreWebView2CapturePreviewCompletedHandler) uintptr {
	return this.impl.AddRef()
}

func _ICoreWebView2CapturePreviewCompletedHandlerIUnknownRelease(this *ICoreWebView2CapturePreviewCompletedHandler) uintptr {
	return this.impl.Release()
}

func _ICoreWebView2CapturePreviewCompletedHandlerInvoke(this *ICoreWebView2CapturePreviewCompletedHandler, result uintptr) uintptr {
	return this.impl.CapturePreviewCompleted(result)
}

type _ICoreWebView2CapturePreviewCompletedHandlerImpl interface {
	_IUnknownImpl
	CapturePreviewCompleted(result uintptr) uintptr
}

var _ICoreWebView2CapturePreviewCompletedHandlerFn = _ICoreWebView2CapturePreviewCompletedHandlerVtbl{
	_IUnknownVtbl{
		NewComProc(_ICoreWebView2CapturePreviewCompletedHandlerIUnknownQueryInterface),
		NewComProc(_ICoreWebView2CapturePreviewCompletedHandlerIUnknownAddRef),
		NewComProc(_ICoreWebView2CapturePreviewCompletedHandlerIUnknownRelease),
	},
	NewComProc(_ICoreWebView2CapturePreviewCompletedHandlerInvoke),
}

func newICoreWebView2CapturePreviewCompletedHandler(impl _ICoreWebView2CapturePreviewCompletedHandlerImpl) *ICoreWebView2CapturePreviewCompletedHandler {
	return &ICoreWebView2CapturePreviewCompletedHandler{
		vtbl: &_ICoreWebView2CapturePreviewCompletedHandlerFn,
		impl: impl,
	}
}
