package rtmp

import (
	"fmt"
	"github.com/chwjbn/livesrv/av"
	"github.com/chwjbn/livesrv/glog"
	"github.com/chwjbn/livesrv/protocol/rtmp/core"
	"github.com/chwjbn/livesrv/utils/uid"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type VirWriter struct {
	Uid    string
	closed bool
	av.RWBaser
	conn        StreamReadWriteCloser
	packetQueue chan *av.Packet
	WriteBWInfo StaticsBW
}

func NewVirWriter(conn StreamReadWriteCloser) *VirWriter {
	ret := &VirWriter{
		Uid:         uid.NewId(),
		conn:        conn,
		RWBaser:     av.NewRWBaser(time.Second * time.Duration(GetWriteTimeout())),
		packetQueue: make(chan *av.Packet, maxQueueNum),
		WriteBWInfo: StaticsBW{0, 0, 0, 0, 0, 0, 0, 0},
	}

	go ret.Check()
	go func() {
		err := ret.SendPacket()
		if err != nil {
			glog.Warn(err)
		}
	}()
	return ret
}

func (v *VirWriter) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
	nowInMS := int64(time.Now().UnixNano() / 1e6)

	v.WriteBWInfo.StreamId = streamid
	if isVideoFlag {
		v.WriteBWInfo.VideoDatainBytes = v.WriteBWInfo.VideoDatainBytes + length
	} else {
		v.WriteBWInfo.AudioDatainBytes = v.WriteBWInfo.AudioDatainBytes + length
	}

	if v.WriteBWInfo.LastTimestamp == 0 {
		v.WriteBWInfo.LastTimestamp = nowInMS
	} else if (nowInMS - v.WriteBWInfo.LastTimestamp) >= SAVE_STATICS_INTERVAL {
		diffTimestamp := (nowInMS - v.WriteBWInfo.LastTimestamp) / 1000

		v.WriteBWInfo.VideoSpeedInBytesperMS = (v.WriteBWInfo.VideoDatainBytes - v.WriteBWInfo.LastVideoDatainBytes) * 8 / uint64(diffTimestamp) / 1000
		v.WriteBWInfo.AudioSpeedInBytesperMS = (v.WriteBWInfo.AudioDatainBytes - v.WriteBWInfo.LastAudioDatainBytes) * 8 / uint64(diffTimestamp) / 1000

		v.WriteBWInfo.LastVideoDatainBytes = v.WriteBWInfo.VideoDatainBytes
		v.WriteBWInfo.LastAudioDatainBytes = v.WriteBWInfo.AudioDatainBytes
		v.WriteBWInfo.LastTimestamp = nowInMS
	}
}

func (v *VirWriter) Check() {
	var c core.ChunkStream
	for {
		if err := v.conn.Read(&c); err != nil {
			v.Close(err)
			return
		}
	}
}

func (v *VirWriter) DropPacket(pktQue chan *av.Packet, info av.Info) {
	glog.WarnF("[%v] packet queue max!!!", info)
	for i := 0; i < maxQueueNum-84; i++ {
		tmpPkt, ok := <-pktQue
		// try to don't drop audio
		if ok && tmpPkt.IsAudio {
			if len(pktQue) > maxQueueNum-2 {
				glog.Info("drop audio pkt")
				<-pktQue
			} else {
				pktQue <- tmpPkt
			}

		}

		if ok && tmpPkt.IsVideo {
			videoPkt, ok := tmpPkt.Header.(av.VideoPacketHeader)
			// dont't drop sps config and dont't drop key frame
			if ok && (videoPkt.IsSeq() || videoPkt.IsKeyFrame()) {
				pktQue <- tmpPkt
			}
			if len(pktQue) > maxQueueNum-10 {
				glog.Info("drop video pkt")
				<-pktQue
			}
		}

	}
	glog.Info("packet queue len: ", len(pktQue))
}

func (v *VirWriter) Write(p *av.Packet) (err error) {
	err = nil

	if v.closed {
		err = fmt.Errorf("VirWriter closed")
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("VirWriter has already been closed:%v", e)
		}
	}()
	if len(v.packetQueue) >= maxQueueNum-24 {
		v.DropPacket(v.packetQueue, v.Info())
	} else {
		v.packetQueue <- p
	}

	return
}

func (v *VirWriter) SendPacket() error {
	Flush := reflect.ValueOf(v.conn).MethodByName("Flush")
	var cs core.ChunkStream
	for {
		p, ok := <-v.packetQueue
		if ok {
			cs.Data = p.Data
			cs.Length = uint32(len(p.Data))
			cs.StreamID = p.StreamID
			cs.Timestamp = p.TimeStamp
			cs.Timestamp += v.BaseTimeStamp()

			if p.IsVideo {
				cs.TypeID = av.TAG_VIDEO
			} else {
				if p.IsMetadata {
					cs.TypeID = av.TAG_SCRIPTDATAAMF0
				} else {
					cs.TypeID = av.TAG_AUDIO
				}
			}

			v.SaveStatics(p.StreamID, uint64(cs.Length), p.IsVideo)
			v.SetPreTime()
			v.RecTimeStamp(cs.Timestamp, cs.TypeID)
			err := v.conn.Write(cs)
			if err != nil {
				v.closed = true
				return err
			}
			Flush.Call(nil)
		} else {
			return fmt.Errorf("closed")
		}

	}
}

func (v *VirWriter) Info() (ret av.Info) {
	ret.UID = v.Uid
	_, _, URL := v.conn.GetInfo()
	ret.URL = URL
	_url, err := url.Parse(URL)
	if err != nil {
		glog.Warn(err)
	}
	ret.Key = strings.TrimLeft(_url.Path, "/")
	ret.Inter = true
	return
}

func (v *VirWriter) Close(err error) {
	glog.Warn("player ", v.Info(), "closed: "+err.Error())
	if !v.closed {
		close(v.packetQueue)
	}
	v.closed = true
	v.conn.Close(err)
}
