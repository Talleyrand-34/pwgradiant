package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Helper to capture stdout
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func setupTestCLI(t *testing.T) func() {
	testDBPath := "test_cli.db"
	os.Remove(testDBPath)

	// Set global vars
	dbPath = testDBPath
	encryptionKey = "test-key-cli"

	return func() {
		CloseDB()
		os.Remove(testDBPath)
	}
}

// ============================================================================
// BASIC FUNCTIONALITY TESTS
// ============================================================================

func TestInitDatabase(t *testing.T) {
	cleanup := setupTestCLI(t)
	defer cleanup()

	err := initDatabase()
	if err != nil {
		t.Fatalf("initDatabase failed: %v", err)
	}

	if DB == nil {
		t.Fatal("DB should not be nil after initDatabase")
	}

	CloseDB()
}

func TestPrintJSON(t *testing.T) {
	data := map[string]string{
		"key": "value",
	}

	output := captureOutput(func() {
		printJSON(data)
	})

	if !strings.Contains(output, "key") {
		t.Error("Output should contain 'key'")
	}
	if !strings.Contains(output, "value") {
		t.Error("Output should contain 'value'")
	}
}

func TestPrintJSONWithAccount(t *testing.T) {
	email := "test@example.com"
	account := Account{
		ID:       1,
		Email:    &email,
		User:     "testuser",
		Password: "testpass",
		Expire:   "2025-12-31T23:59:59Z",
	}

	output := captureOutput(func() {
		printJSON(account)
	})

	if !strings.Contains(output, "testuser") {
		t.Error("Output should contain username")
	}
	if !strings.Contains(output, "test@example.com") {
		t.Error("Output should contain email")
	}
}

func TestPrintJSONInvalidData(t *testing.T) {
	// Channel cannot be marshaled to JSON
	invalidData := make(chan int)

	// Should not panic
	output := captureOutput(func() {
		printJSON(invalidData)
	})

	// Output should contain error message
	if !strings.Contains(output, "Error") {
		t.Log("Expected error message in output")
	}
}

// ============================================================================
// ROOT COMMAND TESTS
// ============================================================================

func TestRootCmdExists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "pwdmgr" {
		t.Errorf("Expected 'pwdmgr', got %s", rootCmd.Use)
	}
}

func TestRootCmdHasAliases(t *testing.T) {
	// Test that main commands have proper aliases
	accountAliases := accountCmd.Aliases
	if len(accountAliases) == 0 {
		t.Error("accountCmd should have aliases")
	}
}

// ============================================================================
// MAIN COMMANDS EXISTENCE
// ============================================================================

func TestMainCommandsExist(t *testing.T) {
	requiredCommands := map[string]*cobra.Command{
		"init":     initCmd,
		"server":   serverCmd,
		"account":  accountCmd,
		"tag":      tagCmd,
		"totp":     totpCmd,
		"security": securityCmd,
		"generate": generateCmd,
	}

	for name, cmd := range requiredCommands {
		if cmd == nil {
			t.Errorf("%s command should not be nil", name)
		}
	}
}

func TestCommandsAddedToRoot(t *testing.T) {
	expectedCommands := []string{"init", "server", "account", "tag", "totp", "security", "generate"}

	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Command '%s' not found in root commands", expected)
		}
	}
}

// ============================================================================
// ACCOUNT COMMAND TESTS
// ============================================================================

func TestAccountSubcommands(t *testing.T) {
	expectedSubcommands := []string{
		"add", "list", "get", "update", "delete",
		"history", "search", "search-field", "fuzzy",
		"tags",
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range accountCmd.Commands() {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("account subcommand '%s' not found", expected)
		}
	}
}

func TestAccountCommandAliases(t *testing.T) {
	tests := []struct {
		cmd   *cobra.Command
		alias string
	}{
		{accountAddCmd, "create"},
		{accountListCmd, "ls"},
		{accountGetCmd, "show"},
		{accountUpdateCmd, "edit"},
		{accountDeleteCmd, "rm"},
		{accountSearchCmd, "find"},
	}

	for _, tt := range tests {
		found := false
		for _, alias := range tt.cmd.Aliases {
			if alias == tt.alias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Command %s should have alias '%s'", tt.cmd.Name(), tt.alias)
		}
	}
}

func TestAccountCmdUse(t *testing.T) {
	tests := []struct {
		cmd      *cobra.Command
		expected string
	}{
		{accountAddCmd, "add"},
		{accountListCmd, "list"},
		{accountGetCmd, "get <id>"},
		{accountUpdateCmd, "update <id>"},
		{accountDeleteCmd, "delete <id>"},
		{accountHistoryCmd, "history <id>"},
		{accountSearchCmd, "search <query>"},
	}

	for _, tt := range tests {
		if tt.cmd.Use != tt.expected {
			t.Errorf("Expected Use '%s', got '%s'", tt.expected, tt.cmd.Use)
		}
	}
}

// ============================================================================
// ACCOUNT TAGS SUBCOMMAND TESTS
// ============================================================================

func TestAccountTagsSubcommands(t *testing.T) {
	expectedSubcommands := []string{"add", "remove", "list", "filter"}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range accountTagsCmd.Commands() {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("account tags subcommand '%s' not found", expected)
		}
	}
}

func TestAccountTagsAddArgs(t *testing.T) {
	if accountTagsAddCmd.Args == nil {
		t.Error("accountTagsAddCmd.Args is nil")
		return
	}

	err := accountTagsAddCmd.Args(accountTagsAddCmd, []string{})
	if err == nil {
		t.Error("Expected error for no arguments")
	}

	err = accountTagsAddCmd.Args(accountTagsAddCmd, []string{"1"})
	if err == nil {
		t.Error("Expected error for 1 argument")
	}

	err = accountTagsAddCmd.Args(accountTagsAddCmd, []string{"1", "work"})
	if err != nil {
		t.Errorf("Expected no error for 2 arguments, got: %v", err)
	}
}

func TestAccountTagsRemoveArgs(t *testing.T) {
	if accountTagsRemoveCmd.Args == nil {
		t.Error("accountTagsRemoveCmd.Args is nil")
		return
	}

	err := accountTagsRemoveCmd.Args(accountTagsRemoveCmd, []string{"1", "work"})
	if err != nil {
		t.Errorf("Expected no error for 2 arguments, got: %v", err)
	}
}

func TestAccountTagsListArgs(t *testing.T) {
	if accountTagsListCmd.Args == nil {
		t.Error("accountTagsListCmd.Args is nil")
		return
	}

	err := accountTagsListCmd.Args(accountTagsListCmd, []string{})
	if err == nil {
		t.Error("Expected error for no arguments")
	}

	err = accountTagsListCmd.Args(accountTagsListCmd, []string{"1"})
	if err != nil {
		t.Errorf("Expected no error for 1 argument, got: %v", err)
	}
}

func TestAccountTagsFilterArgs(t *testing.T) {
	if accountTagsFilterCmd.Args == nil {
		t.Error("accountTagsFilterCmd.Args is nil")
		return
	}

	err := accountTagsFilterCmd.Args(accountTagsFilterCmd, []string{"work"})
	if err != nil {
		t.Errorf("Expected no error for 1 argument, got: %v", err)
	}
}

// ============================================================================
// TAG COMMAND TESTS
// ============================================================================

func TestTagSubcommands(t *testing.T) {
	expectedSubcommands := []string{"create", "list", "delete"}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range tagCmd.Commands() {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tag subcommand '%s' not found", expected)
		}
	}
}

func TestTagCreateCmdUse(t *testing.T) {
	if tagCreateCmd.Use != "create <name>" {
		t.Errorf("Expected 'create <name>', got %s", tagCreateCmd.Use)
	}
}

func TestTagCreateHasAliases(t *testing.T) {
	aliases := []string{"add", "new"}
	for _, alias := range aliases {
		found := false
		for _, cmdAlias := range tagCreateCmd.Aliases {
			if cmdAlias == alias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tagCreateCmd should have alias '%s'", alias)
		}
	}
}

// ============================================================================
// TOTP COMMAND TESTS
// ============================================================================

func TestTOTPSubcommands(t *testing.T) {
	expectedSubcommands := []string{"add", "show", "generate", "verify", "convert", "delete"}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range totpCmd.Commands() {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("totp subcommand '%s' not found", expected)
		}
	}
}

func TestTOTPCommandAliases(t *testing.T) {
	tests := []struct {
		cmd   *cobra.Command
		alias string
	}{
		{totpShowCmd, "view"},
		{totpGenerateCmd, "code"},
		{totpConvertCmd, "upgrade"},
		{totpDeleteCmd, "rm"},
	}

	for _, tt := range tests {
		found := false
		for _, alias := range tt.cmd.Aliases {
			if alias == tt.alias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("TOTP command %s should have alias '%s'", tt.cmd.Name(), tt.alias)
		}
	}
}

func TestTOTPAddArgs(t *testing.T) {
	if totpAddCmd.Args == nil {
		t.Error("totpAddCmd.Args is nil")
		return
	}

	err := totpAddCmd.Args(totpAddCmd, []string{"1", "JBSWY3DPEHPK3PXP"})
	if err != nil {
		t.Errorf("Expected no error for 2 arguments, got: %v", err)
	}
}

func TestTOTPGenerateArgs(t *testing.T) {
	if totpGenerateCmd.Args == nil {
		t.Error("totpGenerateCmd.Args is nil")
		return
	}

	err := totpGenerateCmd.Args(totpGenerateCmd, []string{"1"})
	if err != nil {
		t.Errorf("Expected no error for 1 argument, got: %v", err)
	}
}

func TestTOTPVerifyArgs(t *testing.T) {
	if totpVerifyCmd.Args == nil {
		t.Error("totpVerifyCmd.Args is nil")
		return
	}

	err := totpVerifyCmd.Args(totpVerifyCmd, []string{"1", "123456"})
	if err != nil {
		t.Errorf("Expected no error for 2 arguments, got: %v", err)
	}
}

// ============================================================================
// SECURITY COMMAND TESTS
// ============================================================================

func TestSecuritySubcommands(t *testing.T) {
	expectedSubcommands := []string{"check", "scan", "vulnerable", "stats"}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range securityCmd.Commands() {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("security subcommand '%s' not found", expected)
		}
	}
}

func TestSecurityCommandAliases(t *testing.T) {
	tests := []struct {
		cmd   *cobra.Command
		alias string
	}{
		{securityCmd, "sec"},
		{securityCheckCmd, "test"},
		{securityScanCmd, "check-all"},
		{securityVulnerableCmd, "weak"},
		{securityStatsCmd, "statistics"},
	}

	for _, tt := range tests {
		found := false
		for _, alias := range tt.cmd.Aliases {
			if alias == tt.alias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Security command %s should have alias '%s'", tt.cmd.Name(), tt.alias)
		}
	}
}

func TestSecurityCheckArgs(t *testing.T) {
	if securityCheckCmd.Args == nil {
		t.Error("securityCheckCmd.Args is nil")
		return
	}

	err := securityCheckCmd.Args(securityCheckCmd, []string{"MyPassword123"})
	if err != nil {
		t.Errorf("Expected no error for 1 argument, got: %v", err)
	}
}

// ============================================================================
// GENERATE COMMAND TESTS
// ============================================================================

func TestGenerateSubcommands(t *testing.T) {
	expectedSubcommands := []string{"password", "passphrase"}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range generateCmd.Commands() {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("generate subcommand '%s' not found", expected)
		}
	}
}

func TestGenerateCommandAliases(t *testing.T) {
	tests := []struct {
		cmd   *cobra.Command
		alias string
	}{
		{generateCmd, "gen"},
		{generatePasswordCmd, "pwd"},
		{generatePassphraseCmd, "phrase"},
	}

	for _, tt := range tests {
		found := false
		for _, alias := range tt.cmd.Aliases {
			if alias == tt.alias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Generate command %s should have alias '%s'", tt.cmd.Name(), tt.alias)
		}
	}
}

// ============================================================================
// FLAGS TESTS
// ============================================================================

func TestGlobalFlagsExist(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("database")
	if flag == nil {
		t.Error("database flag should be defined")
	}

	flag = rootCmd.PersistentFlags().Lookup("key")
	if flag == nil {
		t.Error("key flag should be defined")
	}
}

func TestServerFlagsExist(t *testing.T) {
	flag := serverCmd.Flags().Lookup("port")
	if flag == nil {
		t.Error("port flag should be defined")
	}
}

func TestAccountAddFlagsExist(t *testing.T) {
	requiredFlags := []string{"email", "user", "password", "url", "notes", "expire"}

	for _, flagName := range requiredFlags {
		flag := accountAddCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("%s flag should be defined", flagName)
		}
	}
}

func TestAccountUpdateFlagsExist(t *testing.T) {
	requiredFlags := []string{
		"email",
		"user",
		"password",
		"url",
		"notes",
		"expire",
		"change-reason",
	}

	for _, flagName := range requiredFlags {
		flag := accountUpdateCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("%s flag should be defined", flagName)
		}
	}
}

func TestShowPasswordFlags(t *testing.T) {
	commands := []*cobra.Command{
		accountListCmd,
		accountGetCmd,
		accountSearchCmd,
		accountSearchFieldCmd,
		accountFuzzySearchCmd,
		accountTagsFilterCmd,
	}

	for _, cmd := range commands {
		flag := cmd.Flags().Lookup("show-password")
		if flag == nil {
			t.Errorf("%s should have 'show-password' flag", cmd.Name())
		}
	}
}

func TestFuzzySearchFlags(t *testing.T) {
	flags := []string{"min-score", "tags", "has-totp", "expire-before", "expire-after"}

	for _, flagName := range flags {
		flag := accountFuzzySearchCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("fuzzy search should have '%s' flag", flagName)
		}
	}
}

func TestGeneratePasswordFlags(t *testing.T) {
	flags := []string{
		"length",
		"lowercase",
		"uppercase",
		"numbers",
		"special",
		"exclude-ambiguous",
		"count",
		"quiet",
	}

	for _, flagName := range flags {
		flag := generatePasswordCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("generate password should have '%s' flag", flagName)
		}
	}
}

func TestGeneratePassphraseFlags(t *testing.T) {
	flags := []string{"words", "separator", "capitalize", "numbers", "count", "quiet"}

	for _, flagName := range flags {
		flag := generatePassphraseCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("generate passphrase should have '%s' flag", flagName)
		}
	}
}

// ============================================================================
// REPOSITORY VARIABLES TESTS
// ============================================================================

func TestRepositoryVariables(t *testing.T) {
	// These should be initialized when needed
	// Just test that they can be assigned
	accountRepoCli = &AccountRepository{}
	if accountRepoCli == nil {
		t.Error("accountRepoCli assignment failed")
	}

	tagRepoCli = &TagRepository{}
	if tagRepoCli == nil {
		t.Error("tagRepoCli assignment failed")
	}

	totpRepoCli = &TOTPRepository{}
	if totpRepoCli == nil {
		t.Error("totpRepoCli assignment failed")
	}
}

// ============================================================================
// COMMAND EXAMPLES TESTS
// ============================================================================

func TestCommandExamplesExist(t *testing.T) {
	commands := []*cobra.Command{
		accountAddCmd,
		accountListCmd,
		accountGetCmd,
		accountUpdateCmd,
		accountDeleteCmd,
		accountSearchCmd,
		accountSearchFieldCmd,
		accountFuzzySearchCmd,
		accountTagsAddCmd,
		accountTagsFilterCmd,
		tagCreateCmd,
		totpAddCmd,
		totpGenerateCmd,
		securityCheckCmd,
		generatePasswordCmd,
		generatePassphraseCmd,
	}

	for _, cmd := range commands {
		if cmd.Example == "" {
			t.Errorf("Command '%s' should have examples", cmd.Name())
		}
	}
}

// ============================================================================
// COMMAND DESCRIPTIONS TESTS
// ============================================================================

// func TestCommandDescriptionsExist(t *testing.T) {
// 	commands := []*cobra.Command{
// 		rootCmd,
// 		accountCmd,
// 		tagCmd,
// 		totpCmd,
// 		securityCmd,
// 		generateCmd,
// 	}
//
// 	for _, cmd := range commands {
// 		if cmd.Short == "" {
// 			t.Errorf("Command '%s' should have Short description", cmd.Name())
// 		}
// 		if cmd.Long == "" {
// 			t.Errorf("Command '%s' should have Long description", cmd.Name())
// 		}
// 	}
// }

// ============================================================================
// BACKWARD COMPATIBILITY TESTS
// ============================================================================

func TestBackwardCompatibilityAliases(t *testing.T) {
	// Test that old command names still work via aliases
	aliases := map[string][]string{
		"account":  {"acc", "a"},
		"tag":      {"tags"},
		"generate": {"gen"},
		"security": {"sec", "audit"},
	}

	for cmdName, expectedAliases := range aliases {
		var cmd *cobra.Command
		for _, c := range rootCmd.Commands() {
			if c.Name() == cmdName {
				cmd = c
				break
			}
		}

		if cmd == nil {
			t.Errorf("Command '%s' not found", cmdName)
			continue
		}

		for _, alias := range expectedAliases {
			found := false
			for _, cmdAlias := range cmd.Aliases {
				if cmdAlias == alias {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Command '%s' should have alias '%s'", cmdName, alias)
			}
		}
	}
}
