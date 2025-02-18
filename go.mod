module github.com/gavv/webrtc-cli

go 1.12

require (
	github.com/gen2brain/malgo v0.10.35
	github.com/mattn/go-isatty v0.0.10
	github.com/mesilliac/pulse-simple v0.0.0-20170506101341-75ac54e19fdf
	github.com/pion/rtp v1.1.4
	github.com/pion/sdp/v2 v2.3.1
	github.com/pion/webrtc/v2 v2.1.12
	github.com/spf13/pflag v1.0.5
	github.com/youpy/go-riff v0.0.0-20131220112943-557d78c11efb // indirect
	github.com/youpy/go-wav v0.0.0-20160223082350-b63a9887d320
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	gopkg.in/hraban/opus.v2 v2.0.0-20220302220929-eeacdbcb92d0
)

replace gopkg.in/hraban/opus.v2 => github.com/gilwo/opus v0.0.0-20221027073344-2ead4004193d