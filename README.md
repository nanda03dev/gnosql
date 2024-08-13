# GnoSQL Database

This project is an in-memory NoSQL database implemented in Golang. It leverages Golang's concurrency model, using goroutines, channels, and workers to handle multiple operations efficiently and in parallel. The database is designed to handle CRUD operations with high performance and consistency across collections.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Key Components](#key-components)
  - [gRPC Handler](#grpc-handler)
  - [Service Handler](#service-handler)
  - [Worker](#worker)
- [Operation Flow](#operation-flow)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [License](#license)

## Overview

GnoSQL Database is a lightweight, in-memory database built for fast, concurrent processing of data. It uses Golang's native concurrency features, such as goroutines and channels, to process multiple database operations simultaneously. Each collection within the database has its own set of channels and workers, ensuring that operations are properly synchronized and executed in a sequential order.

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

## Prerequisites

To run this application, you will need the following:

-   Golang installed

## Installation

To install this application, follow these steps:

1. Clone this repository:

    ```bash
    git clone https://github.com/nanda03dev/gnosql.git
    ```

2. Run the following command to install the dependencies:
    ```bash
    go mod download
    ```

## Usage

To run the application, you will need to configure the connection to your NoSQL database.

Once configured, run the application with the following command:

```bash
go run main.go
```

To run application using docker

```bash
docker build gnosql .

docker run -p 5454:5454 gnosql
```

If you want to run database in specfic port you can pass PORT number as environment variables,

```bash
docker run -p 5454:3000 -e PORT=3000 gnosql
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
