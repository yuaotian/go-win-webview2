//go:build windows
// +build windows
package edge

import "golang.org/x/sys/windows"

var (
    IID_ICoreWebView2NavigationStartingEventHandler = windows.GUID{
        Data1: 0x9adbe429,
        Data2: 0x37d6,
        Data3: 0x4c40,
        Data4: [8]byte{0x93, 0xcb, 0xcf, 0x95, 0x66, 0x4a, 0x13, 0x41},
    }

    IID_ICoreWebView2NavigationStartingEventArgs = windows.GUID{
        Data1: 0x5b495469,
        Data2: 0xe119,
        Data3: 0x438a,
        Data4: [8]byte{0x9b, 0x18, 0x73, 0xd2, 0x0a, 0x49, 0x2a, 0x5b},
    }
) 