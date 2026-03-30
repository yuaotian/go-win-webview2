//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type iCoreWebView2_2Vtbl struct {
	iCoreWebView2Vtbl
	AddWebResourceResponseReceived    ComProc
	RemoveWebResourceResponseReceived ComProc
	NavigateWithWebResourceRequest    ComProc
	AddDomContentLoaded               ComProc
	RemoveDomContentLoaded            ComProc
	GetCookieManager                  ComProc
	GetEnvironment                    ComProc
}

type ICoreWebView2_2 struct {
	vtbl *iCoreWebView2_2Vtbl
}

func (i *ICoreWebView2_2) AddRef() uintptr {
	r, _, _ := i.vtbl.AddRef.Call()
	return r
}

func (i *ICoreWebView2_2) AddDomContentLoaded(handler *ICoreWebView2DOMContentLoadedEventHandler, token *_EventRegistrationToken) error {
	_, _, err := i.vtbl.AddDomContentLoaded.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(handler)),
		uintptr(unsafe.Pointer(token)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2) GetICoreWebView2_2() *ICoreWebView2_2 {
	var result *ICoreWebView2_2

	iidICoreWebView2_2 := NewGUID("{9E8F0CF8-E670-4B5E-B2BC-73E061E3184C}")
	_, _, _ = i.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2_2)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func (e *Chromium) GetICoreWebView2_2() *ICoreWebView2_2 {
	return e.webview.GetICoreWebView2_2()
}
