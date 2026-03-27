package interfaces

//go:generate mockgen -source=errors.go -destination=../../mocks/errors/mapper.go -package=mockerrors -exclude_interfaces=
type ErrorsMapper interface {
	Map(err error) error
}
