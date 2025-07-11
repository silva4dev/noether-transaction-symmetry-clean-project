package usecase

import (
	"context"

	"github.com/silva4dev/noether-transaction-symmetry-clean-project/internal/domain/entity"
	"github.com/silva4dev/noether-transaction-symmetry-clean-project/internal/domain/repository"
	"github.com/silva4dev/noether-transaction-symmetry-clean-project/pkg/uow"
)

type AddMyTeamInput struct {
	ID    string
	Name  string
	Score int
}

type AddMyTeamUseCase struct {
	Uow uow.UowInterface
}

func (a *AddMyTeamUseCase) Execute(ctx context.Context, input AddMyTeamInput) error {
	myTeamRepository := a.getMyTeamRepository(ctx)
	myTeam := entity.NewMyTeam(input.ID, input.Name)
	err := myTeamRepository.Create(ctx, myTeam)
	if err != nil {
		return err
	}
	a.Uow.CommitOrRollback()
	return nil
}

func (a *AddMyTeamUseCase) getMyTeamRepository(ctx context.Context) repository.MyTeamRepositoryInterface {
	myTeamRepository, err := a.Uow.GetRepository(ctx, "MyTeamRepository")
	if err != nil {
		panic(err)
	}
	return myTeamRepository.(repository.MyTeamRepositoryInterface)
}
