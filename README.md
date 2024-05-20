# Go Chatroom (Updating)

## Overview
A chatroom based on websocket developed by go.
**Currently only TCP console version completed. Full version will be updated soon.**
## Getting Started

### Running the Server
In the terminal, start the server:
```sh
$ go run cmd/tcp/server.go
```

### Running the Clients
Start the client in multiple terminals (e.g., 3 clients):

```sh
$ go run cmd/tcp/client.go
Welcome, 127.0.0.1:49777, UID:1, Enter At:2020-01-31 16:15:24+8000
user:`2` has enter
user:`3` has enter

$ go run cmd/tcp/client.go
Welcome, 127.0.0.1:49781, UID:2, Enter At:2020-01-31 16:15:35+8000
user:`3` has enter

$ go run cmd/tcp/client.go
Welcome, 127.0.0.1:49784, UID:3, Enter At:2020-01-31 16:15:44+8000
```

Then, in the first client, type: `hello, I am first user` and all clients will receive the message.