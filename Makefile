.PHONY: build install

build:
	cd src && go build -o ../build

install:
	cd src && go install
