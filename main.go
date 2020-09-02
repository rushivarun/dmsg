package main

import (
	"dmsg/proto"
	"os"

	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream proto.Deploy_CreateStreamServer `json:"stream"`
	ID     string                          `json:"ID"`
	Active bool                            `json:"Active"`
	Topic  proto.Topic                     `json:"Topic"`
	error  chan error
}

type Server struct {
	Connection []*Connection `json:"connections"`
}

type TopicSub struct {
	Topic    proto.Topic     `json:"topic"`
	Subs     int64           `json:"subs"`
	Messages []proto.Message `json:"messages"`
}

func main() {

}
