all:
	$(error please pick a target)

test:
	staticcheck ./...
	govulncheck .
	gosec ./...
	go test -v ./...

install-tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | \
		sh -s -- -b $(shell go env GOPATH)/bin v2.14.0


# .bashrc or .zshrc
# export PATH=$(go env GOPATH)/bin:${PATH}
