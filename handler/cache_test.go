package handler

import "testing"

func TestValidate(t *testing.T) {
	c := Cache{}
	tests := map[string]bool{
		"1,view":    true,
		"1,view ":   false,
		"1,":        false,
		"1":         false,
		",qwe":      false,
		"555,close": true,
	}
	for key, want := range tests {
		t.Run(key, func(t *testing.T) {
			if got := c.Validate(key); got != want {
				t.Errorf("Validate() = %v, want %v", got, want)
			}
		})
	}
}
