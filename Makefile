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
