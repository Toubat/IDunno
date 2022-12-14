package utils

type Queue[T any] []T

/*
* Push to queue, if failed, not popped, successed,
 */
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (t *Queue[T]) Len() int {
	return len(*t)
}

func (t *Queue[T]) Empty() bool {
	return t.Len() == 0
}

func (t *Queue[T]) Top() T {
	if t.Empty() {
		panic("empty queue")
	}
	return (*t)[0]
}

func (t *Queue[T]) Pop() {
	if t.Empty() {
		return
	}

	*t = (*t)[1:]
}

func (t *Queue[T]) Push(item T) {
	*t = append(*t, item)
}

func (t *Queue[T]) Clear() {
	*t = make(Queue[T], 0)
}

func (t *Queue[T]) ToSlice() []T {
	return *t
}
