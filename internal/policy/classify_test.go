package policy

import "testing"

func TestClassifyWriteCTE(t *testing.T) {
	cases := map[string]Action{
		"WITH x AS (DELETE FROM secrets RETURNING *) SELECT * FROM x": ActionWrite,
		"EXPLAIN ANALYZE INSERT INTO secrets VALUES (2,'x')":          ActionWrite,
		"SELECT set_config('app.x','1',true)":                         ActionAdmin,
		"WITH x AS (SELECT 1) SELECT * FROM x":                        ActionRead,
	}
	for sql, want := range cases {
		if got, _ := ClassifySQL(sql); got != want {
			t.Errorf("%q = %s, want %s", sql, got, want)
		}
	}
}
