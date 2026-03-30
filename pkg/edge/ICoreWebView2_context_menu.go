//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"github.com/yuaotian/go-win-webview2/internal/w32"
	"golang.org/x/sys/windows"
)

type ICoreWebView2_11 struct {
	vtbl *iCoreWebView2_11Vtbl
}

func (i *ICoreWebView2_11) AddContextMenuRequested(eventHandler *ICoreWebView2ContextMenuRequestedEventHandler, token *_EventRegistrationToken) error {
	_, _, err := i.vtbl.AddContextMenuRequested.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(token)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2) GetICoreWebView2_11() *ICoreWebView2_11 {
	var result *ICoreWebView2_11

	iidICoreWebView2_11 := NewGUID("{0be78e56-c193-4051-b943-23b460c08bdb}")
	_, _, _ = i.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2_11)),
		uintptr(unsafe.Pointer(&result)),
	)

	return result
}

func (e *Chromium) GetICoreWebView2_11() *ICoreWebView2_11 {
	return e.webview.GetICoreWebView2_11()
}

type ICoreWebView2ContextMenuRequestedEventArgs struct {
	vtbl *ICoreWebView2ContextMenuRequestedEventArgsVtbl
}

type ICoreWebView2ContextMenuRequestedEventArgsVtbl struct {
	_IUnknownVtbl
	GetMenuItems         ComProc
	GetContextMenuTarget ComProc
	GetLocation          ComProc
	PutSelectedCommandId ComProc
	GetSelectedCommandId ComProc
	PutHandled           ComProc
	GetHandled           ComProc
	GetDeferral          ComProc
}

func (i *ICoreWebView2ContextMenuRequestedEventArgs) GetMenuItems() (*ICoreWebView2ContextMenuItemCollection, error) {
	var items *ICoreWebView2ContextMenuItemCollection
	_, _, err := i.vtbl.GetMenuItems.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&items)),
	)
	if err != windows.ERROR_SUCCESS {
		return nil, err
	}
	return items, nil
}

func (i *ICoreWebView2ContextMenuRequestedEventArgs) GetContextMenuTarget() (*ICoreWebView2ContextMenuTarget, error) {
	var target *ICoreWebView2ContextMenuTarget
	_, _, err := i.vtbl.GetContextMenuTarget.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&target)),
	)
	if err != windows.ERROR_SUCCESS {
		return nil, err
	}
	return target, nil
}

func (i *ICoreWebView2ContextMenuRequestedEventArgs) GetLocation() (w32.Point, error) {
	var point w32.Point
	_, _, err := i.vtbl.GetLocation.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&point)),
	)
	if err != windows.ERROR_SUCCESS {
		return w32.Point{}, err
	}
	return point, nil
}

func (i *ICoreWebView2ContextMenuRequestedEventArgs) PutSelectedCommandId(commandID int32) error {
	_, _, err := i.vtbl.PutSelectedCommandId.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(commandID),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2ContextMenuRequestedEventArgs) GetSelectedCommandId() (int32, error) {
	var commandID int32
	_, _, err := i.vtbl.GetSelectedCommandId.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&commandID)),
	)
	if err != windows.ERROR_SUCCESS {
		return 0, err
	}
	return commandID, nil
}

func (i *ICoreWebView2ContextMenuRequestedEventArgs) PutHandled(handled bool) error {
	_, _, err := i.vtbl.PutHandled.Call(
		uintptr(unsafe.Pointer(i)),
		boolToInt(handled),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2ContextMenuRequestedEventArgs) GetHandled() (bool, error) {
	var handled int32
	_, _, err := i.vtbl.GetHandled.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&handled)),
	)
	if err != windows.ERROR_SUCCESS {
		return false, err
	}
	return handled != 0, nil
}

func (i *ICoreWebView2ContextMenuRequestedEventArgs) GetDeferral() (*ICoreWebView2Deferral, error) {
	var deferral *ICoreWebView2Deferral
	_, _, err := i.vtbl.GetDeferral.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&deferral)),
	)
	if err != windows.ERROR_SUCCESS {
		return nil, err
	}
	return deferral, nil
}

type _ICoreWebView2ContextMenuRequestedEventHandlerVtbl struct {
	_IUnknownVtbl
	Invoke ComProc
}

type ICoreWebView2ContextMenuRequestedEventHandler struct {
	vtbl *_ICoreWebView2ContextMenuRequestedEventHandlerVtbl
	impl _ICoreWebView2ContextMenuRequestedEventHandlerImpl
}

type _ICoreWebView2ContextMenuRequestedEventHandlerImpl interface {
	_IUnknownImpl
	ContextMenuRequested(sender *ICoreWebView2, args *ICoreWebView2ContextMenuRequestedEventArgs) uintptr
}

func _ICoreWebView2ContextMenuRequestedEventHandlerIUnknownQueryInterface(this *ICoreWebView2ContextMenuRequestedEventHandler, refiid, object uintptr) uintptr {
	return this.impl.QueryInterface(refiid, object)
}

func _ICoreWebView2ContextMenuRequestedEventHandlerIUnknownAddRef(this *ICoreWebView2ContextMenuRequestedEventHandler) uintptr {
	return this.impl.AddRef()
}

func _ICoreWebView2ContextMenuRequestedEventHandlerIUnknownRelease(this *ICoreWebView2ContextMenuRequestedEventHandler) uintptr {
	return this.impl.Release()
}

func _ICoreWebView2ContextMenuRequestedEventHandlerInvoke(this *ICoreWebView2ContextMenuRequestedEventHandler, sender *ICoreWebView2, args *ICoreWebView2ContextMenuRequestedEventArgs) uintptr {
	return this.impl.ContextMenuRequested(sender, args)
}

var _ICoreWebView2ContextMenuRequestedEventHandlerFn = _ICoreWebView2ContextMenuRequestedEventHandlerVtbl{
	_IUnknownVtbl{
		NewComProc(_ICoreWebView2ContextMenuRequestedEventHandlerIUnknownQueryInterface),
		NewComProc(_ICoreWebView2ContextMenuRequestedEventHandlerIUnknownAddRef),
		NewComProc(_ICoreWebView2ContextMenuRequestedEventHandlerIUnknownRelease),
	},
	NewComProc(_ICoreWebView2ContextMenuRequestedEventHandlerInvoke),
}

func newICoreWebView2ContextMenuRequestedEventHandler(impl _ICoreWebView2ContextMenuRequestedEventHandlerImpl) *ICoreWebView2ContextMenuRequestedEventHandler {
	return &ICoreWebView2ContextMenuRequestedEventHandler{
		vtbl: &_ICoreWebView2ContextMenuRequestedEventHandlerFn,
		impl: impl,
	}
}

type ICoreWebView2ContextMenuItemCollection struct {
	vtbl *ICoreWebView2ContextMenuItemCollectionVtbl
}

type ICoreWebView2ContextMenuItemCollectionVtbl struct {
	_IUnknownVtbl
	GetCount           ComProc
	GetValueAtIndex    ComProc
	RemoveValueAtIndex ComProc
	InsertValueAtIndex ComProc
}

func (i *ICoreWebView2ContextMenuItemCollection) GetCount() (uint32, error) {
	var count uint32
	_, _, err := i.vtbl.GetCount.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&count)),
	)
	if err != windows.ERROR_SUCCESS {
		return 0, err
	}
	return count, nil
}

func (i *ICoreWebView2ContextMenuItemCollection) GetValueAtIndex(index uint32) (*ICoreWebView2ContextMenuItem, error) {
	var item *ICoreWebView2ContextMenuItem
	_, _, err := i.vtbl.GetValueAtIndex.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(index),
		uintptr(unsafe.Pointer(&item)),
	)
	if err != windows.ERROR_SUCCESS {
		return nil, err
	}
	return item, nil
}

func (i *ICoreWebView2ContextMenuItemCollection) RemoveValueAtIndex(index uint32) error {
	_, _, err := i.vtbl.RemoveValueAtIndex.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(index),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2ContextMenuItemCollection) InsertValueAtIndex(index uint32, item *ICoreWebView2ContextMenuItem) error {
	_, _, err := i.vtbl.InsertValueAtIndex.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(index),
		uintptr(unsafe.Pointer(item)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

type ICoreWebView2ContextMenuItem struct {
	vtbl *ICoreWebView2ContextMenuItemVtbl
}

type ICoreWebView2ContextMenuItemVtbl struct {
	_IUnknownVtbl
	GetName                     ComProc
	GetLabel                    ComProc
	GetCommandId                ComProc
	GetShortcutKeyDescription   ComProc
	GetIcon                     ComProc
	GetKind                     ComProc
	PutIsEnabled                ComProc
	GetIsEnabled                ComProc
	PutIsChecked                ComProc
	GetIsChecked                ComProc
	GetChildren                 ComProc
	AddCustomItemSelected       ComProc
	RemoveCustomItemSelected    ComProc
}

func (i *ICoreWebView2ContextMenuItem) AddRef() uintptr {
	r, _, _ := i.vtbl.AddRef.Call()
	return r
}

func (i *ICoreWebView2ContextMenuItem) GetName() (string, error) {
	return contextMenuCoTaskMemString(i.vtbl.GetName, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuItem) GetLabel() (string, error) {
	return contextMenuCoTaskMemString(i.vtbl.GetLabel, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuItem) GetCommandId() (int32, error) {
	var commandID int32
	_, _, err := i.vtbl.GetCommandId.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&commandID)),
	)
	if err != windows.ERROR_SUCCESS {
		return 0, err
	}
	return commandID, nil
}

func (i *ICoreWebView2ContextMenuItem) GetKind() (COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND, error) {
	var kind COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND
	_, _, err := i.vtbl.GetKind.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&kind)),
	)
	if err != windows.ERROR_SUCCESS {
		return 0, err
	}
	return kind, nil
}

func (i *ICoreWebView2ContextMenuItem) PutIsEnabled(enabled bool) error {
	_, _, err := i.vtbl.PutIsEnabled.Call(
		uintptr(unsafe.Pointer(i)),
		boolToInt(enabled),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2ContextMenuItem) GetIsEnabled() (bool, error) {
	var enabled int32
	_, _, err := i.vtbl.GetIsEnabled.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&enabled)),
	)
	if err != windows.ERROR_SUCCESS {
		return false, err
	}
	return enabled != 0, nil
}

func (i *ICoreWebView2ContextMenuItem) PutIsChecked(checked bool) error {
	_, _, err := i.vtbl.PutIsChecked.Call(
		uintptr(unsafe.Pointer(i)),
		boolToInt(checked),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func (i *ICoreWebView2ContextMenuItem) GetIsChecked() (bool, error) {
	var checked int32
	_, _, err := i.vtbl.GetIsChecked.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&checked)),
	)
	if err != windows.ERROR_SUCCESS {
		return false, err
	}
	return checked != 0, nil
}

func (i *ICoreWebView2ContextMenuItem) GetChildren() (*ICoreWebView2ContextMenuItemCollection, error) {
	var children *ICoreWebView2ContextMenuItemCollection
	_, _, err := i.vtbl.GetChildren.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&children)),
	)
	if err != windows.ERROR_SUCCESS {
		return nil, err
	}
	return children, nil
}

func (i *ICoreWebView2ContextMenuItem) AddCustomItemSelected(eventHandler *ICoreWebView2CustomItemSelectedEventHandler, token *_EventRegistrationToken) error {
	_, _, err := i.vtbl.AddCustomItemSelected.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(eventHandler)),
		uintptr(unsafe.Pointer(token)),
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

type ICoreWebView2ContextMenuTarget struct {
	vtbl *ICoreWebView2ContextMenuTargetVtbl
}

type ICoreWebView2ContextMenuTargetVtbl struct {
	_IUnknownVtbl
	GetKind                   ComProc
	GetIsEditable             ComProc
	GetIsRequestedForMainFrame ComProc
	GetPageURI                ComProc
	GetFrameURI               ComProc
	GetHasLinkURI             ComProc
	GetLinkURI                ComProc
	GetHasLinkText            ComProc
	GetLinkText               ComProc
	GetHasSourceURI           ComProc
	GetSourceURI              ComProc
	GetHasSelection           ComProc
	GetSelectionText          ComProc
}

func (i *ICoreWebView2ContextMenuTarget) GetKind() (COREWEBVIEW2_CONTEXT_MENU_TARGET_KIND, error) {
	var kind COREWEBVIEW2_CONTEXT_MENU_TARGET_KIND
	_, _, err := i.vtbl.GetKind.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&kind)),
	)
	if err != windows.ERROR_SUCCESS {
		return 0, err
	}
	return kind, nil
}

func (i *ICoreWebView2ContextMenuTarget) GetIsEditable() (bool, error) {
	return contextMenuBoolFromCall(i.vtbl.GetIsEditable, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetIsRequestedForMainFrame() (bool, error) {
	return contextMenuBoolFromCall(i.vtbl.GetIsRequestedForMainFrame, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetPageURI() (string, error) {
	return contextMenuCoTaskMemString(i.vtbl.GetPageURI, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetFrameURI() (string, error) {
	return contextMenuCoTaskMemString(i.vtbl.GetFrameURI, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetHasLinkURI() (bool, error) {
	return contextMenuBoolFromCall(i.vtbl.GetHasLinkURI, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetLinkURI() (string, error) {
	return contextMenuCoTaskMemString(i.vtbl.GetLinkURI, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetHasLinkText() (bool, error) {
	return contextMenuBoolFromCall(i.vtbl.GetHasLinkText, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetLinkText() (string, error) {
	return contextMenuCoTaskMemString(i.vtbl.GetLinkText, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetHasSourceURI() (bool, error) {
	return contextMenuBoolFromCall(i.vtbl.GetHasSourceURI, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetSourceURI() (string, error) {
	return contextMenuCoTaskMemString(i.vtbl.GetSourceURI, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetHasSelection() (bool, error) {
	return contextMenuBoolFromCall(i.vtbl.GetHasSelection, unsafe.Pointer(i))
}

func (i *ICoreWebView2ContextMenuTarget) GetSelectionText() (string, error) {
	return contextMenuCoTaskMemString(i.vtbl.GetSelectionText, unsafe.Pointer(i))
}

type _ICoreWebView2CustomItemSelectedEventHandlerVtbl struct {
	_IUnknownVtbl
	Invoke ComProc
}

type ICoreWebView2CustomItemSelectedEventHandler struct {
	vtbl *_ICoreWebView2CustomItemSelectedEventHandlerVtbl
	impl _ICoreWebView2CustomItemSelectedEventHandlerImpl
}

type _ICoreWebView2CustomItemSelectedEventHandlerImpl interface {
	_IUnknownImpl
	CustomItemSelected(sender *ICoreWebView2ContextMenuItem, args *iUnknown) uintptr
}

func _ICoreWebView2CustomItemSelectedEventHandlerIUnknownQueryInterface(this *ICoreWebView2CustomItemSelectedEventHandler, refiid, object uintptr) uintptr {
	return this.impl.QueryInterface(refiid, object)
}

func _ICoreWebView2CustomItemSelectedEventHandlerIUnknownAddRef(this *ICoreWebView2CustomItemSelectedEventHandler) uintptr {
	return this.impl.AddRef()
}

func _ICoreWebView2CustomItemSelectedEventHandlerIUnknownRelease(this *ICoreWebView2CustomItemSelectedEventHandler) uintptr {
	return this.impl.Release()
}

func _ICoreWebView2CustomItemSelectedEventHandlerInvoke(this *ICoreWebView2CustomItemSelectedEventHandler, sender *ICoreWebView2ContextMenuItem, args *iUnknown) uintptr {
	return this.impl.CustomItemSelected(sender, args)
}

var _ICoreWebView2CustomItemSelectedEventHandlerFn = _ICoreWebView2CustomItemSelectedEventHandlerVtbl{
	_IUnknownVtbl{
		NewComProc(_ICoreWebView2CustomItemSelectedEventHandlerIUnknownQueryInterface),
		NewComProc(_ICoreWebView2CustomItemSelectedEventHandlerIUnknownAddRef),
		NewComProc(_ICoreWebView2CustomItemSelectedEventHandlerIUnknownRelease),
	},
	NewComProc(_ICoreWebView2CustomItemSelectedEventHandlerInvoke),
}

func newICoreWebView2CustomItemSelectedEventHandler(impl _ICoreWebView2CustomItemSelectedEventHandlerImpl) *ICoreWebView2CustomItemSelectedEventHandler {
	return &ICoreWebView2CustomItemSelectedEventHandler{
		vtbl: &_ICoreWebView2CustomItemSelectedEventHandlerFn,
		impl: impl,
	}
}

func contextMenuCoTaskMemString(call ComProc, receiver unsafe.Pointer) (string, error) {
	var value *uint16
	_, _, err := call.Call(
		uintptr(receiver),
		uintptr(unsafe.Pointer(&value)),
	)
	if err != windows.ERROR_SUCCESS {
		return "", err
	}
	if value == nil {
		return "", nil
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(value))
	return w32.Utf16PtrToString(value), nil
}

func contextMenuBoolFromCall(call ComProc, receiver unsafe.Pointer) (bool, error) {
	var value int32
	_, _, err := call.Call(
		uintptr(receiver),
		uintptr(unsafe.Pointer(&value)),
	)
	if err != windows.ERROR_SUCCESS {
		return false, err
	}
	return value != 0, nil
}
