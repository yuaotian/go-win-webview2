//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type _ICoreWebView2DevToolsProtocolEventReceiverVtbl struct {
	_IUnknownVtbl
	AddDevToolsProtocolEventReceived    ComProc
	RemoveDevToolsProtocolEventReceived ComProc
}

type ICoreWebView2DevToolsProtocolEventReceiver struct {
	vtbl *_ICoreWebView2DevToolsProtocolEventReceiverVtbl
}

func (i *ICoreWebView2DevToolsProtocolEventReceiver) Release() uintptr {
	r, _, _ := i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))
	return r
}

func (i *ICoreWebView2DevToolsProtocolEventReceiver) AddDevToolsProtocolEventReceived(handler *ICoreWebView2DevToolsProtocolEventReceivedEventHandler, token *_EventRegistrationToken) error {
	hr, _, err := i.vtbl.AddDevToolsProtocolEventReceived.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(handler)),
		uintptr(unsafe.Pointer(token)),
	)
	if int32(hr) < 0 {
		return windows.Errno(hr)
	}
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2DevToolsProtocolEventReceiver) RemoveDevToolsProtocolEventReceived(token _EventRegistrationToken) error {
	hr, _, err := i.vtbl.RemoveDevToolsProtocolEventReceived.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&token)),
	)
	if int32(hr) < 0 {
		return windows.Errno(hr)
	}
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}
