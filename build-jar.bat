@echo off
REM	• 编译 src 目录下 所有 .java（递归），输出到 bin
REM	• 只有一个源码文件时自动把它当作入口（尝试读取 package，否则用路径推断）
REM	• 多个文件时，脚本 第一个参数为入口类的全限定名（如 com.example.Main）
REM 若识别到入口类，会顺便在 bin\app.jar 里打一个可运行 JAR（含 Main-Class）
REM 
REM 用法示例：
REM 	• 单文件项目（例如 src\Hello.java）：
REM build.bat
REM 
REM 	• 多文件项目（指定入口类全限定名）：
REM build.bat com.example.Main
REM 
REM 注意：
REM 	• 需要已安装 JDK，并把 javac / jar 加入 PATH。
REM 	• 脚本会尝试从唯一源码文件里读取 package 声明；若没有，就按目录结构推断入口的全限定名。
REM 	• 若你不想打 JAR，可删掉脚本中“打可运行 JAR”的那一段（jar cfm ...）。

setlocal EnableExtensions EnableDelayedExpansion
REM === 配置 ===
set "SRC=src"
set "OUT=bin"

if not exist "%SRC%" (
  echo [ERR] Source folder "%SRC%" not found.
  exit /b 1
)
REM 确保输出目录存在
if not exist "%OUT%" mkdir "%OUT%"

REM 统计源文件数量
for /f %%A in ('dir /s /b "%SRC%\*.java" ^| find /c /v ""') do set SRC_COUNT=%%A
if "%SRC_COUNT%"=="0" (
  echo [ERR] No .java sources under "%SRC%".
  exit /b 1
)

REM 解析入口类
set "MAIN_CLASS=%~1"
if "%SRC_COUNT%"=="1" if "%MAIN_CLASS%"=="" (
  for /f "delims=" %%F in ('dir /s /b "%SRC%\*.java"') do set "ONLY_FILE=%%~fF"
  REM 优先从源码读取 package 声明
  set "PKG="
  for /f "tokens=2 delims= " %%P in ('findstr /r /c:"^package " "!ONLY_FILE!"') do set "PKG=%%P"
  if defined PKG set "PKG=!PKG:;=!"
  for %%I in ("!ONLY_FILE!") do set "CLASS_NAME=%%~nI"
  if defined PKG (
    set "MAIN_CLASS=!PKG!.!CLASS_NAME!"
  ) else (
    REM 无 package：按路径推断 FQCN
    for %%I in ("%SRC%") do set "SRC_ABS=%%~fI"
    set "REL=!ONLY_FILE:%SRC_ABS%\=!"
    set "REL=!REL:~0,-5!"  REM 去掉 .java
    set "MAIN_CLASS=!REL:\=.!"
  )
)

if "%SRC_COUNT%" NEQ "1" if "%MAIN_CLASS%"=="" (
  echo [ERR] Multiple source files detected.
  echo Usage: build.bat com.example.Main
  exit /b 2
)
REM 写临时源文件列表
set "SOURCES_FILE=%TEMP%\sources-%RANDOM%%RANDOM%.txt"
dir /s /b "%SRC%\*.java" > "%SOURCES_FILE%"

echo [INFO] Compiling %SRC_COUNT% source file^(s^) to "%OUT%"...
javac -d "%OUT%" @"%SOURCES_FILE%"
set COMP_RC=%ERRORLEVEL%
del "%SOURCES_FILE%" >nul 2>&1
if %COMP_RC% NEQ 0 (
  echo [ERR] javac failed.
  exit /b 3
)
REM 如有入口类，打可运行 JAR
if not "%MAIN_CLASS%"=="" (
  set "MANIFEST=%TEMP%\manifest-%RANDOM%%RANDOM%.mf"
  > "%MANIFEST%" echo Main-Class: %MAIN_CLASS%
  >> "%MANIFEST%" echo Class-Path: .
  >> "%MANIFEST%" echo.
  set "JAR=%OUT%\app.jar"
  if exist "%JAR%" del "%JAR%" >nul 2>&1
  jar cfm "%JAR%" "%MANIFEST%" -C "%OUT%" .
  del "%MANIFEST%" >nul 2>&1
  echo [OK] Compiled to "%OUT%" and built runnable JAR: "%JAR%"
  echo [RUN] java -jar "%JAR%"
) else (
  echo [OK] Compiled to "%OUT%".
  echo [RUN] java -cp "%OUT%" YourMainClass
)

exit /b 0
