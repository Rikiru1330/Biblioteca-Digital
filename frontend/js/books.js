// books.js - Funciones específicas para libros
// (Ya están incluidas en los HTMLs, pero puedes moverlas aquí)

// Función para formatear fecha
function formatDate(dateString) {
    if (!dateString) return 'N/A';
    try {
        const date = new Date(dateString);
        return date.toLocaleDateString('es-ES', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    } catch (e) {
        return dateString;
    }
}

// Función para truncar texto
function truncateText(text, maxLength = 100) {
    if (!text) return '';
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
}

// Función para limpiar ISBN
function cleanISBN(isbn) {
    if (!isbn) return '';
    // Remover guiones y espacios
    return isbn.replace(/[-\s]/g, '');
}

// Función para validar ISBN
function isValidISBN(isbn) {
    const cleaned = cleanISBN(isbn);
    // ISBN-10 o ISBN-13
    return cleaned.length === 10 || cleaned.length === 13;
}