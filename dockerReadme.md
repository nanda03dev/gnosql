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

To run application using docker

```bash
docker run -p 8080:8080 gnosql
```

If you want to run database in specfic port you can pass PORT number as environment variables,

```bash
docker run -p 8080:3000 -e PORT=3000 gnosql
```

The application will start and listen for connections on port 8080. Use an HTTP client to send requests to the application.

### Database Endpoints

#### Add Database

    Endpoint: `/database/add`
    Method: POST
    Add a new database to GnoSQL.

#### Delete Database

    Endpoint: `/database/delete`
    Method: POST
    Delete an existing database from GnoSQL.

#### Get All Databases

    Endpoint: `/database/get-all`
    Method: GET
    Retrieve a list of all databases in GnoSQL.

### Collection Endpoints

#### Add Collection

    Endpoint: `/collection/{databaseName}/add`
    Method: POST
    Add a new collection to the specified database in GnoSQL.

#### Delete Collection

    Endpoint: `/collection/{databaseName}/delete`
    Method: POST
    Delete an existing collection from the specified database in GnoSQL.

#### Get All Collections

    Endpoint: `/collection/{databaseName}/get-all`
    Method: GET
    Retrieve a list of all collections in the specified database in GnoSQL.

#### Collection Statistics

    Endpoint: `/collection/{databaseName}/{collectionName}/stats`
    Method: GET
    Retrieve statistics for the specified collection in the specified database.

### Document Endpoints

#### Add Document

    Endpoint: `/document/{databaseName}/{collectionName}/`
    Method: POST
    Add a new document to the specified collection in the specified database.

#### Get Document by ID

    Endpoint: `/document/{databaseName}/{collectionName}/{id}`
    Method: GET
    Retrieve a document by ID from the specified collection in the specified database.

#### Filter Documents

    Endpoint: `/document/{databaseName}/{collectionName}/filter`
    Method: POST
    Filter documents in the specified collection in the specified database.

#### Update Document

    Endpoint: `/document/{databaseName}/{collectionName}/{id}`
    Method: PUT
    Update a document by ID in the specified collection in the specified database.

#### Delete Document

    Endpoint: `/document/{databaseName}/{collectionName}/{id}`
    Method: DELETE
    Delete a document by ID from the specified collection in the specified database.

#### Get All Documents

    Endpoint: `/document/{databaseName}/{collectionName}/all-data`
    Method: GET
    Retrieve all documents from the specified collection in the specified database.

#### Swagger Documentation

    Explore the API using Swagger: [Swagger Documentation](/swagger/index.html#/)
