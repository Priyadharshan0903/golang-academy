//go:build !solution

// Package palindrome works with text one character (rune) at a time.
package palindrome

// Reverse returns s with its characters in reverse order.
// It must work for any UTF-8 text: Reverse("héllo") is "olléh",
// and Reverse("世界") is "界世".
//
// Hint: converting to []rune gives you one element per character.
func Reverse(s string) string {
	// TODO: implement me.
	return ""
}

// IsPalindrome reports whether s reads the same forwards and
// backwards, ignoring case and ignoring every rune that is not a
// letter or a digit.
//
//	IsPalindrome("A man, a plan, a canal: Panama") == true
//	IsPalindrome("Go") == false
//	IsPalindrome("") == true
//
// Hint: look at unicode.IsLetter, unicode.IsDigit and unicode.ToLower.
func IsPalindrome(s string) bool {
	// TODO: implement me.
	return false
}
