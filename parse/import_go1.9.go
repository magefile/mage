// +build go1.9

package parse

import (
	"go/types"
	"go/importer"
)

func getImporter() types.Importer {
	return importer.For("source", nil),
}
