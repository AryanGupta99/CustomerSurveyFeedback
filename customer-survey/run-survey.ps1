<#
Usage: .\run-survey.ps1 [ZOHO_WEBHOOK_URL]

This script checks for Go in PATH, optionally sets ZOHO_WEBHOOK_URL from the first argument,
and runs the server with `go run`.

If Go is not found it prints installation instructions.
#>

param(
    [string]$ZohoUrl
)

function Ensure-Go {
    $g = Get-Command go -ErrorAction SilentlyContinue
    if (-not $g) {
        # Try common install paths
        $candidates = @(
            "C:\Program Files\Go\bin\go.exe",
            "C:\Go\bin\go.exe",
            "$env:LocalAppData\Programs\Go\bin\go.exe"
        )
        $found = $null
        foreach ($p in $candidates) {
            if (Test-Path $p) { $found = (Split-Path $p -Parent); break }
        }
        if ($found) {
            Write-Host "Found go.exe at: $found; adding to PATH for this session." -ForegroundColor Green
            $env:Path = $found + ";" + $env:Path
            return
        }

        Write-Host "Go (go.exe) was not found in PATH." -ForegroundColor Yellow
        Write-Host "Please install Go (https://go.dev/dl/) and re-open PowerShell, or add 'C:\Program Files\Go\bin' to your PATH." -ForegroundColor Cyan
        exit 1
    }
}

Ensure-Go

if ($ZohoUrl) {
    Write-Host "Setting ZOHO_WEBHOOK_URL environment variable for this session..."
    $env:ZOHO_WEBHOOK_URL = $ZohoUrl
    Write-Host "ZOHO_WEBHOOK_URL set to: $ZohoUrl"
} else {
    if (-not $env:ZOHO_WEBHOOK_URL) {
        Write-Host "Note: ZOHO_WEBHOOK_URL is not set. Submissions will be printed to stdout. Provide an argument to set it." -ForegroundColor Yellow
    } else {
        Write-Host "Using existing ZOHO_WEBHOOK_URL from environment."
    }
}

Write-Host "Starting server (go run ./cmd/survey)..." -ForegroundColor Green
go run .\cmd\survey\main.go
