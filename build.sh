set -e 
go build .
GOOS=windows GOARCH=amd64 go build . 