package users

import "testing"

func TestNewUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "normalizes lowercase", input: " Alice ", want: "alice"},
		{name: "one character is valid", input: "a", want: "a"},
		{name: "twenty characters are valid", input: "abcdefghijklmnopqrst", want: "abcdefghijklmnopqrst"},
		{name: "rejects digits", input: "alice1", wantErr: ErrInvalidUsername},
		{name: "rejects underscore", input: "alice_dev", wantErr: ErrInvalidUsername},
		{name: "rejects cyrillic", input: "алиса", wantErr: ErrInvalidUsername},
		{name: "rejects long value", input: "abcdefghijklmnopqrstu", wantErr: ErrInvalidUsername},
		{name: "rejects reserved value", input: "admin", wantErr: ErrReservedUsername},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			username, err := NewUsername(test.input)
			if test.wantErr != nil {
				if err != test.wantErr {
					t.Fatalf("expected %v, got %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if username.String() != test.want {
				t.Fatalf("expected %q, got %q", test.want, username.String())
			}
		})
	}
}
