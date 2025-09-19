@echo off

:: @set JAVA_HOME=%~dp0
:: @set GRAALVM_HOME=%~dp0
:: call :getScriptLocation location

@set JAVA_HOME=C:\green\graalvm-jdk-17.0.8+9.1
@set GRAALVM_HOME=C:\green\graalvm-jdk-17.0.8+9.1

@set PATH=%JAVA_HOME%bin;%JAVA_HOME%jre\bin;%PATH%
:: cmd /k "C:\Program Files (x86)\Microsoft Visual Studio\2017\BuildTools\Common7\Tools\VsDevCmd.bat"
:: @%comspec% /k "C:\Program Files (x86)\Microsoft Visual Studio\2017\BuildTools\VC\Auxiliary\Build\vcvars64.bat"
@%comspec% /k "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat"
:: àÍç±íËà ãZçI
:: call :getScriptLocation location
:: "%location%..\lib\installer\bin\gu.exe" %*
:: exit /b %errorlevel%
exit /b
:: If this script is in `%PATH%` and called quoted without a full path (e.g., `"js"`), `%~dp0` is expanded to `cwd`
:: rather than the path to the script.
:: This does not happen if `%~dp0` is accessed in a subroutine.
:getScriptLocation variableName
     set "%~1=%~dp0"
     exit /b 0