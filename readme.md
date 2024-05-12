# uClipboard
This is a cross-platform clipboard client/server

**Still working...**

# Plan
## Client 
### Platform
- Windows
- Linux
- Mac OS X (fulture)

### How to impl
- Golang 
- clipboard polling checker (easy and cross-platform)

## Server
### Platform
- Same as before 
### How to impl
- Golang
- HTTP API
- WS API (fulture)

## Uniform Application
Single app to implement both client and server

## TODO
- Authentication(randomly password generator)
- Hostname source support
- Web UI
## Bug/Feature
In instant mode, you are not allowed to push clipboard cotent with "" (empty string but specific `-m ""` argument)