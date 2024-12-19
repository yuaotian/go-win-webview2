//go:build windows
// +build windows

package edge

import "syscall"

//go:uintptrescapes
// Call calls a COM procedure.
func (p ComProc) Call(a ...uintptr) (r1, r2 uintptr, lastErr error) {
	//需要神奇的 uintptrescapes 注释来防止移动 uintptr(unsafe.Pointer(p)) 因此也调用 .Call()
	//满足 unsafe.Pointer 规则“(4) 在调用 syscall.Syscall 时将指针转换为 uintptr。”
	//否则，指针可能会被移动，特别是 Go 堆栈上的指针可能会动态增长。
	//请参阅 https://pkg.go.dev/unsafe#Pointer 和 https://github.com/golang/go/issues/34474
	return syscall.SyscallN(uintptr(p), a...)
}
