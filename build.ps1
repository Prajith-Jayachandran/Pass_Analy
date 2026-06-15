# build.ps1 - PowerShell script for cross-compiling the targets
$ErrorActionPreference = "Stop"

$distDir = "dist"
Write-Host "Cleaning and preparing target directory: $distDir/..."
if (Test-Path $distDir) {
    Remove-Item -Recurse -Force $distDir
}
New-Item -ItemType Directory -Path $distDir | Out-Null

$targets = @(
    @{ os = "linux"; arch = "amd64"; out = "pass_analy_linux_amd64" },
    @{ os = "linux"; arch = "arm64"; out = "pass_analy_linux_arm64" },
    @{ os = "darwin"; arch = "amd64"; out = "pass_analy_darwin_amd64" },
    @{ os = "darwin"; arch = "arm64"; out = "pass_analy_darwin_arm64" },
    @{ os = "windows"; arch = "amd64"; out = "pass_analy_windows_amd64.exe" },
    @{ os = "windows"; arch = "arm64"; out = "pass_analy_windows_arm64.exe" }
)

Write-Host "Beginning multi-compilation pipelines..."
foreach ($t in $targets) {
    Write-Host "  -> Building for $($t.os)/$($t.arch)..."
    $env:GOOS = $t.os
    $env:GOARCH = $t.arch
    go build -ldflags="-s -w" -o "$distDir/$($t.out)" .
}

# Reset environment variables to default
$env:GOOS = ""
$env:GOARCH = ""

Write-Host "Cross-compilation pipelines successfully finished. Binaries ready in $distDir/:"
Get-ChildItem $distDir
