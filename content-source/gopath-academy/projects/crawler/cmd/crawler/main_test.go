package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			fmt.Fprint(w, `<a href="/next">next</a>`)
			return
		}
		fmt.Fprint(w, "leaf")
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	code := run(context.Background(), []string{"-url", srv.URL + "/", "-rate", "1000"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), srv.URL+"/next") || !strings.Contains(errOut.String(), "2 pages") {
		t.Errorf("stdout:\n%s\nstderr:\n%s", out.String(), errOut.String())
	}
}

func TestRunRequiresURL(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run(context.Background(), nil, &out, &errOut); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
}
