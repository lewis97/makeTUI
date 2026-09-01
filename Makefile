.PHONY: dev-up
dev-up:
	@docker compose up -d

.PHONY: dev-down
dev-down:
	@docker compose down

.PHONY: dev-shell
dev-shell:
	@docker compose exec dev sh

.PHONY: build
build:
	go build -o dist/ ./cmd/...

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: run
run:
	./dist/barkeep | jq

# ########################### MongoDB ############################
.PHONY: connect-to-mongoDB
connect-to-mongoDB:
	@docker exec -it barkeep-mongoDB bash -c 'mongosh $$(printenv MONGO_INITDB_DATABASE) -u $$(printenv MONGO_BARKEEP_DB_USER) -p $$(printenv MONGO_BARKEEP_DB_PASS)'

.PHONY: seed-mongoDB
seed-mongoDB:
	@docker exec -it barkeep-mongoDB bash -c 'mongosh -f dev-scripts/seed-DB.js'
# ################################################################


# .PHONY: dev-db-connect
# dev-db-connect:
# 	psql -h postgres -U ${DB_USER} ${DB_NAME}
# # TODO: some more db commands to run migrations up and down (+ seed db?)

.PHONY: mocks
mocks:
	rm -rf mocks/
	mockery

.PHONY: unit
unit:
	go test ./...
