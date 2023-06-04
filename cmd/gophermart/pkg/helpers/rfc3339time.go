package helpers

import "time"

type RFC3339Time time.Time

func (ct RFC3339Time) MarshalJSON() ([]byte, error) {
	t := time.Time(ct)
	formatted := t.Format(time.RFC3339)
	return []byte(`"` + formatted + `"`), nil
}
