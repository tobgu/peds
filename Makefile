dev_generate:
	python util/create_templates.py

build: dev_generate
	go build github.com/tobgu/peds/cmd/peds/

install: dev_generate
	go install github.com/tobgu/peds/cmd/peds/

fmt:
	go fmt ./...

test: install
	rm tests/*_gen.go
	cd tests && go generate
	go test ./...

benchmark_vector:
	rm tests/*_gen.go
	cd tests && go generate && go test -bench Iteration -run=^$

benchmark_map:
	rm tests/*_gen.go
	cd tests && go generate && go test -bench Map -run=^$
