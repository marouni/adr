package main

import "embed"

// fs embeds the template into the binary.
//go:embed tpl
var fs embed.FS
