package matching

import (
	"testing"
)

func TestJoinFloats(t *testing.T) {
	v := []float64{0.001, -0.02, 0.5}
	result := joinFloats(v, ",")
	expected := "[0.001,-0.02,0.5]"
	if result != expected {
		t.Errorf("joinFloats(%v, \",\") = %q, want %q", v, result, expected)
	}

	result = joinFloats(nil, ",")
	if result != "[]" {
		t.Errorf("joinFloats(nil, \",\") = %q, want \"[]\"", result)
	}

	result = joinFloats([]float64{}, ",")
	if result != "[]" {
		t.Errorf("joinFloats([], \",\") = %q, want \"[]\"", result)
	}

	v = []float64{3.14}
	result = joinFloats(v, ",")
	expected = "[3.14]"
	if result != expected {
		t.Errorf("joinFloats(%v, \",\") = %q, want %q", v, result, expected)
	}
}

func TestStore_CreateTables(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_UpsertMarket(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_GetUnembeddedMarkets(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_UpsertEmbedding(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_InsertCandidate(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_GetPendingCandidates(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_InsertMatchPair(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_GetMatchPairs(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_UpdateMatchPairStatus(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_GetCandidateByID(t *testing.T) {
	t.Skip("integration — run on EC2")
}

func TestStore_GetMarketByID(t *testing.T) {
	t.Skip("integration — run on EC2")
}
