//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type ICoreWebView2DOMContentLoadedEventArgs struct {
	vtbl *ICoreWebView2DOMContentLoadedEventArgsVtbl
}

type ICoreWebView2DOMContentLoadedEventArgsVtbl struct {
	_IUnknownVtbl
	GetNavigationId ComProc
}

func (i *ICoreWebView2DOMContentLoadedEventArgs) GetNavigationId() (uint64, error) {
	var navigationID uint64
	_, _, err := i.vtbl.GetNavigationId.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&navigationID)),
	)
	if err != nil && err != windows.ERROR_SUCCESS {
		return 0, err
	}
	return navigationID, nil
}
