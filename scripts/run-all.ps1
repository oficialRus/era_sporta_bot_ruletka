# Запуск API, бота и Mini App одной командой.
# Откроются 3 окна консоли. Закройте их или Ctrl+C в каждом для остановки.

$projectRoot = if ($PSScriptRoot) { Split-Path -Parent (Split-Path -Parent $PSScriptRoot) } else { Get-Location }
if (-not (Test-Path "$projectRoot\cmd\api\main.go")) { $projectRoot = Get-Location }

Write-Host "Запуск в трёх окнах: API :8080, Бот, Mini App :5173" -ForegroundColor Cyan
Write-Host "Корень проекта: $projectRoot" -ForegroundColor Gray

Start-Process -FilePath "go" -ArgumentList "run", "./cmd/api" -WorkingDirectory $projectRoot -WindowStyle Normal
Start-Process -FilePath "go" -ArgumentList "run", "./cmd/bot" -WorkingDirectory $projectRoot -WindowStyle Normal
Start-Process -FilePath "go" -ArgumentList "run", "./cmd/serveweb" -WorkingDirectory $projectRoot -WindowStyle Normal

Write-Host "Готово. Откройте в боте ссылку на http://localhost:5173" -ForegroundColor Green
