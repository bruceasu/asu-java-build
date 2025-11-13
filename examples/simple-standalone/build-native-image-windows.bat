@echo off
:: Set up GraalVM environment
:: call ..\..\graalvm-env.bat

:: ..\..\bin\asu-java-build-windows-amd64.exe -jar bin\app.jar -native -native-args "--no-fallback -Ob" 
:: jbuild -jar bin\app.jar -native -native-target windows -native-args "--no-fallback -Ob" 

jbuild -jar bin\app.jar -native -native-driver local -native-target windows -native-args "--no-fallback -Ob" 