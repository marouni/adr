package main

import "testing"

func TestFindLastNumber(t *testing.T) {
	t.Parallel()
	cfg := getConfig()
	n := findLastNumber(cfg)
	_ = n
}
