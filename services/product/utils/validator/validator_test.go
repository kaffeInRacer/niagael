package validator

import (
	"testing"
)

type TestStruct struct {
	Name   string `validate:"required,min=3,max=255"`
	SortBy string `validate:"def_enum=created_at name price"`
	Order  string `validate:"def_enum=desc asc"`
	Page   int32  `validate:"clamp=1 100"`
}

func TestDefEnum(t *testing.T) {
	v := Get()

	tests := []struct {
		name     string
		input    *TestStruct
		wantErr  bool
		expected string
	}{
		{
			name:     "valid value - created_at",
			input:    &TestStruct{Name: "test", SortBy: "created_at", Order: "desc", Page: 10},
			wantErr:  false,
			expected: "created_at",
		},
		{
			name:     "valid value - name",
			input:    &TestStruct{Name: "test", SortBy: "name", Order: "asc", Page: 10},
			wantErr:  false,
			expected: "name",
		},
		{
			name:     "invalid value - default to first",
			input:    &TestStruct{Name: "test", SortBy: "invalid", Order: "desc", Page: 10},
			wantErr:  false,
			expected: "created_at", // default
		},
		{
			name:     "empty value - default to first",
			input:    &TestStruct{Name: "test", SortBy: "", Order: "desc", Page: 10},
			wantErr:  false,
			expected: "created_at", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.input.SortBy != tt.expected {
				t.Errorf("SortBy = %v, want %v", tt.input.SortBy, tt.expected)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	v := Get()

	tests := []struct {
		name     string
		input    *TestStruct
		wantErr  bool
		expected int32
	}{
		{
			name:     "valid value",
			input:    &TestStruct{Name: "test", SortBy: "created_at", Order: "desc", Page: 50},
			wantErr:  false,
			expected: 50,
		},
		{
			name:     "below min - clamp to min",
			input:    &TestStruct{Name: "test", SortBy: "created_at", Order: "desc", Page: 0},
			wantErr:  false,
			expected: 1, // clamped to min
		},
		{
			name:     "above max - clamp to max",
			input:    &TestStruct{Name: "test", SortBy: "created_at", Order: "desc", Page: 200},
			wantErr:  false,
			expected: 100, // clamped to max
		},
		{
			name:     "at min boundary",
			input:    &TestStruct{Name: "test", SortBy: "created_at", Order: "desc", Page: 1},
			wantErr:  false,
			expected: 1,
		},
		{
			name:     "at max boundary",
			input:    &TestStruct{Name: "test", SortBy: "created_at", Order: "desc", Page: 100},
			wantErr:  false,
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.input.Page != tt.expected {
				t.Errorf("Page = %v, want %v", tt.input.Page, tt.expected)
			}
		})
	}
}

func TestValidateStruct(t *testing.T) {
	v := Get()

	tests := []struct {
		name    string
		input   *TestStruct
		wantNil bool
	}{
		{
			name:    "all valid",
			input:   &TestStruct{Name: "test", SortBy: "created_at", Order: "desc", Page: 10},
			wantNil: true,
		},
		{
			name:    "invalid sort - gets default",
			input:   &TestStruct{Name: "test", SortBy: "invalid", Order: "desc", Page: 10},
			wantNil: true, // def_enum sets default, so no error
		},
		{
			name:    "name too short",
			input:   &TestStruct{Name: "ab", SortBy: "created_at", Order: "desc", Page: 10},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateStruct(tt.input)
			if tt.wantNil && errors != nil {
				t.Errorf("ValidateStruct() errors = %v, want nil", errors)
			}
			if !tt.wantNil && errors == nil {
				t.Errorf("ValidateStruct() errors = nil, want errors")
			}
		})
	}
}
