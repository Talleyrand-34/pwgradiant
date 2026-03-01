# Frontend Web para Gestor de Contraseñas

Frontend web en PHP con estilo KeePassXC para el gestor de contraseñas con TOTP homemórfico.

## Características

### ✨ Funcionalidades Principales

- 🔐 **Pantalla de Desbloqueo** - Autenticación con contraseña maestra
- 📋 **Lista de Cuentas** - Vista de todas las cuentas guardadas
- 🔍 **Búsqueda** - Filtrado rápido de cuentas
- ✏️ **CRUD Completo** - Crear, editar y eliminar cuentas
- 👁️ **Toggle de Contraseña** - Mostrar/ocultar contraseñas
- 📋 **Copiar al Portapapeles** - Copiar usuario, contraseña y URL
- 🎲 **Generador de Contraseñas** - Contraseñas aleatorias configurables
- ⏱️ **TOTP con Actualización Automática** - Códigos 2FA que se actualizan cada 30 segundos
- 🔒 **Conversión a Homemórfico** - Convertir TOTP estándar a cifrado homemórfico

### 🎨 Diseño

- Interfaz inspirada en KeePassXC
- Diseño responsivo
- Tema moderno con degradados
- Iconos Font Awesome
- Animaciones suaves
- Notificaciones visuales

## Requisitos

### Servidor

- PHP 7.4 o superior
- Extensión cURL habilitada
- Servidor web (Apache, Nginx, o PHP built-in server)

### Backend

- API REST del gestor de contraseñas ejecutándose en `http://localhost:8080`
- Base de datos SQLCipher inicializada

## Instalación

### 1. Copiar Archivos

```bash
# Copiar todos los archivos del frontend a tu directorio web
cp -r frontend/ /var/www/html/password-manager/
# O usar el servidor integrado de PHP (ver más abajo)
```

### 2. Configurar la API

Editar `config.php`:

```php
// URL de la API
define('API_BASE_URL', 'http://localhost:8080/api/v1');

// Ruta a la base de datos (solo informativa)
define('DB_PATH', 'passwords.db');

// Timeout de sesión (en segundos)
define('SESSION_TIMEOUT', 3600); // 1 hora
```

### 3. Configurar Permisos

```bash
# Dar permisos de lectura/escritura a la sesión
chmod 755 /var/www/html/password-manager
chmod 644 /var/www/html/password-manager/*.php
```

### 4. Iniciar el Backend

Asegúrate de que el servidor API está ejecutándose:

```bash
# En el directorio del backend
./pwdmgr server --port 8080 --key "tu-clave-maestra"
```

### 5. Acceder al Frontend

**Opción A: Servidor Web (Apache/Nginx)**
```
http://localhost/password-manager/
```

**Opción B: Servidor Integrado de PHP**
```bash
cd frontend
php -S localhost:8000
# Acceder a: http://localhost:8000
```

## Estructura del Proyecto

```
frontend/
├── index.php                 # Punto de entrada principal
├── config.php               # Configuración
├── api.php                  # Endpoint AJAX
│
├── includes/
│   └── api.php              # Clase wrapper de la API
│
├── views/
│   ├── main.php            # Vista principal de la aplicación
│   └── modals.php          # Modales (TOTP, contraseñas, etc.)
│
├── assets/
│   ├── css/
│   │   └── style.css       # Estilos principales
│   └── js/
│       └── main.js         # JavaScript de la aplicación
│
└── README.md               # Este archivo
```

## Uso

### Desbloqueo

1. Abre el frontend en tu navegador
2. Introduce tu **contraseña maestra** (la misma usada para el backend)
3. Haz clic en **Desbloquear**

### Gestión de Cuentas

#### Ver Cuentas
- Las cuentas se muestran en el panel izquierdo
- Haz clic en una cuenta para ver sus detalles

#### Búsqueda
- Usa la barra de búsqueda en la parte superior del panel izquierdo
- Filtra por usuario, email o URL

#### Nueva Cuenta
1. Haz clic en el botón **+** en la barra de herramientas
2. Rellena los campos:
   - **Usuario** (requerido)
   - **Contraseña** (requerido)
   - Email (opcional)
   - URL (opcional)
   - Notas (opcional)
   - Fecha de expiración (opcional)
3. Haz clic en **Guardar**

#### Editar Cuenta
1. Selecciona una cuenta
2. Haz clic en el botón **✏️ Editar** en la barra de herramientas
3. Modifica los campos
4. Haz clic en **Actualizar**

#### Eliminar Cuenta
1. Selecciona una cuenta
2. Haz clic en el botón **🗑️ Eliminar** en la barra de herramientas
3. Confirma la eliminación

#### Mostrar/Ocultar Contraseña
- Haz clic en el icono **👁️** junto a la contraseña
- La página se recargará mostrando u ocultando la contraseña

#### Copiar al Portapapeles
- **Usuario**: Botón de copia inline o botón en toolbar
- **Contraseña**: Botón de copia inline o botón en toolbar (requiere que la contraseña esté visible)
- **URL**: Botón de copia inline o botón en toolbar

### Generador de Contraseñas

1. Haz clic en el botón **🎲** en la barra de herramientas
2. Configura las opciones:
   - **Longitud**: Desliza el control (8-64 caracteres)
   - **Minúsculas (a-z)**: Activar/desactivar
   - **Mayúsculas (A-Z)**: Activar/desactivar
   - **Números (0-9)**: Activar/desactivar
   - **Caracteres especiales**: Activar/desactivar
3. Haz clic en **Regenerar** para crear una nueva contraseña
4. Opciones:
   - **Copiar**: Copiar al portapapeles
   - **Usar Contraseña**: Usar en el formulario actual
   - **Cancelar**: Cerrar sin usar

### TOTP (Autenticación de Dos Factores)

#### Ver Código TOTP
1. Selecciona una cuenta con TOTP configurado
2. Haz clic en el botón **⏱️ Ver TOTP** en la barra de herramientas
3. El código se muestra y se actualiza automáticamente cada 30 segundos
4. El círculo de progreso muestra el tiempo restante
5. Haz clic en **Copiar Código** para copiar al portapapeles

#### Información Mostrada
- **Código TOTP**: 6 dígitos que cambian cada 30 segundos
- **Temporizador**: Tiempo restante hasta el próximo código
- **Semilla TOTP**: 
  - Si es estándar: muestra la semilla en Base32
  - Si es homemórfico: muestra "🔒 Cifrada"
- **Estado de Cifrado**: Estándar o Homemórfico

#### Convertir a TOTP Homemórfico
1. Selecciona una cuenta con TOTP estándar
2. Opción A: Desde la vista de detalles, haz clic en el botón de conversión
3. Opción B: Desde el modal de TOTP, haz clic en **Convertir a Homemórfico**
4. Selecciona el tamaño de clave:
   - **2048 bits** (Recomendado) - Balance entre seguridad y rendimiento
   - **3072 bits** (Más seguro) - Mayor seguridad
   - **4096 bits** (Máxima seguridad) - Máxima seguridad pero más lento
5. Haz clic en **Convertir Ahora**
6. **⚠️ IMPORTANTE**: Guarda las claves privadas mostradas
   - **Paillier N (hex)**: Primera clave
   - **Paillier Lambda (hex)**: Segunda clave
   - Estas claves solo se muestran una vez
   - Son necesarias para generar códigos TOTP
   - Guárdalas en un lugar seguro (gestor de contraseñas, HSM, etc.)
7. Haz clic en **He Guardado las Claves** para cerrar

### Bloquear Base de Datos

- Haz clic en el botón **🔒** en la barra de herramientas
- La sesión se cerrará y volverás a la pantalla de desbloqueo
- La contraseña maestra se borra de la memoria

## Atajos de Teclado (Próximamente)

```
Ctrl + N    - Nueva entrada
Ctrl + E    - Editar entrada
Ctrl + D    - Eliminar entrada
Ctrl + B    - Copiar usuario
Ctrl + C    - Copiar contraseña
Ctrl + U    - Copiar URL
Ctrl + G    - Generar contraseña
Ctrl + T    - Ver TOTP
Ctrl + L    - Bloquear base de datos
```

## Seguridad

### Medidas Implementadas

1. **Sesiones Seguras**
   - Cookie HttpOnly
   - Cookie Secure (HTTPS)
   - SameSite Strict
   - Timeout automático (1 hora por defecto)

2. **Contraseña Maestra**
   - Nunca se almacena en disco
   - Solo se mantiene en memoria de sesión
   - Se borra al cerrar sesión o timeout

3. **Comunicación con API**
   - La contraseña maestra se envía en cada request
   - Se usa como clave de cifrado en el backend
   - No se almacena en el servidor API

4. **CSRF Protection**
   - Tokens CSRF en formularios (configurable)
   - Validación de origen

### Recomendaciones

1. **Usar HTTPS en producción**
   ```php
   // En config.php
   define('ENABLE_HTTPS_ONLY', true);
   ```

2. **Configurar timeouts apropiados**
   ```php
   // Timeout más corto para mayor seguridad
   define('SESSION_TIMEOUT', 900); // 15 minutos
   ```

3. **Deshabilitar errores en producción**
   ```php
   error_reporting(0);
   ini_set('display_errors', 0);
   ```

4. **Usar servidor web robusto**
   - Apache con mod_security
   - Nginx con rate limiting

## API REST Utilizada

### Endpoints Principales

```
GET  /api/v1/accounts                    # Listar cuentas
GET  /api/v1/accounts/:id                # Obtener cuenta
POST /api/v1/accounts                    # Crear cuenta
PUT  /api/v1/accounts/:id                # Actualizar cuenta
DELETE /api/v1/accounts/:id              # Eliminar cuenta

GET  /api/v1/accounts/search?q=...       # Buscar cuentas

GET  /api/v1/totp/account/:account_id    # Obtener TOTP
POST /api/v1/totp/:id/generate           # Generar código
POST /api/v1/totp/:id/convert-to-homomorphic  # Convertir
```

### Parámetros de Query

- `show_password=true` - Mostrar contraseñas en la respuesta

### Headers

- `X-Encryption-Key: {master_password}` - Contraseña maestra (en cada request)

## Personalización

### Colores y Tema

Editar `assets/css/style.css`:

```css
:root {
    --primary-color: #3498db;      /* Color primario */
    --primary-dark: #2980b9;       /* Color primario oscuro */
    --danger-color: #e74c3c;       /* Color de peligro */
    --success-color: #27ae60;      /* Color de éxito */
    /* ... más variables */
}
```

### Timeout de Sesión

Editar `config.php`:

```php
define('SESSION_TIMEOUT', 3600); // En segundos
```

### Tamaño de Ventana Modal

Editar `assets/css/style.css`:

```css
.modal-content {
    max-width: 600px;  /* Cambiar ancho */
}
```

## Troubleshooting

### Error: "Not authenticated"
- **Causa**: Sesión expirada o no iniciada
- **Solución**: Volver a desbloquear con la contraseña maestra

### Error: "Failed to connect to API"
- **Causa**: Backend no está ejecutándose
- **Solución**: Iniciar el servidor backend en puerto 8080

### Las contraseñas no se muestran
- **Causa**: No se ha activado "show_password"
- **Solución**: Hacer clic en el botón de ojo 👁️

### TOTP no se genera
- **Causa**: TOTP homemórfico sin clave privada
- **Solución**: Los TOTP homemórficos requieren la clave privada para generar códigos

### Sesión se cierra sola
- **Causa**: Timeout de sesión (1 hora por defecto)
- **Solución**: Aumentar SESSION_TIMEOUT en config.php

### Errores de permisos
```bash
# Dar permisos correctos
chmod 755 frontend/
chmod 644 frontend/*.php
chmod 644 frontend/assets/css/*.css
chmod 644 frontend/assets/js/*.js
```

## Desarrollo

### Agregar Nueva Funcionalidad

1. **Backend**: Agregar endpoint en la API
2. **API Wrapper**: Agregar método en `includes/api.php`
3. **Vista**: Agregar UI en `views/main.php` o `views/modals.php`
4. **JavaScript**: Agregar lógica en `assets/js/main.js`
5. **Estilos**: Agregar CSS en `assets/css/style.css`

### Depuración

```php
// Habilitar en config.php
error_reporting(E_ALL);
ini_set('display_errors', 1);

// Ver logs de PHP
tail -f /var/log/apache2/error.log

// Depurar requests AJAX
// Ver console del navegador (F12)
```

## Licencia

Este proyecto es parte del gestor de contraseñas con TOTP homemórfico.

## Contacto y Soporte

Para reportar bugs o solicitar features, crear un issue en el repositorio.

---

**Versión**: 1.0.0  
**Última actualización**: 2024
