.PHONY: build clean deps test

build:
	go build -o tracking-api.exe ./app

clean:
	rm -f tracking-api.exe

deps:
	cd app && go mod tidy
	cd domain/tracking && go mod tidy
	cd infrastructure && go mod tidy
	cd pkg && go mod tidy
	go work sync

test:
	go test ./... -v
