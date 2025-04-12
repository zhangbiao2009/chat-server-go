package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	nick string
}

func NewClient(conn net.Conn) *Client {
	client := &Client{conn: conn}
	client.nick = RandString(4)
	return client
}

type ClientMgr struct {
	clientMap map[*Client]struct{}
	sync.RWMutex
}

func NewClientMgr() *ClientMgr {
	return &ClientMgr{clientMap: make(map[*Client]struct{})}
}

func (c *ClientMgr) AddClient(client *Client) {
	c.Lock()
	defer c.Unlock()
	c.clientMap[client] = struct{}{}
	fmt.Println("Client added:", client.nick)
}
func (c *ClientMgr) RemoveClient(client *Client) {
	c.Lock()
	defer c.Unlock()
	delete(c.clientMap, client)
	fmt.Println("Client removed:", client.nick)
}

// Send message to all clients
func (c *ClientMgr) SendMessage(self *Client, message string) {
	c.RLock()
	defer c.RUnlock()
	for client := range c.clientMap {
		if client == self {
			continue
		}
		client.conn.Write([]byte(self.nick + ": " + message))
	}
}

// Generate a random string of length n
func RandString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

var clientMgr = NewClientMgr()

func main() {
	// Specify the address and port for the server to listen on
	listenAddr := "127.0.0.1:12345"

	// Create a TCP server
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server is listening on", listenAddr)

	for {
		// Accept client connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		c := NewClient(conn)
		clientMgr.AddClient(c)

		// Start a goroutine to handle the connection
		go handleClient(c)
	}
}

func handleClient(c *Client) {
	defer c.conn.Close()
	defer clientMgr.RemoveClient(c)

	// Create a buffered reader using bufio
	reader := bufio.NewReader(c.conn)

	for {
		// Read a line of data until a newline character is encountered
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed:", err)
			break
		}

		if message[0] == '/' { // If it is a command
			// Handle the command
			parts := strings.SplitN(message, " ", 2)
			cmd := parts[0]
			if cmd == "/nick" && len(parts) > 1 {
				oldNick := c.nick
				c.nick = strings.Trim(parts[1], "\n")
				fmt.Println("Client renamed from", oldNick, "to", c.nick)
			}
			continue
		}

		// Print the received message
		fmt.Print("Received:", message)

		clientMgr.SendMessage(c, message)
	}
}
