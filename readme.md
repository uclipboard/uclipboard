# uClipboard
This is a cross-platform clipboard client/server

**Still developing...**

# Plan
## Client 
### Platform
- [ ] Windows(A win-clipboard)
- [x] Linux(X/xclip)
- [x] Linux(wayland/wl-clipboard)
- [ ] Android APK
- [ ] Mac OS X (fulture)

## Server
### Platform
- Same as before 
### How to impl
- Golang
- [x] HTTP API
- WS API (fulture)

## Single Application
Single app to implement both client and server

## TODOLIST
- [x] Hostname source support
- Authentication(randomly password generator)
- Web UI
## Bug/Feature
In instant mode, you are not allowed to push clipboard cotent with "" (specific `-m ""` argument but in fact it is *empty* string)