dev-up:
	docker compose up


dev-down:
	docker compose down

local-run:
	cd app && go run cmd/main.go
