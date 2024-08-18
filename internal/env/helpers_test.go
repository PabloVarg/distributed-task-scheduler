package env

import (
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestGetRequiredEnvString(t *testing.T) {
	t.Run("test existing env", func(t *testing.T) {
		t.Setenv("KEY", "test_value")
		envValue := GetRequiredEnvString("KEY", slog.New(slog.NewJSONHandler(io.Discard, nil)))

		if envValue != "test_value" {
			t.Fail()
		}
	})

	t.Run("test missing env", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fail()
			}
		}()

		GetRequiredEnvString("KEY", slog.New(slog.NewJSONHandler(io.Discard, nil)))
	})
}

func TestGetEnvString(t *testing.T) {
	tbl := []struct {
		key          string
		value        string
		defaultValue string
	}{
		{key: "EXISTING_KEY", value: "value"},
		{key: "NON_EXISTING_KEY", value: "", defaultValue: "default_value"},
	}

	for _, test := range tbl {
		t.Run(test.key, func(t *testing.T) {
			if test.value != "" {
				t.Setenv(test.key, test.value)
			}

			value := GetEnvString(test.key, test.defaultValue, slog.New(slog.NewJSONHandler(io.Discard, nil)))

			if test.value == "" {
				if test.defaultValue != value {
					t.Fail()
				}
			} else {
				if value != test.value {
					t.Fail()
				}
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tbl := []struct {
		key           string
		value         string
		defaultValue  int
		expectedInt   int
		expectedPanic bool
	}{
		{key: "KEY_1", value: "1", expectedInt: 1},
		{key: "KEY_NEGATIVE", value: "-1", expectedInt: -1},
		{key: "KEY_INVALID", value: "test", expectedPanic: true},
		{key: "KEY_EMPTY", value: "", defaultValue: 20, expectedInt: 20},
	}

	for _, test := range tbl {
		t.Run(test.key, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if test.expectedPanic == false {
						t.Fail()
					}
				}
			}()

			if test.value != "" {
				t.Setenv(test.key, test.value)
			}

			value := GetEnvInt(test.key, test.defaultValue, slog.New(slog.NewJSONHandler(io.Discard, nil)))

			if test.value != "" && value != test.expectedInt {
				t.Fail()
			}
			if test.value == "" && value != test.defaultValue {
				t.Fail()
			}
		})
	}
}

func TestGetRequiredEnvDuration(t *testing.T) {
	tbl := []struct {
		key           string
		value         string
		expectedValue time.Duration
		expectedPanic bool
	}{
		{key: "VALID_DURATION", value: "5s", expectedValue: 5 * time.Second},
		{key: "INVALID_DURATION", value: "test", expectedPanic: true},
		{key: "EMPTY_DURATION", expectedPanic: true},
	}

	for _, test := range tbl {
		t.Run(test.key, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if test.expectedPanic == false {
						t.Fail()
					}
				}
			}()

			if test.value != "" {
				t.Setenv(test.key, test.value)
			}

			value := GetRequiredEnvDuration(test.key, slog.New(slog.NewJSONHandler(io.Discard, nil)))

			if value != test.expectedValue {
				t.Fail()
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tbl := []struct {
		key                string
		value              string
		defaultValue       time.Duration
		defaultStringValue string
		expectedValue      time.Duration
		expectedPanic      bool
	}{
		{key: "VALID_DURATION", value: "5s", expectedValue: 5 * time.Second},
		{key: "INVALID_DURATION", value: "test", expectedPanic: true},
		{key: "EMPTY_DURATION", defaultValue: 10 * time.Second, defaultStringValue: "10s"},
	}

	for _, test := range tbl {
		t.Run(test.key, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if test.expectedPanic == false {
						t.Fail()
					}
				}
			}()

			if test.value != "" {
				t.Setenv(test.key, test.value)
			}

			value := GetEnvDuration(test.key, test.defaultStringValue, slog.New(slog.NewJSONHandler(io.Discard, nil)))

			if test.value != "" && value != test.expectedValue {
				t.Fail()
			}
			if test.value == "" && value != test.defaultValue {
				t.Fail()
			}
		})
	}
}
