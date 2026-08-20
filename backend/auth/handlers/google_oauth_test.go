package handlers

import "testing"

func TestBuildDisabledRedirectURL_EscapesReactivationToken(t *testing.T) {
	got := buildDisabledRedirectURL("https://letslive.app", "abc.def+ghi")
	want := "https://letslive.app/login?accountDisabled=true&reactivationToken=abc.def%2Bghi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
