//go:build windows
// +build windows

package webview2

import (
	"errors"
	"log"

	"github.com/yuaotian/go-win-webview2/pkg/edge"
)

type CapturePreviewFormat uint32

const (
	CapturePreviewFormatPNG  CapturePreviewFormat = CapturePreviewFormat(edge.COREWEBVIEW2_CAPTURE_PREVIEW_IMAGE_FORMAT_PNG)
	CapturePreviewFormatJPEG CapturePreviewFormat = CapturePreviewFormat(edge.COREWEBVIEW2_CAPTURE_PREVIEW_IMAGE_FORMAT_JPEG)
)

type CapturePreviewOptions struct {
	Format CapturePreviewFormat
}

type WebViewDeferral struct {
	inner *edge.ICoreWebView2Deferral
}

func (d *WebViewDeferral) Complete() error {
	if d == nil || d.inner == nil {
		return errors.New("deferral is nil")
	}
	return d.inner.Complete()
}

type NewWindowRequestedEvent struct {
	URI             string
	IsUserInitiated bool

	args  *edge.ICoreWebView2NewWindowRequestedEventArgs
	owner *webview
}

func (e *NewWindowRequestedEvent) SetHandled(handled bool) error {
	if e == nil || e.args == nil {
		return errors.New("new window args is nil")
	}
	return e.args.PutHandled(handled)
}

func (e *NewWindowRequestedEvent) Block() error {
	return e.SetHandled(true)
}

func (e *NewWindowRequestedEvent) OpenInCurrentWindow() error {
	if e == nil || e.owner == nil {
		return errors.New("current webview is nil")
	}
	e.owner.Navigate(e.URI)
	return e.SetHandled(true)
}

func (e *NewWindowRequestedEvent) UseNewWindow(newWindow WebView) error {
	if e == nil || e.args == nil {
		return errors.New("new window args is nil")
	}
	target, ok := newWindow.(*webview)
	if !ok {
		return errors.New("new window must be created by this package")
	}
	chromium, ok := target.browser.(*edge.Chromium)
	if !ok {
		return errors.New("target browser is not Chromium")
	}
	return e.args.PutNewWindow(chromium.CoreWebView2())
}

func (e *NewWindowRequestedEvent) GetDeferral() (*WebViewDeferral, error) {
	if e == nil || e.args == nil {
		return nil, errors.New("new window args is nil")
	}
	deferral, err := e.args.GetDeferral()
	if err != nil {
		return nil, err
	}
	return &WebViewDeferral{inner: deferral}, nil
}

func (w *webview) OnNewWindowRequested(callback func(*NewWindowRequestedEvent)) {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		// 将底层 popup 请求转换成公开事件对象，避免把 COM 细节泄漏到外层。
		chromium.SetNewWindowRequestedCallback(func(_ *edge.ICoreWebView2, args *edge.ICoreWebView2NewWindowRequestedEventArgs) {
			if callback == nil {
				return
			}

			uri, err := args.GetUri()
			if err != nil {
				log.Printf("读取新窗口 URL 失败: %v", err)
				return
			}
			userInitiated, err := args.GetIsUserInitiated()
			if err != nil {
				log.Printf("读取新窗口请求来源失败: %v", err)
				return
			}

			callback(&NewWindowRequestedEvent{
				URI:             uri,
				IsUserInitiated: userInitiated,
				args:            args,
				owner:           w,
			})
		})
	}
}

func (w *webview) OnWindowCloseRequested(callback func()) {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		chromium.SetWindowCloseRequestedCallback(callback)
	}
}

func (w *webview) CallDevToolsProtocolMethod(method string, paramsJSON string, callback func(error, string)) error {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		var err error
		w.dispatchSync(func() {
			err = chromium.CallDevToolsProtocolMethod(method, paramsJSON, callback)
		})
		return err
	}
	return errors.New("browser does not support devtools protocol")
}

func (w *webview) SubscribeDevToolsProtocolEvent(eventName string, handler func(string)) (func(), error) {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		var (
			err         error
			unsubscribe func()
		)
		w.dispatchSync(func() {
			unsubscribe, err = chromium.SubscribeDevToolsProtocolEvent(eventName, handler)
		})
		return unsubscribe, err
	}
	return nil, errors.New("browser does not support devtools protocol events")
}

func (w *webview) CapturePreviewWithOptions(path string, options CapturePreviewOptions) error {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		var err error
		w.dispatchSync(func() {
			err = chromium.CapturePreviewWithFormat(path, edge.COREWEBVIEW2_CAPTURE_PREVIEW_IMAGE_FORMAT(options.Format))
		})
		return err
	}
	return errors.New("browser does not support capture preview")
}
