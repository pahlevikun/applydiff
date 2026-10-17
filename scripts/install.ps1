# Install the latest applydiff release binary for Windows.
$ErrorActionPreference = "Stop"

$Repo = if ($env:APPLYDIFF_INSTALL_REPO) { $env:APPLYDIFF_INSTALL_REPO } else { "pahlevikun/applydiff" }
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:USERPROFILE\.local\bin" }

$Arch = $env:PROCESSOR_ARCHITECTURE
switch ($Arch) {
  "AMD64" { $Arch = "amd64" }
  "ARM64" { $Arch = "arm64" }
  default {
    Write-Error "install.ps1: unsupported CPU: $Arch"
    exit 1
  }
}

if ($env:APPLYDIFF_VERSION) {
  $Version = $env:APPLYDIFF_VERSION -replace '^v', ''
  $Tag = "v$Version"
} else {
  $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
  $Tag = $Release.tag_name
  $Version = $Tag -replace '^v', ''
}

$Archive = "applydiff_${Version}_windows_${Arch}.zip"
$Base = "https://github.com/$Repo/releases/download/$Tag"
$Tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP ("applydiff-install-" + [guid]::NewGuid().ToString())) -Force

$ArchivePath = Join-Path $Tmp.FullName $Archive
$ChecksumsPath = Join-Path $Tmp.FullName "checksums.txt"
Invoke-WebRequest -Uri "$Base/$Archive" -OutFile $ArchivePath
Invoke-WebRequest -Uri "$Base/checksums.txt" -OutFile $ChecksumsPath

$ChecksumLine = Get-Content $ChecksumsPath | Where-Object { $_ -match " $Archive$" } | Select-Object -First 1
if (-not $ChecksumLine) {
  Write-Error "install.ps1: $Archive not found in checksums.txt"
  exit 1
}
$Expected = ($ChecksumLine -split '\s+')[0].ToLower()
$Actual = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLower()
if ($Expected -ne $Actual) {
  Write-Error "install.ps1: checksum mismatch for $Archive"
  exit 1
}

Expand-Archive -Path $ArchivePath -DestinationPath $Tmp.FullName -Force
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
$Dest = Join-Path $InstallDir "applydiff.exe"
Copy-Item -Path (Join-Path $Tmp.FullName "applydiff.exe") -Destination $Dest -Force

Write-Host "Installed applydiff $Version -> $Dest"
Write-Host "Add $InstallDir to your PATH if needed."

Remove-Item -Recurse -Force $Tmp.FullName
