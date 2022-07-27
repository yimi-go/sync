package mu

type node[E any] struct {
	val        *E
	prev, next *node[E]
}
