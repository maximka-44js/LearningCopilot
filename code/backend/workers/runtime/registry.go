package runtime

import (
	"github.com/maximka-44js/lc/workers/storage/file"
	"github.com/maximka-44js/lc/workers/storage/folder"
	"github.com/maximka-44js/lc/workers/storage/library"
	"github.com/riverqueue/river"
)

// RegisterWorkers wires job handlers.
func RegisterWorkers(workers *river.Workers, deps Deps) {
	river.AddWorker(workers, library.NewDeletedWorker(deps.PG, deps.ObjectStorage))
	river.AddWorker(workers, library.NewCompressWorker(deps.LibraryCompress))
	river.AddWorker(workers, folder.NewDeletedWorker(deps.PG, deps.ObjectStorage))
	river.AddWorker(workers, folder.NewCompressWorker(deps.FolderCompress))
	river.AddWorker(workers, folder.NewExtractWorker(deps.FolderExtract))
	river.AddWorker(workers, file.NewDeletedWorker(deps.PG, deps.ObjectStorage))
}
