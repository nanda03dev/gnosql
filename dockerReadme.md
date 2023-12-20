<h1 align="center">GnoSQL Database</h1>

GnoSQL is a lightweight, in-memory NoSQL database implemented in Go. It offers a simple and intuitive API for storing and retrieving data without the need for a persistent backend. With support for concurrent operations through goroutines.

Resources

## Author

-   [Nandakumar](https://github.com/Nandha23311/)

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

### Database Endpoints

    * [POST]   /database/add
    * [POST]   /database/delete
    * [GET]    /database/get-all
    * [POST]   /collection/{databaseName}/add
    * [POST]   /collection/{databaseName}/delete
    * [GET]    /collection/{databaseName}/get-all
    * [GET]    /collection/{databaseName}/{collectionName}/stats
    * [POST]   /document/{databaseName}/{collectionName}/
    * [GET]    /document/{databaseName}/{collectionName}/{id}
    * [POST]   /document/{databaseName}/{collectionName}/filter
    * [PUT]    /document/{databaseName}/{collectionName}/{id}
    * [DELETE] /document/{databaseName}/{collectionName}/{id}
    * [GET]    /document/{databaseName}/{collectionName}/all-data

#### Swagger Documentation

    Explore the API using Swagger: [Swagger Documentation](/swagger/index.html#/)
