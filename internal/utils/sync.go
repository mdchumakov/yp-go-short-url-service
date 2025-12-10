package utils

import "sync"

// Resetter определяет интерфейс для типов, которые можно сбросить
type Resetter interface {
	Reset()
}

// TypedPool - Типизированный пул объектов.
type TypedPool[T Resetter] struct {
	pool *sync.Pool
}

// NewTypedPool - Создает новый типизированный пул объектов.
func NewTypedPool[T Resetter](newFunc func() T) *TypedPool[T] {
	return &TypedPool[T]{
		pool: &sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get - Получает объект из пула.
func (p *TypedPool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put - Возвращает объект в пул.
func (p *TypedPool[T]) Put(x T) {
	p.pool.Put(x)
}
