dev_generate:
	python util/create_templates.py

build: dev_generate
	go build github.com/tobgu/peds/cmd/peds/

install: dev_generate
	go install github.com/tobgu/peds/cmd/peds/

test: install
	-rm tests/vector_test_gen.go
	-rm tests/map_test_gen.go
	cd tests && go generate
	go test ./...

benchmark:
	rm tests/array_test_gen.go
	cd tests && go generate && go test -bench=.