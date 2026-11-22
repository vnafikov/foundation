package tasks

import (
	"context"

	"github.com/vnafikov/foundation/adapters/storages/localfs"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/definitions/backgroundtaskdefinitions"
	"github.com/vnafikov/foundation/features/backgroundtasks/core/usecases/storageusecases"
	"github.com/vnafikov/foundation/json"
)

func DeleteFiles(ctx context.Context, payload []byte, rootPath string) (deleted bool, err error) {
	p := new(backgroundtaskdefinitions.DeleteFilesPayload)
	if err := json.Deserialize(ctx, payload, p); err != nil {
		return false, err
	}

	deleteFilesIfExists := localfs.NewDeleteFilesIfExists(rootPath)
	useCase := storageusecases.NewDeleteFiles(deleteFilesIfExists)
	err = useCase.Handle(ctx, p.Paths)
	return false, err
}
