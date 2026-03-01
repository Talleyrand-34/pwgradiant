// assets/js/main.js

// Global variables
let totpInterval = null;
let currentPassword = '';

// Filter accounts in the list
function filterAccounts() {
    const searchTerm = document.getElementById('search-input').value.toLowerCase();
    const accountItems = document.querySelectorAll('.account-item');
    
    accountItems.forEach(item => {
        const searchData = item.getAttribute('data-search');
        if (searchData.includes(searchTerm)) {
            item.style.display = 'flex';
        } else {
            item.style.display = 'none';
        }
    });
}

// Select an account
function selectAccount(accountId) {
    window.location.href = `?account_id=${accountId}`;
}

// Toggle password visibility
function togglePassword(accountId) {
    const currentUrl = new URL(window.location.href);
    const showPassword = currentUrl.searchParams.get('show_password') !== 'true';
    currentUrl.searchParams.set('account_id', accountId);
    currentUrl.searchParams.set('show_password', showPassword);
    window.location.href = currentUrl.toString();
}

// Copy to clipboard
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        showNotification('Copied to clipboard!', 'success');
    }).catch(err => {
        showNotification('Failed to copy', 'error');
    });
}

// Show notification
function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `alert alert-${type}`;
    notification.style.position = 'fixed';
    notification.style.top = '20px';
    notification.style.right = '20px';
    notification.style.zIndex = '9999';
    notification.style.minWidth = '250px';
    notification.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check-circle' : 'exclamation-circle'}"></i>
        ${message}
    `;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.transition = 'opacity 0.3s';
        notification.style.opacity = '0';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// Modal functions
function showModal(modalId) {
    document.getElementById(modalId).classList.add('active');
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
    
    // Stop TOTP interval if closing TOTP modal
    if (modalId === 'totpModal' && totpInterval) {
        clearInterval(totpInterval);
        totpInterval = null;
    }
}

// Close modal when clicking outside
window.addEventListener('click', (e) => {
    if (e.target.classList.contains('modal')) {
        e.target.classList.remove('active');
    }
});

// Account Management
function showNewEntryModal() {
    document.getElementById('modalTitle').textContent = 'New Entry';
    document.getElementById('accountForm').reset();
    document.getElementById('account_id').value = '';
    document.getElementById('submitBtn').name = 'create_account';
    document.getElementById('submitBtn').innerHTML = '<i class="fas fa-save"></i> Save';
    showModal('accountModal');
}

function editSelectedEntry() {
    if (!selectedAccountData) {
        showNotification('Please select an account first', 'error');
        return;
    }
    
    document.getElementById('modalTitle').textContent = 'Edit Entry';
    document.getElementById('account_id').value = selectedAccountData.id;
    document.getElementById('email').value = selectedAccountData.email || '';
    document.getElementById('user').value = selectedAccountData.user;
    document.getElementById('password').value = selectedAccountData.password || '';
    document.getElementById('url').value = selectedAccountData.url || '';
    document.getElementById('notes').value = selectedAccountData.notes || '';
    
    if (selectedAccountData.expire) {
        const date = new Date(selectedAccountData.expire);
        document.getElementById('expire').value = date.toISOString().slice(0, 16);
    }
    
    document.getElementById('submitBtn').name = 'update_account';
    document.getElementById('submitBtn').innerHTML = '<i class="fas fa-save"></i> Update';
    showModal('accountModal');
}

function deleteSelectedEntry() {
    if (!selectedAccountData) {
        showNotification('Please select an account first', 'error');
        return;
    }
    
    if (confirm(`Are you sure you want to delete "${selectedAccountData.user}"?`)) {
        const form = document.createElement('form');
        form.method = 'POST';
        form.innerHTML = `
            <input type="hidden" name="delete_account" value="1">
            <input type="hidden" name="account_id" value="${selectedAccountData.id}">
        `;
        document.body.appendChild(form);
        form.submit();
    }
}

// Copy functions
function copyUsername() {
    if (!selectedAccountData) {
        showNotification('Please select an account first', 'error');
        return;
    }
    copyToClipboard(selectedAccountData.user);
}

function copyPassword() {
    if (!selectedAccountData || !selectedAccountData.password) {
        showNotification('Password not available. Enable "Show Password" first.', 'error');
        return;
    }
    copyToClipboard(selectedAccountData.password);
}

function copyURL() {
    if (!selectedAccountData || !selectedAccountData.url) {
        showNotification('No URL available', 'error');
        return;
    }
    copyToClipboard(selectedAccountData.url);
}

// Password Generator
function showPasswordGenerator() {
    generatePassword();
    showModal('generatorModal');
}

function generatePassword() {
    const length = parseInt(document.getElementById('passwordLength').value);
    const includeLowercase = document.getElementById('includeLowercase').checked;
    const includeUppercase = document.getElementById('includeUppercase').checked;
    const includeNumbers = document.getElementById('includeNumbers').checked;
    const includeSpecial = document.getElementById('includeSpecial').checked;
    
    let chars = '';
    if (includeLowercase) chars += 'abcdefghijklmnopqrstuvwxyz';
    if (includeUppercase) chars += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    if (includeNumbers) chars += '0123456789';
    if (includeSpecial) chars += '!@#$%^&*()_+-=[]{}|;:,.<>?';
    
    if (chars.length === 0) {
        showNotification('Please select at least one character type', 'error');
        return;
    }
    
    let password = '';
    for (let i = 0; i < length; i++) {
        password += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    
    currentPassword = password;
    document.getElementById('generatedPassword').value = password;
}

function updateLengthDisplay(value) {
    document.getElementById('lengthValue').textContent = value;
}

function copyGeneratedPassword() {
    const password = document.getElementById('generatedPassword').value;
    copyToClipboard(password);
}

function useGeneratedPassword() {
    const password = document.getElementById('generatedPassword').value;
    if (document.getElementById('password')) {
        document.getElementById('password').value = password;
    }
    closeModal('generatorModal');
}

function generateAndFillPassword() {
    const length = 16;
    const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*';
    let password = '';
    for (let i = 0; i < length; i++) {
        password += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    document.getElementById('password').value = password;
    showNotification('Password generated!', 'success');
}

function togglePasswordInput(inputId) {
    const input = document.getElementById(inputId);
    const icon = event.target.closest('button').querySelector('i');
    
    if (input.type === 'password') {
        input.type = 'text';
        icon.classList.remove('fa-eye');
        icon.classList.add('fa-eye-slash');
    } else {
        input.type = 'password';
        icon.classList.remove('fa-eye-slash');
        icon.classList.add('fa-eye');
    }
}

// TOTP Functions
function showTOTP() {
    if (!selectedAccountData || !selectedAccountData.totp) {
        showNotification('This account does not have TOTP configured', 'error');
        return;
    }
    
    const totp = selectedAccountData.totp;
    
    // Display TOTP seed
    if (totp.use_homomorphic) {
        document.getElementById('totpSeed').innerHTML = '<span class="tag"><i class="fas fa-lock"></i> Encrypted</span>';
        document.getElementById('totpStatus').innerHTML = '<span class="tag"><i class="fas fa-shield-alt"></i> Homomorphic Encryption</span>';
        document.getElementById('btnConvertTOTP').style.display = 'none';
    } else {
        document.getElementById('totpSeed').textContent = totp.totp_seed;
        document.getElementById('totpStatus').innerHTML = '<span class="tag"><i class="fas fa-key"></i> Standard TOTP</span>';
        document.getElementById('btnConvertTOTP').style.display = 'inline-flex';
    }
    
    showModal('totpModal');
    startTOTPGeneration();
}

function startTOTPGeneration() {
    // Clear any existing interval
    if (totpInterval) {
        clearInterval(totpInterval);
    }
    
    // Generate TOTP code immediately
    generateTOTPCode();
    
    // Update every second
    totpInterval = setInterval(() => {
        const seconds = Math.floor(Date.now() / 1000) % 30;
        const remaining = 30 - seconds;
        
        document.getElementById('timerText').textContent = remaining;
        
        // Update progress circle
        const percentage = (remaining / 30) * 100;
        document.getElementById('timerPath').setAttribute('stroke-dasharray', `${percentage}, 100`);
        
        // Generate new code when timer resets
        if (remaining === 30) {
            generateTOTPCode();
        }
    }, 1000);
}

async function generateTOTPCode() {
    if (!selectedAccountData || !selectedAccountData.totp) {
        return;
    }
    
    try {
        const response = await fetch(`api.php?action=generate_totp&id=${selectedAccountData.totp.id}`);
        const data = await response.json();
        
        if (data.success) {
            document.getElementById('totpCode').textContent = data.code;
        } else {
            document.getElementById('totpCode').textContent = '------';
            if (!selectedAccountData.totp.use_homomorphic) {
                showNotification('Failed to generate TOTP code', 'error');
            }
        }
    } catch (error) {
        console.error('Error generating TOTP:', error);
    }
}

function copyTOTP() {
    const code = document.getElementById('totpCode').textContent;
    if (code && code !== '------') {
        copyToClipboard(code);
    }
}

function convertTOTPToHomomorphic() {
    if (!selectedAccountData || !selectedAccountData.totp) {
        showNotification('No TOTP to convert', 'error');
        return;
    }
    
    if (selectedAccountData.totp.use_homomorphic) {
        showNotification('TOTP is already using homomorphic encryption', 'error');
        return;
    }
    
    closeModal('totpModal');
    showModal('convertModal');
}

function convertToHomomorphic() {
    if (!selectedAccountData || !selectedAccountData.totp) {
        return;
    }
    
    convertTOTPToHomomorphic();
}

async function performConversion() {
    const keyBits = parseInt(document.getElementById('keyBits').value);
    const totpId = selectedAccountData.totp.id;
    
    try {
        const response = await fetch(`api.php?action=convert_totp&id=${totpId}&key_bits=${keyBits}`, {
            method: 'POST'
        });
        const data = await response.json();
        
        if (data.success) {
            // Show private key
            document.getElementById('paillierN').value = data.paillier_n_hex;
            document.getElementById('paillierLambda').value = data.paillier_lambda_hex;
            
            closeModal('convertModal');
            showModal('privateKeyModal');
            
            showNotification('TOTP converted to homomorphic encryption!', 'success');
        } else {
            showNotification(data.error || 'Conversion failed', 'error');
        }
    } catch (error) {
        console.error('Error converting TOTP:', error);
        showNotification('Conversion failed', 'error');
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    // Enable/disable toolbar buttons based on selection
    const hasSelection = selectedAccountId !== null;
    const buttons = ['btn-edit', 'btn-delete', 'btn-copy-user', 'btn-copy-pass', 'btn-copy-url', 'btn-totp', 'btn-convert'];
    
    buttons.forEach(btnId => {
        const btn = document.getElementById(btnId);
        if (btn) {
            btn.disabled = !hasSelection;
            btn.style.opacity = hasSelection ? '1' : '0.5';
        }
    });
    
    // Disable TOTP buttons if no TOTP
    if (selectedAccountData && !selectedAccountData.totp) {
        ['btn-totp', 'btn-convert'].forEach(btnId => {
            const btn = document.getElementById(btnId);
            if (btn) {
                btn.disabled = true;
                btn.style.opacity = '0.5';
            }
        });
    }
});
