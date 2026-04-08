package main

import (
	"embed"

	"github.com/ZouZhao321/distill/cmd"
)

//go:embed locales/*
var localesFS embed.FS

func main() {
	cmd.SetLocalesFS(localesFS)
	cmd.Execute()
}
