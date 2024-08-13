<h1 align="center">GnoSQL Database</h1>

GnoSQL is a lightweight, in-memory NoSQL database implemented in Go. It offers a simple and intuitive API for storing and retrieving data without the need for a persistent backend. With support for concurrent operations through goroutines.

## Author

-   [Nandakumar](https://github.com/nanda03dev/)

## Github repo

-   [GnoSQL](https://github.com/nanda03dev/gnosql)

## Installation

1. Run the following command to pull gnosql image:

```bash
docker pull gnosql
```

## Usage

To run application using docker img

```bash
docker run -p 5454:5454 gnosql
```

If you want to run database in specfic port you can pass PORT number as environment variables,

```bash
docker run -p 5454:3000 -e PORT=3000 gnosql
```

The application will start and listen for connections on port 5454. Use an HTTP client to send requests to the application.

To run application using docker compose file

Example `docker-compose.yml` file for gnosql
```bash
version: "3.9"
name: gnosql
services:
    gnosql:
        container_name: "gnosql"
        image: gnosql
        build:
            context: .
            dockerfile: Dockerfile
        ports:
            - 5454:5454
        volumes:
            - gnosqldb-data:/root/gnosql/db/
        environment:
            PORT: 5454
volumes:
    gnosqldb-data:
        name: gnosqldb-data
```

The database files are stored in the container path "/root/gnosql/db". To persist this data, please specify a host volume path and the corresponding container path before starting the container.

## GnoSQL Client

[GnoSQL Client](https://pkg.go.dev/github.com/nanda03dev/gnosql_client) is the Go library to connect with the GnoSQL database.

## Architecture

The database is structured to handle multiple operations across various collections in a concurrent yet synchronized manner. The architecture comprises the following key components:

- **gRPC Handler**: Receives gnosql-client requests and forwards them to the service handler.
- **Service Handler**: Validates incoming requests, creates events, and pushes them into the appropriate collection channels.
- **Workers**: Each collection has a dedicated worker that listens to its respective channel, processes events, and performs the corresponding CRUD operation.

## Key Components

### gRPC Handler

The gRPC Handler is the entry point for all client requests. It is responsible for receiving incoming requests and passing them to the service handler for further processing. This handler ensures that the requests are handled asynchronously and efficiently.

### Service Handler

The Service Handler is where the core logic of the request processing happens:

1. **Validation**: The incoming request is first validated to ensure all necessary inputs are correct and complete.
2. **Event Creation**: Once validated, an event is created that encapsulates the operation to be performed.
3. **Event Dispatch**: The event is then pushed into the appropriate channel associated with the collection that the operation targets.

### Worker

Workers are the heart of the concurrency model in this database:

- Each collection within the database has its own worker.
- Workers continuously listen to their respective channels for new events.
- Upon receiving an event, a worker processes it by executing the appropriate operation (create, update, delete) on the collection.
- Operations are processed sequentially within each collection to maintain data consistency.

## Operation Flow

The flow of operations in the GnoSQL Database is as follows:

1. **Client Request**: A request is sent from the gnosql-client and received by the gRPC Handler.
2. **Request Validation**: The Service Handler validates the request.
3. **Event Creation and Dispatch**: A valid request leads to the creation of an event, which is then dispatched to the correct collection's channel, After the event is pushed, the Service Handler sends an acknowledgment to the gRPC Handler, which then responds to the client.
4. **Worker Processing**: The worker associated with the collection processes the event, ensuring operations are performed in sequence.

This flow ensures that while multiple operations can be processed concurrently, they are done so in a way that maintains the integrity and consistency of the data.
