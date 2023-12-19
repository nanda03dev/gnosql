<p align="center">GnoSQL Database App</p>

GnoSQL is a lightweight, in-memory NoSQL database implemented in Go. It offers a simple and intuitive API for storing and retrieving data without the need for a persistent backend. With support for concurrent operations through goroutines.

Resources

## Author

-   [Nandakumar](https://github.com/Nandha23311/)

## Prerequisites

To run this application, you will need the following:

-   Golang installed

## Installation

To install this application, follow these steps:

1. Clone this repository:

    ```bash
    git clone https://github.com/Nandha23311/gnosql.git
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

The application will start and listen for connections on port 5454. Use an HTTP client to send requests to the application.

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

## Contributing

If you would like to contribute to this project, please fork this repository and create a pull request. Ensure that you follow the project's style guide and guidelines.
