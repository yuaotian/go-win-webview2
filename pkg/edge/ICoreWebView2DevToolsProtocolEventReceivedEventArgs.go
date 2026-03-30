//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"github.com/yuaotian/go-win-webview2/internal/w32"
	"golang.org/x/sys/windows"
)

type ICoreWebView2DevToolsProtocolEventReceivedEventArgs struct {
	vtbl *ICoreWebView2DevToolsProtocolEventReceivedEventArgsVtbl
}

type ICoreWebView2DevToolsProtocolEventReceivedEventArgsVtbl struct {
	_IUnknownVtbl
	GetParameterObjectAsJson ComProc
}

func (i *ICoreWebView2DevToolsProtocolEventReceivedEventArgs) GetParameterObjectAsJson() (string, error) {
	var jsonText *uint16
	_, _, err := i.vtbl.GetParameterObjectAsJson.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&jsonText)),
	)
	if err != windows.ERROR_SUCCESS {
		return "", err
	}
	if jsonText == nil {
		return "", nil
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(jsonText))
	return w32.Utf16PtrToString(jsonText), nil
}
