all:
	$(error please pick a target)

test:
	staticcheck ./...
	go test -v ./...

install-tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest