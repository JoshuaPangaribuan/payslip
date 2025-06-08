up-docker:
	docker compose up --build -d
	make migrate

clean-docker:
	make migrate-down
	docker compose down

restart-docker:
	make clean-docker
	make up-docker

colima-spin-up:
	docker-compose up --build -d
	make migrate

colima-clean-up:
	make migrate-down
	docker-compose down

up:
	docker compose up --build -d

clean:
	docker compose down

migrate:
	goose -dir=./migration postgres "user=root password=rootpassword dbname=payslip sslmode=disable" up
	goose -dir=./seed postgres "user=root password=rootpassword dbname=payslip sslmode=disable" up

migrate-down:
	goose -dir=./seed postgres "user=root password=rootpassword dbname=payslip sslmode=disable" down
	goose -dir=./migration postgres "user=root password=rootpassword dbname=payslip sslmode=disable" down

docs:
	swag init -g main.go -o ./docs

.PHONY: up down spin-docker clean-docker migrate migrate-down docs colima-spin-up colima-clean-up restart-docker

