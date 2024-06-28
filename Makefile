fmt:
	gci write . --skip-generated -s standard -s default
	gofumpt -l -w .

check:
	golangci-lint run --timeout 3m ./...
	staticcheck -go 1.22 ./...
