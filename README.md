# dmsg
async pub-sub messaging 

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

### Run server docker container
```
docker run -p 8000:8000 dmsg-server
```

### Run client nodes
* Client nodes can have a username and subscribe to a particular topic in order to have a stream of messages on that topic.
* Clients can declare a username and topic to subscribe using arguments, -N for username and -T for topic to subscribe. 
```
cd client
go run main.go -N Rushi -T politics
```

## Upcoming...
* RBAC
* SYNC Pub sub
* ASYNC references
* Larger message types
* Distributed file system


