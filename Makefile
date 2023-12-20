run: 
	docker-compose -f ./docker-compose.local.yml up --build --remove-orphans

run-hub:
	docker-compose -f ./docker-compose.hub.yml up --build --remove-orphans

run-prod:
	docker-compose -f ./docker-compose.prod.yml up --build --remove-orphans
