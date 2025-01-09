package rtmp

import (
	"github.com/chwjbn/livesrv/av"
	"github.com/chwjbn/livesrv/configure"
	"github.com/chwjbn/livesrv/glog"
	"github.com/chwjbn/livesrv/protocol/rtmp/core"
	"net"
	"reflect"
)

const (
	maxQueueNum           = 1024
	SAVE_STATICS_INTERVAL = 5000
)

var (
	writeTimeout = configure.Config.GetInt("write_timeout")
)

func GetWriteTimeout() int {
	writeTimeout = configure.Config.GetInt("write_timeout")
	return writeTimeout
}

type Server struct {
	handler av.Handler
	getter  av.GetWriter
}

func NewRtmpServer(h av.Handler, getter av.GetWriter) *Server {
	return &Server{
		handler: h,
		getter:  getter,
	}
}

func (s *Server) Serve(listener net.Listener) (err error) {

	defer func() {
		if r := recover(); r != nil {
			glog.Error("RTMP服务器异常: ", r)
		}
	}()

	for {
		var netconn net.Conn
		netconn, err = listener.Accept()
		if err != nil {
			glog.ErrorF("接收RTMP客户端连接错误=[%v]", err.Error())
			return
		}

		conn := core.NewConn(netconn, 4*1024)
		glog.InfoF("新的RTMP客户端连接 远程地址=[%v] 本地地址=[%v]", conn.RemoteAddr().String(), conn.LocalAddr().String())

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *core.Conn) error {

	if err := conn.HandshakeServer(); err != nil {
		conn.Close()
		glog.Error("RTMP客户端握手消息错误: ", err)
		return err
	}

	connServer := core.NewConnServer(conn)

	if err := connServer.ReadMsg(); err != nil {
		conn.Close()
		glog.Error("RTMP客户端读取消息错误: ", err)
		return err
	}

	appname, name, _ := connServer.GetInfo()

	clientRole := "Subscriber"
	if connServer.IsPublisher() {
		clientRole = "Publisher"
	}

	glog.InfoF("RTMP客户端角色=[%v] 请求应用名称=[%v] 流名称=[%v]", clientRole, appname, name)

	if connServer.IsPublisher() {

		connServer.PublishInfo.Name = name

		reader := NewVirReader(connServer)

		glog.InfoF("创建RTMP发布者:%+v", reader.Info())

		s.handler.HandleReader(reader)

		if s.getter != nil {

			writeType := reflect.TypeOf(s.getter)

			glog.InfoF("handleConn:writeType=%v", writeType)

			writer := s.getter.GetWriter(reader.Info())
			s.handler.HandleWriter(writer)
		}

	} else {

		writer := NewVirWriter(connServer)

		glog.InfoF("创建RTMP订阅者:%+v", writer.Info())

		s.handler.HandleWriter(writer)

	}

	return nil
}

type GetInFo interface {
	GetInfo() (string, string, string)
}

type StreamReadWriteCloser interface {
	GetInFo
	Close(error)
	Write(core.ChunkStream) error
	Read(c *core.ChunkStream) error
}

type StaticsBW struct {
	StreamId               uint32
	VideoDatainBytes       uint64
	LastVideoDatainBytes   uint64
	VideoSpeedInBytesperMS uint64

	AudioDatainBytes       uint64
	LastAudioDatainBytes   uint64
	AudioSpeedInBytesperMS uint64

	LastTimestamp int64
}
