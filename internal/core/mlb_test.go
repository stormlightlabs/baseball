package core

import (
	"encoding/json"
	"testing"
)

func TestMLBStringNumber_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	type payload struct {
		Value MLBStringNumber `json:"value"`
	}

	tests := []struct {
		name      string
		jsonInput string
		want      MLBStringNumber
		wantErr   bool
	}{
		{
			name:      "string value",
			jsonInput: `{"value":"+42"}`,
			want:      "+42",
		},
		{
			name:      "number value",
			jsonInput: `{"value":-7}`,
			want:      "-7",
		},
		{
			name:      "float value",
			jsonInput: `{"value":12.5}`,
			want:      "12.5",
		},
		{
			name:      "null value",
			jsonInput: `{"value":null}`,
			want:      "",
		},
		{
			name:      "invalid type",
			jsonInput: `{"value":true}`,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got payload
			err := json.Unmarshal([]byte(tc.jsonInput), &got)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected unmarshal error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			if got.Value != tc.want {
				t.Fatalf("unexpected value: got=%q want=%q", got.Value, tc.want)
			}
		})
	}
}
