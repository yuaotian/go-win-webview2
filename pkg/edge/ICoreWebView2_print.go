//go:build windows
// +build windows

package edge

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// 按官方 WebView2.h 的继承链补齐打印相关接口偏移，避免继续复用错误的 base vtable 尾部。
type iCoreWebView2_4Vtbl struct {
	iCoreWebView2_3Vtbl
	AddFrameCreated    ComProc
	RemoveFrameCreated ComProc
}

type iCoreWebView2_5Vtbl struct {
	iCoreWebView2_4Vtbl
	AddDownloadStarting    ComProc
	RemoveDownloadStarting ComProc
}

type iCoreWebView2_6Vtbl struct {
	iCoreWebView2_5Vtbl
	AddClientCertificateRequested    ComProc
	RemoveClientCertificateRequested ComProc
	OpenTaskManagerWindow            ComProc
}

type iCoreWebView2_7Vtbl struct {
	iCoreWebView2_6Vtbl
	PrintToPdf ComProc
}

type ICoreWebView2_7 struct {
	vtbl *iCoreWebView2_7Vtbl
}

func (i *ICoreWebView2_7) PrintToPdf(path string) error {
	return i.PrintToPdfWithSettings(path, nil, nil)
}

func (i *ICoreWebView2_7) PrintToPdfWithSettings(path string, settings *ICoreWebView2PrintSettings, handler *ICoreWebView2PrintToPdfCompletedHandler) error {
	_path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	hr, _, err := i.vtbl.PrintToPdf.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_path)),
		uintptr(unsafe.Pointer(settings)),
		uintptr(unsafe.Pointer(handler)),
	)
	if int32(hr) < 0 {
		return fmt.Errorf("print to pdf failed: %08x", hr)
	}
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2) GetICoreWebView2_7() *ICoreWebView2_7 {
	var result *ICoreWebView2_7

	iidICoreWebView2_7 := NewGUID("{79C24D83-09A3-45AE-9418-487F32A58740}")
	_, _, _ = i.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2_7)),
		uintptr(unsafe.Pointer(&result)),
	)

	return result
}

func (e *Chromium) GetICoreWebView2_7() *ICoreWebView2_7 {
	return e.webview.GetICoreWebView2_7()
}

type iCoreWebView2_8Vtbl struct {
	iCoreWebView2_7Vtbl
	AddIsMutedChanged                   ComProc
	RemoveIsMutedChanged                ComProc
	GetIsMuted                          ComProc
	PutIsMuted                          ComProc
	AddIsDocumentPlayingAudioChanged    ComProc
	RemoveIsDocumentPlayingAudioChanged ComProc
	GetIsDocumentPlayingAudio           ComProc
}

type iCoreWebView2_9Vtbl struct {
	iCoreWebView2_8Vtbl
	AddIsDefaultDownloadDialogOpenChanged    ComProc
	RemoveIsDefaultDownloadDialogOpenChanged ComProc
	GetIsDefaultDownloadDialogOpen           ComProc
	OpenDefaultDownloadDialog                ComProc
	CloseDefaultDownloadDialog               ComProc
	GetDefaultDownloadDialogCornerAlignment  ComProc
	PutDefaultDownloadDialogCornerAlignment  ComProc
	GetDefaultDownloadDialogMargin           ComProc
	PutDefaultDownloadDialogMargin           ComProc
}

type iCoreWebView2_10Vtbl struct {
	iCoreWebView2_9Vtbl
	AddBasicAuthenticationRequested    ComProc
	RemoveBasicAuthenticationRequested ComProc
}

type iCoreWebView2_11Vtbl struct {
	iCoreWebView2_10Vtbl
	CallDevToolsProtocolMethodForSession ComProc
	AddContextMenuRequested              ComProc
	RemoveContextMenuRequested           ComProc
}

type iCoreWebView2_12Vtbl struct {
	iCoreWebView2_11Vtbl
	AddStatusBarTextChanged    ComProc
	RemoveStatusBarTextChanged ComProc
	GetStatusBarText           ComProc
}

type iCoreWebView2_13Vtbl struct {
	iCoreWebView2_12Vtbl
	GetProfile ComProc
}

type iCoreWebView2_14Vtbl struct {
	iCoreWebView2_13Vtbl
	AddServerCertificateErrorDetected    ComProc
	RemoveServerCertificateErrorDetected ComProc
	ClearServerCertificateErrorActions   ComProc
}

type iCoreWebView2_15Vtbl struct {
	iCoreWebView2_14Vtbl
	AddFaviconChanged    ComProc
	RemoveFaviconChanged ComProc
	GetFaviconUri        ComProc
	GetFavicon           ComProc
}

type iCoreWebView2_16Vtbl struct {
	iCoreWebView2_15Vtbl
	Print            ComProc
	ShowPrintUI      ComProc
	PrintToPdfStream ComProc
}

type ICoreWebView2_16 struct {
	vtbl *iCoreWebView2_16Vtbl
}

func (i *ICoreWebView2_16) Print(settings *ICoreWebView2PrintSettings, handler *ICoreWebView2PrintCompletedHandler) error {
	hr, _, err := i.vtbl.Print.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(settings)),
		uintptr(unsafe.Pointer(handler)),
	)
	if int32(hr) < 0 {
		return fmt.Errorf("print failed: %08x", hr)
	}
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2_16) PrintToPdfStream(settings *ICoreWebView2PrintSettings, handler *ICoreWebView2PrintToPdfStreamCompletedHandler) error {
	hr, _, err := i.vtbl.PrintToPdfStream.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(settings)),
		uintptr(unsafe.Pointer(handler)),
	)
	if int32(hr) < 0 {
		return fmt.Errorf("print to pdf stream failed: %08x", hr)
	}
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2_16) ShowPrintUI(kind COREWEBVIEW2_PRINT_DIALOG_KIND) error {
	hr, _, err := i.vtbl.ShowPrintUI.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(kind),
	)
	if int32(hr) < 0 {
		return fmt.Errorf("show print ui failed: %08x", hr)
	}
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2) GetICoreWebView2_16() *ICoreWebView2_16 {
	var result *ICoreWebView2_16

	iidICoreWebView2_16 := NewGUID("{0EB34DC9-9F91-41E1-8639-95CD5943906B}")
	_, _, _ = i.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2_16)),
		uintptr(unsafe.Pointer(&result)),
	)

	return result
}

func (e *Chromium) GetICoreWebView2_16() *ICoreWebView2_16 {
	return e.webview.GetICoreWebView2_16()
}

type _ICoreWebView2PrintToPdfCompletedHandlerVtbl struct {
	_IUnknownVtbl
	Invoke ComProc
}

type ICoreWebView2PrintToPdfCompletedHandler struct {
	vtbl *_ICoreWebView2PrintToPdfCompletedHandlerVtbl
	impl _ICoreWebView2PrintToPdfCompletedHandlerImpl
}

type _ICoreWebView2PrintToPdfCompletedHandlerImpl interface {
	_IUnknownImpl
	PrintToPdfCompleted(errorCode uintptr, isSuccessful int32) uintptr
}

func _ICoreWebView2PrintToPdfCompletedHandlerIUnknownQueryInterface(this *ICoreWebView2PrintToPdfCompletedHandler, refiid, object uintptr) uintptr {
	return this.impl.QueryInterface(refiid, object)
}

func _ICoreWebView2PrintToPdfCompletedHandlerIUnknownAddRef(this *ICoreWebView2PrintToPdfCompletedHandler) uintptr {
	return this.impl.AddRef()
}

func _ICoreWebView2PrintToPdfCompletedHandlerIUnknownRelease(this *ICoreWebView2PrintToPdfCompletedHandler) uintptr {
	return this.impl.Release()
}

func _ICoreWebView2PrintToPdfCompletedHandlerInvoke(this *ICoreWebView2PrintToPdfCompletedHandler, errorCode uintptr, isSuccessful int32) uintptr {
	return this.impl.PrintToPdfCompleted(errorCode, isSuccessful)
}

var _ICoreWebView2PrintToPdfCompletedHandlerFn = _ICoreWebView2PrintToPdfCompletedHandlerVtbl{
	_IUnknownVtbl{
		NewComProc(_ICoreWebView2PrintToPdfCompletedHandlerIUnknownQueryInterface),
		NewComProc(_ICoreWebView2PrintToPdfCompletedHandlerIUnknownAddRef),
		NewComProc(_ICoreWebView2PrintToPdfCompletedHandlerIUnknownRelease),
	},
	NewComProc(_ICoreWebView2PrintToPdfCompletedHandlerInvoke),
}

func newICoreWebView2PrintToPdfCompletedHandler(impl _ICoreWebView2PrintToPdfCompletedHandlerImpl) *ICoreWebView2PrintToPdfCompletedHandler {
	return &ICoreWebView2PrintToPdfCompletedHandler{
		vtbl: &_ICoreWebView2PrintToPdfCompletedHandlerFn,
		impl: impl,
	}
}

type _ICoreWebView2PrintCompletedHandlerVtbl struct {
	_IUnknownVtbl
	Invoke ComProc
}

type ICoreWebView2PrintCompletedHandler struct {
	vtbl *_ICoreWebView2PrintCompletedHandlerVtbl
	impl _ICoreWebView2PrintCompletedHandlerImpl
}

type _ICoreWebView2PrintCompletedHandlerImpl interface {
	_IUnknownImpl
	PrintCompleted(errorCode uintptr, status COREWEBVIEW2_PRINT_STATUS) uintptr
}

func _ICoreWebView2PrintCompletedHandlerIUnknownQueryInterface(this *ICoreWebView2PrintCompletedHandler, refiid, object uintptr) uintptr {
	return this.impl.QueryInterface(refiid, object)
}

func _ICoreWebView2PrintCompletedHandlerIUnknownAddRef(this *ICoreWebView2PrintCompletedHandler) uintptr {
	return this.impl.AddRef()
}

func _ICoreWebView2PrintCompletedHandlerIUnknownRelease(this *ICoreWebView2PrintCompletedHandler) uintptr {
	return this.impl.Release()
}

func _ICoreWebView2PrintCompletedHandlerInvoke(this *ICoreWebView2PrintCompletedHandler, errorCode uintptr, status COREWEBVIEW2_PRINT_STATUS) uintptr {
	return this.impl.PrintCompleted(errorCode, status)
}

var _ICoreWebView2PrintCompletedHandlerFn = _ICoreWebView2PrintCompletedHandlerVtbl{
	_IUnknownVtbl{
		NewComProc(_ICoreWebView2PrintCompletedHandlerIUnknownQueryInterface),
		NewComProc(_ICoreWebView2PrintCompletedHandlerIUnknownAddRef),
		NewComProc(_ICoreWebView2PrintCompletedHandlerIUnknownRelease),
	},
	NewComProc(_ICoreWebView2PrintCompletedHandlerInvoke),
}

func newICoreWebView2PrintCompletedHandler(impl _ICoreWebView2PrintCompletedHandlerImpl) *ICoreWebView2PrintCompletedHandler {
	return &ICoreWebView2PrintCompletedHandler{
		vtbl: &_ICoreWebView2PrintCompletedHandlerFn,
		impl: impl,
	}
}

type _ICoreWebView2PrintToPdfStreamCompletedHandlerVtbl struct {
	_IUnknownVtbl
	Invoke ComProc
}

type ICoreWebView2PrintToPdfStreamCompletedHandler struct {
	vtbl *_ICoreWebView2PrintToPdfStreamCompletedHandlerVtbl
	impl _ICoreWebView2PrintToPdfStreamCompletedHandlerImpl
}

type _ICoreWebView2PrintToPdfStreamCompletedHandlerImpl interface {
	_IUnknownImpl
	PrintToPdfStreamCompleted(errorCode uintptr, stream uintptr) uintptr
}

func _ICoreWebView2PrintToPdfStreamCompletedHandlerIUnknownQueryInterface(this *ICoreWebView2PrintToPdfStreamCompletedHandler, refiid, object uintptr) uintptr {
	return this.impl.QueryInterface(refiid, object)
}

func _ICoreWebView2PrintToPdfStreamCompletedHandlerIUnknownAddRef(this *ICoreWebView2PrintToPdfStreamCompletedHandler) uintptr {
	return this.impl.AddRef()
}

func _ICoreWebView2PrintToPdfStreamCompletedHandlerIUnknownRelease(this *ICoreWebView2PrintToPdfStreamCompletedHandler) uintptr {
	return this.impl.Release()
}

func _ICoreWebView2PrintToPdfStreamCompletedHandlerInvoke(this *ICoreWebView2PrintToPdfStreamCompletedHandler, errorCode uintptr, stream uintptr) uintptr {
	return this.impl.PrintToPdfStreamCompleted(errorCode, stream)
}

var _ICoreWebView2PrintToPdfStreamCompletedHandlerFn = _ICoreWebView2PrintToPdfStreamCompletedHandlerVtbl{
	_IUnknownVtbl{
		NewComProc(_ICoreWebView2PrintToPdfStreamCompletedHandlerIUnknownQueryInterface),
		NewComProc(_ICoreWebView2PrintToPdfStreamCompletedHandlerIUnknownAddRef),
		NewComProc(_ICoreWebView2PrintToPdfStreamCompletedHandlerIUnknownRelease),
	},
	NewComProc(_ICoreWebView2PrintToPdfStreamCompletedHandlerInvoke),
}

func newICoreWebView2PrintToPdfStreamCompletedHandler(impl _ICoreWebView2PrintToPdfStreamCompletedHandlerImpl) *ICoreWebView2PrintToPdfStreamCompletedHandler {
	return &ICoreWebView2PrintToPdfStreamCompletedHandler{
		vtbl: &_ICoreWebView2PrintToPdfStreamCompletedHandlerFn,
		impl: impl,
	}
}
