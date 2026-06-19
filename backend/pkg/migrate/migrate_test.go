package migrate

import "testing"

func TestVersionFromPath(t *testing.T) {
	if got := versionFromPath("/app/migrations/016_rag_enhancements.up.sql"); got != "016_rag_enhancements" {
		t.Fatalf("got %q", got)
	}
}

func TestMigrationNumber(t *testing.T) {
	cases := map[string]int{
		"001_initial_schema": 1,
		"014_enterprise_kb":    14,
		"015_multilingual":     15,
		"100_future":           100,
	}
	for v, want := range cases {
		if got := migrationNumber(v); got != want {
			t.Fatalf("%s: got %d want %d", v, got, want)
		}
	}
}

func TestLegacyCutoff(t *testing.T) {
	if migrationNumber("014_enterprise_kb") > legacyInitCutoff {
		t.Fatal("014 should be within legacy cutoff")
	}
	if migrationNumber("015_multilingual_search") <= legacyInitCutoff {
		t.Fatal("015 should be after legacy cutoff")
	}
}
