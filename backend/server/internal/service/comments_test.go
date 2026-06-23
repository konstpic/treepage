package service

import "testing"

func TestMentionRegex(t *testing.T) {
	cases := []struct {
		body string
		want []string
	}{
		{"Hi @admin@local please review", []string{"admin@local"}},
		{"cc @user@example.com and @bob@corp.io", []string{"user@example.com", "bob@corp.io"}},
		{"no mentions here", nil},
	}
	for _, tc := range cases {
		got := mentionRe.FindAllStringSubmatch(tc.body, -1)
		var emails []string
		for _, m := range got {
			if len(m) > 1 {
				emails = append(emails, m[1])
			}
		}
		if len(emails) != len(tc.want) {
			t.Fatalf("body %q: got %d mentions %v, want %d %v", tc.body, len(emails), emails, len(tc.want), tc.want)
		}
		for i := range emails {
			if emails[i] != tc.want[i] {
				t.Fatalf("body %q: got %v, want %v", tc.body, emails, tc.want)
			}
		}
	}
}
