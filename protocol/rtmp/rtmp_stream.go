package rtmp

import (
	"github.com/chwjbn/livesrv/av"
	"github.com/chwjbn/livesrv/glog"
	"sync"
	"time"
)

var (
	EmptyID = ""
)

type RtmpStream struct {
	streams *sync.Map //key
}

func NewRtmpStream() *RtmpStream {
	ret := &RtmpStream{
		streams: &sync.Map{},
	}
	go ret.CheckAlive()
	return ret
}

func (rs *RtmpStream) HandleReader(r av.ReadCloser) {

	info := r.Info()

	glog.InfoF("HandleReader: info[%v]", info)

	var stream *Stream
	i, ok := rs.streams.Load(info.Key)
	if stream, ok = i.(*Stream); ok {
		stream.TransStop()
		id := stream.ID()
		if id != EmptyID && id != info.UID {
			ns := NewStream()
			stream.Copy(ns)
			stream = ns
			rs.streams.Store(info.Key, ns)
		}
	} else {
		stream = NewStream()
		rs.streams.Store(info.Key, stream)
		stream.info = info
	}

	stream.AddReader(r)
}

func (rs *RtmpStream) HandleWriter(w av.WriteCloser) {

	info := w.Info()

	glog.InfoF("HandleWriter: info[%v]", info)

	var s *Stream
	item, ok := rs.streams.Load(info.Key)
	if !ok {
		glog.InfoF("HandleWriter: not found create new info[%v]", info)
		s = NewStream()
		rs.streams.Store(info.Key, s)
		s.info = info
	} else {
		s = item.(*Stream)
	}

	s.AddWriter(w)
}

func (rs *RtmpStream) GetStreams() *sync.Map {
	return rs.streams
}

func (rs *RtmpStream) CheckAlive() {

	for {

		<-time.After(5 * time.Second)

		rs.streams.Range(func(key, val interface{}) bool {

			v := val.(*Stream)

			glog.InfoF("检测直播流=[%+v] 活跃端=[%v]", v.info, v.CheckAlive())

			if v.CheckAlive() == 0 {
				rs.streams.Delete(key)
			}

			return true
		})

	}

}
