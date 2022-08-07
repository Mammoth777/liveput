linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./release/liveput-linux main.go

win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./release/liveput-win.exe main.go

mac:
	go build -o ./release/liveput-mac main.go