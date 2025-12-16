// api.js - Funciones compartidas para comunicación con la API
const API_BASE_URL = 'http://localhost:8080';

// Función para hacer requests a la API
async function apiRequest(endpoint, options = {}) {
    const token = localStorage.getItem('library_token');
    const defaultHeaders = {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    };
    
    if (token) {
        defaultHeaders['Authorization'] = `Bearer ${token}`;
    }
    
    const headers = { ...defaultHeaders, ...options.headers };
    
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            ...options,
            headers
        });
        
        // Si es 401 (no autorizado), redirigir a login
        if (response.status === 401) {
            logout();
            return null;
        }
        
        // Intentar parsear JSON
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            return await response.json();
        } else {
            return await response.text();
        }
        
    } catch (error) {
        console.error('API Error:', error);
        showMessage('Error de conexión con el servidor', 'error');
        return null;
    }
}

// Verificar autenticación
function isAuthenticated() {
    const token = localStorage.getItem('library_token');
    const user = localStorage.getItem('library_user');
    return !!(token && user);
}

// Obtener usuario actual
function getCurrentUser() {
    const userStr = localStorage.getItem('library_user');
    return userStr ? JSON.parse(userStr) : null;
}

// Cerrar sesión
function logout() {
    localStorage.removeItem('library_token');
    localStorage.removeItem('library_user');
    window.location.href = 'login.html';
}

// Mostrar mensajes
function showMessage(message, type = 'info') {
    // Eliminar mensajes existentes
    const existingMessages = document.querySelectorAll('.message');
    existingMessages.forEach(msg => msg.remove());
    
    // Crear nuevo mensaje
    const messageEl = document.createElement('div');
    messageEl.className = `message message-${type}`;
    
    // Icono según tipo
    const icons = {
        success: 'fa-check-circle',
        error: 'fa-exclamation-circle',
        info: 'fa-info-circle',
        warning: 'fa-exclamation-triangle'
    };
    
    const icon = icons[type] || icons.info;
    const colors = {
        success: '#00b894',
        error: '#e17055',
        info: '#74b9ff',
        warning: '#fdcb6e'
    };
    
    messageEl.innerHTML = `
        <i class="fas ${icon}" style="color: ${colors[type] || colors.info};"></i>
        <div>
            <strong style="text-transform: capitalize;">${type}:</strong>
            <div style="font-size: 14px; margin-top: 2px;">${message}</div>
        </div>
        <i class="fas fa-times" style="margin-left: 15px; cursor: pointer; color: #999;" 
           onclick="this.parentElement.remove()"></i>
    `;
    
    document.body.appendChild(messageEl);
    
    // Auto-eliminar después de 5 segundos
    setTimeout(() => {
        if (messageEl.parentElement) {
            messageEl.style.animation = 'slideOut 0.3s';
            setTimeout(() => messageEl.remove(), 300);
        }
    }, 5000);
}

// Cargar y aplicar header/nav a todas las páginas
function loadNavigation() {
    const user = getCurrentUser();
    const isAdmin = user && user.role === 'admin';
    
    // Determinar página activa basada en la URL
    const currentPage = window.location.pathname.split('/').pop();
    const pageMap = {
        'dashboard.html': 'dashboard',
        'books.html': 'books',
        'loans.html': 'loans',
        'external-search.html': 'search',
        'add-book.html': 'add-book'
    };
    
    const activePage = pageMap[currentPage] || 'dashboard';
    
    return `
        <div class="header">
            <div class="logo">
                <i class="fas fa-book-open"></i>
                <h1>Biblioteca Digital</h1>
            </div>
            
            <div class="nav">
                <a href="dashboard.html" class="nav-link ${activePage === 'dashboard' ? 'active' : ''}">
                    <i class="fas fa-tachometer-alt"></i> Dashboard
                </a>
                <a href="books.html" class="nav-link ${activePage === 'books' ? 'active' : ''}">
                    <i class="fas fa-book"></i> Libros
                </a>
                <a href="loans.html" class="nav-link ${activePage === 'loans' ? 'active' : ''}">
                    <i class="fas fa-exchange-alt"></i> Préstamos
                </a>
                <a href="external-search.html" class="nav-link ${activePage === 'search' ? 'active' : ''}">
                    <i class="fas fa-search"></i> Buscar Externo
                </a>
                
                ${isAdmin ? `
                    <a href="add-book.html" class="nav-link ${activePage === 'add-book' ? 'active' : ''}">
                        <i class="fas fa-plus-circle"></i> Agregar Libro
                    </a>
                ` : ''}
                
                <button class="logout-btn" onclick="logout()">
                    <i class="fas fa-sign-out-alt"></i> Cerrar Sesión
                </button>
            </div>
        </div>
    `;
}