# MyVault

A local-first, CLI-first secret manager for developers. Your encrypted vault belongs to you.

- **No cloud account required**
- **No third-party password storage**
- **No vendor lock-in**
- **Zero-Passphrase Listing** via unencrypted local metadata index (`vault.index`)
- **Cross-platform OS Keyring session management** (macOS Keychain, Linux DBus SecretService, Windows Credential Manager)
- **Encryption powered by [age](https://filippo.io/age)** (scrypt KDF + ChaCha20-Poly1305 AEAD)

---

## 🚀 Features

- **Flexible Key-Value Secrets**: Store database credentials (`host`, `port`, `username`, `password`), API tokens, or SSH keys under a single resource.
- **Dynamic Syntax Support**: Accepts both `key=value` and `key value` (space-separated) syntaxes across all commands.
- **Smart Concealment**: Sensitive fields (`password`, `secret`, `token`, `key`) are redacted by default in CLI outputs.
- **Shell Autocomplete**: Tab completion for resources, tags, and commands (`zsh`, `bash`, `fish`, `powershell`).
- **Security Audit**: Scans passwords for weak strength, duplication/reuse, and staleness (>180 days).
- **Snapshot Backups & Restore**: Create timestamped encrypted snapshots (`myvault backup`) and restore safely (`myvault restore`).
- **Imports & Exports**: Import from JSON, CSV, Bitwarden, and 1Password; export to JSON or CSV.
- **Crash-Safe Storage**: Pre-write backups (`backup/vault.age`) and atomic writes prevent data corruption.
- **Memory Zeroing**: Sensitive byte buffers are zeroed out after processing via `memguard`.

---

## 📦 Install & Build

```bash
# Build & install from source
git clone https://github.com/rakshithgoud453/myvault.git
cd myvault
make install
```

To cross-compile release binaries for macOS (ARM/Intel), Linux (AMD64/ARM64), and Windows (AMD64):

```bash
make dist
```

---

## 🛠️ Usage

```bash
# Shell Completion Setup
source <(myvault completion zsh)        # Enable zsh autocomplete in current session
myvault completion zsh > "${fpath[1]}/_myvault" # Permanent zsh autocomplete setup

# Reading Secrets
myvault list                           # List all secret resources (Zero-passphrase prompt!)
myvault list --tag work                # List resources matching tag
myvault get mysql                      # Show resource fields (concealed fields redacted)
myvault get mysql --show               # Show all fields including concealed values
myvault get mysql host                 # Output raw value of a single field

# Searching & Auditing
myvault search prod                    # Search across resource names, tags, and fields
myvault audit                          # Audit vault for weak, reused, or stale passwords

# Clipboard
myvault copy mysql                     # Copy default password/secret to clipboard (30s auto-clear)
myvault copy mysql host                # Copy specific field to clipboard

# Adding, Creating & Editing (Supports both 'key=value' and 'key value')
myvault create db1 host 10.0.0.1 user admin  # Create new resource (fails if exists)
myvault set db1 port 5432                    # Set/update fields on existing resource
myvault add mysql host=db.com port=3306      # Add/update fields
myvault rekey                          # Change master vault passphrase

# Backups & Restoration
myvault backup                         # Create timestamped encrypted snapshot in backup/
myvault restore backup/vault-20261007.bak # Restore vault from snapshot

# Imports & Exports
myvault export backup.csv --format csv # Export vault to CSV
myvault import bitwarden.csv --format bitwarden # Import from Bitwarden export

# Deleting
myvault delete mysql port              # Delete a single field
myvault delete mysql                   # Delete an entire resource
```

---

## 🔒 Security Policy

For detailed security boundaries, memory hygiene, and cryptographic design, see [SECURITY.md](SECURITY.md).

---

## 📄 License

MIT
