
build:
	go build -o bs-engine fen.go main.go position.go piece.go move.go eval.go

install: build
	cp bs-engine ~/src/3rdparty/gopath/bin/

