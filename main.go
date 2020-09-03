package main

import (
	"context"
	"dmsg/proto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

// Connection struct is the type of data for a connection
type Connection struct {
	Stream proto.Deploy_CreateStreamServer `json:"stream"`
	User   proto.User                      `json:"User"`
	Active bool                            `json:"Active"`
	Topic  *proto.Topic                    `json:"Topic"`
	error  chan error
}

// Server is a collection of all the succesful connections to create a stream
type Server struct {
	Connection []*Connection `json:"connections"`
}

// TopicOffset is a struct in order to keep track of the latest offset
// This has to kept local to the server...
type TopicOffset struct {
	TopicID string `json:"TopicID"`
	Offset  int64  `json:"Offset"`
}

// LocalStat is an array of all the Topics with their respective latest offsets.
// This has to be kept local to the server...
type LocalStat struct {
	TopicOffset []TopicOffset `json:"TopicOffset"`
}

// LS Instantiate LocalStat structure for topic data to be appended.
var LS LocalStat

// Instantiate Global topic to keep track of jsonable files to be distributed.
var gt proto.GlobalTopic

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []*proto.TopicWise, val string) (int, bool) {
	for i, item := range slice {
		if item.Topic.Name == val {
			return i, true
		}
	}
	return -1, false
}

func getFileStat(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Panic("Error while getting stat : ", err)
	}
	// get the size
	size := fi.Size()
	return size
}

func writeToFile(name string, globalT proto.GlobalTopic) error {
	e, err := json.Marshal(globalT)
	if err != nil {
		return err
	}
	WriteErr := ioutil.WriteFile("test.json", e, 0644)
	return WriteErr
}

func updateOffest(topicID string, offset int64) bool {
	for idx, val := range LS.TopicOffset {
		if val.TopicID == topicID {
			LS.TopicOffset[idx].Offset = offset
		}
	}
	return true
}

func getLatestOffset(topicID string) int64 {
	var offset int64
	for _, val := range LS.TopicOffset {
		if val.TopicID == topicID {
			offset = val.Offset
		}
	}
	return offset
}

func getValidScope(UserScopeArray []string, TopicScopeArray []string) bool {
	var dispatch bool
	for _, val := range TopicScopeArray {
		for _, kal := range UserScopeArray {
			if val == kal {
				dispatch = true
			} else {
				dispatch = false
			}
		}
	}
	return dispatch
}

// CreateStream ensures the successful connection to start stream
func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Deploy_CreateStreamServer) error {
	RbacAuth := getValidScope(pconn.User.Scope, pconn.Topic.Scope)
	var dispatch chan error
	if RbacAuth {
		grpcLog.Info("RBAC : User connected with valid scope")
		conn := &Connection{
			Stream: stream,
			User:   *pconn.User,
			Active: true,
			Topic:  pconn.Topic,
			error:  make(chan error),
		}

		_, result := Find(gt.TopicWise, conn.Topic.Name)
		if result == false {
			payload := &proto.TopicWise{
				Topic: conn.Topic,
			}
			gt.TopicWise = append(gt.TopicWise, payload)
		} else {
			fmt.Println("old topic found")
		}
		s.Connection = append(s.Connection, conn)
		dispatch = conn.error
	} else {
		grpcLog.Info("RBAC: User attempt blocked due to unauthorisation")
	}

	return <-dispatch
}

func (s *Server) DeployMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	return nil, nil
}

func (s *Server) QueueMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	return nil, nil
}

func main() {
	var connections []*Connection

	server := &Server{connections}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal("Error while listening on port 8000: ", err)
	}

	grpcLog.Info("Starting server on port 8000")

	proto.RegisterDeployServer(grpcServer, server)
	grpcServer.Serve(listener)
}
