
build:
	go build -o bs-engine cmd/main.go

install: build
	cp bs-engine ~/src/3rdparty/gopath/bin/

