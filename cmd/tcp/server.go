// Copyright © 2020 polaris@studygolang.com.
// License: Apache Licence

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan string
}

func (u *User) String() string {
	return u.Addr + ", UID:" + strconv.Itoa(u.ID) + ", Enter At:" +
		u.EnterAt.Format("2006-01-02 15:04:05+8000")
}

// Message to be sent to the user
type Message struct {
	OwnerID int
	Content string
}

var (
	// New user arrives and registers through this channel
	enteringChannel = make(chan *User)
	// User leaves and registers through this channel
	leavingChannel = make(chan *User)
	// Channel dedicated to broadcasting user messages; buffered to avoid blocking in exceptional situations
	messageChannel = make(chan Message, 8)
)

// broadcaster records chat room users and broadcasts messages:
// 1. New user joins; 2. User messages; 3. User leaves
func broadcaster() {
	users := make(map[*User]struct{})

	for {
		select {
		case user := <-enteringChannel:
			// New user joins
			users[user] = struct{}{}
		case user := <-leavingChannel:
			// User leaves
			delete(users, user)
			// Avoid goroutine leaks
			close(user.MessageChannel)
		case msg := <-messageChannel:
			// Send messages to all online users
			for user := range users {
				if user.ID == msg.OwnerID {
					continue
				}
				user.MessageChannel <- msg.Content
			}
		}
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// 1. New user joins, create an instance of the user
	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}

	// 2. A new goroutine for reading, so we need another goroutine for writing
	// Communication between read and write goroutines can be done via channels
	go sendMessage(conn, user.MessageChannel)

	// 3. Send a welcome message to the current user; notify all users of the new user joining
	user.MessageChannel <- "Welcome, " + user.String()
	msg := Message{
		OwnerID: user.ID,
		Content: "user:`" + strconv.Itoa(user.ID) + "` has entered",
	}
	messageChannel <- msg

	// 4. Record the user to the global user list without using a lock
	enteringChannel <- user

	// Control timeout users and kick them out
	var userActive = make(chan struct{})
	go func() {
		d := 1 * time.Minute
		timer := time.NewTimer(d)
		for {
			select {
			case <-timer.C:
				conn.Close()
			case <-userActive:
				timer.Reset(d)
			}
		}
	}()

	// 5. Loop to read user input
	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg.Content = strconv.Itoa(user.ID) + ":" + input.Text()
		messageChannel <- msg

		// User active
		userActive <- struct{}{}
	}

	if err := input.Err(); err != nil {
		log.Println("Read error:", err)
	}

	// 6. User leaves
	leavingChannel <- user
	msg.Content = "user:`" + strconv.Itoa(user.ID) + "` has left"
	messageChannel <- msg
}

func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

// Generate user ID
var (
	globalID int
	idLocker sync.Mutex
)

func GenUserID() int {
	idLocker.Lock()
	defer idLocker.Unlock()

	globalID++
	return globalID
}
