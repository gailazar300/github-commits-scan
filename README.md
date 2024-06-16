# AWS exposures scan for GitHub commits

This application scans the commits content of a given GitHub repository to find leaks
of access secrets of AWS .
## Installation

run

```bash 
  go mod tidy
```

## Build and run

```bash 
go build
./commits-scan -owner=<owner or organization name> -token=<github-token> -repo=<repo-name>
```

or directly run the main function
```bash 
  go run main.go -owner=<owner or organization name> -token=<github-token> -repo=<repo-name>
```

## Run tests

```bash 
  go test
```