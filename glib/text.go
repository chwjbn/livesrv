package glib

import (
	"net"
	"reflect"
	"unsafe"
)

func TextIsIP4(data string) bool {

	bRet := false

	ipAddr := net.ParseIP(data)

	if ipAddr != nil {
		bRet = true
	}

	return bRet

}

func TextUintptrToSlice(addr uintptr, size int) []byte {

	header := reflect.SliceHeader{
		Len:  size,
		Cap:  size,
		Data: addr,
	}

	return *(*[]byte)(unsafe.Pointer(&header))
}
