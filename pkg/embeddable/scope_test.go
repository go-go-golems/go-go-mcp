package embeddable

import (
	"reflect"
	"testing"
)

func TestScopeSetNormalizesAndCopies(t *testing.T) {
	set, err := ParseScopeSet([]string{"scope:b", " scope:a ", "scope:b"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := set.Strings(), []string{"scope:a", "scope:b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes = %v, want %v", got, want)
	}
	values := set.Values()
	values[0] = "tampered"
	if set.Contains("tampered") {
		t.Fatal("scope set exposed mutable storage")
	}
}

func TestScopeSetRejectsWhitespaceAndEmptyValues(t *testing.T) {
	for _, values := range [][]string{{""}, {"scope a"}, {"scope\ta"}, {"scope\"a"}} {
		if _, err := ParseScopeSet(values); err == nil {
			t.Fatalf("ParseScopeSet(%q) unexpectedly succeeded", values)
		}
	}
}

func TestScopeSetZeroValueFailsClosed(t *testing.T) {
	var set ScopeSet
	if !set.Empty() || set.Contains("scope:a") {
		t.Fatal("zero scope set granted authority")
	}
}
