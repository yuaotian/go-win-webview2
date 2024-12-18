//go:build windows
// +build windows

package webviewloader

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/jchv/go-winloader"
	"golang.org/x/sys/windows"
)

var (
	nativeModule                                       = windows.NewLazyDLL("WebView2Loader")
	nativeCreate                                       = nativeModule.NewProc("CreateCoreWebView2EnvironmentWithOptions")
	nativeCompareBrowserVersions                       = nativeModule.NewProc("CompareBrowserVersions")
	nativeGetAvailableCoreWebView2BrowserVersionString = nativeModule.NewProc("GetAvailableCoreWebView2BrowserVersionString")

	memOnce                                         sync.Once
	memModule                                       winloader.Module
	memCreate                                       winloader.Proc
	memCompareBrowserVersions                       winloader.Proc
	memGetAvailableCoreWebView2BrowserVersionString winloader.Proc
	memErr                                          error
)

// CompareBrowserVersions 将比较 2 个给定版本并返回：
//
//	-1 = v1 < v2
//	 0 = v1 == v2
//	 1 = v1 > v2
func CompareBrowserVersions(v1 string, v2 string) (int, error) {

	_v1, err := windows.UTF16PtrFromString(v1)
	if err != nil {
		return 0, err
	}
	_v2, err := windows.UTF16PtrFromString(v2)
	if err != nil {
		return 0, err
	}

	nativeErr := nativeModule.Load()
	if nativeErr == nil {
		nativeErr = nativeCompareBrowserVersions.Find()
	}
	var result int
	if nativeErr != nil {
		err = loadFromMemory(nativeErr)
		if err != nil {
			return 0, fmt.Errorf("Unable to load WebView2Loader.dll from disk: %v -- or from memory: %w", nativeErr, memErr)
		}
		_, _, err = memCompareBrowserVersions.Call(
			uint64(uintptr(unsafe.Pointer(_v1))),
			uint64(uintptr(unsafe.Pointer(_v2))),
			uint64(uintptr(unsafe.Pointer(&result))))
	} else {
		_, _, err = nativeCompareBrowserVersions.Call(
			uintptr(unsafe.Pointer(_v1)),
			uintptr(unsafe.Pointer(_v2)),
			uintptr(unsafe.Pointer(&result)))
	}
	if err != windows.ERROR_SUCCESS {
		return result, err
	}
	return result, nil
}

//GetInstalledVersion 返回 webview2 运行时的已安装版本。
//如果没有安装版本，则返回空字符串。
func GetInstalledVersion() (string, error) {
//GetAvailableCoreWebView2BrowserVersionString 记录为：
	//公共 STDAPI GetAvailableCoreWebView2BrowserVersionString(PCWSTR browserExecutableFolder, LPWSTR *versionInfo)
	//其中 winnt.h 将 STDAPI 定义为：
	//EXTERN_C HRESULT STDAPICALLTYPE
	//第一部分 (EXTERN_C) 可以忽略，因为它只与 C++ 相关，
	//HRESULT 是返回类型，这意味着它返回一个整数，如果成功，该整数将为 0 (S_OK)，
//最后 STDAPICALLTYPE 告诉我们该函数使用 stdcall 调用约定（Go 对系统调用的假设）。

	nativeErr := nativeModule.Load()
	if nativeErr == nil {
		nativeErr = nativeGetAvailableCoreWebView2BrowserVersionString.Find()
	}
	var hr uintptr
	var result *uint16
	if nativeErr != nil {
		if err := loadFromMemory(nativeErr); err != nil {
			return "", fmt.Errorf("Unable to load WebView2Loader.dll from disk: %v -- or from memory: %w", nativeErr, memErr)
		}
		hr64, _, _ := memGetAvailableCoreWebView2BrowserVersionString.Call(
			uint64(uintptr(unsafe.Pointer(nil))),
			uint64(uintptr(unsafe.Pointer(&result))))
		hr = uintptr(hr64) // THRESULT 的返回大小将是本机大小（即 uintptr），而不是 32 位系统上的 64 位。  在这两种情况下，它都应该被解释为 32 位（LONG）。
	} else {
		hr, _, _ = nativeGetAvailableCoreWebView2BrowserVersionString.Call(
			uintptr(unsafe.Pointer(nil)),
			uintptr(unsafe.Pointer(&result)))
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(result)) // Safe even if result is nil
	if hr != uintptr(windows.S_OK) {
		if hr&0xFFFF == uintptr(windows.ERROR_FILE_NOT_FOUND) {
			// HRESULT 的低 16 位（错误代码本身）是 ERROR_FILE_NOT_FOUND，这意味着系统未安装。
			return "", nil // 返回一个空白字符串，但没有错误，因为我们成功检测到没有安装。
		}
		return "", fmt.Errorf("GetAvailableCoreWebView2BrowserVersionString returned HRESULT 0x%X", hr)
	}
	version := windows.UTF16PtrToString(result) // 即使结果为零也是安全的
	return version, nil
}

//CreateCoreWebView2EnvironmentWithOptions 尝试加载 WebviewLoader2 并
//调用 CreateCoreWebView2EnvironmentWithOptions 例程。
func CreateCoreWebView2EnvironmentWithOptions(browserExecutableFolder, userDataFolder *uint16, environmentOptions uintptr, environmentCompletedHandle uintptr) (uintptr, error) {
	nativeErr := nativeModule.Load()
	if nativeErr == nil {
		nativeErr = nativeCreate.Find()
	}
	if nativeErr != nil {
		err := loadFromMemory(nativeErr)
		if err != nil {
			return 0, err
		}
		res, _, _ := memCreate.Call(
			uint64(uintptr(unsafe.Pointer(browserExecutableFolder))),
			uint64(uintptr(unsafe.Pointer(userDataFolder))),
			uint64(environmentOptions),
			uint64(environmentCompletedHandle),
		)
		return uintptr(res), nil
	}
	res, _, _ := nativeCreate.Call(
		uintptr(unsafe.Pointer(browserExecutableFolder)),
		uintptr(unsafe.Pointer(userDataFolder)),
		environmentOptions,
		environmentCompletedHandle,
	)
	return res, nil
}

func loadFromMemory(nativeErr error) error {
	var err error
	// DLL 本身不可用。尝试加载嵌入副本。
	memOnce.Do(func() {
		memModule, memErr = winloader.LoadFromMemory(WebView2Loader)
		if memErr != nil {
			err = fmt.Errorf("Unable to load WebView2Loader.dll from disk: %v -- or from memory: %w", nativeErr, memErr)
			return
		}
		memCreate = memModule.Proc("CreateCoreWebView2EnvironmentWithOptions")
		memCompareBrowserVersions = memModule.Proc("CompareBrowserVersions")
		memGetAvailableCoreWebView2BrowserVersionString = memModule.Proc("GetAvailableCoreWebView2BrowserVersionString")
	})
	return err
}
