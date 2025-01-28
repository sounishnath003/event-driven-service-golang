
.PHONY: compose-up
compose-up:
	docker-compose stop
	docker-compose up --build

.PHONY: runtests
runtests:
	k6 run apitests.js