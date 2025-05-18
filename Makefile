build:
	go build -o bin/fs

run: build
	./bin/fs

test:
	go test ./... -v

delfoobar: 
	rm -rf foobar