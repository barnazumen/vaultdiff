package vault

import (
	"testing"
)

func TestResolveVersionPair_ExplicitVersions(t *testing.T) {
	versions := []int{1, 2, 3, 4}
	pair, err := ResolveVersionPair(versions, VersionPair{From: 2, To: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair.From != 2 || pair.To != 4 {
		t.Errorf("expected 2->4, got %d->%d", pair.From, pair.To)
	}
}

func TestResolveVersionPair_AutoLatest(t *testing.T) {
	versions := []int{1, 2, 3}
	pair, err := ResolveVersionPair(versions, VersionPair{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair.To != 3 || pair.From != 2 {
		t.Errorf("expected 2->3, got %d->%d", pair.From, pair.To)
	}
}

func TestResolveVersionPair_EmptyVersions(t *testing.T) {
	_, err := ResolveVersionPair([]int{}, VersionPair{})
	if err == nil {
		t.Fatal("expected error for empty versions")
	}
}

func TestResolveVersionPair_FromEqualsTo(t *testing.T) {
	versions := []int{1, 2, 3}
	_, err := ResolveVersionPair(versions, VersionPair{From: 3, To: 3})
	if err == nil {
		t.Fatal("expected error when from >= to")
	}
}

func TestResolveVersionPair_OnlyOneVersion(t *testing.T) {
	versions := []int{1}
	_, err := ResolveVersionPair(versions, VersionPair{})
	if err == nil {
		t.Fatal("expected error when only one version exists (from would be 0)")
	}
}
