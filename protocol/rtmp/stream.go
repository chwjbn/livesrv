package rtmp

import (
	"fmt"
	"github.com/chwjbn/livesrv/av"
	"github.com/chwjbn/livesrv/glog"
	"github.com/chwjbn/livesrv/protocol/rtmp/cache"
	"sync"
)

type Stream struct {
	isStart bool
	cache   *cache.Cache
	r       av.ReadCloser
	ws      *sync.Map
	info    av.Info
}

type PackWriterCloser struct {
	init bool
	w    av.WriteCloser
}

func (p *PackWriterCloser) GetWriter() av.WriteCloser {
	return p.w
}

func NewStream() *Stream {
	return &Stream{
		cache: cache.NewCache(),
		ws:    &sync.Map{},
	}
}

func (s *Stream) ID() string {
	if s.r != nil {
		return s.r.Info().UID
	}
	return EmptyID
}

func (s *Stream) GetReader() av.ReadCloser {
	return s.r
}

func (s *Stream) GetWs() *sync.Map {
	return s.ws
}

func (s *Stream) Copy(dst *Stream) {
	dst.info = s.info
	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		s.ws.Delete(key)
		v.w.CalcBaseTimestamp()
		dst.AddWriter(v.w)
		return true
	})
}

func (s *Stream) AddReader(r av.ReadCloser) {
	s.r = r
	go s.TransStart()
}

func (s *Stream) AddWriter(w av.WriteCloser) {
	info := w.Info()
	pw := &PackWriterCloser{w: w}
	s.ws.Store(info.UID, pw)
}

func (s *Stream) TransStart() {
	s.isStart = true
	var p av.Packet

	glog.InfoF("TransStart: %v", s.info)

	for {
		if !s.isStart {
			s.closeInter()
			return
		}
		err := s.r.Read(&p)
		if err != nil {
			s.closeInter()
			s.isStart = false
			return
		}

		s.cache.Write(p)

		s.ws.Range(func(key, val interface{}) bool {
			v := val.(*PackWriterCloser)
			if !v.init {
				//glog.InfoF("cache.send: %v", v.w.Info())
				if err = s.cache.Send(v.w); err != nil {
					glog.InfoF("[%s] send cache packet error: %v, remove", v.w.Info(), err)
					s.ws.Delete(key)
					return true
				}
				v.init = true
			} else {
				newPacket := p
				//writeType := reflect.TypeOf(v.w)
				//glog.InfoF("w.Write: type=%v, %v", writeType, v.w.Info())
				if err = v.w.Write(&newPacket); err != nil {
					glog.InfoF("[%s] write packet error: %v, remove", v.w.Info(), err)
					s.ws.Delete(key)
				}
			}
			return true
		})
	}
}

func (s *Stream) TransStop() {
	glog.InfoF("TransStop: %s", s.info.Key)

	if s.isStart && s.r != nil {
		s.r.Close(fmt.Errorf("stop old"))
	}

	s.isStart = false
}

func (s *Stream) CheckAlive() (n int) {

	if s.r != nil && s.isStart {
		if s.r.Alive() {
			n++
		} else {
			s.r.Close(fmt.Errorf("read timeout"))
		}
	}

	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		if v.w != nil {
			//Alive from RWBaser, check last frame now - timestamp, if > timeout then Remove it
			if !v.w.Alive() {
				glog.InfoF("write timeout remove")
				s.ws.Delete(key)
				v.w.Close(fmt.Errorf("write timeout"))
				return true
			}
			n++
		}
		return true
	})

	return
}

func (s *Stream) closeInter() {
	if s.r != nil {
		glog.InfoF("[%v] publisher closed", s.r.Info())
	}

	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		if v.w != nil {
			v.w.Close(fmt.Errorf("closed"))
			if v.w.Info().IsInterval() {
				s.ws.Delete(key)
				glog.InfoF("[%v] player closed and remove\n", v.w.Info())
			}
		}
		return true
	})
}
