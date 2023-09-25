.PHONY: create-keypair

PWD = $(shell pwd)
ACCTPATH = $(PWD)/account

create-keypair:
	@echo "Creating an rsa 256 key pair"
	openssl genpkey -algorithm RSA -out $(ACCTPATH)/rsa_private_$(ENV).pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -in $(ACCTPATH)/rsa_private_$(ENV).pem -pubout -out $(ACCTPATH)/rsa_public_$(ENV).pem

.PHONY: migrate-create migrate-up migrate-down migrate-force

MPATH = $(ACCTPATH)/migrations
PORT=5432

# Default number of migrations to execute up or down
N = 1

migrate-create:
	@echo "---Creating migration files---"
	migrate create -ext sql -dir $(MPATH) -seq -digits 5 $(NAME)

migrate-up:
	migrate -database postgres://postgres:password@localhost:$(PORT)/postgres?sslmode=disable -path $(MPATH) up $(N)

migrate-down:
	migrate -database postgres://postgres:password@localhost:$(PORT)/postgres?sslmode=disable -path $(MPATH) down $(N)

migrate-force:
	migrate -database postgres://postgres:password@localhost:$(PORT)/postgres?sslmode=disable -path $(MPATH) force $(VERSION)