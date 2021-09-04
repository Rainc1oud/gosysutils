.ONESHELL:
SHELL := $(shell which bash)

.PHONY: all
all: test

.PHONY: test
test: clean
	go test -v ./...

.PHONY: clean
clean:
	rm -rf .test-*
