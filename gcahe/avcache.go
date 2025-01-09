package gcahe

import (
	"bytes"
	"fmt"
	"github.com/chwjbn/livesrv/glib"
	"sync"
)

type AvCache struct {
	mWritePos    int
	mWriteLocker sync.Locker

	mReadPos    int
	mReadLocker sync.Locker

	mMgr *Manager
}

func NewAvCache() (*AvCache, error) {

	xThis := new(AvCache)
	xErr := xThis.init()
	if xErr != nil {
		return nil, xErr
	}

	return xThis, xErr

}

func (this *AvCache) init() error {

	var xErr error

	this.mMgr, xErr = NewManager(glib.EncryptNewId("AvCache"), 1024*256, 4096)
	if xErr != nil {
		return xErr
	}

	return xErr

}

func (this *AvCache) Release() error {

	this.mWriteLocker.Lock()
	defer this.mWriteLocker.Unlock()

	this.mReadLocker.Lock()
	defer this.mReadLocker.Unlock()

	if this.mMgr == nil {
		return nil
	}

	return this.mMgr.Release()
}

func (this *AvCache) WriteNode(data Node) error {

	var xErr error

	this.mWriteLocker.Lock()
	defer this.mWriteLocker.Unlock()

	dataBuffer := bytes.Buffer{}

	dataErr := data.Encode(&dataBuffer)
	if dataErr != nil {
		xErr = fmt.Errorf("encode Node data error:[%v]", dataErr.Error())
		return xErr
	}

	dataErr = this.mMgr.SetData(this.mWritePos, dataBuffer.Bytes())
	if dataErr != nil {
		xErr = fmt.Errorf("set Node data error:[%v]", dataErr.Error())
		return xErr
	}

	if this.mWritePos < (this.mMgr.GetSize() - 1) {
		this.mWritePos++
	} else {
		this.mWritePos = 0
	}

	return xErr
}

func (this *AvCache) ReadNode() (Node, error) {

	var xErr error
	var xData Node

	this.mReadLocker.Lock()
	defer this.mReadLocker.Unlock()

	dataBytes, dataErr := this.mMgr.GetData(this.mReadPos)
	if dataErr != nil {
		xErr = fmt.Errorf("get Node data error:[%v]", dataErr.Error())
		return xData, xErr
	}

	dataBuffer := bytes.NewBuffer(dataBytes)

	dataErr = xData.Decode(dataBuffer)
	if dataErr != nil {
		xErr = fmt.Errorf("decode Node data error:[%v]", dataErr.Error())
		return xData, xErr
	}

	if xData.Type <= 0 {
		return xData, xErr
	}

	if this.mReadPos < (this.mMgr.GetSize() - 1) {
		this.mReadPos++
	} else {
		this.mReadPos = 0
	}

	return xData, xErr

}
