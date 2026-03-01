<!-- New/Edit Account Modal -->
<div id="accountModal" class="modal">
    <div class="modal-content">
        <div class="modal-header">
            <h2 id="modalTitle">New Entry</h2>
            <button class="modal-close" onclick="closeModal('accountModal')">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <form method="POST" id="accountForm">
            <input type="hidden" name="account_id" id="account_id">
            
            <div class="form-group">
                <label for="email"><i class="fas fa-envelope"></i> Email</label>
                <input type="email" name="email" id="email" class="form-control">
            </div>

            <div class="form-group">
                <label for="user"><i class="fas fa-user"></i> Username *</label>
                <input type="text" name="user" id="user" class="form-control" required>
            </div>

            <div class="form-group">
                <label for="password"><i class="fas fa-lock"></i> Password *</label>
                <div class="input-group">
                    <input type="password" name="password" id="password" class="form-control" required>
                    <button type="button" class="btn btn-icon" onclick="togglePasswordInput('password')">
                        <i class="fas fa-eye"></i>
                    </button>
                    <button type="button" class="btn btn-icon" onclick="generateAndFillPassword()">
                        <i class="fas fa-random"></i>
                    </button>
                </div>
            </div>

            <div class="form-group">
                <label for="url"><i class="fas fa-link"></i> URL</label>
                <input type="url" name="url" id="url" class="form-control">
            </div>

            <div class="form-group">
                <label for="notes"><i class="fas fa-sticky-note"></i> Notes</label>
                <textarea name="notes" id="notes" class="form-control" rows="4"></textarea>
            </div>

            <div class="form-group">
                <label for="expire"><i class="fas fa-calendar"></i> Expiration Date</label>
                <input type="datetime-local" name="expire" id="expire" class="form-control">
            </div>

            <div class="modal-footer">
                <button type="button" class="btn btn-secondary" onclick="closeModal('accountModal')">
                    Cancel
                </button>
                <button type="submit" name="create_account" id="submitBtn" class="btn btn-primary">
                    <i class="fas fa-save"></i> Save
                </button>
            </div>
        </form>
    </div>
</div>

<!-- TOTP Modal -->
<div id="totpModal" class="modal">
    <div class="modal-content modal-totp">
        <div class="modal-header">
            <h2><i class="fas fa-clock"></i> TOTP Code</h2>
            <button class="modal-close" onclick="closeModal('totpModal')">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <div class="totp-display">
            <div class="totp-code-container">
                <div class="totp-code" id="totpCode">------</div>
                <div class="totp-timer">
                    <div class="timer-circle">
                        <svg class="timer-svg" viewBox="0 0 36 36">
                            <path class="timer-bg" d="M18 2.0845
                                a 15.9155 15.9155 0 0 1 0 31.831
                                a 15.9155 15.9155 0 0 1 0 -31.831"/>
                            <path id="timerPath" class="timer-progress" stroke-dasharray="0, 100" d="M18 2.0845
                                a 15.9155 15.9155 0 0 1 0 31.831
                                a 15.9155 15.9155 0 0 1 0 -31.831"/>
                        </svg>
                        <span class="timer-text" id="timerText">30</span>
                    </div>
                </div>
            </div>
            
            <button class="btn btn-primary btn-copy-totp" onclick="copyTOTP()">
                <i class="fas fa-copy"></i> Copy Code
            </button>
        </div>

        <div class="totp-info">
            <div class="detail-group">
                <label><i class="fas fa-key"></i> TOTP Seed</label>
                <div class="detail-value" id="totpSeed">
                    <span class="loading">Loading...</span>
                </div>
            </div>

            <div class="detail-group">
                <label><i class="fas fa-shield-alt"></i> Encryption Status</label>
                <div class="detail-value" id="totpStatus">
                    <span class="loading">Loading...</span>
                </div>
            </div>
        </div>

        <div class="modal-footer">
            <button type="button" class="btn btn-secondary" onclick="closeModal('totpModal')">
                Close
            </button>
            <button type="button" class="btn btn-primary" onclick="convertTOTPToHomomorphic()" id="btnConvertTOTP">
                <i class="fas fa-shield-alt"></i> Convert to Homomorphic
            </button>
        </div>
    </div>
</div>

<!-- Password Generator Modal -->
<div id="generatorModal" class="modal">
    <div class="modal-content">
        <div class="modal-header">
            <h2><i class="fas fa-random"></i> Password Generator</h2>
            <button class="modal-close" onclick="closeModal('generatorModal')">
                <i class="fas fa-times"></i>
            </button>
        </div>
        
        <div class="generator-content">
            <div class="generated-password-display">
                <input type="text" id="generatedPassword" class="form-control" readonly>
                <button class="btn btn-icon" onclick="copyGeneratedPassword()">
                    <i class="fas fa-copy"></i>
                </button>
            </div>

            <div class="form-group">
                <label for="passwordLength">
                    Password Length: <span id="lengthValue">16</span>
                </label>
                <input 
                    type="range" 
                    id="passwordLength" 
                    min="8" 
                    max="64" 
                    value="16" 
                    class="range-slider"
                    oninput="updateLengthDisplay(this.value); generatePassword()"
                >
            </div>

            <div class="form-group">
                <label class="checkbox-label">
                    <input type="checkbox" id="includeLowercase" checked onchange="generatePassword()">
                    <span>Lowercase (a-z)</span>
                </label>
            </div>

            <div class="form-group">
                <label class="checkbox-label">
                    <input type="checkbox" id="includeUppercase" checked onchange="generatePassword()">
                    <span>Uppercase (A-Z)</span>
                </label>
            </div>

            <div class="form-group">
                <label class="checkbox-label">
                    <input type="checkbox" id="includeNumbers" checked onchange="generatePassword()">
                    <span>Numbers (0-9)</span>
                </label>
            </div>

            <div class="form-group">
                <label class="checkbox-label">
                    <input type="checkbox" id="includeSpecial" checked onchange="generatePassword()">
                    <span>Special Characters (!@#$%^&*)</span>
                </label>
            </div>

            <button class="btn btn-secondary btn-block" onclick="generatePassword()">
                <i class="fas fa-sync"></i> Regenerate
            </button>
        </div>

        <div class="modal-footer">
            <button type="button" class="btn btn-secondary" onclick="closeModal('generatorModal')">
                Cancel
            </button>
            <button type="button" class="btn btn-primary" onclick="useGeneratedPassword()">
                <i class="fas fa-check"></i> Use Password
            </button>
        </div>
    </div>
</div>

<!-- Convert to Homomorphic Confirmation Modal -->
<div id="convertModal" class="modal">
    <div class="modal-content">
        <div class="modal-header">
            <h2><i class="fas fa-shield-alt"></i> Convert to Homomorphic TOTP</h2>
            <button class="modal-close" onclick="closeModal('convertModal')">
                <i class="fas fa-times"></i>
            </button>
        </div>
        
        <div class="modal-body">
            <div class="warning-box">
                <i class="fas fa-exclamation-triangle"></i>
                <p><strong>Important:</strong> After conversion, you will need to store the private key securely.</p>
                <p>The private key (N and Lambda) will be displayed only once. Make sure to save it in a secure location.</p>
            </div>

            <div class="form-group">
                <label for="keyBits">Key Size (bits)</label>
                <select id="keyBits" class="form-control">
                    <option value="2048" selected>2048 bits (Recommended)</option>
                    <option value="3072">3072 bits (More Secure)</option>
                    <option value="4096">4096 bits (Maximum Security)</option>
                </select>
                <small>Larger keys are more secure but slower to generate.</small>
            </div>
        </div>

        <div class="modal-footer">
            <button type="button" class="btn btn-secondary" onclick="closeModal('convertModal')">
                Cancel
            </button>
            <button type="button" class="btn btn-primary" onclick="performConversion()">
                <i class="fas fa-shield-alt"></i> Convert Now
            </button>
        </div>
    </div>
</div>

<!-- Private Key Display Modal -->
<div id="privateKeyModal" class="modal">
    <div class="modal-content">
        <div class="modal-header">
            <h2><i class="fas fa-key"></i> Private Key - Save Securely!</h2>
            <button class="modal-close" onclick="closeModal('privateKeyModal')">
                <i class="fas fa-times"></i>
            </button>
        </div>
        
        <div class="modal-body">
            <div class="alert alert-error">
                <i class="fas fa-exclamation-triangle"></i>
                <strong>Warning:</strong> This private key will only be shown once. 
                Copy and store it in a secure location (password manager, HSM, etc.)
            </div>

            <div class="form-group">
                <label>Paillier N (hex)</label>
                <textarea id="paillierN" class="form-control monospace" rows="4" readonly></textarea>
                <button class="btn btn-sm" onclick="copyToClipboard(document.getElementById('paillierN').value)">
                    <i class="fas fa-copy"></i> Copy
                </button>
            </div>

            <div class="form-group">
                <label>Paillier Lambda (hex)</label>
                <textarea id="paillierLambda" class="form-control monospace" rows="4" readonly></textarea>
                <button class="btn btn-sm" onclick="copyToClipboard(document.getElementById('paillierLambda').value)">
                    <i class="fas fa-copy"></i> Copy
                </button>
            </div>
        </div>

        <div class="modal-footer">
            <button type="button" class="btn btn-primary" onclick="closeModal('privateKeyModal')">
                <i class="fas fa-check"></i> I've Saved the Keys
            </button>
        </div>
    </div>
</div>
