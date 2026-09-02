package embeddable

import (
	"reflect"
	"testing"
)

func TestScopeSetNormalizesAndCopies(t *testing.T) {
	set, err := ParseScopeSet([]string{"scope:b", "scope:a", "scope:b"})
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

func TestScopeSetRejectsValuesOutsideRFC6749ScopeTokenGrammar(t *testing.T) {
	for _, values := range [][]string{{""}, {" scope:a"}, {"scope:a "}, {"scope a"}, {"scope\\a"}, {"scope\x00a"}, {"scope\x1fa"}, {"scope\x7fa"}, {"scopé"}, {"scope\"a"}} {
		if _, err := ParseScopeSet(values); err == nil {
			t.Fatalf("ParseScopeSet(%q) unexpectedly succeeded", values)
		}
	}
	validBoundary, err := ParseScopeSet([]string{string([]byte{0x21, 0x23, 0x5b, 0x5d, 0x7e})})
	if err != nil || !validBoundary.Contains("!#[]~") {
		t.Fatalf("valid RFC 6749 boundary token rejected: %v", err)
	}
}

func TestScopeSetZeroValueFailsClosed(t *testing.T) {
	var set ScopeSet
	if !set.Empty() || set.Contains("scope:a") {
		t.Fatal("zero scope set granted authority")
	}
}
