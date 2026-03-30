//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"github.com/yuaotian/go-win-webview2/internal/w32"
	"golang.org/x/sys/windows"
)

type ICoreWebView2NewWindowRequestedEventArgs struct {
	vtbl *ICoreWebView2NewWindowRequestedEventArgsVtbl
}

type ICoreWebView2NewWindowRequestedEventArgsVtbl struct {
	_IUnknownVtbl
	GetUri             ComProc
	PutNewWindow       ComProc
	GetNewWindow       ComProc
	PutHandled         ComProc
	GetHandled         ComProc
	GetIsUserInitiated ComProc
	GetDeferral        ComProc
	GetWindowFeatures  ComProc
}

func (i *ICoreWebView2NewWindowRequestedEventArgs) GetUri() (string, error) {
	var uri *uint16
	_, _, err := i.vtbl.GetUri.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&uri)),
	)
	if err != windows.ERROR_SUCCESS {
		return "", err
	}
	if uri == nil {
		return "", nil
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(uri))
	return w32.Utf16PtrToString(uri), nil
}

func (i *ICoreWebView2NewWindowRequestedEventArgs) PutNewWindow(newWindow *ICoreWebView2) error {
	_, _, err := i.vtbl.PutNewWindow.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(newWindow)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2NewWindowRequestedEventArgs) PutHandled(handled bool) error {
	_, _, err := i.vtbl.PutHandled.Call(
		uintptr(unsafe.Pointer(i)),
		boolToInt(handled),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2NewWindowRequestedEventArgs) GetIsUserInitiated() (bool, error) {
	var userInitiated int32
	_, _, err := i.vtbl.GetIsUserInitiated.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&userInitiated)),
	)
	if err != windows.ERROR_SUCCESS {
		return false, err
	}
	return userInitiated != 0, nil
}

func (i *ICoreWebView2NewWindowRequestedEventArgs) GetDeferral() (*ICoreWebView2Deferral, error) {
	var deferral *ICoreWebView2Deferral
	_, _, err := i.vtbl.GetDeferral.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&deferral)),
	)
	if err != windows.ERROR_SUCCESS {
		return nil, err
	}
	return deferral, nil
}
