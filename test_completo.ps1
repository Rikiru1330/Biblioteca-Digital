# test_completo.ps1
Write-Host "=== PRUEBA COMPLETA API BIBLIOTECA ===" -ForegroundColor Cyan

# 1. Verificar servidor
Write-Host "`n1. Verificando servidor..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
    Write-Host "   ✅ Servidor funcionando: $($health.status)" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Servidor no disponible" -ForegroundColor Red
    exit
}

# 2. Crear libro
Write-Host "`n2. Creando nuevo libro..." -ForegroundColor Yellow
$newBook = @{
    title = "Harry Potter y la piedra filosofal"
    author = "J.K. Rowling"
    isbn = "978-8478884452"
    published = 1997
    genre = "Fantasia"
    description = "Primer libro de la serie Harry Potter"
}

$createdBook = Invoke-RestMethod -Uri "http://localhost:8080/books" `
    -Method Post `
    -Body ($newBook | ConvertTo-Json) `
    -ContentType "application/json"

Write-Host "   ✅ Libro creado con ID: $($createdBook.id)" -ForegroundColor Green

# 3. Prestar libro
Write-Host "`n3. Prestando libro..." -ForegroundColor Yellow
$loanRequest = @{
    book_id = $createdBook.id
    user = "Diego Sanchez"
}

$loan = Invoke-RestMethod -Uri "http://localhost:8080/books/$($createdBook.id)/borrow" `
    -Method Post `
    -Body ($loanRequest | ConvertTo-Json) `
    -ContentType "application/json"

Write-Host "   ✅ Libro prestado. ID préstamo: $($loan.id)" -ForegroundColor Green

# 4. Ver libros (debería mostrar Available: false)
Write-Host "`n4. Listando libros (el prestado debe mostrar Available: false)..." -ForegroundColor Yellow
$books = Invoke-RestMethod -Uri "http://localhost:8080/books" -Method Get
$books | ForEach-Object {
    Write-Host "   - $($_.title) por $($_.author) | Disponible: $($_.available)"
}

# 5. Ver préstamos activos
Write-Host "`n5. Préstamos activos:" -ForegroundColor Yellow
$loans = Invoke-RestMethod -Uri "http://localhost:8080/loans?status=active" -Method Get
if ($loans.Count -eq 0) {
    Write-Host "   No hay préstamos activos" -ForegroundColor Gray
} else {
    $loans | ForEach-Object {
        Write-Host "   - Libro ID: $($_.book_id) | Usuario: $($_.user) | Fecha: $($_.loan_date)"
    }
}

Write-Host "`n=== PRUEBA COMPLETADA ===" -ForegroundColor Cyan