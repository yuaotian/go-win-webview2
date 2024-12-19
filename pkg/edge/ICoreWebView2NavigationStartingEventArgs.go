//go:build windows
// +build windows
package edge

import (
	"syscall"
	"unsafe"
)

type ICoreWebView2NavigationStartingEventArgs struct {
	vtbl *ICoreWebView2NavigationStartingEventArgsVtbl
}

type ICoreWebView2NavigationStartingEventArgsVtbl struct {
	QueryInterface     uintptr
	AddRef             uintptr
	Release            uintptr
	GetUri             uintptr
	GetNavigationId    uintptr
	GetIsUserInitiated uintptr
	GetIsRedirected    uintptr
	GetRequestHeaders  uintptr
	GetCancel          uintptr
	PutCancel          uintptr
}

func (i *ICoreWebView2NavigationStartingEventArgs) GetUri(uri **uint16) uintptr {
	ret, _, _ := syscall.SyscallN(
		i.vtbl.GetUri,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(uri)),
	)
	return ret
}

func (i *ICoreWebView2NavigationStartingEventArgs) GetNavigationId(navigationId *uint64) uintptr {
	ret, _, _ := syscall.SyscallN(
		i.vtbl.GetNavigationId,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(navigationId)),
	)
	return ret
}
