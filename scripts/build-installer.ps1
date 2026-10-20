# Construye el .exe portable y el instalador per-user con Inno Setup.
#
# Uso:
#   powershell -ExecutionPolicy Bypass -File scripts\build-installer.ps1
#   powershell -ExecutionPolicy Bypass -File scripts\build-installer.ps1 -Version 1.3.0
#
# Requisitos:
#   - Wails CLI + Go + pnpm (para `wails build`).
#   - Inno Setup 6.3+ (ISCC.exe). Si no está en el PATH, se busca en Program Files.
param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    $wailsJson = Get-Content "wails.json" -Raw | ConvertFrom-Json
    $projectVersion = $wailsJson.Info.productVersion
    if (-not $Version) {
        $Version = $projectVersion
    }
    if ($Version -ne $projectVersion) {
        Write-Warning "La versión del instalador ($Version) no coincide con wails.json ($projectVersion). Actualiza Info.productVersion para que el .exe también la refleje."
    }

    $iscc = Get-Command iscc -ErrorAction SilentlyContinue
    $isccPath = $null
    if ($iscc) {
        $isccPath = $iscc.Source
    } else {
        $candidates = @(
            "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
            "$env:ProgramFiles\Inno Setup 6\ISCC.exe",
            "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe"
        )
        $isccPath = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
    }
    if (-not $isccPath) {
        throw "No se encontró ISCC.exe. Instala Inno Setup 6.3+ desde https://jrsoftware.org/isinfo.php"
    }

    if (Get-Process opencode-tray -ErrorAction SilentlyContinue) {
        Write-Warning "OpenCode Tray está en ejecución. Ciérralo desde el menú de la bandeja antes de compilar para poder reemplazar el .exe."
    }

    Write-Host "==> wails build" -ForegroundColor Cyan
    wails build -clean -webview2 download
    if ($LASTEXITCODE -ne 0) {
        throw "wails build falló (código $LASTEXITCODE)."
    }

    Write-Host "==> Inno Setup ($Version)" -ForegroundColor Cyan
    & $isccPath "/DMyAppVersion=$Version" "build\windows\installer\opencode-tray.iss"
    if ($LASTEXITCODE -ne 0) {
        throw "ISCC falló (código $LASTEXITCODE)."
    }

    Write-Host "`nArtefactos:" -ForegroundColor Green
    Write-Host "  Portable : build\bin\opencode-tray.exe"
    Write-Host "  Instalador: build\bin\OpenCodeTray-Setup-$Version.exe"
}
finally {
    Pop-Location
}
