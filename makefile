all:
	tsc
	go build .
	feed.exe

run:
	go build .
	feed.exe

build:
	tsc
	go build .