package embeddings

import "testing"

func TestCosineSimilarity(t *testing.T) {
	a := Vector{1, 0, 0}
	b := Vector{1, 0, 0}
	if CosineSimilarity(a, b) != 1 {
		t.Fatal("expected 1")
	}
	c := Vector{0, 1, 0}
	if CosineSimilarity(a, c) != 0 {
		t.Fatal("expected 0")
	}
}
