//go:build windows
// +build windows

package webview2

import (
	"errors"
	"fmt"
	"log"

	"github.com/yuaotian/go-win-webview2/pkg/edge"
)

type WebResourceContext uint32

const (
	WebResourceContextAll                WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_ALL)
	WebResourceContextDocument           WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_DOCUMENT)
	WebResourceContextStylesheet         WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_STYLESHEET)
	WebResourceContextImage              WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_IMAGE)
	WebResourceContextMedia              WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_MEDIA)
	WebResourceContextFont               WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_FONT)
	WebResourceContextScript             WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_SCRIPT)
	WebResourceContextXMLHttpRequest     WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_XML_HTTP_REQUEST)
	WebResourceContextFetch              WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_FETCH)
	WebResourceContextTextTrack          WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_TEXT_TRACK)
	WebResourceContextEventSource        WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_EVENT_SOURCE)
	WebResourceContextWebSocket          WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_WEBSOCKET)
	WebResourceContextManifest           WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_MANIFEST)
	WebResourceContextSignedExchange     WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_SIGNED_EXCHANGE)
	WebResourceContextPing               WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_PING)
	WebResourceContextCSPViolationReport WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_CSP_VIOLATION_REPORT)
	WebResourceContextOther              WebResourceContext = WebResourceContext(edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_OTHER)
)

type HostResourceAccessKind uint32

const (
	HostResourceAccessKindDeny     HostResourceAccessKind = HostResourceAccessKind(edge.COREWEBVIEW2_HOST_RESOURCE_ACCESS_KIND_DENY)
	HostResourceAccessKindAllow    HostResourceAccessKind = HostResourceAccessKind(edge.COREWEBVIEW2_HOST_RESOURCE_ACCESS_KIND_ALLOW)
	HostResourceAccessKindDenyCORS HostResourceAccessKind = HostResourceAccessKind(edge.COREWEBVIEW2_HOST_RESOURCE_ACCESS_KIND_DENY_CORS)
)

type DOMContentLoadedEvent struct {
	NavigationID uint64
}

type WebResourceResponse struct {
	Content      []byte
	FilePath     string
	Stream       *Stream
	StatusCode   int
	ReasonPhrase string
	Headers      string
}

type WebResourceRequestedEvent struct {
	URI     string
	Method  string
	Context WebResourceContext

	args *edge.ICoreWebView2WebResourceRequestedEventArgs
	env  *edge.ICoreWebView2Environment
}

func (e *WebResourceRequestedEvent) SetResponse(response WebResourceResponse) error {
	if e == nil || e.args == nil || e.env == nil {
		return errors.New("web resource event is not initialized")
	}

	sourceCount := 0
	if len(response.Content) > 0 {
		sourceCount++
	}
	if response.FilePath != "" {
		sourceCount++
	}
	if response.Stream != nil {
		sourceCount++
	}
	if sourceCount > 1 {
		return errors.New("web resource response content, file path and stream can only set one source")
	}

	var (
		err                 error
		webResourceResponse *edge.ICoreWebView2WebResourceResponse
	)
	if response.FilePath != "" {
		// 大文件/多媒体优先走文件 IStream，避免强制拷入 []byte。
		webResourceResponse, err = e.env.CreateWebResourceResponseFromFile(response.FilePath, response.StatusCode, response.ReasonPhrase, response.Headers)
	} else if response.Stream != nil {
		webResourceResponse, err = e.env.CreateWebResourceResponseFromStream(response.Stream.cloneInner(), response.StatusCode, response.ReasonPhrase, response.Headers)
	} else {
		webResourceResponse, err = e.env.CreateWebResourceResponse(response.Content, response.StatusCode, response.ReasonPhrase, response.Headers)
	}
	if err != nil {
		return fmt.Errorf("create web resource response failed: %w", err)
	}
	if err := e.args.PutResponse(webResourceResponse); err != nil {
		return fmt.Errorf("put web resource response failed: %w", err)
	}
	return nil
}

func (e *WebResourceRequestedEvent) SetFileResponse(path string, statusCode int, reasonPhrase string, headers string) error {
	return e.SetResponse(WebResourceResponse{
		FilePath:     path,
		StatusCode:   statusCode,
		ReasonPhrase: reasonPhrase,
		Headers:      headers,
	})
}

func (e *WebResourceRequestedEvent) SetStreamResponse(stream *Stream, statusCode int, reasonPhrase string, headers string) error {
	return e.SetResponse(WebResourceResponse{
		Stream:       stream,
		StatusCode:   statusCode,
		ReasonPhrase: reasonPhrase,
		Headers:      headers,
	})
}

func (e *WebResourceRequestedEvent) GetDeferral() (*WebViewDeferral, error) {
	if e == nil || e.args == nil {
		return nil, errors.New("web resource event is not initialized")
	}
	deferral, err := e.args.GetDeferral()
	if err != nil {
		return nil, err
	}
	return &WebViewDeferral{inner: deferral}, nil
}

func (w *webview) AddWebResourceRequestedFilter(uri string, context WebResourceContext) error {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		chromium.AddWebResourceRequestedFilter(uri, edge.COREWEBVIEW2_WEB_RESOURCE_CONTEXT(context))
		return nil
	}
	return errors.New("browser does not support web resource filters")
}

func (w *webview) OnWebResourceRequested(callback func(*WebResourceRequestedEvent)) {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		chromium.WebResourceRequestedCallback = func(request *edge.ICoreWebView2WebResourceRequest, args *edge.ICoreWebView2WebResourceRequestedEventArgs) {
			if callback == nil {
				return
			}

			uri, err := request.GetUri()
			if err != nil {
				log.Printf("读取请求 URI 失败: %v", err)
				return
			}
			method, err := request.GetMethod()
			if err != nil {
				log.Printf("读取请求 Method 失败: %v", err)
				return
			}
			context, err := args.GetResourceContext()
			if err != nil {
				log.Printf("读取请求 Context 失败: %v", err)
				return
			}

			callback(&WebResourceRequestedEvent{
				URI:     uri,
				Method:  method,
				Context: WebResourceContext(context),
				args:    args,
				env:     chromium.Environment(),
			})
		}
	}
}

func (w *webview) SetVirtualHostNameToFolderMapping(hostName string, folderPath string, accessKind HostResourceAccessKind) error {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		var err error
		// 扩展接口和实际调用都放到主线程里，避免跨线程复用 COM 指针。
		w.dispatchSync(func() {
			webview2_3 := chromium.GetICoreWebView2_3()
			if webview2_3 == nil {
				err = errors.New("ICoreWebView2_3 is not available")
				return
			}
			err = webview2_3.SetVirtualHostNameToFolderMapping(hostName, folderPath, edge.COREWEBVIEW2_HOST_RESOURCE_ACCESS_KIND(accessKind))
		})
		return err
	}
	return errors.New("browser does not support virtual host mapping")
}

func (w *webview) ClearVirtualHostNameToFolderMapping(hostName string) error {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		var err error
		// 扩展接口和实际调用都放到主线程里，避免跨线程复用 COM 指针。
		w.dispatchSync(func() {
			webview2_3 := chromium.GetICoreWebView2_3()
			if webview2_3 == nil {
				err = errors.New("ICoreWebView2_3 is not available")
				return
			}
			err = webview2_3.ClearVirtualHostNameToFolderMapping(hostName)
		})
		return err
	}
	return errors.New("browser does not support virtual host mapping")
}

func (w *webview) OnDOMContentLoaded(callback func(DOMContentLoadedEvent)) {
	if chromium, ok := w.browser.(*edge.Chromium); ok {
		chromium.SetDOMContentLoadedCallback(func(_ *edge.ICoreWebView2, args *edge.ICoreWebView2DOMContentLoadedEventArgs) {
			if callback == nil {
				return
			}

			navigationID, err := args.GetNavigationId()
			if err != nil {
				log.Printf("读取 DOMContentLoaded navigation id 失败: %v", err)
				return
			}
			callback(DOMContentLoadedEvent{NavigationID: navigationID})
		})
	}
}
