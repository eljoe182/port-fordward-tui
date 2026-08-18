package ports

type LocalPortChecker interface {
	Available(port int) bool
}
