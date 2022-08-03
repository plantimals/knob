all: build
	@echo "all"

build:
	go build -o knob ./knob.go
