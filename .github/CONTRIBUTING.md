Thank you for contributing to [Sir](https://github.com/komuw/sir/pull/11).                     


## To contribute:             

1. open an issue to discuss what you would like to contribute
2. upon approval of the said issue:
- fork this repo.
- make the changes you want on your fork.
- your changes should have backward compatibility in mind unless it is impossible to do so.
- add release notes to .github/RELEASE_NOTES.md
- add tests and benchmarks
- format your code using gofmt:                                          
- run tests(with race flag) and static analysis:
```bash
go test -timeout 1m -race -cover -v ./...
go test -timeout 1m -race -run=XXXX -bench=. ./...
go vet -v -all -shadow ./...
staticcheck -tests -show-ignored ./...
go build --race -o sir cmd/main.go
```
- open a pull request on this repo.          
          
NB: I make no commitment of accepting your pull requests.                 

