.ONESHELL:
SHELL := $(shell which bash)

.PHONY: all
all: test

.PHONY: test
test: clean
	sudo go test -v ./...

.PHONY: clean
clean:
	sudo find .test-* -type d -exec umount {} \;
	sudo rm -rf .test-*
