//go:generate $HOME/go/bin/mockgen -source=interfaces.go -destination=../mock/mock_observer.go -package=mock

package base

import "context"

// Observer определяет интерфейс для наблюдателей в паттерне Observer.
// Наблюдатели получают уведомления о событиях через метод Notify.
type Observer[Event any] interface {
	Notify(ctx context.Context, event Event) error
	GetID() string
	Stop() error
}

// Subject определяет интерфейс для субъекта в паттерне Observer.
// Позволяет подписывать и отписывать наблюдателей, а также уведомлять всех наблюдателей о событиях.
type Subject[Event any] interface {
	Subscribe(observer Observer[Event])
	Unsubscribe(observer Observer[Event])
	NotifyAll(ctx context.Context, event Event) error
	UnsubscribeAll()
}
