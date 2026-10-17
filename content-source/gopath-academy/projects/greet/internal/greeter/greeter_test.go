package greeter

import "testing"

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want string
	}{
		{"defaults", Options{}, "Hello, Gopher!"},
		{"name", Options{Name: "Ana"}, "Hello, Ana!"},
		{"trim", Options{Name: "  Ana "}, "Hello, Ana!"},
		{"times", Options{Name: "Go", Times: 2}, "Hello, Go!\nHello, Go!"},
		{"negative times", Options{Name: "Go", Times: -5}, "Hello, Go!"},
		{"shout", Options{Name: "Go", Shout: true}, "HELLO, GO!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Render(tt.opts); got != tt.want {
				t.Errorf("Render(%+v) = %q, want %q", tt.opts, got, tt.want)
			}
		})
	}
}
