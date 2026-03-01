<?php
// config.php - Configuration file

// API Configuration
define('API_BASE_URL', 'http://www.pwg.t34.dev/api/v1');

// Database Configuration
define('DB_PATH', 'passwords.db');

// Session Configuration
define('SESSION_TIMEOUT', 3600); // 1 hour in seconds

// Security Configuration
define('ENABLE_HTTPS_ONLY', false); // Set to true in production
define('ENABLE_CSRF_PROTECTION', true);

// Application Settings
define('APP_NAME', 'Password Manager');
define('APP_VERSION', '1.0.0');

// Timezone
date_default_timezone_set('UTC');

// Error Reporting (disable in production)
error_reporting(E_ALL);
ini_set('display_errors', 1);

// Start session with secure settings
if (session_status() === PHP_SESSION_NONE) {
    session_start([
        'cookie_httponly' => true,
        'cookie_secure' => ENABLE_HTTPS_ONLY,
        'cookie_samesite' => 'Strict',
        'use_strict_mode' => true,
        'gc_maxlifetime' => SESSION_TIMEOUT
    ]);
}

// CSRF Token Generation
function generateCSRFToken() {
    if (!isset($_SESSION['csrf_token'])) {
        $_SESSION['csrf_token'] = bin2hex(random_bytes(32));
    }
    return $_SESSION['csrf_token'];
}

// CSRF Token Validation
function validateCSRFToken($token) {
    return isset($_SESSION['csrf_token']) && hash_equals($_SESSION['csrf_token'], $token);
}

// Session Timeout Check
function checkSessionTimeout() {
    if (isset($_SESSION['last_activity']) && 
        (time() - $_SESSION['last_activity'] > SESSION_TIMEOUT)) {
        session_unset();
        session_destroy();
        return false;
    }
    $_SESSION['last_activity'] = time();
    return true;
}
?>
