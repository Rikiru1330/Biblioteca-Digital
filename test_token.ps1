# test_token.ps1
$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4iLCJpc3MiOiJsaWJyYXJ5LWFwaSIsImV4cCI6MTc2NDg0MjQwMywibmJmIjoxNzY0NzU2MDAzLCJpYXQiOjE3NjQ3NTYwMDN9.4Dze84EcSI5rvgahCKsHsUAHQmax97Rr86NQdYzl3R8"

Write-Host "=== PRUEBA CON TOKEN JWT ===" -ForegroundColor Cyan

# Configurar headers
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# 1. Verificar token
Write-Host "`n1. Verificando token (/me)..." -ForegroundColor Yellow
try {
    $me = Invoke-RestMethod -Uri "http://localhost:8080/me" -Method Get -Headers $headers
    Write-Host "   ✅ Token válido" -ForegroundColor Green
    Write-Host "   👤 Usuario: $($me.user_id)" -ForegroundColor Green
    Write-Host "   🎭 Rol: $($me.role)" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Token inválido: $_" -ForegroundColor Red
    exit
}

# 2. Crear libro con token
Write-Host "`n2. Creando libro protegido..." -ForegroundColor Yellow
$newBook = @{
    title = "Libro creado con JWT"
    author = "Autor Token"
    isbn = "TOKEN-$(Get-Random -Minimum 1000 -Maximum 9999)"
    published = 2024
    genre = "Testing"
    description = "Libro creado usando autenticación JWT"
}

try {
    $created = Invoke-RestMethod -Uri "http://localhost:8080/books" `
        -Method Post `
        -Body ($newBook | ConvertTo-Json) `
        -Headers $headers
    
    Write-Host "   ✅ Libro creado: $($created.title)" -ForegroundColor Green
    Write-Host "   📖 ID: $($created.id)" -ForegroundColor Green
    
    # Guardar ID para pruebas posteriores
    $bookId = $created.id
} catch {
    Write-Host "   ❌ Error creando libro: $_" -ForegroundColor Red
}

# 3. Prestar libro
Write-Host "`n3. Prestando libro..." -ForegroundColor Yellow
if ($bookId) {
    $loanRequest = @{
        book_id = $bookId
        user = "Lector con Token"
    }
    
    try {
        $loan = Invoke-RestMethod -Uri "http://localhost:8080/books/$bookId/borrow" `
            -Method Post `
            -Body ($loanRequest | ConvertTo-Json) `
            -Headers $headers
        
        Write-Host "   ✅ Libro prestado" -ForegroundColor Green
        Write-Host "   📝 ID Préstamo: $($loan.id)" -ForegroundColor Green
        $loanId = $loan.id
    } catch {
        Write-Host "   ❌ Error prestando libro: $_" -ForegroundColor Red
    }
}

# 4. Ver préstamos activos
Write-Host "`n4. Listando préstamos activos..." -ForegroundColor Yellow
try {
    $loans = Invoke-RestMethod -Uri "http://localhost:8080/loans?status=active" `
        -Method Get `
        -Headers $headers
    
    Write-Host "   ✅ Préstamos activos: $($loans.Count)" -ForegroundColor Green
    foreach ($loan in $loans) {
        Write-Host "   - Libro: $($loan.book_id) | Usuario: $($loan.user)" -ForegroundColor Gray
    }
} catch {
    Write-Host "   ❌ Error obteniendo préstamos: $_" -ForegroundColor Red
}

# 5. Devolver libro
Write-Host "`n5. Devolviendo libro..." -ForegroundColor Yellow
if ($loanId) {
    try {
        Invoke-RestMethod -Uri "http://localhost:8080/loans/$loanId/return" `
            -Method Post `
            -Headers $headers
        
        Write-Host "   ✅ Libro devuelto" -ForegroundColor Green
    } catch {
        Write-Host "   ❌ Error devolviendo libro: $_" -ForegroundColor Red
    }
}

Write-Host "`n=== PRUEBA COMPLETADA ===" -ForegroundColor Cyan