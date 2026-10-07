package cli

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(auditCmd)
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit vault passwords for strength, reuse, and staleness",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		resources := v.List()
		if len(resources) == 0 {
			fmt.Println("Vault is empty. Nothing to audit.")
			return nil
		}

		var weakPasswords []string
		var stalePasswords []string
		passwordMap := make(map[string][]string) // password -> []resourceName

		now := time.Now()
		staleThreshold := 180 * 24 * time.Hour // 180 days

		for _, name := range resources {
			entry, canonName, _ := v.Get(name)

			for fKey, fVal := range entry.Fields {
				// Only audit concealed/password fields
				if !fVal.Concealed && !strings.Contains(strings.ToLower(fKey), "pass") {
					continue
				}

				pw := fVal.Value
				if pw == "" {
					continue
				}

				// Check strength
				isWeak, reason := checkPasswordStrength(pw)
				if isWeak {
					weakPasswords = append(weakPasswords, fmt.Sprintf("%s (%s: %s)", canonName, fKey, reason))
				}

				// Track duplicates
				passwordMap[pw] = append(passwordMap[pw], fmt.Sprintf("%s.%s", canonName, fKey))

				// Check staleness (only for valid non-zero timestamps)
				if !entry.ModifiedAt.IsZero() && entry.ModifiedAt.Year() > 2000 && entry.ModifiedAt.Before(now.Add(-staleThreshold)) {
					daysOld := int(now.Sub(entry.ModifiedAt).Hours() / 24)
					stalePasswords = append(stalePasswords, fmt.Sprintf("%s (%d days old)", canonName, daysOld))
				}
			}
		}

		fmt.Println("🔍 Vault Password Health Audit Report")
		fmt.Println("=======================================")
		fmt.Printf("Total Resources Audited: %d\n\n", len(resources))

		// 1. Weak Passwords
		if len(weakPasswords) > 0 {
			fmt.Printf("⚠️  Weak Passwords (%d):\n", len(weakPasswords))
			for _, w := range weakPasswords {
				fmt.Printf("   • %s\n", w)
			}
			fmt.Println()
		} else {
			fmt.Println("✅ No weak passwords found.")
		}

		// 2. Reused Passwords
		reusedCount := 0
		for _, usages := range passwordMap {
			if len(usages) > 1 {
				if reusedCount == 0 {
					fmt.Println("⚠️  Reused Passwords:")
				}
				reusedCount++
				fmt.Printf("   • Shared across: %s\n", strings.Join(usages, ", "))
			}
		}
		if reusedCount == 0 {
			fmt.Println("✅ No reused passwords found.")
		}
		fmt.Println()

		// 3. Stale Passwords
		if len(stalePasswords) > 0 {
			fmt.Printf("ℹ️  Stale Passwords > 180 days (%d):\n", len(stalePasswords))
			for _, s := range stalePasswords {
				fmt.Printf("   • %s\n", s)
			}
			fmt.Println()
		} else {
			fmt.Println("✅ All passwords are up to date.")
		}

		return nil
	},
}

func checkPasswordStrength(pw string) (isWeak bool, reason string) {
	if len(pw) < 12 {
		return true, fmt.Sprintf("short length (%d chars, min 12 recommended)", len(pw))
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	var missing []string
	if !hasUpper {
		missing = append(missing, "uppercase")
	}
	if !hasLower {
		missing = append(missing, "lowercase")
	}
	if !hasDigit {
		missing = append(missing, "digits")
	}
	if !hasSymbol {
		missing = append(missing, "symbols")
	}

	if len(missing) > 1 {
		return true, fmt.Sprintf("missing %s", strings.Join(missing, ", "))
	}

	return false, ""
}
