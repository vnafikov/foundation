package uploadingfileusecases

import (
	"context"
	"time"

	"github.com/vnafikov/common/crypto"
	"github.com/vnafikov/common/retry"

	"github.com/vnafikov/foundation/adapters/storages/localfs"
	"github.com/vnafikov/foundation/features/uploadingfiles/adapters/queries/uploadingfiles"
	"github.com/vnafikov/foundation/features/uploadingfiles/core/models"
	"github.com/vnafikov/foundation/features/uploadingfiles/core/queries/uploadingfileparameters"
)

var retryConfiguration = retry.ExponentialConfiguration{
	MinDelay:   2 * time.Second,
	MaxDelay:   8 * time.Second,
	MaxRetries: 3,
}

type DeleteExpired struct {
	uploadingFileQueries uploadingfiles.Queries
	deleteFiles          localfs.DeleteFilesIfExists
}

// Handle удаляет незавершённые загрузки (пустые папки останутся) и возвращает количество удалённых записей.
func (useCase *DeleteExpired) Handle(ctx context.Context, params *DeleteExpiredParameters) (int64, error) {
	leaseToken := crypto.GenerateHex(models.UploadingFileLeaseTokenSize)
	var total int64
	for {
		uploadingFiles, err := useCase.uploadingFileQueries.LeaseExpiredBatch(
			ctx,
			&uploadingfileparameters.LeaseExpiredBatch{
				TTLInterval:   params.TTLInterval,
				LeaseInterval: params.LeaseInterval,
				BatchSize:     params.BatchSize,
				LeaseToken:    leaseToken,
			},
		)
		if err != nil {
			return total, err
		}

		uploadingFileCount := len(uploadingFiles)
		if uploadingFileCount == 0 {
			return total, nil
		}

		paths := make([]string, 0, uploadingFileCount*2)
		for _, uploadingFile := range uploadingFiles {
			paths = append(paths, uploadingFile.Path+models.UploadingFileTempExtension, uploadingFile.Path)
		}
		if err := useCase.deleteFiles.Handle(ctx, paths); err != nil {
			return total, err
		}

		if err := retry.Exponential(ctx, retryConfiguration, func() error {
			count, err := useCase.uploadingFileQueries.DeleteByLeaseToken(ctx, leaseToken)
			if err != nil {
				return err
			}

			total += count
			return nil
		}); err != nil {
			return total, err
		}
		if uploadingFileCount < params.BatchSize {
			return total, nil
		}
	}
}

func NewDeleteExpired(uploadingFileQueries uploadingfiles.Queries, deleteFiles localfs.DeleteFilesIfExists) DeleteExpired {
	return DeleteExpired{
		uploadingFileQueries: uploadingFileQueries,
		deleteFiles:          deleteFiles,
	}
}
