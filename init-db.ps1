Write-Host "=== INICIALIZACIÃ“N DE BASE DE DATOS ===" -ForegroundColor Cyan
Write-Host ""

# Verificar directorio
if (-not (Test-Path "data")) {
    New-Item -ItemType Directory -Path "data" | Out-Null
    Write-Host "Carpeta 'data' creada" -ForegroundColor Green
}

# Verificar si ya existe
if (Test-Path "data/library.db") {
    Write-Host "ADVERTENCIA: La base de datos ya existe" -ForegroundColor Yellow
    $choice = Read-Host "Â¿Deseas borrarla y crear una nueva? (s/n)"
    if ($choice -ne 's') {
        Write-Host "OperaciÃ³n cancelada" -ForegroundColor Red
        exit
    }
    Remove-Item "data/library.db" -Force
    Write-Host "Base de datos anterior eliminada" -ForegroundColor Yellow
}

Write-Host "Creando nueva base de datos..." -ForegroundColor Yellow

# SQL para crear la base de datos
$sql = @"
-- Tabla de usuarios
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de libros
CREATE TABLE IF NOT EXISTS books (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    isbn TEXT UNIQUE NOT NULL,
    published INTEGER,
    genre TEXT,
    description TEXT,
    available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de prÃ©stamos
CREATE TABLE IF NOT EXISTS loans (
    id TEXT PRIMARY KEY,
    book_id TEXT NOT NULL,
    user TEXT NOT NULL,
    loan_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    return_date TIMESTAMP,
    returned BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE CASCADE
);

-- Ãndices
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_genre ON books(genre);
CREATE INDEX IF NOT EXISTS idx_books_available ON books(available);
CREATE INDEX IF NOT EXISTS idx_loans_book_id ON loans(book_id);
CREATE INDEX IF NOT EXISTS idx_loans_returned ON loans(returned);

-- Usuario admin por defecto
INSERT OR IGNORE INTO users (id, username, password, role) 
VALUES ('1', 'admin', 'admin123', 'admin');

-- Algunos libros de ejemplo
INSERT OR IGNORE INTO books (id, title, author, isbn, published, genre, description, available) VALUES
('b1', 'Cien aÃ±os de soledad', 'Gabriel GarcÃ­a MÃ¡rquez', '9780307474728', 1967, 'Realismo mÃ¡gico', 'Una obra maestra de la literatura latinoamericana', 1),
('b2', '1984', 'George Orwell', '9780451524935', 1949, 'DistopÃ­a', 'Novela sobre vigilancia y control totalitario', 1),
('b3', 'El principito', 'Antoine de Saint-ExupÃ©ry', '9780156012195', 1943, 'FÃ¡bula', 'Cuento filosÃ³fico para niÃ±os y adultos', 1);
"@

# Guardar SQL en archivo temporal
$tempFile = "temp_sql.sql"
$sql | Out-File -FilePath $tempFile -Encoding UTF8

# Ejecutar SQLite
sqlite3 data/library.db < $tempFile

# Eliminar archivo temporal
Remove-Item $tempFile -Force

if ($LASTEXITCODE -eq 0) {
    Write-Host "Â¡Base de datos creada exitosamente!" -ForegroundColor Green
    Write-Host ""
    
    # Mostrar resumen
    sqlite3 data/library.db @"
SELECT '=== RESUMEN DE LA BASE DE DATOS ===' as info;
SELECT 'Tablas creadas:' as mensaje; 
SELECT name as tabla FROM sqlite_master WHERE type='table' ORDER BY name;

SELECT '--- Conteo de registros ---' as info;
SELECT 'usuarios' as tabla, COUNT(*) as total FROM users 
UNION SELECT 'libros', COUNT(*) FROM books 
UNION SELECT 'prÃ©stamos', COUNT(*) FROM loans;
"@
} else {
    Write-Host "Error al crear la base de datos" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== PROCESO COMPLETADO ===" -ForegroundColor Green
Write-Host "Base de datos: data/library.db" -ForegroundColor Cyan
Write-Host "Usuario: admin / admin123" -ForegroundColor Cyan
