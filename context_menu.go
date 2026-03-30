//go:build windows
// +build windows

package webview2

import (
	"errors"
	"log"

	"github.com/yuaotian/go-win-webview2/pkg/edge"
)

type ContextMenuItemKind uint32

const (
	ContextMenuItemKindCommand   ContextMenuItemKind = ContextMenuItemKind(edge.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_COMMAND)
	ContextMenuItemKindCheckBox  ContextMenuItemKind = ContextMenuItemKind(edge.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_CHECK_BOX)
	ContextMenuItemKindRadio     ContextMenuItemKind = ContextMenuItemKind(edge.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_RADIO)
	ContextMenuItemKindSeparator ContextMenuItemKind = ContextMenuItemKind(edge.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_SEPARATOR)
	ContextMenuItemKindSubmenu   ContextMenuItemKind = ContextMenuItemKind(edge.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_SUBMENU)
)

type ContextMenuTargetKind uint32

const (
	ContextMenuTargetKindPage         ContextMenuTargetKind = ContextMenuTargetKind(edge.COREWEBVIEW2_CONTEXT_MENU_TARGET_KIND_PAGE)
	ContextMenuTargetKindImage        ContextMenuTargetKind = ContextMenuTargetKind(edge.COREWEBVIEW2_CONTEXT_MENU_TARGET_KIND_IMAGE)
	ContextMenuTargetKindSelectedText ContextMenuTargetKind = ContextMenuTargetKind(edge.COREWEBVIEW2_CONTEXT_MENU_TARGET_KIND_SELECTED_TEXT)
	ContextMenuTargetKindAudio        ContextMenuTargetKind = ContextMenuTargetKind(edge.COREWEBVIEW2_CONTEXT_MENU_TARGET_KIND_AUDIO)
	ContextMenuTargetKindVideo        ContextMenuTargetKind = ContextMenuTargetKind(edge.COREWEBVIEW2_CONTEXT_MENU_TARGET_KIND_VIDEO)
)

type ContextMenuLocation struct {
	X int32
	Y int32
}

type ContextMenuTarget struct {
	Kind                    ContextMenuTargetKind
	IsEditable              bool
	IsRequestedForMainFrame bool
	PageURI                 string
	FrameURI                string
	HasLinkURI              bool
	LinkURI                 string
	HasLinkText             bool
	LinkText                string
	HasSourceURI            bool
	SourceURI               string
	HasSelection            bool
	SelectionText           string
}

type ContextMenuRequestedEvent struct {
	Location ContextMenuLocation
	Target   ContextMenuTarget

	args     *edge.ICoreWebView2ContextMenuRequestedEventArgs
	env      *edge.ICoreWebView2Environment
	chromium *edge.Chromium
}

func (e *ContextMenuRequestedEvent) MenuItems() (*ContextMenuItemCollection, error) {
	if e == nil || e.args == nil {
		return nil, errors.New("context menu event is not initialized")
	}
	items, err := e.args.GetMenuItems()
	if err != nil {
		return nil, err
	}
	return &ContextMenuItemCollection{
		inner:    items,
		env:      e.env,
		chromium: e.chromium,
	}, nil
}

func (e *ContextMenuRequestedEvent) SetHandled(handled bool) error {
	if e == nil || e.args == nil {
		return errors.New("context menu event is not initialized")
	}
	return e.args.PutHandled(handled)
}

func (e *ContextMenuRequestedEvent) GetHandled() (bool, error) {
	if e == nil || e.args == nil {
		return false, errors.New("context menu event is not initialized")
	}
	return e.args.GetHandled()
}

func (e *ContextMenuRequestedEvent) SetSelectedCommandID(commandID int32) error {
	if e == nil || e.args == nil {
		return errors.New("context menu event is not initialized")
	}
	return e.args.PutSelectedCommandId(commandID)
}

func (e *ContextMenuRequestedEvent) GetSelectedCommandID() (int32, error) {
	if e == nil || e.args == nil {
		return 0, errors.New("context menu event is not initialized")
	}
	return e.args.GetSelectedCommandId()
}

func (e *ContextMenuRequestedEvent) GetDeferral() (*WebViewDeferral, error) {
	if e == nil || e.args == nil {
		return nil, errors.New("context menu event is not initialized")
	}
	deferral, err := e.args.GetDeferral()
	if err != nil {
		return nil, err
	}
	return &WebViewDeferral{inner: deferral}, nil
}

type ContextMenuItemCollection struct {
	inner    *edge.ICoreWebView2ContextMenuItemCollection
	env      *edge.ICoreWebView2Environment
	chromium *edge.Chromium
}

func (c *ContextMenuItemCollection) Count() (uint32, error) {
	if c == nil || c.inner == nil {
		return 0, errors.New("context menu items are not initialized")
	}
	return c.inner.GetCount()
}

func (c *ContextMenuItemCollection) Item(index uint32) (*ContextMenuItem, error) {
	if c == nil || c.inner == nil {
		return nil, errors.New("context menu items are not initialized")
	}
	item, err := c.inner.GetValueAtIndex(index)
	if err != nil {
		return nil, err
	}
	return &ContextMenuItem{
		inner:    item,
		env:      c.env,
		chromium: c.chromium,
	}, nil
}

func (c *ContextMenuItemCollection) Remove(index uint32) error {
	if c == nil || c.inner == nil {
		return errors.New("context menu items are not initialized")
	}
	return c.inner.RemoveValueAtIndex(index)
}

func (c *ContextMenuItemCollection) Clear() error {
	if c == nil || c.inner == nil {
		return errors.New("context menu items are not initialized")
	}
	count, err := c.inner.GetCount()
	if err != nil {
		return err
	}
	for index := int(count) - 1; index >= 0; index-- {
		if err := c.inner.RemoveValueAtIndex(uint32(index)); err != nil {
			return err
		}
	}
	return nil
}

func (c *ContextMenuItemCollection) AppendCommand(label string, onSelected func()) error {
	count, err := c.Count()
	if err != nil {
		return err
	}
	return c.InsertCommand(count, label, onSelected)
}

func (c *ContextMenuItemCollection) InsertCommand(index uint32, label string, onSelected func()) error {
	return c.insertCustomItem(index, label, edge.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_COMMAND, onSelected)
}

func (c *ContextMenuItemCollection) AppendSeparator() error {
	count, err := c.Count()
	if err != nil {
		return err
	}
	return c.InsertSeparator(count)
}

func (c *ContextMenuItemCollection) InsertSeparator(index uint32) error {
	return c.insertCustomItem(index, "", edge.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_SEPARATOR, nil)
}

func (c *ContextMenuItemCollection) insertCustomItem(index uint32, label string, kind edge.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND, onSelected func()) error {
	if c == nil || c.inner == nil || c.env == nil {
		return errors.New("context menu items are not initialized")
	}
	env9 := c.env.GetICoreWebView2Environment9()
	if env9 == nil {
		return errors.New("ICoreWebView2Environment9 is not available")
	}

	item, err := env9.CreateContextMenuItem(label, nil, kind)
	if err != nil {
		return err
	}
	if onSelected != nil {
		if c.chromium == nil {
			return errors.New("chromium is not initialized")
		}
		if err := c.chromium.RegisterContextMenuCommand(item, onSelected); err != nil {
			return err
		}
	}
	return c.inner.InsertValueAtIndex(index, item)
}

type ContextMenuItem struct {
	inner    *edge.ICoreWebView2ContextMenuItem
	env      *edge.ICoreWebView2Environment
	chromium *edge.Chromium
}

func (i *ContextMenuItem) Name() (string, error) {
	if i == nil || i.inner == nil {
		return "", errors.New("context menu item is not initialized")
	}
	return i.inner.GetName()
}

func (i *ContextMenuItem) Label() (string, error) {
	if i == nil || i.inner == nil {
		return "", errors.New("context menu item is not initialized")
	}
	return i.inner.GetLabel()
}

func (i *ContextMenuItem) CommandID() (int32, error) {
	if i == nil || i.inner == nil {
		return 0, errors.New("context menu item is not initialized")
	}
	return i.inner.GetCommandId()
}

func (i *ContextMenuItem) Kind() (ContextMenuItemKind, error) {
	if i == nil || i.inner == nil {
		return 0, errors.New("context menu item is not initialized")
	}
	kind, err := i.inner.GetKind()
	if err != nil {
		return 0, err
	}
	return ContextMenuItemKind(kind), nil
}

func (i *ContextMenuItem) IsEnabled() (bool, error) {
	if i == nil || i.inner == nil {
		return false, errors.New("context menu item is not initialized")
	}
	return i.inner.GetIsEnabled()
}

func (i *ContextMenuItem) IsChecked() (bool, error) {
	if i == nil || i.inner == nil {
		return false, errors.New("context menu item is not initialized")
	}
	return i.inner.GetIsChecked()
}

func (i *ContextMenuItem) Children() (*ContextMenuItemCollection, error) {
	if i == nil || i.inner == nil {
		return nil, errors.New("context menu item is not initialized")
	}
	children, err := i.inner.GetChildren()
	if err != nil {
		return nil, err
	}
	return &ContextMenuItemCollection{
		inner:    children,
		env:      i.env,
		chromium: i.chromium,
	}, nil
}

func (w *webview) OnContextMenuRequested(callback func(*ContextMenuRequestedEvent)) {
	chromium, ok := w.browser.(*edge.Chromium)
	if !ok {
		return
	}

	// WebView2 只有在默认上下文菜单启用时才会派发 ContextMenuRequested。
	if err := w.EnableContextMenu(); err != nil {
		log.Printf("启用默认右键菜单失败: %v", err)
	}

	chromium.SetContextMenuRequestedCallback(func(_ *edge.ICoreWebView2, args *edge.ICoreWebView2ContextMenuRequestedEventArgs) {
		if callback == nil {
			return
		}

		location, err := args.GetLocation()
		if err != nil {
			log.Printf("读取右键菜单位置失败: %v", err)
			return
		}
		target, err := args.GetContextMenuTarget()
		if err != nil {
			log.Printf("读取右键菜单目标失败: %v", err)
			return
		}

		callback(&ContextMenuRequestedEvent{
			Location: ContextMenuLocation{
				X: location.X,
				Y: location.Y,
			},
			Target:   snapshotContextMenuTarget(target),
			args:     args,
			env:      chromium.Environment(),
			chromium: chromium,
		})
	})
}

func snapshotContextMenuTarget(target *edge.ICoreWebView2ContextMenuTarget) ContextMenuTarget {
	if target == nil {
		return ContextMenuTarget{}
	}

	result := ContextMenuTarget{}
	if kind, err := target.GetKind(); err == nil {
		result.Kind = ContextMenuTargetKind(kind)
	}
	if value, err := target.GetIsEditable(); err == nil {
		result.IsEditable = value
	}
	if value, err := target.GetIsRequestedForMainFrame(); err == nil {
		result.IsRequestedForMainFrame = value
	}
	if value, err := target.GetPageURI(); err == nil {
		result.PageURI = value
	}
	if value, err := target.GetFrameURI(); err == nil {
		result.FrameURI = value
	}
	if value, err := target.GetHasLinkURI(); err == nil {
		result.HasLinkURI = value
	}
	if value, err := target.GetLinkURI(); err == nil {
		result.LinkURI = value
	}
	if value, err := target.GetHasLinkText(); err == nil {
		result.HasLinkText = value
	}
	if value, err := target.GetLinkText(); err == nil {
		result.LinkText = value
	}
	if value, err := target.GetHasSourceURI(); err == nil {
		result.HasSourceURI = value
	}
	if value, err := target.GetSourceURI(); err == nil {
		result.SourceURI = value
	}
	if value, err := target.GetHasSelection(); err == nil {
		result.HasSelection = value
	}
	if value, err := target.GetSelectionText(); err == nil {
		result.SelectionText = value
	}
	return result
}
