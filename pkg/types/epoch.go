package types

import (
	"strings"
	"time"
)

type Epoch int64

func (t Epoch) MarshalJSON() ([]byte, error) {
	strDate := time.Time(time.Unix(int64(t), 0)).Format(time.RFC3339)
	out := []byte(`"` + strDate + `"`)
	return out, nil
}

func (t *Epoch) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	q, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = Epoch(q.Unix())
	return nil
}
