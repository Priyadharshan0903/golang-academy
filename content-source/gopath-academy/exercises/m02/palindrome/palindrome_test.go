package palindrome

import "testing"

func TestReverse(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"a", "a"},
		{"Go", "oG"},
		{"héllo", "olléh"},
		{"世界", "界世"},
		{"ab cd", "dc ba"},
	}
	for _, tt := range tests {
		if got := Reverse(tt.in); got != tt.want {
			t.Errorf("Reverse(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestReverseTwiceIsIdentity(t *testing.T) {
	for _, s := range []string{"gopher", "héllo, 世界", "🙂👍"} {
		if got := Reverse(Reverse(s)); got != s {
			t.Errorf("Reverse(Reverse(%q)) = %q", s, got)
		}
	}
}

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"a", true},
		{"racecar", true},
		{"RaceCar", true},
		{"A man, a plan, a canal: Panama", true},
		{"No 'x' in Nixon", true},
		{"12321", true},
		{"Go", false},
		{"hello", false},
		{"été", true},
		{"ab", false},
	}
	for _, tt := range tests {
		if got := IsPalindrome(tt.in); got != tt.want {
			t.Errorf("IsPalindrome(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
