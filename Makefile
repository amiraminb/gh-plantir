BINARY_NAME := gh-plantir

.PHONY: build install clean

build:
	go build -o $(BINARY_NAME) .

install: build
	gh extension install .

clean:
	rm -f $(BINARY_NAME)
