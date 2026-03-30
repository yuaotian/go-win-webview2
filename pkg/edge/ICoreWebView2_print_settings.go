//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type ICoreWebView2PrintSettings struct {
	vtbl *iCoreWebView2PrintSettingsVtbl
}

type iCoreWebView2PrintSettingsVtbl struct {
	_IUnknownVtbl
	GetOrientation               ComProc
	PutOrientation               ComProc
	GetScaleFactor               ComProc
	PutScaleFactor               ComProc
	GetPageWidth                 ComProc
	PutPageWidth                 ComProc
	GetPageHeight                ComProc
	PutPageHeight                ComProc
	GetMarginTop                 ComProc
	PutMarginTop                 ComProc
	GetMarginBottom              ComProc
	PutMarginBottom              ComProc
	GetMarginLeft                ComProc
	PutMarginLeft                ComProc
	GetMarginRight               ComProc
	PutMarginRight               ComProc
	GetShouldPrintBackgrounds    ComProc
	PutShouldPrintBackgrounds    ComProc
	GetShouldPrintSelectionOnly  ComProc
	PutShouldPrintSelectionOnly  ComProc
	GetShouldPrintHeaderAndFooter ComProc
	PutShouldPrintHeaderAndFooter ComProc
	GetHeaderTitle               ComProc
	PutHeaderTitle               ComProc
	GetFooterURI                 ComProc
	PutFooterURI                 ComProc
}

type ICoreWebView2PrintSettings2 struct {
	vtbl *iCoreWebView2PrintSettings2Vtbl
}

type iCoreWebView2PrintSettings2Vtbl struct {
	iCoreWebView2PrintSettingsVtbl
	GetPageRanges  ComProc
	PutPageRanges  ComProc
	GetPagesPerSide ComProc
	PutPagesPerSide ComProc
	GetCopies      ComProc
	PutCopies      ComProc
	GetCollation   ComProc
	PutCollation   ComProc
	GetColorMode   ComProc
	PutColorMode   ComProc
	GetDuplex      ComProc
	PutDuplex      ComProc
	GetMediaSize   ComProc
	PutMediaSize   ComProc
	GetPrinterName ComProc
	PutPrinterName ComProc
}

func (i *ICoreWebView2PrintSettings) AddRef() uintptr {
	r, _, _ := i.vtbl.AddRef.Call()
	return r
}

func (i *ICoreWebView2PrintSettings) PutOrientation(orientation COREWEBVIEW2_PRINT_ORIENTATION) error {
	_, _, err := i.vtbl.PutOrientation.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(orientation),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2PrintSettings) PutScaleFactor(scaleFactor float64) error {
	_, _, err := i.vtbl.PutScaleFactor.Call(
		uintptr(unsafe.Pointer(i)),
		*(*uintptr)(unsafe.Pointer(&scaleFactor)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2PrintSettings) PutShouldPrintBackgrounds(enabled bool) error {
	_, _, err := i.vtbl.PutShouldPrintBackgrounds.Call(
		uintptr(unsafe.Pointer(i)),
		boolToInt(enabled),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2PrintSettings) PutShouldPrintHeaderAndFooter(enabled bool) error {
	_, _, err := i.vtbl.PutShouldPrintHeaderAndFooter.Call(
		uintptr(unsafe.Pointer(i)),
		boolToInt(enabled),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2PrintSettings) PutHeaderTitle(value string) error {
	_value, err := windows.UTF16PtrFromString(value)
	if err != nil {
		return err
	}
	_, _, err = i.vtbl.PutHeaderTitle.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_value)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2PrintSettings) PutFooterURI(value string) error {
	_value, err := windows.UTF16PtrFromString(value)
	if err != nil {
		return err
	}
	_, _, err = i.vtbl.PutFooterURI.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_value)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2PrintSettings) GetICoreWebView2PrintSettings2() *ICoreWebView2PrintSettings2 {
	var result *ICoreWebView2PrintSettings2

	iidICoreWebView2PrintSettings2 := NewGUID("{ca7f0e1f-3484-41d1-8c1a-65cd44a63f8d}")
	_, _, _ = i.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2PrintSettings2)),
		uintptr(unsafe.Pointer(&result)),
	)

	return result
}

func (i *ICoreWebView2PrintSettings2) PutPageRanges(value string) error {
	_value, err := windows.UTF16PtrFromString(value)
	if err != nil {
		return err
	}
	_, _, err = i.vtbl.PutPageRanges.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_value)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2PrintSettings2) PutCopies(value int32) error {
	_, _, err := i.vtbl.PutCopies.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(value),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2PrintSettings2) PutPrinterName(value string) error {
	_value, err := windows.UTF16PtrFromString(value)
	if err != nil {
		return err
	}
	_, _, err = i.vtbl.PutPrinterName.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(_value)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}
