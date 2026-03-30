//go:build windows
// +build windows

package webview2

import (
	"errors"

	"github.com/yuaotian/go-win-webview2/pkg/edge"
)

type PrintDialogKind uint32

const (
	PrintDialogKindBrowser PrintDialogKind = PrintDialogKind(edge.COREWEBVIEW2_PRINT_DIALOG_KIND_BROWSER)
	PrintDialogKindSystem  PrintDialogKind = PrintDialogKind(edge.COREWEBVIEW2_PRINT_DIALOG_KIND_SYSTEM)
)

type PrintOrientation uint32

const (
	PrintOrientationPortrait  PrintOrientation = PrintOrientation(edge.COREWEBVIEW2_PRINT_ORIENTATION_PORTRAIT)
	PrintOrientationLandscape PrintOrientation = PrintOrientation(edge.COREWEBVIEW2_PRINT_ORIENTATION_LANDSCAPE)
)

type PrintStatus uint32

const (
	PrintStatusSucceeded          PrintStatus = PrintStatus(edge.COREWEBVIEW2_PRINT_STATUS_SUCCEEDED)
	PrintStatusPrinterUnavailable PrintStatus = PrintStatus(edge.COREWEBVIEW2_PRINT_STATUS_PRINTER_UNAVAILABLE)
	PrintStatusOtherError         PrintStatus = PrintStatus(edge.COREWEBVIEW2_PRINT_STATUS_OTHER_ERROR)
)

type PrintSettings struct {
	inner  *edge.ICoreWebView2PrintSettings
	inner2 *edge.ICoreWebView2PrintSettings2
}

func (s *PrintSettings) SetOrientation(orientation PrintOrientation) error {
	if s == nil || s.inner == nil {
		return errors.New("print settings is not initialized")
	}
	return s.inner.PutOrientation(edge.COREWEBVIEW2_PRINT_ORIENTATION(orientation))
}

func (s *PrintSettings) SetScaleFactor(scaleFactor float64) error {
	if s == nil || s.inner == nil {
		return errors.New("print settings is not initialized")
	}
	return s.inner.PutScaleFactor(scaleFactor)
}

func (s *PrintSettings) SetShouldPrintBackgrounds(enabled bool) error {
	if s == nil || s.inner == nil {
		return errors.New("print settings is not initialized")
	}
	return s.inner.PutShouldPrintBackgrounds(enabled)
}

func (s *PrintSettings) SetShouldPrintHeaderAndFooter(enabled bool) error {
	if s == nil || s.inner == nil {
		return errors.New("print settings is not initialized")
	}
	return s.inner.PutShouldPrintHeaderAndFooter(enabled)
}

func (s *PrintSettings) SetHeaderTitle(value string) error {
	if s == nil || s.inner == nil {
		return errors.New("print settings is not initialized")
	}
	return s.inner.PutHeaderTitle(value)
}

func (s *PrintSettings) SetFooterURI(value string) error {
	if s == nil || s.inner == nil {
		return errors.New("print settings is not initialized")
	}
	return s.inner.PutFooterURI(value)
}

func (s *PrintSettings) SetPageRanges(value string) error {
	if s == nil || s.inner2 == nil {
		return errors.New("ICoreWebView2PrintSettings2 is not available")
	}
	return s.inner2.PutPageRanges(value)
}

func (s *PrintSettings) SetCopies(value int32) error {
	if s == nil || s.inner2 == nil {
		return errors.New("ICoreWebView2PrintSettings2 is not available")
	}
	return s.inner2.PutCopies(value)
}

func (s *PrintSettings) SetPrinterName(value string) error {
	if s == nil || s.inner2 == nil {
		return errors.New("ICoreWebView2PrintSettings2 is not available")
	}
	return s.inner2.PutPrinterName(value)
}

type PrintToPDFResult struct {
	Error      error
	Successful bool
}

type PrintJobResult struct {
	Error  error
	Status PrintStatus
}

func (w *webview) NewPrintSettings() (*PrintSettings, error) {
	chromium, ok := w.browser.(*edge.Chromium)
	if !ok {
		return nil, errors.New("browser does not support print settings")
	}

	var (
		settings *edge.ICoreWebView2PrintSettings
		err      error
	)
	w.dispatchSync(func() {
		settings, err = chromium.CreatePrintSettings()
	})
	if err != nil {
		return nil, err
	}
	return &PrintSettings{
		inner:  settings,
		inner2: settings.GetICoreWebView2PrintSettings2(),
	}, nil
}

func (w *webview) PrintWithSettings(settings *PrintSettings, callback func(PrintJobResult)) error {
	chromium, ok := w.browser.(*edge.Chromium)
	if !ok {
		return errors.New("browser does not support print")
	}

	var inner *edge.ICoreWebView2PrintSettings
	if settings != nil {
		inner = settings.inner
	}

	var err error
	w.dispatchSync(func() {
		err = chromium.PrintWithSettings(inner, func(callbackErr error, status edge.COREWEBVIEW2_PRINT_STATUS) {
			if callback == nil {
				return
			}
			callback(PrintJobResult{
				Error:  callbackErr,
				Status: PrintStatus(status),
			})
		})
	})
	return err
}

func (w *webview) PrintToPDFWithSettings(path string, settings *PrintSettings, callback func(PrintToPDFResult)) error {
	chromium, ok := w.browser.(*edge.Chromium)
	if !ok {
		return errors.New("browser does not support print to pdf")
	}

	var inner *edge.ICoreWebView2PrintSettings
	if settings != nil {
		inner = settings.inner
	}

	var err error
	w.dispatchSync(func() {
		err = chromium.PrintToPDFWithSettings(path, inner, func(callbackErr error, successful bool) {
			if callback == nil {
				return
			}
			callback(PrintToPDFResult{
				Error:      callbackErr,
				Successful: successful,
			})
		})
	})
	return err
}

func (w *webview) ShowPrintDialogWithKind(kind PrintDialogKind) error {
	chromium, ok := w.browser.(*edge.Chromium)
	if !ok {
		return errors.New("browser does not support print dialog")
	}

	var err error
	w.dispatchSync(func() {
		err = chromium.ShowPrintUI(edge.COREWEBVIEW2_PRINT_DIALOG_KIND(kind))
	})
	return err
}
