<p align="center">GnoSQL Database App</p>

This is a simple NoSQL database application written in Golang. It demonstrates how to create, read, update, and delete data in a NoSQL database.

## Prerequisites

To run this application, you will need the following:

- Golang installed

## Installation

To install this application, follow these steps:

1. Clone this repository:
   ```bash
   git clone https://github.com/your-username/your-repo.git
   ```

2. Run the following command to install the dependencies:
   ```bash
   go mod download
   ```

## Usage

To run the application, you will need to configure the connection to your NoSQL database. Edit the `config.json` file, replacing the placeholder values with the actual connection details for your database.

Once configured, run the application with the following command:
```bash
go run main.go
```

To run application using docker 
```bash
docker build gnosql .

docker run -p 8080:8080 gnosql
```
If you want to run database in specfic port you can pass PORT number as environment variables, 
```bash
docker run -p 8080:3000 -e PORT=3000 gnosql
```

The application will start and listen for connections on port 8080. Use an HTTP client to send requests to the application. For example, to create a new user, send the following POST request:
```bash
curl -X POST http://localhost:8080/users -d '{ "name": "John Doe", "email": "johndoe@example.com" }'
```

To get a list of all users, send the following GET request:
```bash
curl http://localhost:8080/users
```

## Contributing

If you would like to contribute to this project, please fork this repository and create a pull request. Ensure that you follow the project's style guide and guidelines.
