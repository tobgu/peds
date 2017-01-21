test:
	rm tests/array_test_gen.go
	cd tests && go generate
	go test ./...