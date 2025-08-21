package eval

import (
	"testing"
)

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		want       bool
		wantErr    bool
	}{
		{
			name:       "simple true expression",
			expression: "1 == 1",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "simple false expression",
			expression: "1 != 1",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "strlen function",
			expression: `strlen("hello") == 5`,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "dur function",
			expression: `dur("1h") == 3600`,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "contains function",
			expression: `contains("hello", "ll")`,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "size function",
			expression: `size("1K") == 1000`,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "matches function",
			expression: `matches("hello", "^h")`,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "invalid expression",
			expression: "1 ==",
			want:       false,
			wantErr:    true,
		},
		{
			name:       "function with invalid arguments",
			expression: "strlen()",
			want:       false,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Evaluate(tc.expression)
			if (err != nil) != tc.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if got != tc.want {
				t.Errorf("Evaluate() = %v, want %v", got, tc.want)
			}
		})
	}
}
