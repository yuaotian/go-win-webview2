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
	_path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	hr, _, err := i.vtbl.PrintToPdf.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_path)),
		0,
		0,
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
