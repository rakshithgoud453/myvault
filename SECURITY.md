# Security Policy & Threat Model — myvault

`myvault` is a lightweight, CLI-first password manager built in Go, designed with Unix security principles and robust cryptography.

---

## 1. Cryptographic Design

- **Vault Format**: Encrypted using **[age](https://filippo.io/age)** format (`age1...` standard).
- **Key Derivation (KDF)**: Scrypt (configured via `age` default parameters for passphrase protection).
- **Authenticated Encryption**: ChaCha20-Poly1305 AEAD.
- **Randomness**: Cryptographically secure random strings generated via `crypto/rand` for passwords.

---

## 2. Threat Model & Security Boundaries

### Assumed Assets
- Vault storage file (`~/.password-vault/vault.age`).
- Cached session passphrase during unlock windows.
- In-memory decrypted vault contents during command execution.
- Password strings copied to system clipboard.

### In-Scope Threat Actors & Vectors
1. **Stolen Vault File / Cold Disk Theft**: Attacker gains access to `vault.age` via disk theft, cloud backup leak, or secondary filesystem read.
   - *Mitigation*: Scrypt KDF enforces high CPU/RAM computational cost; vault data remains encrypted at rest via ChaCha20-Poly1305.
2. **Session Persistence Leakage**: Attacker reads `.session` file from disk.
   - *Mitigation*: The session file contains **zero secrets or passphrases**. It stores only expiry timestamps (`created_at`, `expires_at`). The actual passphrase is exclusively held in the system Keychain (`security` service on macOS) protected by OS user session controls.
3. **Power Loss / Mid-Write Vault Corruption**: Crash or power outage during `add` or `delete`.
   - *Mitigation*: Atomic writes via temporary files and automatic pre-write backups (`vault.age.bak`).
4. **Memory Scraping / Core Dumps**: Inspection of process memory or swap space.
   - *Mitigation*: Sensitive byte buffers (passphrases, raw decrypted INI payloads) are explicitly zero-filled (`memguard.ZeroBytes`) immediately after parsing.
5. **Persistent Clipboard Exposure**: Sensitive passwords left in system clipboard.
   - *Mitigation*: Automatic clipboard clearance after 30 seconds via detached background daemon process.

### Out-of-Scope Threats
- **Root / Kernel Level Malware**: Malware with administrative (`root`/`sudo`) privileges or kernel access can inspect arbitrary process memory and bypass Keychain prompts.
- **Hardware Keyloggers / Screen Recorders**: Physical or malware-based input monitoring before CLI input reaches terminal input stream.

---

## 3. Implemented Security Controls

| Domain | Control | Description |
|---|---|---|
| **Storage at Rest** | Age Encryption | Passphrase-encrypted payload using scrypt KDF and ChaCha20-Poly1305 AEAD |
| **Session Model** | OS Keychain Integration | Passphrases stored in macOS Keychain; `.session` file stores only non-sensitive expiration metadata |
| **Integrity** | Atomic Write & Backup | Pre-write `.bak` creation + temporary file write + atomic `os.Rename` |
| **Memory Hygiene** | Zeroization | Explicit byte wiping (`memguard.ZeroBytes`) for plaintext buffers |
| **Permissions** | POSIX File Security | Directories enforced to `0700` (`rwx------`), files to `0600` (`rw-------`) |
| **Clipboard** | Auto-Wipe Daemon | Detached timer process clears clipboard after 30 seconds |

---

## 4. Security Reporting

If you discover a vulnerability or security issue in `myvault`, please open a private security advisory or report it directly to the maintainers. Do not post unpatched vulnerabilities in public issues.
