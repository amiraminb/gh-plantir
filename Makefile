BINARY_NAME := gh-plantir

.PHONY: build install clean

build:
	go build -o $(BINARY_NAME) .

install: build
	@gh extension list | grep -q plantir && echo "Extension already linked, build is sufficient." || gh extension install .

clean:
	rm -f $(BINARY_NAME)
