# uclipboard
A self-hosted cross-platform unified clipboard sychronize program.

**Warning: this program is still under development, API is not stable**

## Feature
- Copy a piece of text on one device, synchronize it to all connected devices.
- Upload a file and synchronize the download link to every connected device

## Supported Platforms
- [x] Windows: clipboard access supported by [win-clip](https://github.com/uclipboard/win-clip)
- [x] X Display Protocol: clipboard access supported by xclip
- [x] Wayland Display Protocol: clipboard access supported by wl-clipboard
- [x] web UI (built-in)
- [ ] Mac OS X (fulture, but I don't have any Apple device :-\( )

*To now, the tested platforms are Windows and Linux (with X and wayland). I don't have BSD device with disply service so can't test this combination.*

## How to Start
1. Specify and install the clipboard adapter
    - Windows: [win-clip](https://github.com/uclipboard/win-clip)(please add it's directory to PATH for executing it anywhere.)
    - Wayland Display Protocol: install `wl-clipboard` by package manager or something 
    - X Display Protocol: install `xclip` by package manager or something
2. Download this applicaton for your platform from release.
3. The program contains client and server implementation, so you have to download it on both your server and client.
4. Write the minimum config file `conf.toml`
```toml
token = "token" #necessary, modify it for your security!
[client]
server_url = "http://example:4533/" #necessary for clipent
adapter = "wl" #necessary, wl/xc/wc wl: wl-clipboard, xc: xclip, wc: win-clip
# X_selection = "clipboard" # select the clipboard which you want to copy/paste in xc mode
# check out `conf.full.toml` for more infomation
```
5. Run `server mode` on your server by execute `./uclipboard --mode server` 
    - If the `conf.toml` doesn't exist in the same directory of uclipboard program, you can indicate uclipboard to load a specific config file by execute `./uclipboard --mode server --conf /path/to/conf`
6. Run `client mode` on your laptop or PC by execute `./uclipboard --mode client`
    - you can directly copy previous `conf.toml` file as the `client mode` config file

## Supported Instant Operations
- pull the latest clipboard msg: `./uclipboard --pull`
- push clipboard msg to all devices: `./uclipboard --msg "hello world"`
- push clipboard msg by reading from stdin: `echo hello world|./uclipboard`
- upload file: `./uclipboard --upload /path/to/file`
- download file: `./uclipboard --download [file name]`
- download latest file: `./uclipboard --latest `
- download file by id: `./uclipboard --download @[id]`

## Contribute
**BUILDING**
## API supports
- [x] HTTP API
- [ ] WS API (fulture)

## TODOLIST
- [x] Hostname source support
- [x] better error and trace message 
- [x] better status code /better client msg 
- [x] not full cover test of file upload on X11 and Windows 
- [x] test initial situation of DB 
- [x] Authentication
- [x] project icon (Thanks to GitHub `identicon`)
- [x] web integration development && better gmake build
- [x] add version
- [x] long interval when network connection is down and recover the interval after connection recovered
- [ ] upload docker container automatically after every release
- [ ] more strict level debug log and hint
- [ ] history page
- [ ] cross-platform test

## Bug/Feature
- In instant mode, you are not allowed to push clipboard cotent with "" (specific `-m ""` argument but in fact it is *empty* string)
- uclipboard only supports TEXT copy/paste,**DO NOT** print binary file to the stdin of uclipboard. it is undefined behavior. But uclipboard supports upload file.  
- There is only `\n` in our db as newline rather than `\r\n`. So win-clip adapter will automaticlly switch them on Windows
