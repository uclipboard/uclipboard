set -e 
go build -o uclipboard -ldflags="-s -w" 
GOOS=windows GOARCH=amd64 go build -o uclipboard -ldflags="-s -w"
