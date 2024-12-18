//go:build windows
// +build windows

package w32

import (
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ole32               = windows.NewLazySystemDLL("ole32")
	Ole32CoInitializeEx = ole32.NewProc("CoInitializeEx")

	kernel32                   = windows.NewLazySystemDLL("kernel32")
	Kernel32GetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")

	shlwapi                  = windows.NewLazySystemDLL("shlwapi")
	shlwapiSHCreateMemStream = shlwapi.NewProc("SHCreateMemStream")

	user32                   = windows.NewLazySystemDLL("user32")
	User32LoadImageW         = user32.NewProc("LoadImageW")
	User32GetSystemMetrics   = user32.NewProc("GetSystemMetrics")
	User32RegisterClassExW   = user32.NewProc("RegisterClassExW")
	User32CreateWindowExW    = user32.NewProc("CreateWindowExW")
	User32DestroyWindow      = user32.NewProc("DestroyWindow")
	User32ShowWindow         = user32.NewProc("ShowWindow")
	User32UpdateWindow       = user32.NewProc("UpdateWindow")
	User32SetFocus           = user32.NewProc("SetFocus")
	User32GetMessageW        = user32.NewProc("GetMessageW")
	User32TranslateMessage   = user32.NewProc("TranslateMessage")
	User32DispatchMessageW   = user32.NewProc("DispatchMessageW")
	User32DefWindowProcW     = user32.NewProc("DefWindowProcW")
	User32GetClientRect      = user32.NewProc("GetClientRect")
	User32PostQuitMessage    = user32.NewProc("PostQuitMessage")
	User32PostMessageW       = user32.NewProc("PostMessageW")
	User32SetWindowTextW     = user32.NewProc("SetWindowTextW")
	User32PostThreadMessageW = user32.NewProc("PostThreadMessageW")
	User32GetWindowLongPtrW  = user32.NewProc("GetWindowLongPtrW")
	User32SetWindowLongPtrW  = user32.NewProc("SetWindowLongPtrW")
	User32AdjustWindowRect   = user32.NewProc("AdjustWindowRect")
	User32SetWindowPos       = user32.NewProc("SetWindowPos")
	User32IsDialogMessage    = user32.NewProc("IsDialogMessage")
	User32GetAncestor        = user32.NewProc("GetAncestor")
	User32GetWindowRect      = user32.NewProc("GetWindowRect")
	User32SendMessageW       = user32.NewProc("SendMessageW")
	User32RegisterHotKey   = user32.NewProc("RegisterHotKey")
	User32UnregisterHotKey = user32.NewProc("UnregisterHotKey")
	User32SystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
	User32SetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
)

const (
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
)

const (
	CW_USEDEFAULT = 0x80000000
)

const (
	LR_DEFAULTCOLOR     = 0x0000
	LR_MONOCHROME       = 0x0001
	LR_LOADFROMFILE     = 0x0010
	LR_LOADTRANSPARENT  = 0x0020
	LR_DEFAULTSIZE      = 0x0040
	LR_VGACOLOR         = 0x0080
	LR_LOADMAP3DCOLORS  = 0x1000
	LR_CREATEDIBSECTION = 0x2000
	LR_SHARED           = 0x8000
)

const (
	SystemMetricsCxIcon = 11
	SystemMetricsCyIcon = 12
)

const (
	SW_SHOW     = 5
	SW_MINIMIZE = 6
	SW_MAXIMIZE = 3
	SW_RESTORE  = 9
	SW_HIDE     = 0
)

const (
	HWND_TOP       = 0
	HWND_TOPMOST   = -1
	HWND_NOTOPMOST = -2

	SWP_NOMOVE       = 0x0002
	SWP_NOSIZE       = 0x0001
	SWP_SHOWWINDOW   = 0x0040
	SWP_NOZORDER     = 0x0004
	SWP_NOACTIVATE   = 0x0010
	SWP_FRAMECHANGED = 0x0020
)

const (
	WMDestroy       = 0x0002
	WMMove          = 0x0003
	WMSize          = 0x0005
	WMActivate      = 0x0006
	WMClose         = 0x0010
	WMQuit          = 0x0012
	WMGetMinMaxInfo = 0x0024
	WMNCLButtonDown = 0x00A1
	WMMoving        = 0x0216
	WMApp           = 0x8000
)

const (
	GAParent    = 1
	GARoot      = 2
	GARootOwner = 3
)

const (
	GWLStyle = ^(16 - 1)
)

const (
	WSOverlapped       = 0x00000000
	WSMaximizeBox      = 0x00010000
	WSThickFrame       = 0x00040000
	WSCaption          = 0x00C00000
	WSSysMenu          = 0x00080000
	WSMinimizeBox      = 0x00020000
	WSOverlappedWindow = (WSOverlapped | WSCaption | WSSysMenu | WSThickFrame | WSMinimizeBox | WSMaximizeBox)
)

const (
	WAInactive    = 0
	WAActive      = 1
	WAActiveClick = 2
)

const (
	WSPopup   = 0x80000000
	WSVisible = 0x10000000
)

const (
	HTCaption     = 2
	HTLeft        = 10
	HTRight       = 11
	HTTop         = 12
	HTTopLeft     = 13
	HTTopRight    = 14
	HTBottom      = 15
	HTBottomLeft  = 16
	HTBottomRight = 17
)

const (
	WMLButtonDown = 0x0201
	WMNCHitTest   = 0x0084
)

const (
	// Modifiers
	MOD_ALT      = 0x0001
	MOD_CONTROL  = 0x0002
	MOD_SHIFT    = 0x0004
	MOD_WIN      = 0x0008
	MOD_NOREPEAT = 0x4000

	// Messages
	WMHotKey = 0x0312
)

const (
	VK_ESCAPE = 0x1B
	VK_SPACE  = 0x20
	VK_TAB    = 0x09
	VK_F1     = 0x70
	VK_F2     = 0x71
	VK_F3     = 0x72
	VK_F4     = 0x73
	VK_F5     = 0x74
	VK_F6     = 0x75
	VK_F7     = 0x76
	VK_F8     = 0x77
	VK_F9     = 0x78
	VK_F10    = 0x79
	VK_F11    = 0x7A
	VK_F12    = 0x7B
)

const (
	SPI_GETWORKAREA = 0x0030
)

const (
	GWL_EXSTYLE = -20
	WS_EX_LAYERED = 0x00080000
	LWA_ALPHA = 0x00000002
)

type WndClassExW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CnClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

type Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type MinMaxInfo struct {
	PtReserved     Point
	PtMaxSize      Point
	PtMaxPosition  Point
	PtMinTrackSize Point
	PtMaxTrackSize Point
}

type Point struct {
	X, Y int32
}

type Msg struct {
	Hwnd     windows.Handle
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       Point
	LPrivate uint32
}

func Utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	// Find NUL terminator.
	end := unsafe.Pointer(p)
	n := 0
	for *(*uint16)(end) != 0 {
		end = unsafe.Pointer(uintptr(end) + unsafe.Sizeof(*p))
		n++
	}
	s := (*[(1 << 30) - 1]uint16)(unsafe.Pointer(p))[:n:n]
	return string(utf16.Decode(s))
}

func SHCreateMemStream(data []byte) (uintptr, error) {
	ret, _, err := shlwapiSHCreateMemStream.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	if ret == 0 {
		return 0, err
	}

	return ret, nil
}
