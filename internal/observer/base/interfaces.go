//go:generate $HOME/go/bin/mockgen -source=interfaces.go -destination=../mock/mock_observer.go -package=mock

package base

import "context"

type Observer[Event any] interface {
	Notify(ctx context.Context, event Event) error
	GetID() string
}

type Subject[Event any] interface {
	Subscribe(observer Observer[Event])
	Unsubscribe(observer Observer[Event])
	NotifyAll(ctx context.Context, event Event) error
}
