// Package vault defines the data model for secret vault entries and provides
// parsing/serialization for the flexible key-value vault format.
//
// v0.2 supports:
//   - Flexible Key-Value fields per secret resource (e.g. host, port, username, password)
//   - Field concealment flag (concealed fields like passwords/tokens are redacted by default)
//   - Full backward compatibility with INI-style (v0.1) and legacy JSON (v0.2 draft)
package vault

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rakshithgoud453/myvault/internal/storage"
)

// Field represents a single key-value field inside a secret resource.
type Field struct {
	Value     string `json:"value"`
	Concealed bool   `json:"concealed"` // If true, redacted by default in 'myvault get'
}

// Entry represents a secret resource in the vault with flexible key-value fields.
type Entry struct {
	Fields     map[string]Field `json:"fields"`
	Tags       []string         `json:"tags,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	ModifiedAt time.Time        `json:"modified_at"`
}

// SearchMatch describes a search match result.
type SearchMatch struct {
	ResourceName string
	MatchedOn    string // "name", "tag", "field:<key>", "value:<key>"
	Snippet      string
}

// Vault holds the full collection of secret entries.
type Vault struct {
	Version string           `json:"version"`
	Entries map[string]Entry `json:"entries"`
}

// New creates a new empty vault.
func New() *Vault {
	return &Vault{
		Version: "0.2",
		Entries: make(map[string]Entry),
	}
}

// IsConcealedKey determines whether a field key name should default to concealed.
func IsConcealedKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "pass") ||
		strings.Contains(k, "secret") ||
		strings.Contains(k, "token") ||
		strings.Contains(k, "key") ||
		strings.Contains(k, "pin") ||
		strings.Contains(k, "code")
}

// BuildIndex generates the unencrypted VaultIndex metadata struct for fast listing/search.
func (v *Vault) BuildIndex() *storage.VaultIndex {
	resources := v.List()
	entries := make([]storage.ResourceIndexEntry, 0, len(resources))
	for _, resName := range resources {
		entry := v.Entries[resName]
		entries = append(entries, storage.ResourceIndexEntry{
			Name: resName,
			Tags: entry.Tags,
		})
	}
	return &storage.VaultIndex{Resources: entries}
}

// List returns all entry resource names sorted alphabetically.
func (v *Vault) List() []string {
	names := make([]string, 0, len(v.Entries))
	for name := range v.Entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListByTag returns entry resource names that contain a specific tag (case-insensitive).
func (v *Vault) ListByTag(tag string) []string {
	target := strings.ToLower(strings.TrimPrefix(tag, "#"))
	var names []string
	for name, entry := range v.Entries {
		for _, t := range entry.Tags {
			if strings.ToLower(strings.TrimPrefix(t, "#")) == target {
				names = append(names, name)
				break
			}
		}
	}
	sort.Strings(names)
	return names
}

// SetTags assigns tags to a resource.
func (v *Vault) SetTags(resourceName string, tags []string) error {
	entry, canonicalResource, err := v.Get(resourceName)
	if err != nil {
		return err
	}

	cleanedTags := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			if !strings.HasPrefix(t, "#") {
				t = "#" + t
			}
			cleanedTags = append(cleanedTags, t)
		}
	}

	entry.Tags = cleanedTags
	entry.ModifiedAt = time.Now()
	v.Entries[canonicalResource] = entry
	return nil
}

// Search queries across resource names, field keys, values, and tags (case-insensitive).
func (v *Vault) Search(query string) []SearchMatch {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}

	var matches []SearchMatch
	resources := v.List()

	for _, name := range resources {
		entry := v.Entries[name]

		// 1. Match resource name
		if strings.Contains(strings.ToLower(name), q) {
			matches = append(matches, SearchMatch{
				ResourceName: name,
				MatchedOn:    "name",
				Snippet:      name,
			})
			continue
		}

		// 2. Match tags
		tagMatched := false
		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), q) {
				matches = append(matches, SearchMatch{
					ResourceName: name,
					MatchedOn:    "tag",
					Snippet:      tag,
				})
				tagMatched = true
				break
			}
		}
		if tagMatched {
			continue
		}

		// 3. Match field keys & values
		for fKey, fVal := range entry.Fields {
			if strings.Contains(strings.ToLower(fKey), q) {
				matches = append(matches, SearchMatch{
					ResourceName: name,
					MatchedOn:    "field:" + fKey,
					Snippet:      fKey,
				})
				break
			}
			if !fVal.Concealed && strings.Contains(strings.ToLower(fVal.Value), q) {
				matches = append(matches, SearchMatch{
					ResourceName: name,
					MatchedOn:    "value:" + fKey,
					Snippet:      fVal.Value,
				})
				break
			}
		}
	}

	return matches
}

// Get retrieves an entry by resource name (case-insensitive).
// Returns the entry and the canonical name, or an error if not found.
func (v *Vault) Get(name string) (Entry, string, error) {
	upper := strings.ToUpper(name)
	for key, entry := range v.Entries {
		if strings.ToUpper(key) == upper {
			return entry, key, nil
		}
	}
	return Entry{}, "", fmt.Errorf("resource %q not found", name)
}

// GetField retrieves a specific field from an entry (case-insensitive for both).
// Returns the field, canonical resource name, canonical field name, or an error.
func (v *Vault) GetField(resourceName, fieldName string) (Field, string, string, error) {
	entry, canonicalResource, err := v.Get(resourceName)
	if err != nil {
		return Field{}, "", "", err
	}

	upperField := strings.ToUpper(fieldName)
	for fKey, fVal := range entry.Fields {
		if strings.ToUpper(fKey) == upperField {
			return fVal, canonicalResource, fKey, nil
		}
	}

	return Field{}, canonicalResource, "", fmt.Errorf("field %q not found in resource %q", fieldName, canonicalResource)
}

// DefaultField finds the primary secret field for an entry (e.g. password, secret, token, or first concealed field).
// Returns the field, field name, and true if found.
func (e Entry) DefaultField() (Field, string, bool) {
	if len(e.Fields) == 0 {
		return Field{}, "", false
	}

	// Priority 1: Field named "password", "secret", "token", "api_key"
	for _, priority := range []string{"password", "secret", "token", "api_key", "pin"} {
		for fKey, fVal := range e.Fields {
			if strings.EqualFold(fKey, priority) {
				return fVal, fKey, true
			}
		}
	}

	// Priority 2: Any concealed field
	for fKey, fVal := range e.Fields {
		if fVal.Concealed {
			return fVal, fKey, true
		}
	}

	// Priority 3: First field alphabetically
	fieldNames := make([]string, 0, len(e.Fields))
	for fKey := range e.Fields {
		fieldNames = append(fieldNames, fKey)
	}
	sort.Strings(fieldNames)
	firstKey := fieldNames[0]
	return e.Fields[firstKey], firstKey, true
}

// SetField creates or updates a field on a resource.
// Returns true if the field was newly created or its value/concealment changed.
func (v *Vault) SetField(resourceName, fieldName, fieldValue string, concealed bool) bool {
	upperResource := strings.ToUpper(resourceName)
	now := time.Now()

	entry, canonicalResource, err := v.Get(resourceName)
	if err != nil {
		// Create new resource
		entry = Entry{
			Fields:     make(map[string]Field),
			CreatedAt:  now,
			ModifiedAt: now,
		}
		canonicalResource = upperResource
	}

	if entry.Fields == nil {
		entry.Fields = make(map[string]Field)
	}

	// Case-insensitive update of existing key or preserve case of new key
	targetKey := fieldName
	for fKey := range entry.Fields {
		if strings.EqualFold(fKey, fieldName) {
			targetKey = fKey
			break
		}
	}

	existing, found := entry.Fields[targetKey]
	if found && existing.Value == fieldValue && existing.Concealed == concealed {
		return false // No change
	}

	entry.Fields[targetKey] = Field{
		Value:     fieldValue,
		Concealed: concealed,
	}
	entry.ModifiedAt = now

	v.Entries[canonicalResource] = entry
	return true
}

// Delete removes an entire resource by name (case-insensitive).
func (v *Vault) Delete(name string) error {
	_, canonicalResource, err := v.Get(name)
	if err != nil {
		return err
	}
	delete(v.Entries, canonicalResource)
	return nil
}

// DeleteField removes a specific field from a resource (case-insensitive).
func (v *Vault) DeleteField(resourceName, fieldName string) error {
	entry, canonicalResource, err := v.Get(resourceName)
	if err != nil {
		return err
	}

	upperField := strings.ToUpper(fieldName)
	deleted := false
	for fKey := range entry.Fields {
		if strings.ToUpper(fKey) == upperField {
			delete(entry.Fields, fKey)
			deleted = true
			break
		}
	}

	if !deleted {
		return fmt.Errorf("field %q not found in resource %q", fieldName, canonicalResource)
	}

	entry.ModifiedAt = time.Now()
	v.Entries[canonicalResource] = entry
	return nil
}

// Parse decrypted plaintext into a Vault.
// Automatically detects whether the data is JSON (v0.2) or INI-style (v0.1).
func Parse(data []byte) (*Vault, error) {
	text := strings.TrimSpace(string(data))
	if len(text) == 0 {
		return New(), nil
	}

	// Try JSON first (v0.2 format).
	if text[0] == '{' {
		return parseJSON(data)
	}

	// Fall back to INI-style (v0.1 format).
	return parseINI(text)
}

// Serialize converts the vault to JSON bytes for encryption.
func (v *Vault) Serialize() ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serializing vault: %w", err)
	}
	return append(data, '\n'), nil
}

// parseJSON parses the v0.2 JSON format, with fallback for old username/password JSON.
func parseJSON(data []byte) (*Vault, error) {
	var raw struct {
		Version string                     `json:"version"`
		Entries map[string]json.RawMessage `json:"entries"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing JSON vault: %w", err)
	}

	v := New()
	if raw.Version != "" {
		v.Version = raw.Version
	}

	for name, rawEntry := range raw.Entries {
		var entry Entry
		if err := json.Unmarshal(rawEntry, &entry); err == nil && entry.Fields != nil {
			v.Entries[name] = entry
			continue
		}

		// Fallback for draft v0.2 JSON containing legacy username/password struct
		var legacy struct {
			Username   string    `json:"username"`
			Password   string    `json:"password"`
			CreatedAt  time.Time `json:"created_at"`
			ModifiedAt time.Time `json:"modified_at"`
		}
		if err := json.Unmarshal(rawEntry, &legacy); err == nil {
			e := Entry{
				Fields:     make(map[string]Field),
				CreatedAt:  legacy.CreatedAt,
				ModifiedAt: legacy.ModifiedAt,
			}
			if legacy.Username != "" {
				e.Fields["username"] = Field{Value: legacy.Username, Concealed: false}
			}
			if legacy.Password != "" {
				e.Fields["password"] = Field{Value: legacy.Password, Concealed: true}
			}
			v.Entries[name] = e
		}
	}

	return v, nil
}

// parseINI parses the v0.1 INI-style format into key-value fields:
//
//	[NAME]
//	username=value
//	password=value
func parseINI(text string) (*Vault, error) {
	v := New()

	var currentName string
	var currentFields map[string]Field

	saveCurrent := func() {
		if currentName != "" && len(currentFields) > 0 {
			now := time.Now()
			v.Entries[currentName] = Entry{
				Fields:     currentFields,
				CreatedAt:  now,
				ModifiedAt: now,
			}
		}
	}

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			saveCurrent()
			currentName = ""
			currentFields = nil
			continue
		}

		// Section header: [NAME]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			saveCurrent()
			currentName = strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			currentFields = make(map[string]Field)
			continue
		}

		// Key=value pair.
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if currentFields != nil {
			concealed := IsConcealedKey(key)
			currentFields[key] = Field{
				Value:     value,
				Concealed: concealed,
			}
		}
	}

	saveCurrent()
	return v, nil
}
