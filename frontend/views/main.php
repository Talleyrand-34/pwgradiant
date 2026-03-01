<?php
// views/main.php
require_once 'includes/api.php';

$api = new PasswordManagerAPI($_SESSION['master_password']);

// Handle actions
$message = '';
$messageType = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (isset($_POST['create_account'])) {
        $result = $api->createAccount($_POST);
        $message = $result['success'] ? 'Account created successfully' : $result['error'];
        $messageType = $result['success'] ? 'success' : 'error';
    } elseif (isset($_POST['update_account'])) {
        $result = $api->updateAccount($_POST['account_id'], $_POST);
        $message = $result['success'] ? 'Account updated successfully' : $result['error'];
        $messageType = $result['success'] ? 'success' : 'error';
    } elseif (isset($_POST['delete_account'])) {
        $result = $api->deleteAccount($_POST['account_id']);
        $message = $result['success'] ? 'Account deleted successfully' : $result['error'];
        $messageType = $result['success'] ? 'success' : 'error';
    }
}

// Get accounts list
$showPassword = isset($_GET['show_password']) && $_GET['show_password'] === 'true';
$accounts = $api->getAccounts($showPassword);
$selectedAccountId = $_GET['account_id'] ?? null;
$selectedAccount = null;

if ($selectedAccountId) {
    $selectedAccount = $api->getAccount($selectedAccountId, $showPassword);
}
?>

<div class="app-container">
    <!-- New Database Header -->
    <header class="db-header">
        <div class="logo-container">
            <img src="../images/logo-only-icon-white.svg" alt="Gradiant Logo">
        </div>
        
        <div class="config-icon" title="Configuración">
            <i class="fas fa-cog"></i>
        </div>
        
        <div class="vertical-separator"></div>
        
        <!-- Database Actions -->
        <div class="icon-group">
            <button class="header-btn" title="Abrir base de datos">
                <i class="fas fa-folder-open"></i>
            </button>
            <button class="header-btn" onclick="location.href='?action=lock'" title="Bloquear base de datos">
                <i class="fas fa-lock"></i>
            </button>
            <button class="header-btn" title="Añadir base de datos">
                <i class="fas fa-folder-plus"></i>
            </button>
        </div>
        
        <!-- Entry Actions -->
        <div class="icon-group">
            <button class="header-btn" onclick="showNewEntryModal()" title="Nueva entrada">
                <i class="fas fa-plus"></i>
            </button>
            <button class="header-btn" onclick="editSelectedEntry()" id="btn-edit" title="Ver/editar entrada">
                <i class="fas fa-edit"></i>
            </button>
            <button class="header-btn" onclick="deleteSelectedEntry()" id="btn-delete" title="Borrar entrada">
                <i class="fas fa-trash"></i>
            </button>
        </div>
        
        <!-- Copy Actions -->
        <div class="icon-group">
            <button class="header-btn" onclick="copyUsername()" id="btn-copy-user" title="Copiar usuario">
                <i class="fas fa-user"></i>
            </button>
            <button class="header-btn" onclick="copyPassword()" id="btn-copy-pass" title="Copiar contraseña">
                <i class="fas fa-copy"></i>
            </button>
            <button class="header-btn" onclick="copyURL()" id="btn-copy-url" title="Copiar URL">
                <i class="fas fa-link"></i>
            </button>
        </div>
        
        <!-- Generator -->
        <div class="icon-group" style="margin-left: 10px;">
            <button class="header-btn" onclick="showPasswordGenerator()" title="Generar contraseña aleatoria">
                <i class="fas fa-random"></i>
            </button>
        </div>
        
        <!-- Search Bar -->
        <div class="search-container">
            <form onsubmit="return false;" style="padding: 0; display: flex; align-items: center;">
                <input 
                    type="text" 
                    id="search-input" 
                    class="search-input-header" 
                    placeholder="Buscar..." 
                    onkeyup="filterAccounts()"
                >
                <img src="../images/busqueda-icon-01.svg" class="search-icon-header" alt="Buscar">
            </form>
        </div>
    </header>

    <?php if ($message): ?>
...
        <div class="alert alert-<?php echo $messageType; ?>">
            <i class="fas fa-<?php echo $messageType === 'success' ? 'check-circle' : 'exclamation-circle'; ?>"></i>
            <?php echo htmlspecialchars($message); ?>
        </div>
    <?php endif; ?>

    <!-- Main Content -->
    <div class="main-content">
        <!-- Left Panel: Accounts List -->
        <div class="accounts-panel">
            <div class="accounts-list" id="accounts-list">
                <?php if ($accounts['success']): ?>
                    <?php foreach ($accounts['data'] as $account): ?>
                        <div class="account-item <?php echo $selectedAccountId == $account['id'] ? 'selected' : ''; ?>" 
                             onclick="selectAccount(<?php echo $account['id']; ?>)"
                             data-search="<?php echo strtolower($account['user'] . ' ' . ($account['url'] ?? '') . ' ' . ($account['email'] ?? '')); ?>">
                            <div class="account-icon">
                                <i class="fas fa-user-circle"></i>
                            </div>
                            <div class="account-info">
                                <div class="account-user"><?php echo htmlspecialchars($account['user']); ?></div>
                                <div class="account-url"><?php echo htmlspecialchars($account['url'] ?? 'No URL'); ?></div>
                            </div>
                        </div>
                    <?php endforeach; ?>
                <?php else: ?>
                    <div class="empty-state">
                        <i class="fas fa-folder-open"></i>
                        <p>No accounts found</p>
                    </div>
                <?php endif; ?>
            </div>
        </div>

        <!-- Right Panel: Account Details -->
        <div class="details-panel">
            <?php if ($selectedAccount && $selectedAccount['success']): ?>
                <?php $account = $selectedAccount['data']; ?>
                <div class="details-header">
                    <h2><?php echo htmlspecialchars($account['user']); ?></h2>
                </div>
                
                <div class="details-content">
                    <div class="detail-group">
                        <label><i class="fas fa-hashtag"></i> ID</label>
                        <div class="detail-value"><?php echo htmlspecialchars($account['id']); ?></div>
                    </div>

                    <?php if (!empty($account['email'])): ?>
                    <div class="detail-group">
                        <label><i class="fas fa-envelope"></i> Email</label>
                        <div class="detail-value">
                            <?php echo htmlspecialchars($account['email']); ?>
                            <button class="btn-copy-inline" onclick="copyToClipboard('<?php echo htmlspecialchars($account['email']); ?>')">
                                <i class="fas fa-copy"></i>
                            </button>
                        </div>
                    </div>
                    <?php endif; ?>

                    <div class="detail-group">
                        <label><i class="fas fa-user"></i> Username</label>
                        <div class="detail-value">
                            <?php echo htmlspecialchars($account['user']); ?>
                            <button class="btn-copy-inline" onclick="copyToClipboard('<?php echo htmlspecialchars($account['user']); ?>')">
                                <i class="fas fa-copy"></i>
                            </button>
                        </div>
                    </div>

                    <div class="detail-group">
                        <label><i class="fas fa-lock"></i> Password</label>
                        <div class="detail-value password-field">
                            <span id="password-display">
                                <?php echo $showPassword ? htmlspecialchars($account['password']) : '••••••••'; ?>
                            </span>
                            <button class="btn-toggle-password" onclick="togglePassword(<?php echo $account['id']; ?>)">
                                <i class="fas fa-eye<?php echo $showPassword ? '-slash' : ''; ?>"></i>
                            </button>
                            <?php if ($showPassword): ?>
                            <button class="btn-copy-inline" onclick="copyToClipboard('<?php echo htmlspecialchars($account['password']); ?>')">
                                <i class="fas fa-copy"></i>
                            </button>
                            <?php endif; ?>
                        </div>
                    </div>

                    <?php if (!empty($account['url'])): ?>
                    <div class="detail-group">
                        <label><i class="fas fa-link"></i> URL</label>
                        <div class="detail-value">
                            <a href="<?php echo htmlspecialchars($account['url']); ?>" target="_blank">
                                <?php echo htmlspecialchars($account['url']); ?>
                            </a>
                            <button class="btn-copy-inline" onclick="copyToClipboard('<?php echo htmlspecialchars($account['url']); ?>')">
                                <i class="fas fa-copy"></i>
                            </button>
                        </div>
                    </div>
                    <?php endif; ?>

                    <?php if (!empty($account['notes'])): ?>
                    <div class="detail-group">
                        <label><i class="fas fa-sticky-note"></i> Notes</label>
                        <div class="detail-value notes">
                            <?php echo nl2br(htmlspecialchars($account['notes'])); ?>
                        </div>
                    </div>
                    <?php endif; ?>

                    <?php if (!empty($account['expire'])): ?>
                    <div class="detail-group">
                        <label><i class="fas fa-calendar"></i> Expires</label>
                        <div class="detail-value">
                            <?php echo htmlspecialchars(date('Y-m-d H:i', strtotime($account['expire']))); ?>
                        </div>
                    </div>
                    <?php endif; ?>

                    <?php if (!empty($account['tags'])): ?>
                    <div class="detail-group">
                        <label><i class="fas fa-tags"></i> Tags</label>
                        <div class="detail-value">
                            <?php foreach ($account['tags'] as $tag): ?>
                                <span class="tag"><?php echo htmlspecialchars($tag['name']); ?></span>
                            <?php endforeach; ?>
                        </div>
                    </div>
                    <?php endif; ?>

                    <?php if (!empty($account['totp'])): ?>
                    <div class="detail-group totp-group">
                        <label><i class="fas fa-clock"></i> TOTP</label>
                        <div class="detail-value">
                            <span class="totp-status">
                                <?php if ($account['totp']['use_homomorphic']): ?>
                                    <i class="fas fa-shield-alt"></i> Homomorphic Encryption Enabled
                                <?php else: ?>
                                    <i class="fas fa-key"></i> Standard TOTP
                                <?php endif; ?>
                            </span>
                            <button class="btn btn-sm" onclick="showTOTP()">
                                <i class="fas fa-clock"></i> View Code
                            </button>
                        </div>
                    </div>
                    <?php endif; ?>
                </div>
            <?php else: ?>
                <div class="empty-state-large">
                    <i class="fas fa-mouse-pointer"></i>
                    <h3>Select an account</h3>
                    <p>Choose an account from the list to view its details</p>
                </div>
            <?php endif; ?>
        </div>
    </div>
</div>

<!-- Modals -->
<?php include 'views/modals.php'; ?>

<script>
const selectedAccountId = <?php echo $selectedAccountId ?? 'null'; ?>;
const selectedAccountData = <?php echo $selectedAccount && $selectedAccount['success'] ? json_encode($selectedAccount['data']) : 'null'; ?>;
</script>
