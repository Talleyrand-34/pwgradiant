# Guía de Inicio Rápido - Frontend Web

## 🚀 Inicio en 5 Minutos

### 1. Requisitos Previos

```bash
# Verificar PHP instalado
php --version
# Debe ser PHP 7.4 o superior

# Verificar cURL habilitado
php -m | grep curl
# Debe mostrar "curl"
```

### 2. Iniciar el Backend (API)

```bash
# En una terminal, desde el directorio del backend
./pwdmgr server --port 8080 --key "tu-clave-maestra"

# Debería mostrar:
# Server starting on :8080
```

### 3. Iniciar el Frontend

**Opción A: Servidor Integrado de PHP (Más Fácil)**

```bash
# Desde el directorio frontend
cd frontend
php -S localhost:8000

# Abrir en navegador:
# http://localhost:8000
```

**Opción B: Apache/Nginx**

```bash
# Copiar archivos a directorio web
sudo cp -r frontend/ /var/www/html/password-manager/

# Configurar permisos
sudo chown -R www-data:www-data /var/www/html/password-manager
sudo chmod 755 /var/www/html/password-manager

# Abrir en navegador:
# http://localhost/password-manager/
```

### 4. Primer Uso

1. **Desbloquear**: Introduce tu contraseña maestra (la misma del backend)
2. **Crear Primera Cuenta**: 
   - Clic en el botón **+**
   - Usuario: `test@example.com`
   - Contraseña: `Test123!`
   - Guardar
3. **Probar Funciones**:
   - Copiar usuario/contraseña
   - Generar contraseña aleatoria
   - Buscar cuentas

## 📋 Checklist de Verificación

- [ ] Backend ejecutándose en puerto 8080
- [ ] Frontend accesible en navegador
- [ ] Puedes desbloquear con la contraseña maestra
- [ ] Puedes crear una cuenta de prueba
- [ ] Puedes ver y copiar la contraseña
- [ ] El generador de contraseñas funciona

## 🔧 Configuración Básica

### Cambiar Puerto del Frontend

```bash
# Servidor PHP en otro puerto
php -S localhost:3000
```

### Cambiar URL de la API

Editar `config.php`:

```php
define('API_BASE_URL', 'http://tu-servidor:8080/api/v1');
```

### Cambiar Timeout de Sesión

Editar `config.php`:

```php
// 30 minutos en lugar de 1 hora
define('SESSION_TIMEOUT', 1800);
```

## 🎯 Funciones Principales

### Barra de Herramientas

```
🔒 Bloquear     - Cierra sesión
➕ Nueva        - Crear cuenta
✏️ Editar       - Modificar cuenta seleccionada
🗑️ Eliminar     - Borrar cuenta seleccionada
👤 Copiar User  - Al portapapeles
📋 Copiar Pass  - Al portapapeles (requiere mostrar contraseña)
🔗 Copiar URL   - Al portapapeles
🎲 Generar Pass - Generador de contraseñas
⏱️ Ver TOTP     - Ver código 2FA
🛡️ Convertir   - A TOTP homemórfico
```

### Atajos Rápidos

- **Buscar**: Escribe en la barra de búsqueda
- **Seleccionar cuenta**: Clic en la lista
- **Mostrar contraseña**: Clic en el ojo 👁️
- **Copiar campo**: Clic en el botón de copia junto al campo

## 🐛 Solución de Problemas

### Error: "Not authenticated"
```bash
# Solución: Volver a desbloquear
# 1. Haz clic en 🔒 para bloquear
# 2. Introduce la contraseña maestra de nuevo
```

### Error: "Cannot connect to API"
```bash
# Verificar que el backend está ejecutándose
curl http://localhost:8080/health

# Debería responder: {"status":"ok"}
```

### La página está en blanco
```bash
# Verificar errores de PHP
tail -f /var/log/apache2/error.log
# o
tail -f /var/log/nginx/error.log

# O activar errores temporalmente en config.php
error_reporting(E_ALL);
ini_set('display_errors', 1);
```

### Las contraseñas no se copian
```bash
# El navegador debe soportar Clipboard API
# Probar en Chrome, Firefox o Edge moderno
# No funciona en HTTP, solo HTTPS o localhost
```

## 📚 Próximos Pasos

1. **Leer el README completo**: `frontend/README.md`
2. **Configurar HTTPS** (producción)
3. **Importar cuentas existentes**
4. **Configurar TOTP** en tus cuentas
5. **Hacer backup** de la base de datos regularmente

## 🔐 Seguridad

### ⚠️ IMPORTANTE

1. **Nunca** compartas tu contraseña maestra
2. **Usa HTTPS** en producción
3. **Haz backups** regularmente
4. **Guarda las claves TOTP homemórficas** de forma segura
5. **Cierra sesión** cuando no uses el gestor

### Configuración Segura

```php
// En config.php (producción)
define('ENABLE_HTTPS_ONLY', true);
define('SESSION_TIMEOUT', 900); // 15 minutos
error_reporting(0);
ini_set('display_errors', 0);
```

## 📞 Soporte

- **README completo**: `frontend/README.md`
- **Documentación API**: Ver archivos en `/outputs/`
- **Logs de error**: Revisar logs de PHP y servidor web

## ✅ Todo Listo!

Si llegaste hasta aquí, tu frontend debería estar funcionando.

**Disfruta de tu gestor de contraseñas seguro! 🎉**

---

**Versión**: 1.0.0  
**Actualizado**: 2024
