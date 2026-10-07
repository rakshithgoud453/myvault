// Package policy defines the security and authentication boundaries for CLI operations.
//
// Operations can be configured to require full passphrase authentication/decryption
// (e.g. get, copy, set, delete, rekey, export) or read unencrypted index metadata
// without prompting (e.g. list, search).
package policy

type Operation string

const (
	OpList   Operation = "list"
	OpSearch Operation = "search"
	OpGet    Operation = "get"
	OpCopy   Operation = "copy"
	OpAdd    Operation = "add"
	OpSet    Operation = "set"
	OpDelete Operation = "delete"
	OpCreate Operation = "create"
	OpRekey  Operation = "rekey"
	OpExport Operation = "export"
	OpImport Operation = "import"
)

// DefaultPolicy specifies authentication requirements for all CLI operations.
var DefaultPolicy = map[Operation]bool{
	OpList:   false, // Listing resource names does NOT require passphrase
	OpSearch: false, // Index search does NOT require passphrase
	OpGet:    true,  // Retrieving secret fields REQUIRES passphrase
	OpCopy:   true,  // Copying secrets REQUIRES passphrase
	OpAdd:    true,  // Adding secrets REQUIRES passphrase
	OpSet:    true,  // Modifying secrets REQUIRES passphrase
	OpCreate: true,  // Creating secrets REQUIRES passphrase
	OpDelete: true,  // Deleting secrets REQUIRES passphrase
	OpRekey:  true,  // Rekeying REQUIRES passphrase
	OpExport: true,  // Exporting secret payload REQUIRES passphrase
	OpImport: true,  // Importing REQUIRES passphrase
}

// RequiresPassphrase returns true if the operation requires vault decryption.
func RequiresPassphrase(op Operation) bool {
	req, ok := DefaultPolicy[op]
	if !ok {
		return true // Fail secure: default to requiring passphrase
	}
	return req
}
