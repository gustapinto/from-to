build/linux:
	GOOS=linux GOARCH=amd64 go build -o bin/from_to_linux_amd64 ./cmd/from_to
	GOOS=linux GOARCH=arm64 go build -o bin/from_to_linux_arm64 ./cmd/from_to

build/windows:
	GOOS=windows GOARCH=amd64 go build -o bin/from_to_windows_amd64.exe ./cmd/from_to
	GOOS=windows GOARCH=arm64 go build -o bin/from_to_windows_arm64.exe ./cmd/from_to

build/macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/from_to_macos_amd64 ./cmd/from_to
	GOOS=darwin GOARCH=arm64 go build -o bin/from_to_macos_arm64 ./cmd/from_to

build: build/linux build/windows build/macos
