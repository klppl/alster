package main

import (
	"testing"
)

func TestCLI(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		bad  bool
	}{
		{"help", []string{"help"}, false}, {"command help", []string{"build", "-h"}, false}, {"unknown", []string{"deploy"}, true}, {"extra argument", []string{"build", "unexpected"}, true}, {"bad flag", []string{"serve", "-bad"}, true}, {"missing config", []string{"build", "-config", "missing-test-config.yaml"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := run(tc.args)
			if (err != nil) != tc.bad {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
