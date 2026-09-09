package auction

type Usecase struct {
	storage storage
}

func NewUsecase(storage storage) *Usecase {
	return &Usecase{
		storage: storage,
	}
}

func (u *Usecase) RunAuction() error {
	return nil
}
