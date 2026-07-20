/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadFileAndExpandEnv verifies that configuration values containing a
// literal "$" (for example a password like "uhQhu4watyTgn$Q$") are not
// corrupted, while the ${VAR} form is still expanded.
//
// It exercises readFileAndExpandEnv, the function that actually parses the
// on-disk config, so it fails on the previous os.ExpandEnv implementation
// (which stripped the "$Q$" portion) and passes on the braced-only expander.
func TestReadFileAndExpandEnv(t *testing.T) {
	os.Setenv("REMCO_TEST_TOKEN", "s3cr3t")
	defer os.Unsetenv("REMCO_TEST_TOKEN")

	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "braced var is expanded",
			in:   "password = \"${REMCO_TEST_TOKEN}\"",
			want: "password = \"s3cr3t\"",
		},
		{
			name: "literal dollar signs are preserved",
			in:   "password = \"uhQhu4watyTgn$Q$\"",
			want: "password = \"uhQhu4watyTgn$Q$\"",
		},
		{
			name: "bare dollar without braces is preserved",
			in:   "value = \"costs $5 today\"",
			want: "value = \"costs $5 today\"",
		},
		{
			name: "mixing braced var and literal dollar",
			in:   "conn = \"${REMCO_TEST_TOKEN}:uhQhu4watyTgn$Q$\"",
			want: "conn = \"s3cr3t:uhQhu4watyTgn$Q$\"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "config.toml")
			if err := os.WriteFile(p, []byte(tc.in), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := readFileAndExpandEnv(p)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Errorf("readFileAndExpandEnv(%q) = %q, want %q", tc.in, string(got), tc.want)
			}
		})
	}
}
