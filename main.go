package main

import (
	"context"
	"dmsg/proto"
	"fmt"
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
	id       string
	active   bool
	topic    proto.Topic
	messages []proto.Message
	error    chan error
}

// Server is a collection of all the succesful connections to create a stream
type Server struct {
	Connection []*Connection
}

// TopicSub keeps a track of subs on a particular topic
type TopicSub struct {
	topic    proto.Topic
	subs     int64
	messages []proto.Message
}

// GlobalTopic hold all topic data
type GlobalTopic struct {
	topics []TopicSub
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []TopicSub, val string) (int, bool) {
	for i, item := range slice {
		if item.topic.Name == val {
			return i, true
		}
	}
	return -1, false
}

var gt GlobalTopic

// CreateStream ensures the successful connection to start stream
func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		topic:  *pconn.Topic,
		error:  make(chan error),
	}

	_, result := Find(gt.topics, conn.topic.Name)
	if result == false {
		fmt.Println("New topic discovered", conn.topic.Name)
		payload := TopicSub{
			topic: conn.topic,
			subs:  0,
		}
		gt.topics = append(gt.topics, payload)
	} else {
		fmt.Println("old topic found")
		payload := TopicSub{
			subs: 1,
		}
		gt.topics = append(gt.topics, payload)
	}
	s.Connection = append(s.Connection, conn)
	fmt.Println("CONN is HERE", &conn)
	fmt.Println(gt)
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

			if conn.active && conn.topic.Name == msg.Topic.Name {
				for idx, ts := range gt.topics {
					if ts.topic.Id == conn.topic.Id {
						gt.topics[idx].messages = append(gt.topics[idx].messages, *msg)
					}
				}
				err := conn.stream.Send(msg)
				conn.messages = append(conn.messages, *msg)
				if err != nil {
					grpcLog.Errorf("Error with Stream: %v , on topic: %v - Error: %v", conn.stream, conn.topic, err)
					conn.active = false
					conn.error <- err

				}
				fmt.Println(gt)
				grpcLog.Info("Sending message to topic: ", conn.topic)

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
