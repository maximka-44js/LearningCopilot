package library

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/entity"
	"github.com/maximka-44js/lc/internal/storage/repo"
)

//UseCase

type UseCase struct {
	repo repo.LibraryRepo
	jobs LibraryJobPublisher
	tx   TxManager
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type LibraryJobPublisher interface {
	EnqueueLibraryDeleted(ctx context.Context, libraryID, userID string) error
	EnqueueLibraryCompress(ctx context.Context, libraryID, userID string) error
}

func NewLibraryUseCase(r repo.LibraryRepo, jobs LibraryJobPublisher, tx TxManager) *UseCase {
	return &UseCase{repo: r, jobs: jobs, tx: tx}
}

func (uc *UseCase) Create(
	ctx context.Context,
	userID, name, description string,
	public bool,
	accesses []entity.LibraryAccess,
) (entity.Library, error) {

	lib, err := entity.NewLibrary(uuid.New().String(),
		userID,
		name,
		description,
		public,
		accesses,
	)

	if err != nil {
		return entity.Library{}, err
	}

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.repo.Store(ctx, &lib); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return entity.Library{}, err
	}
	return lib, nil
}

func (uc *UseCase) GetByID(ctx context.Context, userID, libraryID string) (entity.Library, error) {
	lib, err := uc.repo.GetByID(ctx, userID, libraryID)
	if err != nil {
		return entity.Library{}, err
	}
	return lib, nil
}

func (uc *UseCase) Update(
	ctx context.Context,
	userID, libraryID, title, description string,
	public bool,
	accesses []entity.LibraryAccess,
) (entity.Library, error) {

	lib, err := uc.repo.GetByID(ctx, userID, libraryID)
	if err != nil {
		return entity.Library{}, err
	}

	if lib.DeletedAt != nil {
		return entity.Library{}, d.NewNotFound(entity.CodeLibNotFound, "library not found")
	}

	if lib.CanRead(userID) && !lib.CanWrite(userID) {
		return entity.Library{}, d.NewForbidden(entity.CodeLibraryAccessInvalid, "access denied")
	}

	isOwner := lib.UserID == userID
	if !isOwner {
		if public != lib.Public {
			return entity.Library{}, errors.New("only owner can change public flag")
		}
		if !reflect.DeepEqual(accesses, lib.Accesses) {
			return entity.Library{}, errors.New("only owner can change access list")
		}
	}

	lib.Title = title
	lib.Description = description

	if isOwner {
		lib.Public = public
		lib.Accesses = accesses
	}

	lib.UpdatedAt = time.Now().UTC()

	if err := lib.Validate(); err != nil {
		return entity.Library{}, err
	}

	if err := uc.repo.Update(ctx, &lib); err != nil {
		return entity.Library{}, err
	}

	return lib, nil
}

func (uc *UseCase) Delete(ctx context.Context, userID, libraryID string) (entity.Library, error) {
	lib, err := uc.repo.GetByID(ctx, userID, libraryID)
	if err != nil {
		return entity.Library{}, err
	}

	if lib.DeletedAt != nil {
		return entity.Library{}, d.NewNotFound(entity.CodeLibNotFound, "library not found")
	}

	if lib.UserID != userID {
		if lib.CanWrite(userID) || lib.CanRead(userID) {
			return entity.Library{}, d.NewForbidden(entity.CodeLibraryAccessInvalid, "access denied")
		}
		return entity.Library{}, d.NewNotFound(entity.CodeLibNotFound, "library not found")
	}
	now := time.Now().UTC()
	lib.DeletedAt = &now

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := uc.repo.Delete(ctx, userID, libraryID); err != nil {
			return err
		}
		if err := uc.jobs.EnqueueLibraryDeleted(ctx, libraryID, userID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return entity.Library{}, err
	}
	return lib, nil
}

func (uc *UseCase) RequestCompress(ctx context.Context, userID, libraryID string) error {
	lib, err := uc.repo.GetByID(ctx, userID, libraryID)
	if err != nil {
		return err
	}

	if lib.DeletedAt != nil {
		return d.NewNotFound(entity.CodeLibNotFound, "library not found")
	}

	if !lib.CanRead(userID) {
		return d.NewNotFound(entity.CodeLibNotFound, "library not found")
	} else if !lib.CanWrite(userID) {
		return d.NewForbidden(entity.CodeLibraryAccessInvalid, "access denied")
	}

	if err := uc.jobs.EnqueueLibraryCompress(ctx, libraryID, userID); err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) List(ctx context.Context, userID string) ([]entity.Library, error) {
	return uc.repo.List(ctx, userID)
}
