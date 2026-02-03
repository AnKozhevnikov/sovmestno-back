.PHONY: swagger
swagger:
	swag init --parseDependency --parseInternal

.PHONY: run
run:
	go run .

.PHONY: build
build:
	go build -o gateway .

.PHONY: test
test:
	go test -v ./...
