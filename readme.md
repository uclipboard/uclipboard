# uClipboard
This is a cross-platform clipboard client/server

**Still developing...**

# Plan
## Client 
### Platform
- [x] Windows (my win-clipboard) https://github.com/dangjinghao/win-clip
- [x] **Linux**/BSD/Any other go supports platform with X display Protocol(xclip)
- [x] **Linux**/BSD/Any other go supports platform with Wayland display Protocol(wl-clipboard)
- [ ] Android APK
- [ ] Chrome OS (fulture)
- [ ] Mac OS X (fulture / I don't have any Apple device :-( )
*To now, the tested platforms are Windows and Linux (with X or wayland) I don't have BSD device with disply service*
## Server
### Platform
- Same as before 
### API supports
- [x] HTTP API
- WS API (fulture)

## Single Application
Single app to implement both client and server

## TODOLIST
- [x] Hostname source support
- [ ] Web UI
- [ ] Authentication(randomly password generator)
## Bug/Feature
In instant mode, you are not allowed to push clipboard cotent with "" (specific `-m ""` argument but in fact it is *empty* string)