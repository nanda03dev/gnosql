run: 
	docker-compose -f ./docker-compose.local.yml up --build --remove-orphans

run-hub:
	docker-compose -f ./docker-compose.hub.yml up --build --remove-orphans

run-prod:
	docker-compose -f ./docker-compose.prod.yml up --build --remove-orphans

run-build-img:
	docker build -f Dockerfile.prod -t gnosql:1.0.0 .

run-gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/gnosql.proto


# Define variables
DOCKER_IMAGE_NAME = gnosql-local:latest
DOCKER_COMPOSE_FILE = docker-compose.local.yml
STACK_NAME = gnosql

# Target to build the Docker image
build:
	docker build -t $(DOCKER_IMAGE_NAME) -f Dockerfile.local .

# Target to deploy the stack
deploy: build
	docker stack deploy -c $(DOCKER_COMPOSE_FILE) $(STACK_NAME)

# Target to remove the stack
remove:
	docker stack rm $(STACK_NAME)

# Target to rebuild and redeploy the stack
redeploy: remove deploy
˳