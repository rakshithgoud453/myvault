# MyVault

## 1. Project Goal

Build a personal password manager that is:

- Local-first
- No Apple Passwords
- No cloud dependency
- No third-party password manager
- Open to being distributed to friends
- Eventually distributable/sellable as real software
- Cross-platform eventually
- Secure by design rather than relying on custom encryption

Core philosophy:

> Passwords belong to the user. The software should not require a cloud account or external password-storage service.

---

# 2. Current Prototype — v0.1

The first prototype was built directly on a Mac using Bash and the `age` encryption tool.

Current machine:

```text
MacBook Air
macOS
Apple Silicon
```

Current shell:

```text
zsh
```

---

# 3. Current Storage

Password vault directory:

```text
~/.password-vault/
```

Main encrypted vault:

```text
~/.password-vault/vault.age
```

Backup:

```text
~/.password-vault/backup/vault.age
```

Current permissions:

```text
~/.password-vault       → 700
vault.age               → 600
backup/                 → 700
backup/vault.age        → 600
```

The vault itself is encrypted using `age`.

There should NOT be a permanent plaintext password file.

---

# 4. Current CLI

The executable is:

```text
~/.local/bin/vault
```

PATH configuration was added to:

```text
~/.zshrc
```

with:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

The command is available as:

```bash
vault
```

Current version:

```text
vault 0.1.0
```

---

# 5. Current Commands

Implemented:

```bash
vault list
```

Lists entries.

Example:

```text
GITHUB
RASPBERRY_PI
JUMP_SERVER
GREYTHR
test
```

---

```bash
vault get <name>
```

Retrieves username and password.

Example:

```bash
vault get github
```

Output:

```text
Username: ...
Password: ...
```

Entry names are case-insensitive.

---

```bash
vault copy <name>
```

Copies only the password to the macOS clipboard.

Example:

```bash
vault copy github
```

Current behavior:

```text
Password copied to clipboard.
Clipboard will be cleared in 30 seconds.
```

The clipboard is automatically cleared after 30 seconds.

---

```bash
vault add <name>
```

Adds a credential.

Example:

```bash
vault add github
```

It asks for:

```text
Username:
Password:
```

Password input is hidden.

---

```bash
vault add <name> --generate
```

Creates an entry with a generated password.

Example:

```bash
vault add aws --generate
```

The generated password is copied to the clipboard and cleared after 30 seconds.

---

```bash
vault generate
```

Generates a random password.

Current implementation uses:

```text
openssl rand
```

---

```bash
vault delete <name>
```

Deletes an entry after confirmation.

---

```bash
vault --version
```

Current output:

```text
vault 0.1.0
```

---

# 6. Current Vault Format

The encrypted vault contains plaintext structured roughly like:

```text
[GITHUB]
username=...
password=...

[RASPBERRY_PI]
username=...
password=...

[JUMP_SERVER]
username=...
password=...

[GREYTHR]
username=...
password=...
```

This plaintext exists only when the vault is decrypted for an operation.

The stored file is:

```text
vault.age
```

and is encrypted.

---

# 7. Current Encryption Architecture

Current prototype:

```text
             vault.age
                │
                │ age decrypt
                ▼
        temporary plaintext
                │
                ▼
          CLI operation
                │
                ▼
        temporary plaintext
                │
                │ age encrypt
                ▼
             vault.age
```

The current implementation uses:

```bash
age -p
```

when re-encrypting the vault.

Therefore modifying the vault currently asks for the vault encryption passphrase again.

This is known technical debt.

---

# 8. Important Security Event

During development, one Raspberry Pi password was accidentally displayed in the chat.

That password must be considered compromised.

Action required:

```text
Rotate Raspberry Pi password.
```

Do not reuse that password.

---

# 9. Current Backup

Before freezing v0.1, the encrypted vault was backed up:

```text
~/.password-vault/backup/vault.age
```

The backup was verified using:

```bash
cmp ~/.password-vault/vault.age \
    ~/.password-vault/backup/vault.age
```

No output was produced, meaning the files were identical.

v0.1 is therefore considered frozen.

---

# 10. v0.1 Status

Current status:

```text
v0.1 — FROZEN
```

Do not redesign the current Bash implementation while migrating passwords.

The purpose of v0.1 is to be a working proof of concept.

---

# 11. Why We Are Moving to v0.2

The Bash prototype proves the core idea.

However, it should NOT be the final product.

Problems/limitations include:

1. Shell script implementation
2. External `age` executable dependency
3. External OpenSSL dependency
4. Re-entering the vault passphrase during modifications
5. Temporary plaintext handling
6. No proper locked/unlocked session model
7. No robust cross-platform implementation
8. No formal testing
9. No proper packaging
10. No GUI
11. No signed/notarized distribution
12. No formal security threat model

Therefore v0.2 will be a real application.

---

# 12. v0.2 Goal

Rewrite MyVault as a compiled application.

Preferred implementation language:

```text
Go
```

Reason:

- Simple static binaries
- Excellent CLI support
- Easy cross-compilation
- Small deployment footprint
- Good fit for security-oriented system software
- Can embed encryption functionality
- Easier distribution than a shell/Python application

---

# 13. v0.2 Architecture

Target:

```text
                  MYVAULT
                     │
        ┌────────────┼────────────┐
        │            │            │
       CLI         Crypto       Storage
        │            │            │
        │            │            │
        └────────────┼────────────┘
                     │
                vault.age
```

The final application should NOT require the user to separately install:

```text
age
openssl
Homebrew
Python
Node
Java
```

The goal is:

```text
myvault
```

as a self-contained executable.

---

# 14. Encryption Strategy

Do NOT implement custom cryptography.

Use a well-tested cryptographic implementation.

The `age` format is preferred because the existing v0.1 vault already uses it.

Goal:

```text
v0.1 vault.age
       │
       ▼
v0.2 MyVault
       │
       ▼
same vault.age
```

Existing user passwords should survive the migration.

We should investigate embedding/using the age Go library rather than invoking the external `age` executable.

---

# 15. v0.2 Session Model

Current:

```text
vault add
    ↓
ask vault password
    ↓
decrypt
    ↓
modify
    ↓
ask vault password again
    ↓
encrypt
```

Target:

```text
vault unlock
       │
       ▼
master password entered once
       │
       ▼
secure unlocked session
       │
       ├── vault list
       ├── vault get github
       ├── vault copy github
       ├── vault add aws
       └── vault delete test
       │
       ▼
vault lock
```

Eventually:

```bash
myvault unlock
myvault list
myvault copy github
myvault lock
```

The exact secure session/key-storage mechanism must be designed carefully before implementation.

---

# 16. Target CLI

Eventually:

```bash
myvault --version

myvault init

myvault unlock

myvault lock

myvault status

myvault list

myvault get github

myvault copy github

myvault add github

myvault add github --generate

myvault generate

myvault delete github

myvault backup

myvault restore
```

Possible future commands:

```bash
myvault edit
myvault search
myvault import
myvault export
```

---

# 17. Product Direction

The eventual product should be more than a Bash script.

Target:

```text
                 MyVault
                    │
       ┌────────────┼────────────┐
       │            │            │
    Security      CLI/API     Storage
       │
       ├── encryption
       ├── locking
       ├── password generation
       ├── clipboard protection
       └── secure deletion
```

Potential future interfaces:

```text
CLI
 ↓
Desktop application
 ↓
Cross-platform application
```

But CLI comes first.

---

# 18. Distribution Journey

## Stage 1 — Current

Friend distribution:

```text
MyVault binary
```

No cloud required.

---

## Stage 2

Package:

```text
myvault-macos-arm64.tar.gz
```

or:

```text
MyVault.zip
```

---

## Stage 3

Cross-platform builds:

```text
macOS ARM64
macOS Intel
Linux ARM64
Linux AMD64
Windows AMD64
```

---

## Stage 4

Professional macOS distribution:

```text
MyVault.app
       ↓
Developer ID signing
       ↓
Apple notarization
       ↓
DMG
```

---

## Stage 5

Commercial product.

Potential positioning:

> A local-first password manager where your encrypted vault belongs to you.

No mandatory cloud account.

No mandatory cloud synchronization.

No vendor-controlled password storage.

---

# 19. Development Workspace

Create a new project/workspace called:

```text
MyVault
```

Recommended project structure:

```text
MyVault/
│
├── README.md
├── LICENSE
├── go.mod
│
├── cmd/
│   └── myvault/
│       └── main.go
│
├── internal/
│   ├── vault/
│   ├── crypto/
│   ├── clipboard/
│   ├── password/
│   ├── storage/
│   └── session/
│
├── tests/
│
├── docs/
│   ├── architecture.md
│   ├── security.md
│   └── threat-model.md
│
└── scripts/
```

The existing prototype should be preserved separately as:

```text
prototype-v0.1/
```

Do NOT overwrite it.

---

# 20. Development Milestones

## Milestone 1

Create Go project.

```text
MyVault
└── hello-world compiled binary
```

---

## Milestone 2

Read existing:

```text
vault.age
```

---

## Milestone 3

Implement:

```bash
myvault list
myvault get
```

---

## Milestone 4

Implement:

```bash
myvault copy
```

with clipboard timeout.

---

## Milestone 5

Implement:

```bash
myvault add
myvault delete
myvault generate
```

---

## Milestone 6

Implement proper:

```bash
myvault unlock
myvault lock
```

session architecture.

---

## Milestone 7

Security review.

Threat model:

```text
Who can attack the vault?
What happens if someone gets vault.age?
What happens if someone gets the Mac?
What happens if clipboard is inspected?
What happens if the process crashes?
What plaintext exists on disk?
What plaintext exists in memory?
How are backups protected?
```

---

## Milestone 8

Automated tests.

---

## Milestone 9

Cross-platform builds.

---

## Milestone 10

Package and distribute to friends.

---

# 21. Important Design Principle

Never invent cryptography.

MyVault owns:

```text
UX
CLI
vault format
session management
password generation
clipboard management
backup management
distribution
```

A proven cryptographic library owns:

```text
encryption
decryption
key derivation
cryptographic primitives
```

---

# 22. Current State

We are here:

```text
                    MyVault
                       │
                       ▼
              ┌─────────────────┐
              │ Bash Prototype   │
              │     v0.1.0       │
              └────────┬────────┘
                       │
             ┌─────────┴─────────┐
             ▼                   ▼
        vault.age             CLI
        encrypted             commands
             │
             ▼
        BACKUP VERIFIED
```

Next:

```text
             v0.1.0
                │
                ▼
        ┌────────────────┐
        │ Go rewrite     │
        │     v0.2       │
        └───────┬────────┘
                │
                ▼
       Embedded encryption
                │
                ▼
        unlock / lock
                │
                ▼
       secure CLI product
                │
                ▼
       cross-platform build
                │
                ▼
        distributable app
                │
                ▼
             MyVault
```

## Immediate next task

Do NOT modify the v0.1 Bash vault.

Start a new Go project and build:

```bash
myvault --version
```

as the first v0.2 milestone.

Then make v0.2 capable of reading the existing v0.1 `vault.age` without migrating or destroying the original vault.