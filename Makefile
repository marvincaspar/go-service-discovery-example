start: start-infra start-services

start-infra: stop-infra
	docker network create web || true
	cd consul && docker-compose up -d
	cd traefik && docker-compose up -d

start-services: stop-services
	cd greeting-service && go generate && docker-compose up -d --build
	cd user-service && go generate && docker-compose up -d --build

stop: stop-services stop-infra 

stop-infra:
	cd consul && docker-compose down
	cd traefik && docker-compose down

stop-services:
	cd greeting-service && docker-compose down
	cd user-service && docker-compose down