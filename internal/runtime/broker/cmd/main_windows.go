package main

import (
	"github.com/Hans-Einar/gh-tree/internal/runtime/broker"
	"os"
)

func main() { os.Exit(broker.RunWindowsPrivate()) }
