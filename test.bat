@echo off

:: Running this build file from a different cwd will work because %~dp0 holds the directory that this file is in, so /src is always found correctly.

go test -C %~dp0src -v ./main
