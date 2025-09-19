@echo off
setlocal
set CWD=%CD%
cd /d %~dp0
if "%~1" == "all" (
	go-build -project java-build
	cpFile.bat bin\java-build-windows-amd64.exe c:\green\java-build.exe
) else (
	:: windows, linux, darwin
	set GOOS=windows
	:: amd64,  386
	set GOARCH=amd64
	go build -o java-build.exe .
	call cpFile.bat java-build.exe c:\green\java-build.exe
)

:: GOOS=windows GOARCH=amd64 go build -o asu-java-build.exe .
:: GOOS=linux GOARCH=amd64 go build -o asu-java-build-linux

cd /d %CWD%
endlocal

pause