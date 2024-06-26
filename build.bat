@echo off
@echo on

:: Running this build file from a different cwd will work because %~dp0 holds the directory that this file is in, so /src is always found correctly.

go build -C %~dp0src -o ../run-server.exe ./main
:: note: -C should always be the first flag, it doesn't work otherwise

:: -C moves the cwd of the go compiler to the specified directory

:: user authentication for sqlite: `go build -tags "sqlite_userauth"`
:: from: https://github.com/mattn/go-sqlite3?tab=readme-ov-file#user-authentication

:: Release-build in Go: `go build -ldflags "-s -w"`
:: For list of what -s and -w do, among other flags, see: https://pkg.go.dev/cmd/link
:: from: https://stackoverflow.com/questions/29599209/how-to-build-a-release-version-binary-in-go
