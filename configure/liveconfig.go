package configure

import (
	"bytes"
	"encoding/json"
	"github.com/chwjbn/livesrv/glib"
	"github.com/chwjbn/livesrv/glog"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ServerCfg struct {
	Level      string `mapstructure:"level"`
	ConfigFile string `mapstructure:"config_file"`

	RTMPAddr string `mapstructure:"rtmp_addr"`

	HLSAddr         string `mapstructure:"hls_addr"`
	HLSKeepAfterEnd bool   `mapstructure:"hls_keep_after_end"`

	WriteTimeout    int  `mapstructure:"write_timeout"`
	EnableTLSVerify bool `mapstructure:"enable_tls_verify"`
	GopNum          int  `mapstructure:"gop_num"`
}

// default config
var defaultConf = ServerCfg{
	ConfigFile: "livehub.yaml",
	RTMPAddr:   ":1935",

	HLSAddr:         ":7002",
	HLSKeepAfterEnd: false,

	WriteTimeout:    10,
	EnableTLSVerify: true,
	GopNum:          1,
}

var (
	Config = viper.New()
)

func InitConfig() {
	initDefault()
}

func initDefault() {

	// Default config
	b, _ := json.Marshal(defaultConf)
	defaultConfig := bytes.NewReader(b)
	viper.SetConfigType("json")
	viper.ReadConfig(defaultConfig)
	Config.MergeConfigMap(viper.AllSettings())

	// Flags
	pflag.String("level", "info", "Log level")

	pflag.String("rtmp_addr", ":1935", "RTMP server listen address")
	pflag.Bool("enable_rtmps", false, "enable server session RTMPS")
	pflag.String("rtmps_cert", "server.crt", "cert file path required for RTMPS")
	pflag.String("rtmps_key", "server.key", "key file path required for RTMPS")
	pflag.Bool("enable_tls_verify", true, "Use system root CA to verify RTMPS connection, set this flag to false on Windows")

	pflag.String("hls_addr", ":7002", "HLS server listen address")
	pflag.Bool("hls_keep_after_end", false, "Maintains the HLS after the stream ends")

	pflag.Int("write_timeout", 10, "write time out")
	pflag.Int("gop_num", 1, "gop num")

	pflag.Parse()
	Config.BindPFlags(pflag.CommandLine)

	// File
	Config.SetConfigFile("livehub.yml")
	Config.AddConfigPath(glib.AppBaseDir())
	err := Config.ReadInConfig()
	if err != nil {
		glog.WarnF("读取配置文件失败=[%v]", err.Error())
		glog.Info("使用默认配置")
	} else {
		Config.MergeInConfig()
	}

	// Environment
	replacer := strings.NewReplacer(".", "_")
	Config.SetEnvKeyReplacer(replacer)
	Config.AllowEmptyEnv(true)
	Config.AutomaticEnv()

	// Print final config
	c := ServerCfg{}
	Config.Unmarshal(&c)
	//glog.InfoF("Current configurations: \n%# v", pretty.Formatter(c))
}
