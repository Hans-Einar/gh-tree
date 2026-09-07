package brokerassets

import _ "embed"

//go:embed broker-arm64.gz
var arm64Bytes []byte

//go:embed manifest.json
var manifestBytes []byte

func payload(arch string) []byte {
	if arch == "arm64" {
		return arm64Bytes
	}
	return nil
}
