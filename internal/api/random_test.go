package api

import (
	"math/rand"
	"strings"
	"testing"
)

func TestRandomStationsFromList(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	stations, err := randomStationsFromList(r, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stations) != 3 {
		t.Fatalf("expected 3 stations, got %d", len(stations))
	}
	seen := make(map[string]struct{})
	for _, s := range stations {
		seen[s] = struct{}{}
	}
	if len(seen) != 3 {
		t.Errorf("stations not unique: %v", stations)
	}
}

func TestRandomStationsInvalid(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	if _, err := randomStationsFromList(r, len(majorStations)); err == nil {
		t.Fatal("expected error for invalid count")
	}
}

func TestRandomStationsExclude(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	exclude := map[string]struct{}{strings.ToLower(majorStations[0]): {}}
	stations, err := randomStationsExclude(r, exclude, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stations) != 2 {
		t.Fatalf("expected 2 stations got %d", len(stations))
	}
	for _, s := range stations {
		if strings.ToLower(s) == strings.ToLower(majorStations[0]) {
			t.Errorf("excluded station returned")
		}
	}
}

func TestRandomStationsExcludeInvalid(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	exclude := map[string]struct{}{}
	if _, err := randomStationsExclude(r, exclude, len(majorStations)+1); err == nil {
		t.Fatal("expected error for too many stations")
	}
}
