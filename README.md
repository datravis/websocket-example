# websocket-example
This repo contains a websocket server and client, that implements a simple pub/sub model. A Client is able to subscribe to a topic, and received JSON messages published to that topic

## Running the Server

The server can be run via `go run cmd/server/main.go`. Output should look similar to below.

```
$ go run cmd/server/main.go 
Listening on localhost:8081
HTTP Endpoints:
GET /subscribe
POST /publish
```

## Running the Client

The client can be run via `go run cmd/client/main.go [topic]`. See an example below.

```
$ go run cmd/client/main.go pipeline_123
Connecting to ws://localhost:8081/subscribe?topic=pipeline_123
```

## Publishing a Message

Messages can be published via curl. See an example below. Note that the message body is required to be valid JSON. This validated server side.

```
curl -X POST 'http://localhost:8081/publish?topic=pipeline_123' -H "Content-Type: application/json" -d '{"type":"status", "message":"pipeline execution step 4 completed"}'
```

The submitted message should be received and logged by the client.

```
$ go run cmd/client/main.go pipeline_123
Connecting to ws://localhost:8081/subscribe?topic=pipeline_123
Received message [topic=pipeline_123]: {"message":"pipeline execution step 4 completed","type":"status"}
```