param(
  [string]$InstallDir = "$HOME\bin"
)

Write-Host "Installing SBBuddy.exe to $InstallDir…"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item -Path SBBuddy.exe -Destination $InstallDir -Force
[Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$env:Path", "User")
Write-Host "Done! Please restart your shell and run 'SBBuddy'."
