package token

import (
	"testing"
)

// Test looking up values succeeds, then fails
func TestLookup(t *testing.T) {

	for key, val := range known {

		// Obviously this will pass.
		if LookupIdentifier(key) != val {
			t.Errorf("Lookup of %s failed", key)
		}

		// Once the keywords are "doubled" they'll no longer
		// match - so we find them as identifiers.
		if LookupIdentifier(key+key) != IDENTIFIER {
			t.Errorf("Lookup of %s failed", key)
		}
	}
}
