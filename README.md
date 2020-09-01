# dmsg
distributed msg system

## To get started
Install go 1.15, git

### Clone the repo
```
git clone https://github.com/rushivarun/dmsg.git
cd dmsg
```

### Build the server dockerfile
```
docker build -t dmsg-server .
```

### Run client nodes
Client nodes can have a username and subscribe to a particular topic in order to have a stream of messages on that topic.
```
cd client
go run main.go -N Rushi -T politics
```


