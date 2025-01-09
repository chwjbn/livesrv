package gcahe

import (
	"fmt"
	"github.com/chwjbn/livesrv/glib"
	"path"
	"syscall"
)

type Manager struct {
	mId       string
	mUnitSize int
	mSize     int

	mMapFilePath   string
	mMapFileHandle syscall.Handle

	mMapName   string
	mMapHandle syscall.Handle
}

func NewManager(id string, unitSize int, size int) (*Manager, error) {

	xThis := new(Manager)

	xThis.mUnitSize = unitSize
	xThis.mSize = size
	xThis.mId = glib.EncryptMd5(fmt.Sprintf("%v-%v-%v-%v", id, unitSize, size, glib.EncryptNewId("gcache")))

	xErr := xThis.init()

	if xErr != nil {
		return nil, xErr
	}

	return xThis, xErr

}

func (this *Manager) init() error {

	var xErr error

	xCacheDir := path.Join(glib.AppBaseDir(), "gcache")
	if !glib.DirExists(xCacheDir) {
		glib.DirCreate(xCacheDir)
	}

	if !glib.DirExists(xCacheDir) {
		xErr = fmt.Errorf("init gcache dir=[%v] faild", xCacheDir)
		return xErr
	}

	this.mMapFilePath = path.Join(xCacheDir, fmt.Sprintf("%v.dat", this.mId))

	var xCreateErr error

	this.mMapFileHandle, xCreateErr = syscall.CreateFile(
		syscall.StringToUTF16Ptr(this.mMapFilePath),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_ALWAYS,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)

	if xCreateErr != nil {
		xErr = fmt.Errorf("create gcache file=[%v] with error=[%v]", this.mMapFilePath, xCreateErr.Error())
		return xErr
	}

	this.mMapName = fmt.Sprintf("GCACHE_%v", this.mId)
	xCacheMapFileSize := int64(this.mSize) * int64(this.mUnitSize)

	//主意超过4G的情况
	highSize := uint32(xCacheMapFileSize >> 32)
	lowSize := uint32(xCacheMapFileSize & 0xFFFFFFFF)

	this.mMapHandle, xCreateErr = syscall.CreateFileMapping(
		this.mMapFileHandle,
		nil,
		syscall.PAGE_READWRITE,
		highSize,
		lowSize,
		syscall.StringToUTF16Ptr(this.mMapName),
	)

	if xCreateErr != nil {
		xErr = fmt.Errorf("create gcache filemapping file=[%v] name=[%v] with error=[%v]", this.mMapFilePath, this.mMapName, xCreateErr.Error())
		return xErr
	}

	return xErr

}

func (this *Manager) GetSize() int {
	return this.mSize
}

func (this *Manager) GetUnitSize() int {
	return this.mUnitSize
}

func (this *Manager) SetData(dataIndex int, data []byte) error {

	var xErr error

	if len(data) > this.mUnitSize {
		xErr = fmt.Errorf("data size must blow %v", this.mUnitSize)
		return xErr
	}

	xDataOffset := int64(dataIndex) * int64(this.mUnitSize)
	xDataSize := int64(this.mUnitSize)

	highOffset := uint32(xDataOffset >> 32)
	lowOffset := uint32(xDataOffset & 0xFFFFFFFF)

	xDataAddr, xDataErr := syscall.MapViewOfFile(this.mMapHandle, syscall.FILE_MAP_READ|syscall.FILE_MAP_WRITE|syscall.FILE_MAP_COPY, highOffset, lowOffset, uintptr(xDataSize))
	if xDataErr != nil {
		xErr = fmt.Errorf("MapViewOfFile dataIndex=[%v] error=[%v]", dataIndex, xDataErr.Error())
		return xErr
	}

	defer syscall.UnmapViewOfFile(xDataAddr)

	xDataBuff := glib.TextUintptrToSlice(xDataAddr, int(xDataSize))

	//清空数据
	for i := 0; i < int(xDataSize); i++ {
		xDataBuff[i] = 0
	}

	copy(xDataBuff, data)

	syscall.FlushViewOfFile(xDataAddr, uintptr(xDataSize))

	return xErr

}

func (this *Manager) GetData(dataIndex int) ([]byte, error) {

	var xData []byte
	var xErr error

	xDataOffset := int64(dataIndex) * int64(this.mUnitSize)
	xDataSize := int64(this.mUnitSize)

	highOffset := uint32(xDataOffset >> 32)
	lowOffset := uint32(xDataOffset & 0xFFFFFFFF)

	xDataAddr, xDataErr := syscall.MapViewOfFile(this.mMapHandle, syscall.FILE_MAP_READ|syscall.FILE_MAP_WRITE|syscall.FILE_MAP_COPY, highOffset, lowOffset, uintptr(xDataSize))
	if xDataErr != nil {
		xErr = fmt.Errorf("MapViewOfFile dataIndex=[%v] error=[%v]", dataIndex, xDataErr.Error())
		return xData, xErr
	}

	defer syscall.UnmapViewOfFile(xDataAddr)

	xDataBuff := glib.TextUintptrToSlice(xDataAddr, int(xDataSize))

	xData = make([]byte, this.mUnitSize)

	copy(xData, xDataBuff)

	return xData, xErr

}

func (this *Manager) Release() error {

	var xError error

	if this.mMapHandle != syscall.InvalidHandle {
		syscall.CloseHandle(this.mMapHandle)
	}

	if this.mMapFileHandle != syscall.InvalidHandle {
		syscall.CloseHandle(this.mMapFileHandle)
	}

	if glib.FileExists(this.mMapFilePath) {
		glib.FileDelete(this.mMapFilePath)
	}

	return xError

}
