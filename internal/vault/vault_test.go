package vault_test

import (
	"testing"

	"github.com/rakshithgoud453/myvault/internal/vault"
)

func TestParseINIAndSerialize(t *testing.T) {
	raw := `[GITHUB]
username = alice@example.com
password = supersecret

[MYSQL]
host = db.internal
port = 3306
password = dbpass
`

	v, err := vault.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(v.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(v.Entries))
	}

	// Test GITHUB entry fields
	githubUser, _, _, err := v.GetField("github", "username")
	if err != nil || githubUser.Value != "alice@example.com" {
		t.Errorf("unexpected username for GITHUB: %+v, err: %v", githubUser, err)
	}

	githubPass, _, _, err := v.GetField("github", "password")
	if err != nil || githubPass.Value != "supersecret" || !githubPass.Concealed {
		t.Errorf("unexpected password for GITHUB: %+v, err: %v", githubPass, err)
	}

	// Test MYSQL entry fields
	mysqlHost, _, _, err := v.GetField("mysql", "host")
	if err != nil || mysqlHost.Value != "db.internal" || mysqlHost.Concealed {
		t.Errorf("unexpected host for MYSQL: %+v, err: %v", mysqlHost, err)
	}

	// Test roundtrip serialization to JSON and back
	serialized, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	v2, err := vault.Parse(serialized)
	if err != nil {
		t.Fatalf("Parse of serialized vault failed: %v", err)
	}

	if len(v2.Entries) != 2 {
		t.Fatalf("expected 2 entries in re-parsed vault, got %d", len(v2.Entries))
	}
}

func TestSetFieldAndGetField(t *testing.T) {
	v := vault.New()

	v.SetField("mysql", "host", "db.example.com", false)
	v.SetField("mysql", "password", "secret123", true)

	if len(v.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(v.Entries))
	}

	host, _, _, err := v.GetField("MYSQL", "HOST")
	if err != nil || host.Value != "db.example.com" {
		t.Errorf("failed to get HOST field: %+v, err: %v", host, err)
	}

	pass, _, _, err := v.GetField("mysql", "password")
	if err != nil || pass.Value != "secret123" || !pass.Concealed {
		t.Errorf("failed to get password field: %+v, err: %v", pass, err)
	}
}

func TestDeleteResourceAndField(t *testing.T) {
	v := vault.New()
	v.SetField("mysql", "host", "db.example.com", false)
	v.SetField("mysql", "port", "3306", false)

	// Delete single field
	if err := v.DeleteField("mysql", "port"); err != nil {
		t.Fatalf("DeleteField failed: %v", err)
	}

	if _, _, _, err := v.GetField("mysql", "port"); err == nil {
		t.Errorf("port field still exists after DeleteField")
	}

	// Host should still exist
	if _, _, _, err := v.GetField("mysql", "host"); err != nil {
		t.Errorf("host field was unexpectedly removed")
	}

	// Delete entire resource
	if err := v.Delete("MYSQL"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, _, err := v.Get("mysql"); err == nil {
		t.Errorf("resource mysql still exists after Delete")
	}
}

func TestINIMigrationAndAddNewResource(t *testing.T) {
	iniData := `[GITHUB]
username = alice@example.com
password = supersecret

[GREYTHR]
username = greyuser
password = greypass
`

	// 1. Parse legacy INI vault
	v, err := vault.Parse([]byte(iniData))
	if err != nil {
		t.Fatalf("Parse INI failed: %v", err)
	}

	// 2. Add new resource "MYSQL"
	v.SetField("mysql", "username", "test", false)
	v.SetField("mysql", "password", "pass123", true)

	// 3. Serialize to JSON
	jsonBytes, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// 4. Re-parse JSON vault
	v2, err := vault.Parse(jsonBytes)
	if err != nil {
		t.Fatalf("Parse JSON failed: %v", err)
	}

	names := v2.List()
	expected := []string{"GITHUB", "GREYTHR", "MYSQL"}
	if len(names) != len(expected) {
		t.Fatalf("expected entries %v, got %v", expected, names)
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expected[i], name)
		}
	}
}

func TestTagsAndSearch(t *testing.T) {
	v := vault.New()
	v.SetField("mysql_prod", "host", "db.prod.internal", false)
	_ = v.SetTags("mysql_prod", []string{"work", "db"})

	v.SetField("personal_blog", "url", "https://myblog.com", false)
	_ = v.SetTags("personal_blog", []string{"personal"})

	// Test Tag Filtering
	workItems := v.ListByTag("work")
	if len(workItems) != 1 || workItems[0] != "MYSQL_PROD" {
		t.Errorf("expected MYSQL_PROD for tag work, got: %v", workItems)
	}

	// Test Search
	matches := v.Search("prod")
	if len(matches) != 1 || matches[0].ResourceName != "MYSQL_PROD" {
		t.Errorf("search for 'prod' failed: %+v", matches)
	}

	matchesTag := v.Search("personal")
	if len(matchesTag) != 1 || matchesTag[0].ResourceName != "PERSONAL_BLOG" {
		t.Errorf("search for 'personal' failed: %+v", matchesTag)
	}
}
