# uClipboard
This is a cross-platform clipboard client/server

**Still developing...**

# Plan
## Client 
### Platform
- [x] Windows (my win-clipboard) https://github.com/dangjinghao/win-clip
- [x] **Linux**/BSD/Any other go supports platform with X display Protocol(xclip)
- [x] **Linux**/BSD/Any other go supports platform with Wayland display Protocol(wl-clipboard)
- [ ] web UI
- [ ] Android APK (fulture)
- [ ] Mac OS X (fulture / I don't have any Apple device :-( )
- [ ] Chrome OS (fulture)
*To now, the tested platforms are Windows and Linux (with X or wayland) I don't have BSD device with disply service*
## Server
### Platform
- Same as before 
### API supports
- [x] HTTP API
- [ ] WS API (fulture)

## Single Application
Single app to implement both client and server

## TODOLIST
- [x] Hostname source support
- [x] better error and trace message 
- [x] better status code /better client msg 
- [x] not full cover test of file upload on X11 and Windows 
- [x] test initial situation of DB 
- [x] Authentication
- [ ] more strict level debug log and hint
- [ ] project icon
## Bug/Feature
- In instant mode, you are not allowed to push clipboard cotent with "" (specific `-m ""` argument but in fact it is *empty* string)
- Only support TEXT copy/paste, but support upload file.  
- On X11, if the system clipboard is empty, this app may exit with msg: `adapter.Paste error:exit status 1`. You can copy any text to your clipboard then rerun it
- There is only `\n` in our db as newline rather than `\r\n`. So win-clip adapter will automaticlly switch them on Windows
