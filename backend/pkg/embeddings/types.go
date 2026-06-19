package embeddings

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
)

type Vector []float32

func (v Vector) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (v *Vector) Scan(value any) error {
	if value == nil {
		*v = nil
		return nil
	}
	var raw []byte
	switch x := value.(type) {
	case []byte:
		raw = x
	case string:
		raw = []byte(x)
	default:
		return fmt.Errorf("embeddings: unsupported type %T", value)
	}
	return json.Unmarshal(raw, v)
}

func CosineSimilarity(a, b Vector) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
