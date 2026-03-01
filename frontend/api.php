<?php
// api.php - AJAX proxy handler for JavaScript requests
require_once 'config.php';

header('Content-Type: application/json');

if (!isset($_SESSION['master_password']) || empty($_SESSION['master_password'])) {
    echo json_encode(['success' => false, 'error' => 'Not authenticated']);
    exit;
}

require_once 'includes/api.php';
$api = new PasswordManagerAPI($_SESSION['master_password']);

$action = $_GET['action'] ?? '';
$id = isset($_GET['id']) ? (int)$_GET['id'] : 0;

switch ($action) {
    case 'generate_totp':
        if (!$id) {
            echo json_encode(['success' => false, 'error' => 'Missing TOTP id']);
            exit;
        }
        $result = $api->generateTOTPCode($id);
        if ($result['success']) {
            echo json_encode(array_merge(['success' => true], $result['data'] ?? []));
        } else {
            echo json_encode(['success' => false, 'error' => $result['error'] ?? 'Failed to generate TOTP code']);
        }
        break;

    case 'convert_totp':
        if (!$id) {
            echo json_encode(['success' => false, 'error' => 'Missing TOTP id']);
            exit;
        }
        $keyBits = isset($_GET['key_bits']) ? (int)$_GET['key_bits'] : 2048;
        $result = $api->convertTOTPToHomomorphic($id, $keyBits);
        if ($result['success']) {
            echo json_encode(array_merge(['success' => true], $result['data'] ?? []));
        } else {
            echo json_encode(['success' => false, 'error' => $result['error'] ?? 'Conversion failed']);
        }
        break;

    default:
        echo json_encode(['success' => false, 'error' => 'Unknown action']);
        break;
}
