package main

import (
	"flag"
	"github.com/chwjbn/livesrv/zmobile"
)

func main() {

	//zmobile.SetLogDir("D:\\")

	rtmpAddr := flag.String("rtmp_addr", "0.0.0.0:11101", "RTMP server address")

	zmobile.RunMobile(*rtmpAddr)

}
