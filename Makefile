PHONY=test clean all

build:
	go build -o ~/.local/bin/gocal
run:
	gocal
