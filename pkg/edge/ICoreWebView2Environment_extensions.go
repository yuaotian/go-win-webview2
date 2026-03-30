//go:build windows
// +build windows

package edge

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/yuaotian/go-win-webview2/internal/w32"
	"golang.org/x/sys/windows"
)

type iCoreWebView2Environment2Vtbl struct {
	iCoreWebView2EnvironmentVtbl
	CreateWebResourceRequest ComProc
}

type iCoreWebView2Environment3Vtbl struct {
	iCoreWebView2Environment2Vtbl
	CreateCoreWebView2CompositionController ComProc
	CreateCoreWebView2PointerInfo           ComProc
}

type iCoreWebView2Environment4Vtbl struct {
	iCoreWebView2Environment3Vtbl
	GetAutomationProviderForWindow ComProc
}

type iCoreWebView2Environment5Vtbl struct {
	iCoreWebView2Environment4Vtbl
	AddBrowserProcessExited    ComProc
	RemoveBrowserProcessExited ComProc
}

type iCoreWebView2Environment6Vtbl struct {
	iCoreWebView2Environment5Vtbl
	CreatePrintSettings ComProc
}

type iCoreWebView2Environment7Vtbl struct {
	iCoreWebView2Environment6Vtbl
	GetUserDataFolder ComProc
}

type iCoreWebView2Environment8Vtbl struct {
	iCoreWebView2Environment7Vtbl
	AddProcessInfosChanged    ComProc
	RemoveProcessInfosChanged ComProc
	GetProcessInfos           ComProc
}

type iCoreWebView2Environment9Vtbl struct {
	iCoreWebView2Environment8Vtbl
	CreateContextMenuItem ComProc
}

type ICoreWebView2Environment6 struct {
	vtbl *iCoreWebView2Environment6Vtbl
}

type ICoreWebView2Environment9 struct {
	vtbl *iCoreWebView2Environment9Vtbl
}

func (e *ICoreWebView2Environment6) CreatePrintSettings() (*ICoreWebView2PrintSettings, error) {
	var settings *ICoreWebView2PrintSettings
	hr, _, err := e.vtbl.CreatePrintSettings.Call(
		uintptr(unsafe.Pointer(e)),
		uintptr(unsafe.Pointer(&settings)),
	)
	if int32(hr) < 0 {
		return nil, fmt.Errorf("create print settings failed: %08x", hr)
	}
	if err != windows.ERROR_SUCCESS {
		return nil, err
	}
	if settings == nil {
		return nil, errors.New("create print settings returned nil")
	}
	return settings, nil
}

func (e *ICoreWebView2Environment9) CreateContextMenuItem(label string, iconStream *w32.IUnknown, kind COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND) (*ICoreWebView2ContextMenuItem, error) {
	_label, err := windows.UTF16PtrFromString(label)
	if err != nil {
		return nil, err
	}

	var item *ICoreWebView2ContextMenuItem
	hr, _, err := e.vtbl.CreateContextMenuItem.Call(
		uintptr(unsafe.Pointer(e)),
		uintptr(unsafe.Pointer(_label)),
		uintptr(unsafe.Pointer(iconStream)),
		uintptr(kind),
		uintptr(unsafe.Pointer(&item)),
	)
	if int32(hr) < 0 {
		return nil, fmt.Errorf("create context menu item failed: %08x", hr)
	}
	if err != windows.ERROR_SUCCESS {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("create context menu item returned nil")
	}
	return item, nil
}

func (e *ICoreWebView2Environment) GetICoreWebView2Environment6() *ICoreWebView2Environment6 {
	var result *ICoreWebView2Environment6

	iidICoreWebView2Environment6 := NewGUID("{e59ee362-acbd-4857-9a8e-d3644d9459a9}")
	_, _, _ = e.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(e)),
		uintptr(unsafe.Pointer(iidICoreWebView2Environment6)),
		uintptr(unsafe.Pointer(&result)),
	)

	return result
}

func (e *ICoreWebView2Environment) GetICoreWebView2Environment9() *ICoreWebView2Environment9 {
	var result *ICoreWebView2Environment9

	iidICoreWebView2Environment9 := NewGUID("{f06f41bf-4b5a-49d8-b9f6-fa16cd29f274}")
	_, _, _ = e.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(e)),
		uintptr(unsafe.Pointer(iidICoreWebView2Environment9)),
		uintptr(unsafe.Pointer(&result)),
	)

	return result
}

func (e *Chromium) GetICoreWebView2Environment6() *ICoreWebView2Environment6 {
	if e.environment == nil {
		return nil
	}
	return e.environment.GetICoreWebView2Environment6()
}

func (e *Chromium) GetICoreWebView2Environment9() *ICoreWebView2Environment9 {
	if e.environment == nil {
		return nil
	}
	return e.environment.GetICoreWebView2Environment9()
}
