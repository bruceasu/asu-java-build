@echo off
setlocal
set CWD=%CD%
cd /d %~dp0

if /i "%~1"=="all" (
	call go-build -project java-build
	call cpFile.bat bin\java-build-windows-amd64.exe c:\green\jbuild.exe
) else (
	rem windows, linux, darwin
	set GOOS=windows

	rem amd64, 386
	set GOARCH=amd64

	set CGO_ENABLED=0

	go build -ldflags="-s -w" -v -o java-build.exe cmd\asu-java-build\main.go

	call cpFile.bat java-build.exe c:\green\jbuild.exe
)

:: GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o asu-java-build.exe .
:: GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o asu-java-build-linux
:: gbuild
cd /d %CWD%
endlocal

pause