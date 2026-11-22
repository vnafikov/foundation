//go:generate go tool modelmeta-gen -type=UploadingFile

package models

import "github.com/vnafikov/foundation/infrastructure/databases/pg/queries"

const (
	UploadingFilePathMaxLen     = 1000
	UploadingFileLeaseTokenSize = 16

	UploadingFileTempExtension = ".tmp"
)

var _ queries.Hydratable[UploadingFile] = UploadingFile{}

type UploadingFile struct {
	Path string `db:"uf.path"`
}
