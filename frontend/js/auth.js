// auth.js - Funciones de autenticación
async function login(username, password) {
    try {
        const response = await apiRequest('/login', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        });
        
        if (!response || !response.token) {
            throw new Error('Credenciales incorrectas');
        }
        
        // Guardar token y usuario
        localStorage.setItem('library_token', response.token);
        localStorage.setItem('library_user', JSON.stringify(response.user || { username, role: 'user' }));
        
        showMessage('Sesión iniciada correctamente', 'success');
        
        // Redirigir después de un breve delay
        setTimeout(() => {
            window.location.href = 'dashboard.html';
        }, 1000);
        
        return true;
        
    } catch (error) {
        showMessage(error.message || 'Error al iniciar sesión', 'error');
        return false;
    }
}

// Proteger páginas que requieren autenticación
function protectPage() {
    if (!isAuthenticated()) {
        window.location.href = 'login.html';
        return false;
    }
    return true;
}

// Obtener información del usuario actual para mostrar en UI
function getUserInfo() {
    const user = getCurrentUser();
    return user ? {
        name: user.username || 'Usuario',
        role: user.role || 'user',
        isAdmin: user.role === 'admin'
    } : null;
}

// Función para registrar nuevo usuario
async function register(username, email, password) {
    try {
        const response = await apiRequest('/register', {
            method: 'POST',
            body: JSON.stringify({ username, email, password })
        });
        
        if (!response || response.error) {
            throw new Error(response?.error || 'Error en el registro');
        }
        
        showMessage('Usuario registrado exitosamente', 'success');
        return true;
        
    } catch (error) {
        showMessage(error.message || 'Error al registrar usuario', 'error');
        return false;
    }
}