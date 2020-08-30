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
	"sync"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)

}

// Connection struct is the type of data for a connection
type Connection struct {
	stream   proto.Broadcast_CreateStreamServer
	Id       string          `json:"stream"`
	Active   bool            `json:"active"`
	Topic    proto.Topic     `json:"topic"`
	Messages []proto.Message `json:"messages"`
	error    chan error
}

// Server is a collection of all the succesful connections to create a stream
type Server struct {
	Connection []*Connection `json:"connections"`
}

// TopicSub keeps a track of subs on a particular topic
type TopicSub struct {
	Topic    proto.Topic     `json:"topic"`
	Subs     int64           `json:"subs"`
	Messages []proto.Message `json:"messages"`
}

// GlobalTopic hold all topic data
type GlobalTopic struct {
	Topics []TopicSub `json:"topics"`
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []TopicSub, val string) (int, bool) {
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

func writeToFile(name string, globalT GlobalTopic) error {
	e, err := json.Marshal(globalT)
	if err != nil {
		return err
	}
	WriteErr := ioutil.WriteFile("test.json", e, 0644)
	return WriteErr
}

var gt GlobalTopic

// CreateStream ensures the successful connection to start stream
func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		Id:     pconn.User.Id,
		Active: true,
		Topic:  *pconn.Topic,
		error:  make(chan error),
	}

	_, result := Find(gt.Topics, conn.Topic.Name)
	if result == false {
		fmt.Println("New topic discovered", conn.Topic.Name)
		payload := TopicSub{
			Topic: conn.Topic,
			Subs:  0,
		}
		gt.Topics = append(gt.Topics, payload)
	} else {
		fmt.Println("old topic found")
		// payload := TopicSub{
		// 	subs: 1,
		// }
		// gt.topics = append(gt.topics, payload)
	}
	s.Connection = append(s.Connection, conn)
	// fmt.Println("CONN is HERE", &conn)
	// fmt.Println(gt)
	return <-conn.error
}

// BroadcastMessage is responsible for sending out messages
func (s *Server) BroadcastMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *proto.Message, conn *Connection) {
			defer wait.Done()

			if conn.Active && conn.Topic.Name == msg.Topic.Name {
				for idx, ts := range gt.Topics {
					if ts.Topic.Id == conn.Topic.Id {
						initSub := gt.Topics[idx].Subs
						gt.Topics[idx].Messages = append(gt.Topics[idx].Messages, *msg)
						gt.Topics[idx].Subs = initSub + msg.Id
						// fmt.Println(gt.topics[idx])
					}
				}
				err := conn.stream.Send(msg)
				conn.Messages = append(conn.Messages, *msg)
				if err != nil {
					grpcLog.Errorf("Error with Stream: %v , on topic: %v - Error: %v", conn.stream, conn.Topic, err)
					conn.Active = false
					conn.error <- err

				}

				fmt.Println(gt)
				writeToFile("first.json", gt)
				grpcLog.Info("Sending message to topic: ", conn.Topic)

			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	return &proto.Close{}, nil
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

	proto.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)

}
