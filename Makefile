.PHONY: build run

run:
	templ generate
	go run cmd/main.go
