package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	dicedb "github.com/dicedb/dicedb-go"
)

const (
	messageKey   = "chatroom:message"
	updatedAtKey = "chatroom:updated_at"
	pollInterval = 500 * time.Millisecond
	serverPort   = ":8080"
	diceDBAddr   = "localhost:7379"
)

var (
	ddb      *dicedb.Client
	isServer = flag.Bool("server", false, "Run as server")
	username = flag.String("user", "", "Username for chat")
)

func init() {
	ddb = dicedb.NewClient(&dicedb.Options{
		Addr: diceDBAddr,
	})
}

type Chatroom struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewChatroom() *Chatroom {
	ctx, cancel := context.WithCancel(context.Background())
	return &Chatroom{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Chatroom) RunServer() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("initializing chatroom...")
	err := ddb.Set(c.ctx, messageKey, "", 0).Err()
	if err != nil {
		return fmt.Errorf("initialization failed: failed to set the message key: %v", err)
	}

	err = ddb.Set(c.ctx, updatedAtKey, time.Now().Unix(), 0).Err()
	if err != nil {
		return fmt.Errorf("initialization failed: failed to set the updated_at key: %v", err)
	}

	fmt.Printf("chatroom is running on %s. waiting for clients...\n", serverPort)

	<-sigChan
	fmt.Println("shutting down server...")
	c.cancel()
	return nil
}

func (c *Chatroom) inputPromptLoop(username string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			fmt.Printf("(%s)> ", username)
			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("error reading input: %v\n", err)
				continue
			}
			text = strings.TrimSpace(text)

			if text != "" {
				message := fmt.Sprintf("%s:%d:%s",
					username, time.Now().Unix(), text,
				)

				err := ddb.Set(c.ctx, messageKey, message, 0).Err()
				if err != nil {
					fmt.Printf("error sending message: %v\n", err)
					continue
				}

				err = ddb.Set(c.ctx, updatedAtKey, time.Now().Unix(), 0).Err()
				if err != nil {
					fmt.Printf("error updating last update timestamp: %v\n", err)
				}
			}
		}
	}
}

func renderMessage(username string, msg string) {
	if msg == "" {
		return
	}

	parts := strings.Split(msg, ":")
	if len(parts) >= 3 {
		if parts[0] == username {
			return
		}

		msgTime, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			fmt.Printf("failed to parse timestamp: %v\n", err)
			return
		}
		timestamp := time.Unix(msgTime, 0)

		fmt.Printf("\n[%s] %s: %s\n",
			timestamp.Format("15:04:05"), parts[0],
			strings.Join(parts[2:], ":"))
	}
}

func (c *Chatroom) pollMessages() {
	var lastUpdatedAt int64

	if _update, err := ddb.Get(c.ctx, updatedAtKey).Result(); err == nil {
		if update, err := strconv.ParseInt(_update, 10, 64); err == nil {
			lastUpdatedAt = update
		}
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_update, err := ddb.Get(c.ctx, updatedAtKey).Result()
			if err != nil {
				fmt.Printf("could not check the status from the database: %v\n", err)
				continue
			}
			update, err := strconv.ParseInt(_update, 10, 64)
			if err != nil {
				update = 0
			}

			if update != lastUpdatedAt {
				msg, err := ddb.Get(c.ctx, messageKey).Result()
				if err != nil {
					fmt.Printf("failed to get message from the database: %v", err)
					continue
				}
				renderMessage(*username, msg)
				lastUpdatedAt = update
			}
		}
		time.Sleep(pollInterval)
	}
}

func (c *Chatroom) RunClient(username string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		c.cancel()
	}()

	go c.inputPromptLoop(username)
	// go c.pollMessages()  // poll
	go watchMessages() // push

	<-c.ctx.Done()
}

func watchMessages() {
	ctx := context.Background()
	watchConn := ddb.WatchConn(ctx)
	_, err := watchConn.GetWatch(ctx, messageKey)
	if err != nil {
		log.Println("failed to create watch connection:", err)
		return
	}
	defer watchConn.Close()

	ch := watchConn.Channel()
	for {
		select {
		case msg := <-ch:
			renderMessage(*username, msg.Data.(string))
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	flag.Parse()

	if *username == "" && !*isServer {
		fmt.Println("run either with -server or -user=<username>")
		os.Exit(1)
	}

	chatroom := NewChatroom()
	if !*isServer {
		chatroom.RunClient(*username)
		return
	}

	if err := chatroom.RunServer(); err != nil {
		fmt.Printf("error running server: %v\n", err)
		os.Exit(1)
	}
}
