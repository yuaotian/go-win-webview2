//go:build windows
// +build windows

package webview2

import (
	"errors"
	"os"
	"unsafe"

	"github.com/yuaotian/go-win-webview2/internal/w32"
)

type Stream struct {
	inner *w32.IStream
}

func NewStreamFromBytes(data []byte) (*Stream, error) {
	stream, err := w32.SHCreateMemoryStream(data)
	if err != nil {
		return nil, err
	}
	return newOwnedStream(stream), nil
}

func NewStreamFromFile(path string) (*Stream, error) {
	stream, err := w32.SHCreateReadFileStream(path)
	if err != nil {
		return nil, err
	}
	return newOwnedStream(stream), nil
}

// NewStreamFromHandle 为高级场景包装一个现有原生 IStream 句柄。
func NewStreamFromHandle(handle uintptr) *Stream {
	if handle == 0 {
		return nil
	}
	stream := (*w32.IStream)(unsafe.Pointer(handle))
	stream.AddRef()
	return &Stream{inner: stream}
}

func newOwnedStream(stream *w32.IStream) *Stream {
	if stream == nil {
		return nil
	}
	return &Stream{inner: stream}
}

func (s *Stream) cloneInner() *w32.IStream {
	if s == nil || s.inner == nil {
		return nil
	}
	s.inner.AddRef()
	return s.inner
}

func (s *Stream) Handle() uintptr {
	if s == nil || s.inner == nil {
		return 0
	}
	return uintptr(unsafe.Pointer(s.inner))
}

func (s *Stream) Release() {
	if s == nil || s.inner == nil {
		return
	}
	s.inner.Release()
	s.inner = nil
}

func (s *Stream) ReadAll() ([]byte, error) {
	if s == nil || s.inner == nil {
		return nil, errors.New("stream is not initialized")
	}

	buffer := make([]byte, 32*1024)
	result := make([]byte, 0, 32*1024)
	for {
		n, err := s.inner.Read(buffer)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			break
		}
		result = append(result, buffer[:n]...)
	}
	return result, nil
}

func (s *Stream) SaveToFile(path string) error {
	data, err := s.ReadAll()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
