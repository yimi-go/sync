package sync

type Iterator[E any] interface {
	// Next returns the next element and true if existed, or it returns any value and false.
	Next() (E, bool)
}
