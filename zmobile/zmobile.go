package zmobile

import (
	"crypto/tls"
	"fmt"
	"github.com/chwjbn/livesrv/configure"
	"github.com/chwjbn/livesrv/glib"
	"github.com/chwjbn/livesrv/glog"
	"github.com/chwjbn/livesrv/protocol/hls"
	"github.com/chwjbn/livesrv/protocol/rtmp"
	"net"
	"time"
)

func startRtmp(stream *rtmp.RtmpStream, hlsServer *hls.Server) error {

	rtmpAddr := configure.Config.GetString("rtmp_addr")
	isRtmps := configure.Config.GetBool("enable_rtmps")

	var rtmpListen net.Listener
	if isRtmps {
		certPath := configure.Config.GetString("rtmps_cert")
		keyPath := configure.Config.GetString("rtmps_key")
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			glog.Error(err)
		}

		rtmpListen, err = tls.Listen("tcp", rtmpAddr, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
		if err != nil {
			glog.Error(err)
		}
	} else {
		var err error
		rtmpListen, err = net.Listen("tcp", rtmpAddr)
		if err != nil {
			glog.Error(err)
		}
	}

	var rtmpServer *rtmp.Server

	if hlsServer == nil {
		rtmpServer = rtmp.NewRtmpServer(stream, nil)
		glog.Info("HLS server disable....")
	} else {
		rtmpServer = rtmp.NewRtmpServer(stream, hlsServer)
		glog.Info("HLS server enable....")
	}

	defer func() {
		if r := recover(); r != nil {
			glog.Error("RTMP server panic: ", r)
		}
	}()

	if isRtmps {
		glog.Info("RTMPS Listen On ", rtmpAddr)
	} else {
		glog.Info("RTMP Listen On ", rtmpAddr)
	}

	return rtmpServer.Serve(rtmpListen)
}

func SetLogDir(dirPath string) {

	glog.StdInfo(fmt.Sprintf("设置日志目录=[%v]开始", dirPath))

	if len(dirPath) < 1 {
		return
	}

	if !glib.DirExists(dirPath) {
		return
	}

	glog.StdInfo(fmt.Sprintf("设置日志目录=[%v]成功", dirPath))

	glog.SetLogDir(dirPath)

}

func RunMobile(rtmpAddr string) {

	configure.InitConfig()

	defer func() {
		if r := recover(); r != nil {
			glog.Error("LiveHub panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()

	configure.Config.Set("rtmp_addr", rtmpAddr)
	configure.Config.Set("enable_rtmps", false)
	configure.Config.Set("write_timeout", 10)

	glog.InfoF("开始启动LiveHub,版本号=[%v]", glib.AppVersion())

	stream := rtmp.NewRtmpStream()

	runErr := startRtmp(stream, nil)

	if runErr != nil {
		glog.Error(runErr.Error())
	}

}
