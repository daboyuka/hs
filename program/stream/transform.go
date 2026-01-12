package stream

type Mapper[T any] = Mapper2[T, T]
type Operator[T any] = Operator2[T, T]

type Mapper2[T1, T2 any] func(T1) (T2, error)
type Operator2[T1, T2 any] func(T1, Sink[T2]) error

func (m Mapper2[T1, T2]) Operate(v1 T1, yield Sink[T2]) error {
	if v2, err := m(v1); err != nil {
		return err
	} else {
		return yield(v2)
	}
}

func Map[T1, T2 any](stream Stream[T1], mapper Mapper2[T1, T2]) Stream[T2] {
	return Apply(stream, mapper.Operate)
}

func Apply[T1, T2 any](stream Stream[T1], op Operator2[T1, T2]) Stream[T2] {
	return func(yield2 func(T2) error) error {
		yield1 := func(v1 T1) error {
			return op(v1, yield2)
		}
		return stream(yield1)
	}
}
