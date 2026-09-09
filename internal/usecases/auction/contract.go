package auction

type storage interface {
	ListPartners() ([]string, error)
}
