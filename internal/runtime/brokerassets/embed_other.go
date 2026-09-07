//go:build !windows || (!386 && !amd64)

package brokerassets

var manifestBytes []byte

func payload(string) []byte { return nil }
