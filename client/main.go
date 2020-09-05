package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"dmsg/proto"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type offset struct {
	idx int64
}

// MessageID has a Recieved list keeps a lost of all the recieved messages in order to avoid duplication.
type MessageID struct {
	Recieved []string
}

var client proto.DeployClient
var wait *sync.WaitGroup
var o offset
var recieved MessageID

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *proto.User, topic *proto.Topic) error {
	var streamerror error

	stream, err := client.CreateStream(context.Background(), &proto.Connect{
		User:   user,
		Topic:  topic,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	wait.Add(1)
	go func(str proto.Deploy_CreateStreamClient) {
		defer wait.Done()
		for {
			msg, err := str.Recv()
			if err != nil {
				streamerror = fmt.Errorf("message recieve failed: %v", err)
				break
			}

			_, Duplicates := Find(recieved.Recieved, msg.Id)

			if Duplicates {
				fmt.Println("Found Duplicates")
			} else {
				recieved.Recieved = append(recieved.Recieved, msg.Id)
				fmt.Printf("%v : %s\n", msg.Id, msg.Content)
			}
		}

	}(stream)

	return streamerror
}

func main() {
	timestamp := time.Now()
	o.idx = 1

	done := make(chan int)

	name := flag.String("N", "Guest", "name of the access")
	topicName := flag.String("T", "NewTopic", "Topic of message")
	UserScope := flag.String("US", "User", "Access Scope")
	TopicScope := flag.String("TS", "User", "Access scope for topic")

	flag.Parse()

	UserID := sha256.Sum256([]byte("USER" + *name))
	TopicID := sha256.Sum256([]byte("TOPIC" + *topicName))

	var UserScopeArray []string
	var TopicScopeArray []string

	UserScopeArray = append(UserScopeArray, *UserScope)
	TopicScopeArray = append(TopicScopeArray, *TopicScope)

	conn, err := grpc.Dial("localhost:8000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldnt connect to service: %v", err)
	}

	client = proto.NewDeployClient(conn)

	user := &proto.User{
		Id:    hex.EncodeToString(UserID[:]),
		Scope: UserScopeArray,
	}

	topic := &proto.Topic{
		Id:    hex.EncodeToString(TopicID[:]),
		Name:  *topicName,
		Scope: TopicScopeArray,
	}

	connect(user, topic)

	wait.Add(1)
	go func() {
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := &proto.Message{
				Id:        scanner.Text(),
				Offset:    o.idx,
				Content:   scanner.Text(),
				Timestamp: timestamp.String(),
				Topic:     topic,
			}

			_, err := client.DeployMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("Error Sending Message: %v", err)
				break
			}
			_, Qerr := client.QueueMessage(context.Background(), msg)
			if Qerr != nil {
				fmt.Printf("Error Queueing Message: %v", err)
				break
			}
			o.idx = o.idx + 1
		}

	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done

}
