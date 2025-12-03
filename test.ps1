# test.ps1 - Script para ejecutar tests sin problemas de CGO

Write-Host "=== EJECUTANDO TESTS SIN CGO ===" -ForegroundColor Cyan

# Deshabilitar CGO
$env:CGO_ENABLED = "0"

Write-Host "`n1. Ejecutando tests de storage..." -ForegroundColor Yellow
go test ./storage -v

Write-Host "`n2. Ejecutando tests de handlers..." -ForegroundColor Yellow
go test ./handlers -v

Write-Host "`n3. Ejecutando todos los tests..." -ForegroundColor Yellow
go test ./... -v

Write-Host "`n=== TESTS COMPLETADOS ===" -ForegroundColor Green