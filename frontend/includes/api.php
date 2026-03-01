<?php
// includes/api.php

class PasswordManagerAPI {
    private $baseUrl;
    private $masterPassword;
    
    public function __construct($masterPassword, $baseUrl = 'http://localhost:8080/api/v1') {
        $this->masterPassword = $masterPassword;
        $this->baseUrl = $baseUrl;
    }
    
    private function makeRequest($endpoint, $method = 'GET', $data = null, $queryParams = []) {
        $url = $this->baseUrl . $endpoint;
        
        if (!empty($queryParams)) {
            $url .= '?' . http_build_query($queryParams);
        }
        
        $headers = [
            'Content-Type: application/json',
            'X-Encryption-Key: ' . $this->masterPassword
        ];

        // Use cURL if available, otherwise fallback to file_get_contents
        if (function_exists('curl_init')) {
            $ch = curl_init($url);
            curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
            curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
            
            if ($method !== 'GET') {
                curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
            }
            
            if ($data !== null) {
                curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
            }
            
            $response = curl_exec($ch);
            $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
            curl_close($ch);
        } else {
            // Fallback using stream context (file_get_contents)
            $options = [
                'http' => [
                    'method'  => $method,
                    'header'  => implode("\r\n", $headers),
                    'content' => $data !== null ? json_encode($data) : null,
                    'ignore_errors' => true // To get the response body even for 4xx/5xx errors
                ]
            ];
            
            $context = stream_context_create($options);
            $response = @file_get_contents($url, false, $context);
            
            // Extract HTTP status code from $http_response_header
            $httpCode = 0;
            if (isset($http_response_header) && !empty($http_response_header)) {
                if (preg_match('{HTTP\/\S*\s(\d{3})}', $http_response_header[0], $matches)) {
                    $httpCode = (int)$matches[1];
                }
            }
        }
        
        $result = json_decode($response, true);
        
        return [
            'success' => $httpCode >= 200 && $httpCode < 300,
            'code' => $httpCode,
            'data' => $result,
            'error' => $httpCode >= 400 ? ($result['error'] ?? 'Unknown error') : null
        ];
    }
    
    public function getAccounts($showPassword = false) {
        $params = $showPassword ? ['show_password' => 'true'] : [];
        return $this->makeRequest('/accounts', 'GET', null, $params);
    }
    
    public function getAccount($id, $showPassword = false) {
        $params = $showPassword ? ['show_password' => 'true'] : [];
        return $this->makeRequest('/accounts/' . $id, 'GET', null, $params);
    }
    
    public function createAccount($data) {
        $payload = [
            'email' => $data['email'] ?? null,
            'user' => $data['user'],
            'password' => $data['password'],
            'url' => $data['url'] ?? null,
            'notes' => $data['notes'] ?? null,
            'expire' => $data['expire'] ?? null
        ];
        
        return $this->makeRequest('/accounts', 'POST', $payload);
    }
    
    public function updateAccount($id, $data) {
        $payload = [
            'email' => $data['email'] ?? null,
            'user' => $data['user'],
            'password' => $data['password'],
            'url' => $data['url'] ?? null,
            'notes' => $data['notes'] ?? null,
            'expire' => $data['expire'] ?? null,
        ];

        $queryParams = [];
        if (!empty($data['change_reason'])) {
            $queryParams['change_reason'] = $data['change_reason'];
        }

        return $this->makeRequest('/accounts/' . $id, 'PUT', $payload, $queryParams);
    }
    
    public function deleteAccount($id) {
        return $this->makeRequest('/accounts/' . $id, 'DELETE');
    }
    
    public function searchAccounts($query, $showPassword = false) {
        $params = ['q' => $query];
        if ($showPassword) {
            $params['show_password'] = 'true';
        }
        return $this->makeRequest('/accounts/search', 'GET', null, $params);
    }
    
    // TOTP Methods
    public function getTOTPByAccount($accountId) {
        return $this->makeRequest('/totp/account/' . $accountId);
    }
    
    public function generateTOTPCode($totpId, $privateKey = null) {
        $data = $privateKey ?? [];
        return $this->makeRequest('/totp/' . $totpId . '/generate', 'POST', $data);
    }
    
    public function verifyTOTPCode($totpId, $code) {
        return $this->makeRequest('/totp/' . $totpId . '/verify', 'POST', ['code' => $code]);
    }
    
    public function convertTOTPToHomomorphic($totpId, $keyBits = 2048) {
        return $this->makeRequest('/totp/' . $totpId . '/convert-to-homomorphic', 'POST', ['key_bits' => $keyBits]);
    }
    
    public function createTOTP($accountId, $seed) {
        return $this->makeRequest('/totp', 'POST', [
            'account_id' => $accountId,
            'totp_seed' => $seed
        ]);
    }
    
    public function deleteTOTP($totpId) {
        return $this->makeRequest('/totp/' . $totpId, 'DELETE');
    }
    
    // Tag Methods
    public function getTags() {
        return $this->makeRequest('/tags');
    }
    
    public function createTag($name) {
        return $this->makeRequest('/tags', 'POST', ['name' => $name]);
    }
    
    public function addTagToAccount($accountId, $tagName) {
        return $this->makeRequest('/accounts/' . $accountId . '/tag/' . urlencode($tagName), 'POST');
    }
    
    public function removeTagFromAccount($accountId, $tagName) {
        return $this->makeRequest('/accounts/' . $accountId . '/tag/' . urlencode($tagName), 'DELETE');
    }
    
    public function getAccountsByTag($tagName) {
        return $this->makeRequest('/accounts/by-tag/' . urlencode($tagName));
    }
}
?>
