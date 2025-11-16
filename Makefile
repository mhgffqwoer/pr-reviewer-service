.PHONY: docker-up docker-down

docker-up:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
	fi
	docker-compose -f deployment/docker-compose.yaml --env-file .env up --build -d

docker-down:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
	fi
	docker-compose -f deployment/docker-compose.yaml --env-file .env down -v
