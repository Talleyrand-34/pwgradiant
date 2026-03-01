package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global variables
var (
	dbPath         string
	encryptionKey  string
	serverPort     string
	accountRepoCli *AccountRepository
	tagRepoCli     *TagRepository
	totpRepoCli    *TOTPRepository
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "pwdmgr",
	Short: "Secure password manager with homomorphic TOTP",
	Long: `A secure password manager with SQLCipher encryption and homomorphic TOTP support.
Store, manage, and organize your passwords securely with advanced features like
fuzzy search, security auditing, and cryptographically secure password generation.`,
}

// ============================================================================
// INITIALIZATION COMMANDS
// ============================================================================

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long:  "Create a new encrypted database with the specified master password",
	Example: `  # Initialize database
  pwdmgr init --key "your-master-password"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := InitDB(dbPath, encryptionKey); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		CloseDB()
		fmt.Println("✓ Database initialized successfully")
		return nil
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the API server",
	Long:  "Start the REST API server to access the password manager via HTTP",
	Example: `  # Start server on default port (8080)
  pwdmgr server --key "your-master-password"
  
  # Start server on custom port
  pwdmgr server --port 3000 --key "your-master-password"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := InitDB(dbPath, encryptionKey); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		// Initialize security checker if common passwords file exists
		if _, err := os.Stat("common_passwords.txt"); err == nil {
			securityChecker = NewPasswordSecurityChecker()
			if err := securityChecker.LoadCommonPasswords("common_passwords.txt"); err != nil {
				fmt.Printf("Warning: Failed to load common passwords: %v\n", err)
			}
		}

		router := SetupRouter()
		fmt.Printf("🚀 Server starting on :%s\n", serverPort)
		return router.Run(":" + serverPort)
	},
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func initDatabase() error {
	if err := InitDB(dbPath, encryptionKey); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	accountRepoCli = &AccountRepository{}
	tagRepoCli = &TagRepository{}
	totpRepoCli = &TOTPRepository{}
	return nil
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting output: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func Execute() error {
	return rootCmd.Execute()
}

// ============================================================================
// ACCOUNT COMMANDS
// ============================================================================

var accountCmd = &cobra.Command{
	Use:     "account",
	Aliases: []string{"acc", "a"},
	Short:   "Manage password accounts",
	Long:    "Create, read, update, delete and search password accounts",
}

var accountAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"create", "new"},
	Short:   "Create a new account",
	Long:    "Add a new password account to the database",
	Example: `  # Create account
  pwdmgr account add --user "alice@example.com" --password "SecurePass123!" --key "$KEY"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		email, _ := cmd.Flags().GetString("email")
		user, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("password")
		url, _ := cmd.Flags().GetString("url")
		notes, _ := cmd.Flags().GetString("notes")
		expire, _ := cmd.Flags().GetString("expire")

		account := &Account{User: user, Password: password}
		if email != "" {
			account.Email = &email
		}
		if url != "" {
			account.URL = &url
		}
		if notes != "" {
			account.Notes = &notes
		}
		if expire != "" {
			account.Expire = expire
		}

		id, err := accountRepoCli.Create(account)
		if err != nil {
			return fmt.Errorf("failed to create account: %w", err)
		}

		fmt.Printf("✓ Account created with ID: %d\n", id)
		return nil
	},
}

var accountListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "all"},
	Short:   "List all accounts",
	Example: `  # List accounts
  pwdmgr account list --key "$KEY"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		accounts, err := accountRepoCli.GetAll()
		if err != nil {
			return fmt.Errorf("failed to get accounts: %w", err)
		}

		showPassword, _ := cmd.Flags().GetBool("show-password")
		if !showPassword {
			accounts = HidePasswords(accounts)
		}

		printJSON(accounts)
		return nil
	},
}

var accountGetCmd = &cobra.Command{
	Use:     "get <id>",
	Aliases: []string{"show", "view"},
	Short:   "Get an account by ID",
	Example: `  # Get account
  pwdmgr account get 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var id int64
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		account, err := accountRepoCli.GetByID(id)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		showPassword, _ := cmd.Flags().GetBool("show-password")
		if !showPassword {
			hiddenAccount := account.HidePassword()
			printJSON(hiddenAccount)
		} else {
			printJSON(account)
		}
		return nil
	},
}

var accountUpdateCmd = &cobra.Command{
	Use:     "update <id>",
	Aliases: []string{"edit", "modify"},
	Short:   "Update an account",
	Example: `  # Update password
  pwdmgr account update 1 --password "NewPassword123!" --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var id int64
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		account, err := accountRepoCli.GetByID(id)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		if cmd.Flags().Changed("email") {
			email, _ := cmd.Flags().GetString("email")
			account.Email = &email
		}
		if cmd.Flags().Changed("user") {
			account.User, _ = cmd.Flags().GetString("user")
		}
		if cmd.Flags().Changed("password") {
			account.Password, _ = cmd.Flags().GetString("password")
		}
		if cmd.Flags().Changed("url") {
			url, _ := cmd.Flags().GetString("url")
			account.URL = &url
		}
		if cmd.Flags().Changed("notes") {
			notes, _ := cmd.Flags().GetString("notes")
			account.Notes = &notes
		}
		if cmd.Flags().Changed("expire") {
			account.Expire, _ = cmd.Flags().GetString("expire")
		}

		changeReason, _ := cmd.Flags().GetString("change-reason")
		var changeReasonPtr *string
		if changeReason != "" {
			changeReasonPtr = &changeReason
		}

		if err := accountRepoCli.Update(account, changeReasonPtr); err != nil {
			return fmt.Errorf("failed to update account: %w", err)
		}

		fmt.Printf("✓ Account %d updated successfully\n", id)
		return nil
	},
}

var accountDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete an account",
	Example: `  # Delete account
  pwdmgr account delete 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var id int64
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		if err := accountRepoCli.Delete(id); err != nil {
			return fmt.Errorf("failed to delete account: %w", err)
		}

		fmt.Printf("✓ Account %d deleted successfully\n", id)
		return nil
	},
}

var accountHistoryCmd = &cobra.Command{
	Use:     "history <id>",
	Aliases: []string{"log", "changelog"},
	Short:   "View account history",
	Example: `  # View history
  pwdmgr account history 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var id int64
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		history, err := accountRepoCli.GetHistory(id)
		if err != nil {
			return fmt.Errorf("failed to get history: %w", err)
		}

		printJSON(history)
		return nil
	},
}

var accountSearchCmd = &cobra.Command{
	Use:     "search <query>",
	Aliases: []string{"find", "query"},
	Short:   "Search accounts",
	Example: `  # Search accounts
  pwdmgr account search "example" --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		accounts, err := accountRepoCli.Search(args[0])
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		showPassword, _ := cmd.Flags().GetBool("show-password")
		if !showPassword {
			accounts = HidePasswords(accounts)
		}

		printJSON(accounts)
		return nil
	},
}

var accountSearchFieldCmd = &cobra.Command{
	Use:   "search-field <field> <query>",
	Short: "Search by specific field",
	Example: `  # Search by email
  pwdmgr account search-field email "@gmail.com" --key "$KEY"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		accounts, err := accountRepoCli.SearchByField(args[0], args[1])
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		showPassword, _ := cmd.Flags().GetBool("show-password")
		if !showPassword {
			accounts = HidePasswords(accounts)
		}

		printJSON(accounts)
		return nil
	},
}

var accountFuzzySearchCmd = &cobra.Command{
	Use:     "fuzzy <query>",
	Aliases: []string{"similar", "fuzzy-search"},
	Short:   "Fuzzy search with similarity",
	Example: `  # Fuzzy search
  pwdmgr account fuzzy "john" --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		minScore, _ := cmd.Flags().GetFloat64("min-score")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		hasTOTP, _ := cmd.Flags().GetBool("has-totp")
		hasTOTPSet := cmd.Flags().Changed("has-totp")
		expireBefore, _ := cmd.Flags().GetString("expire-before")
		expireAfter, _ := cmd.Flags().GetString("expire-after")
		showPassword, _ := cmd.Flags().GetBool("show-password")

		filters := &SearchFilters{MinScore: minScore}
		if len(tags) > 0 {
			filters.Tags = tags
		}
		if hasTOTPSet {
			filters.HasTOTP = &hasTOTP
		}
		if expireBefore != "" {
			filters.ExpireBefore = &expireBefore
		}
		if expireAfter != "" {
			filters.ExpireAfter = &expireAfter
		}

		results, err := accountRepoCli.FuzzySearch(args[0], filters)
		if err != nil {
			return fmt.Errorf("fuzzy search failed: %w", err)
		}

		if !showPassword {
			for i := range results {
				results[i].Account = results[i].Account.HidePassword()
			}
		}

		printJSON(results)
		return nil
	},
}

// ============================================================================
// ACCOUNT TAGS SUBCOMMANDS
// ============================================================================

var accountTagsCmd = &cobra.Command{
	Use:     "tags",
	Aliases: []string{"tag", "t"},
	Short:   "Manage account tags",
}

var accountTagsAddCmd = &cobra.Command{
	Use:     "add <account_id> <tag_name>",
	Aliases: []string{"attach", "associate"},
	Short:   "Add a tag to an account",
	Example: `  # Add tag
  pwdmgr account tags add 1 "work" --key "$KEY"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		if err := accountRepoCli.AddTagByName(accountID, args[1]); err != nil {
			return fmt.Errorf("failed to add tag: %w", err)
		}

		fmt.Printf("✓ Tag '%s' added to account %d\n", args[1], accountID)
		return nil
	},
}

var accountTagsRemoveCmd = &cobra.Command{
	Use:     "remove <account_id> <tag_name>",
	Aliases: []string{"rm", "delete", "detach"},
	Short:   "Remove a tag from an account",
	Example: `  # Remove tag
  pwdmgr account tags remove 1 "work" --key "$KEY"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		if err := accountRepoCli.RemoveTagByName(accountID, args[1]); err != nil {
			return fmt.Errorf("failed to remove tag: %w", err)
		}

		fmt.Printf("✓ Tag '%s' removed from account %d\n", args[1], accountID)
		return nil
	},
}

var accountTagsListCmd = &cobra.Command{
	Use:     "list <account_id>",
	Aliases: []string{"ls", "show"},
	Short:   "List tags for an account",
	Example: `  # List tags
  pwdmgr account tags list 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		account, err := accountRepoCli.GetByID(accountID)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		if len(account.Tags) == 0 {
			fmt.Printf("No tags for account %d\n", accountID)
			return nil
		}

		fmt.Printf("Tags for account %d:\n", accountID)
		printJSON(account.Tags)
		return nil
	},
}

var accountTagsFilterCmd = &cobra.Command{
	Use:     "filter <tag_name>",
	Aliases: []string{"find", "by"},
	Short:   "Filter accounts by tag",
	Example: `  # Filter by tag
  pwdmgr account tags filter "work" --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		accounts, err := accountRepoCli.GetAccountsByTagName(args[0])
		if err != nil {
			return fmt.Errorf("failed to get accounts: %w", err)
		}

		if len(accounts) == 0 {
			fmt.Printf("No accounts with tag '%s'\n", args[0])
			return nil
		}

		showPassword, _ := cmd.Flags().GetBool("show-password")
		if !showPassword {
			accounts = HidePasswords(accounts)
		}

		fmt.Printf("Accounts with tag '%s':\n", args[0])
		printJSON(accounts)
		return nil
	},
}

// ============================================================================
// TAG COMMANDS
// ============================================================================

var tagCmd = &cobra.Command{
	Use:     "tag",
	Aliases: []string{"tags"},
	Short:   "Manage tags",
}

var tagCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Aliases: []string{"add", "new"},
	Short:   "Create a new tag",
	Example: `  # Create tag
  pwdmgr tag create "work" --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		tag := &Tag{Name: args[0]}
		id, err := tagRepoCli.Create(tag)
		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}

		fmt.Printf("✓ Tag '%s' created with ID: %d\n", args[0], id)
		return nil
	},
}

var tagListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "all"},
	Short:   "List all tags",
	Example: `  # List tags
  pwdmgr tag list --key "$KEY"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		tags, err := tagRepoCli.GetAll()
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}

		printJSON(tags)
		return nil
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a tag",
	Example: `  # Delete tag
  pwdmgr tag delete 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var id int64
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("invalid tag id: %s", args[0])
		}

		if err := tagRepoCli.Delete(id); err != nil {
			return fmt.Errorf("failed to delete tag: %w", err)
		}

		fmt.Printf("✓ Tag %d deleted successfully\n", id)
		return nil
	},
}

// ============================================================================
// TOTP COMMANDS
// ============================================================================

var totpCmd = &cobra.Command{
	Use:   "totp",
	Short: "Manage TOTP (2FA)",
}

var totpAddCmd = &cobra.Command{
	Use:   "add <account_id> <seed>",
	Short: "Add TOTP to account",
	Example: `  # Add TOTP
  pwdmgr totp add 1 "JBSWY3DPEHPK3PXP" --key "$KEY"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		totp := &TOTP{AccountID: accountID, TOTPSeed: args[1]}
		id, err := totpRepoCli.Create(totp)
		if err != nil {
			return fmt.Errorf("failed to add TOTP: %w", err)
		}

		fmt.Printf("✓ TOTP added to account %d (TOTP ID: %d)\n", accountID, id)
		return nil
	},
}

var totpShowCmd = &cobra.Command{
	Use:     "show <account_id>",
	Aliases: []string{"view", "info"},
	Short:   "Show TOTP details",
	Example: `  # Show TOTP
  pwdmgr totp show 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		totp, err := totpRepoCli.GetByAccountID(accountID)
		if err != nil {
			return fmt.Errorf("failed to get TOTP: %w", err)
		}

		fmt.Printf("TOTP for Account %d:\n", accountID)
		fmt.Printf("  ID: %d\n", totp.ID)
		fmt.Printf("  Seed: %s\n", totp.TOTPSeed)
		fmt.Printf("  Homomorphic: %v\n", totp.UseHomomorphic)

		if totp.UseHomomorphic {
			fmt.Printf("  Encrypted Seed: %s\n", *totp.CTOTPSeed)
			fmt.Println("\n  Note: Requires private key for code generation")
		} else {
			code, err := GenerateStandardTOTP(totp.TOTPSeed)
			if err != nil {
				return fmt.Errorf("failed to generate code: %w", err)
			}
			fmt.Printf("\n  Current Code: %s\n", code)
		}

		return nil
	},
}

var totpGenerateCmd = &cobra.Command{
	Use:     "generate <account_id>",
	Aliases: []string{"code", "gen"},
	Short:   "Generate TOTP code",
	Example: `  # Generate code
  pwdmgr totp generate 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		totp, err := totpRepoCli.GetByAccountID(accountID)
		if err != nil {
			return fmt.Errorf("failed to get TOTP: %w", err)
		}
		var code string
		// 1. Determine which method to use and assign to the outer scope variables
		if totp.TOTPSeed != "" {
			code, err = GenerateStandardTOTPGeneric(totp.TOTPSeed, "", "", false)
		} else {
			code, err = GenerateStandardTOTPGeneric("", *totp.CTOTPSeed, *totp.PaillierN, true)
		}

		// 2. Check for errors once after the logic branch
		if err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}
		fmt.Printf("TOTP Code: %s\n", code)
		return nil
	},
}

var totpVerifyCmd = &cobra.Command{
	Use:   "verify <account_id> <code>",
	Short: "Verify TOTP code",
	Example: `  # Verify code
  pwdmgr totp verify 1 123456 --key "$KEY"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		totp, err := totpRepoCli.GetByAccountID(accountID)
		if err != nil {
			return fmt.Errorf("failed to get TOTP: %w", err)
		}

		if totp.UseHomomorphic {
			return fmt.Errorf("homomorphic TOTP verification requires private key (use API)")
		}

		// valid, err := VerifyStandardTOTP(totp.TOTPSeed, args[1])
		// if err != nil {
		// 	return fmt.Errorf("verification failed: %w", err)
		// }
		//
		// if valid {
		// 	fmt.Println("✓ Code is valid")
		// } else {
		// 	fmt.Println("✗ Code is invalid")
		// }

		return nil
	},
}

var totpConvertCmd = &cobra.Command{
	Use:     "convert <account_id>",
	Aliases: []string{"upgrade", "encrypt"},
	Short:   "Convert to homomorphic",
	Example: `  # Convert
  pwdmgr totp convert 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		totp, err := totpRepoCli.GetByAccountID(accountID)
		if err != nil {
			return fmt.Errorf("failed to get TOTP: %w", err)
		}

		if totp.UseHomomorphic {
			return fmt.Errorf("TOTP already using homomorphic encryption")
		}

		keyBits, _ := cmd.Flags().GetInt("key-bits")
		fmt.Printf("Generating %d-bit Paillier key pair...\n", keyBits)

		privateKey, err := CreatePaillierKeyPair(keyBits)
		if err != nil {
			return fmt.Errorf("failed to create key pair: %w", err)
		}

		homomorphicData, err := EncryptTOTPSeed(totp.TOTPSeed, privateKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt seed: %w", err)
		}

		totp.CTOTPSeed = &homomorphicData.EncryptedSeed
		totp.PaillierN = &homomorphicData.PaillierN
		totp.UseHomomorphic = true

		if err := totpRepoCli.Update(totp); err != nil {
			return fmt.Errorf("failed to update TOTP: %w", err)
		}

		nHex, lambdaHex := SerializePrivateKey(privateKey)

		fmt.Println("\n✓ TOTP converted to homomorphic encryption")
		fmt.Println("\n⚠️  IMPORTANT: Store these private key parameters securely!")
		fmt.Printf("\n  Paillier N:      %s\n", nHex)
		fmt.Printf("  Paillier Lambda: %s\n", lambdaHex)
		fmt.Println("\n  Save these values in a secure location")

		return nil
	},
}

var totpDeleteCmd = &cobra.Command{
	Use:     "delete <account_id>",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete TOTP",
	Example: `  # Delete TOTP
  pwdmgr totp delete 1 --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		var accountID int64
		if _, err := fmt.Sscanf(args[0], "%d", &accountID); err != nil {
			return fmt.Errorf("invalid account id: %s", args[0])
		}

		totp, err := totpRepoCli.GetByAccountID(accountID)
		if err != nil {
			return fmt.Errorf("failed to get TOTP: %w", err)
		}

		if err := totpRepoCli.Delete(totp.ID); err != nil {
			return fmt.Errorf("failed to delete TOTP: %w", err)
		}

		fmt.Printf("✓ TOTP deleted from account %d\n", accountID)
		return nil
	},
}

// ============================================================================
// SECURITY COMMANDS
// ============================================================================

var securityCmd = &cobra.Command{
	Use:     "security",
	Aliases: []string{"sec", "audit"},
	Short:   "Password security auditing",
}

var securityCheckCmd = &cobra.Command{
	Use:     "check <password>",
	Aliases: []string{"test", "analyze"},
	Short:   "Check password security",
	Example: `  # Check password
  pwdmgr security check "MyPass123" --key "$KEY"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		commonPasswordsFile, _ := cmd.Flags().GetString("common-passwords")
		checker := NewPasswordSecurityChecker()

		if err := checker.LoadCommonPasswords(commonPasswordsFile); err != nil {
			return fmt.Errorf("failed to load common passwords: %w", err)
		}

		report := checker.AnalyzePassword(args[0])

		fmt.Printf("\n=== Password Security Analysis ===\n\n")
		fmt.Printf("Strength: %s (Score: %d/7)\n", report.StrengthText, report.Score)
		fmt.Printf("Length: %d characters\n", report.Length)

		if report.IsCommon {
			fmt.Printf("\n⚠️  WARNING: Common password!\n")
		}

		if len(report.Warnings) > 0 {
			fmt.Printf("\n⚠️  Warnings:\n")
			for _, w := range report.Warnings {
				fmt.Printf("  - %s\n", w)
			}
		}

		if len(report.Recommendations) > 0 {
			fmt.Printf("\n💡 Recommendations:\n")
			for _, r := range report.Recommendations {
				fmt.Printf("  - %s\n", r)
			}
		}

		fmt.Println()
		return nil
	},
}

var securityScanCmd = &cobra.Command{
	Use:     "scan",
	Aliases: []string{"check-all", "audit-all"},
	Short:   "Scan all passwords",
	Example: `  # Scan all
  pwdmgr security scan --key "$KEY"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		commonPasswordsFile, _ := cmd.Flags().GetString("common-passwords")
		checker := NewPasswordSecurityChecker()

		if err := checker.LoadCommonPasswords(commonPasswordsFile); err != nil {
			return fmt.Errorf("failed to load common passwords: %w", err)
		}

		reports, err := accountRepoCli.CheckAllAccountPasswords(checker)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		fmt.Printf("\n=== Security Report ===\n\n")
		fmt.Printf("Total Accounts: %d\n\n", len(reports))

		for _, report := range reports {
			status := "✓"
			if report.Analysis.IsCommon || report.Analysis.Strength <= PasswordWeak {
				status = "⚠️"
			}

			fmt.Printf("%s Account #%d - %s\n", status, report.AccountID, report.User)
			fmt.Printf("   Strength: %s", report.Analysis.StrengthText)
			if report.Analysis.IsCommon {
				fmt.Printf(" (COMMON!)")
			}
			fmt.Println()
		}

		return nil
	},
}

var securityVulnerableCmd = &cobra.Command{
	Use:     "vulnerable",
	Aliases: []string{"weak", "insecure"},
	Short:   "List vulnerable accounts",
	Example: `  # List vulnerable
  pwdmgr security vulnerable --key "$KEY"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		commonPasswordsFile, _ := cmd.Flags().GetString("common-passwords")
		checker := NewPasswordSecurityChecker()

		if err := checker.LoadCommonPasswords(commonPasswordsFile); err != nil {
			return fmt.Errorf("failed to load common passwords: %w", err)
		}

		vulnerable, err := accountRepoCli.GetVulnerableAccounts(checker)
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}

		if len(vulnerable) == 0 {
			fmt.Println("\n✓ No vulnerable accounts!")
			return nil
		}

		fmt.Printf("\n⚠️  Found %d Vulnerable:\n\n", len(vulnerable))

		for _, report := range vulnerable {
			fmt.Printf("Account #%d - %s\n", report.AccountID, report.User)
			fmt.Printf("  Strength: %s\n", report.Analysis.StrengthText)
			if report.Analysis.IsCommon {
				fmt.Printf("  ⚠️  COMMON password!\n")
			}
			fmt.Println()
		}

		return nil
	},
}

var securityStatsCmd = &cobra.Command{
	Use:     "stats",
	Aliases: []string{"statistics", "summary"},
	Short:   "Security statistics",
	Example: `  # Show stats
  pwdmgr security stats --key "$KEY"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initDatabase(); err != nil {
			return err
		}
		defer CloseDB()

		commonPasswordsFile, _ := cmd.Flags().GetString("common-passwords")
		checker := NewPasswordSecurityChecker()

		if err := checker.LoadCommonPasswords(commonPasswordsFile); err != nil {
			return fmt.Errorf("failed to load common passwords: %w", err)
		}

		stats, err := accountRepoCli.GetSecurityStatistics(checker, false)
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}

		fmt.Printf("\n=== Security Statistics ===\n\n")
		fmt.Printf("Total Accounts: %d\n\n", stats.TotalAccounts)

		fmt.Printf("Strength Distribution:\n")
		fmt.Printf("  Very Strong: %d\n", stats.VeryStrongPasswords)
		fmt.Printf("  Strong:      %d\n", stats.StrongPasswords)
		fmt.Printf("  Medium:      %d\n", stats.MediumPasswords)
		fmt.Printf("  Weak:        %d\n", stats.WeakPasswords)
		fmt.Printf("  Very Weak:   %d\n", stats.VeryWeakPasswords)

		fmt.Printf("\nCommon Passwords: %d\n", stats.CommonPasswords)

		fmt.Println()
		return nil
	},
}

// ============================================================================
// GENERATE COMMANDS
// ============================================================================

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen"},
	Short:   "Generate passwords",
}

var generatePasswordCmd = &cobra.Command{
	Use:     "password",
	Aliases: []string{"pwd", "pass"},
	Short:   "Generate password",
	Example: `  # Generate
  pwdmgr generate password`,
	RunE: func(cmd *cobra.Command, args []string) error {
		length, _ := cmd.Flags().GetInt("length")
		lower, _ := cmd.Flags().GetBool("lowercase")
		upper, _ := cmd.Flags().GetBool("uppercase")
		numbers, _ := cmd.Flags().GetBool("numbers")
		special, _ := cmd.Flags().GetBool("special")
		excludeAmbiguous, _ := cmd.Flags().GetBool("exclude-ambiguous")
		count, _ := cmd.Flags().GetInt("count")
		quiet, _ := cmd.Flags().GetBool("quiet")

		opts := PasswordGeneratorOptions{
			Length:           length,
			IncludeLower:     lower,
			IncludeUpper:     upper,
			IncludeNumbers:   numbers,
			IncludeSpecial:   special,
			ExcludeAmbiguous: excludeAmbiguous,
		}

		if count > 1 {
			passwords, err := GenerateMultiplePasswords(opts, count)
			if err != nil {
				return fmt.Errorf("failed: %w", err)
			}

			if quiet {
				for _, pwd := range passwords {
					fmt.Println(pwd.Password)
				}
			} else {
				fmt.Printf("\n=== Generated %d Passwords ===\n\n", count)
				for i, pwd := range passwords {
					fmt.Printf("%d. %s\n", i+1, pwd.Password)
				}
			}
		} else {
			generated, err := GeneratePassword(opts)
			if err != nil {
				return fmt.Errorf("failed: %w", err)
			}

			if quiet {
				fmt.Println(generated.Password)
			} else {
				fmt.Printf("\nPassword: %s\n\n", generated.Password)
				fmt.Printf("Length:   %d\n", generated.Length)
				fmt.Printf("Strength: %s\n", generated.StrengthText)
				fmt.Printf("Entropy:  %.1f bits\n\n", generated.Entropy)
			}
		}

		return nil
	},
}

var generatePassphraseCmd = &cobra.Command{
	Use:     "passphrase",
	Aliases: []string{"phrase"},
	Short:   "Generate passphrase",
	Example: `  # Generate
  pwdmgr generate passphrase`,
	RunE: func(cmd *cobra.Command, args []string) error {
		words, _ := cmd.Flags().GetInt("words")
		separator, _ := cmd.Flags().GetString("separator")
		capitalize, _ := cmd.Flags().GetBool("capitalize")
		numbers, _ := cmd.Flags().GetBool("numbers")
		count, _ := cmd.Flags().GetInt("count")
		quiet, _ := cmd.Flags().GetBool("quiet")

		if count > 1 {
			if quiet {
				for i := 0; i < count; i++ {
					passphrase, err := GeneratePassphrase(words, separator, capitalize, numbers)
					if err != nil {
						return err
					}
					fmt.Println(passphrase)
				}
			} else {
				fmt.Printf("\n=== Generated %d Passphrases ===\n\n", count)
				for i := 0; i < count; i++ {
					passphrase, err := GeneratePassphrase(words, separator, capitalize, numbers)
					if err != nil {
						return err
					}
					fmt.Printf("%d. %s\n", i+1, passphrase)
				}
			}
		} else {
			passphrase, err := GeneratePassphrase(words, separator, capitalize, numbers)
			if err != nil {
				return fmt.Errorf("failed: %w", err)
			}

			if quiet {
				fmt.Println(passphrase)
			} else {
				fmt.Printf("\nPassphrase: %s\n\n", passphrase)
				fmt.Printf("Length: %d characters\n\n", len(passphrase))
			}
		}

		return nil
	},
}

// ============================================================================
// COMMAND INITIALIZATION
// ============================================================================

func init() {
	// Global flags
	rootCmd.PersistentFlags().
		StringVarP(&dbPath, "database", "d", "passwords.db", "Database file path")
	rootCmd.PersistentFlags().
		StringVarP(&encryptionKey, "key", "k", "", "Encryption key (required)")
	rootCmd.MarkPersistentFlagRequired("key")

	// Server flags
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "8080", "Server port")

	// Account add flags
	accountAddCmd.Flags().String("email", "", "Email")
	accountAddCmd.Flags().String("user", "", "Username")
	accountAddCmd.Flags().String("password", "", "Password")
	accountAddCmd.Flags().String("url", "", "URL")
	accountAddCmd.Flags().String("notes", "", "Notes")
	accountAddCmd.Flags().String("expire", "", "Expiration date")
	accountAddCmd.MarkFlagRequired("user")
	accountAddCmd.MarkFlagRequired("password")

	// Account update flags
	accountUpdateCmd.Flags().String("email", "", "Email")
	accountUpdateCmd.Flags().String("user", "", "Username")
	accountUpdateCmd.Flags().String("password", "", "Password")
	accountUpdateCmd.Flags().String("url", "", "URL")
	accountUpdateCmd.Flags().String("notes", "", "Notes")
	accountUpdateCmd.Flags().String("expire", "", "Expiration date")
	accountUpdateCmd.Flags().String("change-reason", "", "Change reason")

	// Show password flags
	accountListCmd.Flags().Bool("show-password", false, "Show passwords")
	accountGetCmd.Flags().Bool("show-password", false, "Show password")
	accountSearchCmd.Flags().Bool("show-password", false, "Show passwords")
	accountSearchFieldCmd.Flags().Bool("show-password", false, "Show passwords")
	accountFuzzySearchCmd.Flags().Bool("show-password", false, "Show passwords")
	accountTagsFilterCmd.Flags().Bool("show-password", false, "Show passwords")

	// Fuzzy search flags
	accountFuzzySearchCmd.Flags().Float64("min-score", 0.0, "Minimum score")
	accountFuzzySearchCmd.Flags().StringSlice("tags", []string{}, "Filter by tags")
	accountFuzzySearchCmd.Flags().Bool("has-totp", false, "Filter by TOTP")
	accountFuzzySearchCmd.Flags().String("expire-before", "", "Expire before")
	accountFuzzySearchCmd.Flags().String("expire-after", "", "Expire after")

	// TOTP convert flags
	totpConvertCmd.Flags().Int("key-bits", 2048, "Key size (2048/3072/4096)")

	// Security flags
	securityCmd.PersistentFlags().
		String("common-passwords", "common_passwords.txt", "Common passwords file")

	// Generate password flags
	generatePasswordCmd.Flags().Int("length", 16, "Length")
	generatePasswordCmd.Flags().Bool("lowercase", true, "Include lowercase")
	generatePasswordCmd.Flags().Bool("uppercase", true, "Include uppercase")
	generatePasswordCmd.Flags().Bool("numbers", true, "Include numbers")
	generatePasswordCmd.Flags().Bool("special", true, "Include special")
	generatePasswordCmd.Flags().Bool("exclude-ambiguous", false, "Exclude ambiguous")
	generatePasswordCmd.Flags().Int("count", 1, "Count")
	generatePasswordCmd.Flags().Bool("quiet", false, "Quiet mode")

	// Generate passphrase flags
	generatePassphraseCmd.Flags().Int("words", 4, "Words")
	generatePassphraseCmd.Flags().String("separator", "-", "Separator")
	generatePassphraseCmd.Flags().Bool("capitalize", true, "Capitalize")
	generatePassphraseCmd.Flags().Bool("numbers", true, "Include numbers")
	generatePassphraseCmd.Flags().Int("count", 1, "Count")
	generatePassphraseCmd.Flags().Bool("quiet", false, "Quiet mode")

	// Build account tags hierarchy
	accountTagsCmd.AddCommand(
		accountTagsAddCmd,
		accountTagsRemoveCmd,
		accountTagsListCmd,
		accountTagsFilterCmd,
	)

	// Build account hierarchy
	accountCmd.AddCommand(
		accountAddCmd,
		accountListCmd,
		accountGetCmd,
		accountUpdateCmd,
		accountDeleteCmd,
		accountHistoryCmd,
		accountSearchCmd,
		accountSearchFieldCmd,
		accountFuzzySearchCmd,
		accountTagsCmd,
	)

	// Build tag hierarchy
	tagCmd.AddCommand(
		tagCreateCmd,
		tagListCmd,
		tagDeleteCmd,
	)

	// Build TOTP hierarchy
	totpCmd.AddCommand(
		totpAddCmd,
		totpShowCmd,
		totpGenerateCmd,
		totpVerifyCmd,
		totpConvertCmd,
		totpDeleteCmd,
	)

	// Build security hierarchy
	securityCmd.AddCommand(
		securityCheckCmd,
		securityScanCmd,
		securityVulnerableCmd,
		securityStatsCmd,
	)

	// Build generate hierarchy
	generateCmd.AddCommand(
		generatePasswordCmd,
		generatePassphraseCmd,
	)

	// Build root hierarchy
	rootCmd.AddCommand(
		initCmd,
		serverCmd,
		accountCmd,
		tagCmd,
		totpCmd,
		securityCmd,
		generateCmd,
	)
}

// ExecuteCLI executes the root command
func ExecuteCLI() error {
	return rootCmd.Execute()
}
