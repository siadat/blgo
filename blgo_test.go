package main

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	for text, want := range map[string]map[string]string{
		"---\ndate: 2000-10-20\ntitle: my post title\n---\n": map[string]string{
			"title":  "my post title",
			"date":   "2000-10-20",
			"_after": "",
		},
		"before\n---\ndate: 2001-10-20\ntitle: random title\n---\nafter frontmatter\nend": map[string]string{
			"title":  "random title",
			"date":   "2001-10-20",
			"_after": "after frontmatter\nend",
		},
	} {
		body := []byte(text)
		got, err := parseFrontmatter(&body)
		if err != nil {
			t.Error(err)
		}
		for wantKey, wantVal := range want {
			if strings.HasPrefix(wantKey, "_") {
				continue
			}
			if got[wantKey] != wantVal {
				t.Errorf("for key %q got %q; want %q", wantKey, got[wantKey], want)
			}
		}
		if string(body) != want["_after"] {
			t.Errorf("got %q; want %q", string(body), want["_after"])
		}
	}
}
