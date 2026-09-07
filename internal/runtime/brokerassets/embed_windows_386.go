package brokerassets

import _ "embed"

//go:embed broker-amd64.gz
var amd64Bytes []byte

//go:embed broker-arm64.gz
var arm64Bytes []byte

//go:embed manifest.json
var manifestBytes []byte

func payload(arch string) []byte {
	switch arch {
	case "amd64":
		return amd64Bytes
	case "arm64":
		return arm64Bytes
	}
	return nil
}
