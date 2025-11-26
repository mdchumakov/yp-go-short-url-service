package observer

import (
	"context"
	"sync"
	"yp-go-short-url-service/internal/observer/base"
)

// EventBus реализует паттерн Observer для асинхронной обработки событий.
// Позволяет подписывать наблюдателей и уведомлять их о событиях параллельно.
type EventBus[Event any] struct {
	observers map[string]base.Observer[Event]
	mutex     sync.RWMutex
}

// NewEventBus создает новый шину событий для обработки событий через паттерн Observer.
// Возвращает реализацию интерфейса Subject с поддержкой параллельного уведомления наблюдателей.
func NewEventBus[Event any]() base.Subject[Event] {
	return &EventBus[Event]{
		observers: make(map[string]base.Observer[Event]),
	}
}

// Subscribe подписывает наблюдателя на получение событий.
// Если наблюдатель с таким ID уже подписан, он будет заменен новым.
func (eb *EventBus[Event]) Subscribe(observer base.Observer[Event]) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	eb.observers[observer.GetID()] = observer
}

// Unsubscribe отписывает наблюдателя от получения событий.
// Удаляет наблюдателя из списка подписчиков по его ID.
func (eb *EventBus[Event]) Unsubscribe(observer base.Observer[Event]) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	delete(eb.observers, observer.GetID())
}

// NotifyAll уведомляет всех подписанных наблюдателей о событии параллельно.
// Использует горутины для параллельной обработки и возвращает первую ошибку, если она возникла.
func (eb *EventBus[Event]) NotifyAll(ctx context.Context, event Event) error {
	eb.mutex.RLock()
	observers := make([]base.Observer[Event], 0, len(eb.observers))
	for _, obs := range eb.observers {
		observers = append(observers, obs)
	}
	eb.mutex.RUnlock()

	if len(observers) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(observers))

	for _, observer := range observers {
		wg.Add(1)
		go func(obs base.Observer[Event]) {
			defer wg.Done()
			// Проверяем, не отменен ли контекст перед вызовом
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err := obs.Notify(ctx, event); err != nil {
				select {
				case errChan <- err:
				case <-ctx.Done():
					return
				}
			}
		}(observer)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
