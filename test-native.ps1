#!/usr/bin/env pwsh
# Native Image Test Script for asu-java-build
# Tests native image building features with environment checks
# Run: powershell -ExecutionPolicy Bypass -File test-native.ps1

$ErrorActionPreference = "Stop"
$global:TestsPassed = 0
$global:TestsFailed = 0

# Setup GraalVM environment if available
$graalvmPath = "C:\green\graalvm-jdk-21.0.7"
if (Test-Path $graalvmPath) {
    $env:JAVA_HOME = $graalvmPath
    $env:GRAALVM_HOME = $graalvmPath
    $env:PATH = "$graalvmPath\bin;$env:PATH"
    
    # Setup Visual Studio environment for native-image on Windows
    $vsPath = "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat"
    if (Test-Path $vsPath) {
        # Import VS environment variables into PowerShell
        $tempFile = [IO.Path]::GetTempFileName()
        cmd /c "`"$vsPath`" && set" > $tempFile
        Get-Content $tempFile | ForEach-Object {
            if ($_ -match "^(.*?)=(.*)$") {
                Set-Item -Path "env:$($matches[1])" -Value $matches[2]
            }
        }
        Remove-Item $tempFile
    }
}

# Color output functions
function Write-TestHeader {
    param([string]$Message)
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host $Message -ForegroundColor Cyan
    Write-Host "========================================`n" -ForegroundColor Cyan
}

function Write-TestSuccess {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
    $global:TestsPassed++
}

function Write-TestFailure {
    param([string]$Message, [string]$Error)
    Write-Host "✗ $Message" -ForegroundColor Red
    Write-Host "  Error: $Error" -ForegroundColor Red
    $global:TestsFailed++
}

function Write-TestInfo {
    param([string]$Message)
    Write-Host "→ $Message" -ForegroundColor Yellow
}

function Write-TestSkip {
    param([string]$Message, [string]$Reason)
    Write-Host "⊘ $Message - SKIPPED" -ForegroundColor Gray
    Write-Host "  Reason: $Reason" -ForegroundColor Gray
}

# Check prerequisites
Write-TestHeader "Checking Prerequisites"

# Check Docker
$dockerAvailable = $false
try {
    $null = docker --version 2>&1
    if ($LASTEXITCODE -eq 0) {
        $dockerVersion = docker --version 2>&1 | Select-String "Docker version" | Out-String
        Write-TestInfo "Docker found: $($dockerVersion.Trim())"
        
        # Check if Docker is running - ignore stderr (warnings are normal)
        $ErrorActionPreference = "Continue"
        $null = docker ps 2>&1
        $ErrorActionPreference = "Stop"
        
        if ($LASTEXITCODE -eq 0) {
            $dockerAvailable = $true
            Write-TestSuccess "Docker is running"
        } else {
            Write-TestInfo "Docker is installed but not running"
        }
    }
} catch {
    Write-TestInfo "Docker not found"
}

# Check GraalVM native-image
$graalvmAvailable = $false
try {
    $nativeImageVersion = native-image --version 2>&1 | Out-String
    if ($LASTEXITCODE -eq 0) {
        Write-TestInfo "GraalVM native-image found: $($nativeImageVersion.Trim())"
        $graalvmAvailable = $true
        Write-TestSuccess "GraalVM native-image is available"
    }
} catch {
    Write-TestInfo "GraalVM native-image not found"
}

# Summary of prerequisites
if (-not $dockerAvailable -and -not $graalvmAvailable) {
    Write-Host "`n⚠️  No native image build environment available!" -ForegroundColor Yellow
    Write-Host "   Please install one of the following:" -ForegroundColor Yellow
    Write-Host "   - Docker Desktop (for Docker-based native builds)" -ForegroundColor Yellow
    Write-Host "   - GraalVM with native-image (for local native builds)" -ForegroundColor Yellow
    Write-Host "`nExiting without running tests.`n" -ForegroundColor Yellow
    exit 0
}

# Build the tool first
Write-TestHeader "Building asu-java-build"
try {
    Write-TestInfo "Running: go build -o asu-java-build.exe ."
    go build -o asu-java-build.exe .
    if ($LASTEXITCODE -ne 0) {
        throw "Build failed with exit code $LASTEXITCODE"
    }
    Write-TestSuccess "Tool built successfully"
} catch {
    Write-TestFailure "Tool build" $_.Exception.Message
    exit 1
}

# Prepare test project (Fat JAR required for native-image)
Write-TestHeader "Preparing Test Project"
try {
    $testDir = "examples\simple-standalone"
    Push-Location $testDir
    
    # Clean previous builds
    Remove-Item -Path "bin" -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item -Path "classes" -Recurse -Force -ErrorAction SilentlyContinue
    
    Write-TestInfo "Building Fat JAR for native image testing"
    ..\..\asu-java-build.exe -main com.example.Main -static
    
    if ($LASTEXITCODE -ne 0) {
        throw "Fat JAR build failed"
    }
    
    if (-not (Test-Path "bin\app.jar")) {
        throw "JAR file not created"
    }
    
    Write-TestSuccess "Fat JAR prepared"
    Pop-Location
} catch {
    Pop-Location -ErrorAction SilentlyContinue
    Write-TestFailure "Test project preparation" $_.Exception.Message
    exit 1
}

# Test 1: Docker Native Build (Linux)
if ($dockerAvailable) {
    Write-TestHeader "Test 1: Docker Native Build (Linux)"
    try {
        $testDir = "examples\simple-standalone"
        Push-Location $testDir
        
        Write-TestInfo "Building native image via Docker (this may take several minutes...)"
        Write-TestInfo "Command: asu-java-build.exe -jar bin\app.jar -native"
        
        $startTime = Get-Date
        ..\..\asu-java-build.exe -jar bin\app.jar -native
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalSeconds
        
        if ($LASTEXITCODE -ne 0) {
            throw "Native image build failed"
        }
        
        if (-not (Test-Path "bin\app-linux-amd64")) {
            throw "Native binary not created"
        }
        
        $fileSize = (Get-Item "bin\app-linux-amd64").Length / 1MB
        Write-TestInfo "Native binary size: $([math]::Round($fileSize, 2)) MB"
        Write-TestInfo "Build time: $([math]::Round($duration, 1)) seconds"
        
        Pop-Location
        Write-TestSuccess "Docker native build (Linux)"
    } catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-TestFailure "Docker native build (Linux)" $_.Exception.Message
    }
} else {
    Write-TestSkip "Test 1: Docker Native Build (Linux)" "Docker not available"
}

# Test 2: Docker Native Build with Custom Arguments
if ($dockerAvailable) {
    Write-TestHeader "Test 2: Docker Native Build with Custom Arguments"
    try {
        $testDir = "examples\simple-standalone"
        Push-Location $testDir
        
        # Clean previous native binary
        Remove-Item -Path "bin\app-linux-amd64" -Force -ErrorAction SilentlyContinue
        Remove-Item -Path "bin\myapp-native" -Force -ErrorAction SilentlyContinue
        
        Write-TestInfo "Building with custom native-image arguments"
        Write-TestInfo "Command: -jar bin\app.jar -native -native-name myapp-native -native-args '-Ob'"
        
        ..\..\asu-java-build.exe -jar bin\app.jar -native -native-name myapp-native -native-args "-Ob"
        
        if ($LASTEXITCODE -ne 0) {
            throw "Native image build failed"
        }
        
        if (-not (Test-Path "bin\myapp-native")) {
            throw "Native binary with custom name not created"
        }
        
        Pop-Location
        Write-TestSuccess "Docker native build with custom arguments"
    } catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-TestFailure "Docker native build with custom arguments" $_.Exception.Message
    }
} else {
    Write-TestSkip "Test 2: Docker Native Build with Custom Arguments" "Docker not available"
}

# Test 3: Local GraalVM Native Build
if ($graalvmAvailable) {
    Write-TestHeader "Test 3: Local GraalVM Native Build"
    try {
        $testDir = "examples\simple-standalone"
        Push-Location $testDir
        
        # Clean previous native binaries
        Remove-Item -Path "bin\app-windows-amd64.exe" -Force -ErrorAction SilentlyContinue
        Remove-Item -Path "bin\app-linux-amd64" -Force -ErrorAction SilentlyContinue
        
        Write-TestInfo "Building native image with local GraalVM (this may take several minutes...)"
        Write-TestInfo "Command: asu-java-build.exe -jar bin\app.jar -native -native-driver local"
        
        $startTime = Get-Date
        ..\..\asu-java-build.exe -jar bin\app.jar -native -native-driver local
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalSeconds
        
        if ($LASTEXITCODE -ne 0) {
            throw "Local native image build failed"
        }
        
        # Check for Windows or Linux binary depending on OS
        $expectedBinary = if ($IsWindows -or $env:OS -match "Windows") {
            "bin\app-windows-amd64.exe"
        } else {
            "bin\app-linux-amd64"
        }
        
        if (-not (Test-Path $expectedBinary)) {
            throw "Native binary not created at $expectedBinary"
        }
        
        $fileSize = (Get-Item $expectedBinary).Length / 1MB
        Write-TestInfo "Native binary size: $([math]::Round($fileSize, 2)) MB"
        Write-TestInfo "Build time: $([math]::Round($duration, 1)) seconds"
        
        # Test execution
        Write-TestInfo "Testing native binary execution"
        $output = & ".\$expectedBinary" 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "Native binary execution failed"
        }
        
        Pop-Location
        Write-TestSuccess "Local GraalVM native build and execution"
    } catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-TestFailure "Local GraalVM native build" $_.Exception.Message
    }
} else {
    Write-TestSkip "Test 3: Local GraalVM Native Build" "GraalVM native-image not available"
}

# Test 4: Build from Source + Native (one-step)
if ($dockerAvailable) {
    Write-TestHeader "Test 4: Source to Native Image (One-Step Build)"
    try {
        $testDir = "examples\simple-standalone"
        Push-Location $testDir
        
        # Clean everything
        Remove-Item -Path "bin" -Recurse -Force -ErrorAction SilentlyContinue
        Remove-Item -Path "classes" -Recurse -Force -ErrorAction SilentlyContinue
        
        Write-TestInfo "Building from source directly to native image"
        Write-TestInfo "Command: -main com.example.Main -static -native"
        
        ..\..\asu-java-build.exe -main com.example.Main -static -native
        
        if ($LASTEXITCODE -ne 0) {
            throw "One-step build failed"
        }
        
        if (-not (Test-Path "bin\app.jar")) {
            throw "Intermediate JAR not created"
        }
        
        if (-not (Test-Path "bin\app-linux-amd64")) {
            throw "Native binary not created"
        }
        
        Pop-Location
        Write-TestSuccess "Source to native image (one-step)"
    } catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-TestFailure "Source to native image (one-step)" $_.Exception.Message
    }
} else {
    Write-TestSkip "Test 4: Source to Native Image (One-Step Build)" "Docker not available"
}

# Test 5: Maven Project Native Build
if ($dockerAvailable) {
    Write-TestHeader "Test 5: Maven Project Native Build"
    try {
        $testDir = "examples\maven-test"
        Push-Location $testDir
        
        # Clean previous builds
        Remove-Item -Path "bin" -Recurse -Force -ErrorAction SilentlyContinue
        Remove-Item -Path "target" -Recurse -Force -ErrorAction SilentlyContinue
        
        Write-TestInfo "Building Maven project to native image"
        Write-TestInfo "Command: -main com.example.MavenApp -static -native"
        
        ..\..\asu-java-build.exe -main com.example.MavenApp -static -native
        
        if ($LASTEXITCODE -ne 0) {
            throw "Maven native build failed"
        }
        
        if (-not (Test-Path "bin\app.jar")) {
            throw "JAR not created"
        }
        
        if (-not (Test-Path "bin\app-linux-amd64")) {
            throw "Native binary not created"
        }
        
        Pop-Location
        Write-TestSuccess "Maven project native build"
    } catch {
        Pop-Location -ErrorAction SilentlyContinue
        Write-TestFailure "Maven project native build" $_.Exception.Message
    }
} else {
    Write-TestSkip "Test 5: Maven Project Native Build" "Docker not available"
}

# Print final summary
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "NATIVE IMAGE TEST SUMMARY" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Total Tests Passed: $global:TestsPassed" -ForegroundColor Green
Write-Host "Total Tests Failed: $global:TestsFailed" -ForegroundColor $(if ($global:TestsFailed -eq 0) { "Green" } else { "Red" })
Write-Host "========================================`n" -ForegroundColor Cyan

if ($global:TestsFailed -gt 0) {
    exit 1
}
