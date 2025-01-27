
.PHONY: compose-up
compose-up:
	docker-compose stop
	docker-compose up --build