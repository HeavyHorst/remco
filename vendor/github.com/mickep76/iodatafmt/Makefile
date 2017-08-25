all: test readme

format:
	gofmt -w=true .

test: format
	golint .
	go vet .
	go build

readme:
	godoc2md github.com/mickep76/iodatafmt >README.md
