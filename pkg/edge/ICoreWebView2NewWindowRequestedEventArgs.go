//go:build windows
// +build windows
package edge

// ICoreWebView2NewWindowRequestedEventArgs
type ICoreWebView2NewWindowRequestedEventArgs struct {
    vtbl *iCoreWebView2NewWindowRequestedEventArgsVtbl
}

type iCoreWebView2NewWindowRequestedEventArgsVtbl struct {
    _IUnknownVtbl
    GetUri              ComProc
    PutHandled         ComProc
    GetIsUserInitiated ComProc
    GetDeferral       ComProc
    GetNewWindow      ComProc
    PutNewWindow      ComProc
} 