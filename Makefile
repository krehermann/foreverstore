build:
	go build -buildvcs=false -o ./bin/fs

all:
	go build ./... 

run: build
	./bin/fs

test:
	go test  ./... -race
