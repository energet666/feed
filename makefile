all:
	tsc
	go build .
	feed.exe

allmac:
	tsc
	go build .
	./feed

run:
	go build .
	feed.exe

build:
	tsc
	go build .