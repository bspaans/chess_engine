PHONE: build install benchmark test test-all pprof

build:
	go build -o bs-engine cmd/main.go

install: build
	cp bs-engine ~/src/3rdparty/gopath/bin/

benchmark:
	go test -v -bench . -count 1 -cpuprofile cpu.out

test:
	go test -v .

test-all:
	INTEGRATION=1 go test -v

pprof: benchmark
	go tool pprof ./bs-engine cpu.out
