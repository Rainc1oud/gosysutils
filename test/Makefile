.ONESHELL:
SHELL := $(shell which bash)

.PHONY: all
all: clean test

.PHONY: c
c: fallocate-test
fallocate-test:
	gcc -o $@ fallocate-test.c

.PHONY: test
test: c
	rm -f creserved
	./fallocate-test
	echo $$?
	ls -lah creserved

.PHONY: clean
clean:
	rm -f creserved fallocate-test
