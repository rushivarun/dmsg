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

var client proto.BroadcastClient
var wait *sync.WaitGroup

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
	go func(str proto.Broadcast_CreateStreamClient) {
		defer wait.Done()
		for {
			msg, err := str.Recv()
			if err != nil {
				streamerror = fmt.Errorf("message recieve failed: %v", err)
				break
			}

			fmt.Printf("%v : %s\n", msg.Id, msg.Content)
		}

	}(stream)

	return streamerror
}

func main() {
	timestamp := time.Now()

	done := make(chan int)

	name := flag.String("N", "Guest", "name of the access")
	topicName := flag.String("T", "NewTopic", "Topic of message")

	flag.Parse()

	UserID := sha256.Sum256([]byte("USER" + *name))
	TopicID := sha256.Sum256([]byte("TOPIC" + *topicName))

	conn, err := grpc.Dial("localhost:8000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldnt connect to service: %v", err)
	}

	client = proto.NewBroadcastClient(conn)

	user := &proto.User{
		Id:   hex.EncodeToString(UserID[:]),
		Name: *name,
	}

	topic := &proto.Topic{
		Id:   hex.EncodeToString(TopicID[:]),
		Name: *topicName,
	}

	connect(user, topic)

	wait.Add(1)
	go func() {
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := &proto.Message{
				Id:        user.Id,
				Content:   scanner.Text(),
				Timestamp: timestamp.String(),
				Topic:     topic,
			}

			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("Error Sending Message: %v", err)
				break
			}
		}

	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done

}
