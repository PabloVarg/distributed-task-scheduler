package api

import (
	"net/http/httptest"
	"testing"
)

func TestIntPathValue(t *testing.T) {
	tbl := []struct {
		pathValue     string
		expectedValue int
		expectErr     bool
	}{
		{pathValue: "5", expectedValue: 5},
	}

	for _, test := range tbl {
		t.Run(test.pathValue, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.SetPathValue("id", test.pathValue)

			value, err := IntPathValue(r, "id")
			if err != nil && !test.expectErr {
				t.Errorf("expected no errors")
			}
			if err == nil && test.expectErr {
				t.Errorf("expected an error")
			}

			if value != test.expectedValue {
				t.Errorf("received %d but expected %d", value, test.expectedValue)
			}
		})
	}
}
