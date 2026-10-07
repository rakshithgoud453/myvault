# Security Architecture & Roadmap

This document outlines the security model, cryptographic design principles, and implemented features for `myvault`.

---

## 1. Core Principles

1. **Zero-Knowledge Architecture**: `myvault` stores zero master secrets or passphrase hashes in software code or binaries. The master key is derived strictly at runtime from the user's passphrase.
2. **Passphrase-Derived Encryption**: All vault data on disk (`~/.password-vault/vault.age`) and snapshot backups (`vault-*.bak`) are encrypted using the `age` format (Scrypt key derivation + AEAD encryption).
3. **Stateless Software**: The `myvault` binary is public and stateless. A fresh installation carries no state or key material.

---

## 2. Onboarding & Fresh Installation Architecture

### A. Initialization Modes
When `myvault` is installed on a fresh machine:
* **`myvault init`**: Explicitly initializes an empty vault and sets the master passphrase.
* **Auto-Initialize on `myvault create` / `myvault import`**: If no vault exists at `~/.password-vault/vault.age`, `myvault` detects the missing vault and prompts:
  ```text
  No vault found. Set a master passphrase for your new vault:
  Confirm passphrase:
  ```
* **Snapshot Restoration (`myvault restore <snapshot>`)**: Prompts for the passphrase of the snapshot, decrypts it, and sets up `~/.password-vault/vault.age` without requiring prior initialization.

---

## 3. Encrypted Export & Import Specification

### A. Exporting (`myvault export`)
* **Default Unencrypted Export**: `myvault export secrets.json` exports unencrypted JSON or CSV data.
* **Encrypted Export (`--encrypt` / `-e`)**:
  * By default, encrypts using the user's active **main vault passphrase**.
  * If `-p` / `--passphrase` is passed, prompts for a **custom export passphrase**, allowing safe sharing/transfer without revealing the main vault passphrase.

### B. Importing (`myvault import`)
* Automatically detects whether an import file is raw JSON/CSV or `age`-encrypted ciphertext.
* If encrypted, prompts for the import file passphrase (`Enter passphrase for imported file:`).
* Decrypts in memory and safely merges entries into the local vault.

---

## 4. Advanced Security Options (Roadmap)

### Option 1: Dual-Factor Backups (Passphrase + Machine Keyfile)
* Require both a passphrase **and** a local 256-bit keyfile (`myvault.keyfile`) to unlock/restore vault snapshots.

### Option 2: Hardware & Secure Enclave Binding
* Integrate with macOS Keychain / Secure Enclave (or Linux Secret Service / Windows TPM) to bind key release to local hardware authentication.
