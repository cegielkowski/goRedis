package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Redis struct {
	storage       map[string]any
	storageCancel map[string]context.CancelFunc
	mu            sync.Mutex
}

func newRedis() *Redis {
	return &Redis{
		storage:       make(map[string]any),
		storageCancel: make(map[string]context.CancelFunc),
		mu:            sync.Mutex{},
	}
}

func (r *Redis) Add(ctx context.Context, key string, value any, expirationInMs int) {
	r.mu.Lock()

	_, ok := r.storage[key]
	r.mu.Unlock()

	if ok {
		r.mu.Lock()
		r.storageCancel[key]()
		delete(r.storageCancel, key)
		r.mu.Unlock()
	}

	cancelDelete := make(chan struct{})
	r.mu.Lock()

	r.storage[key] = value
	r.storageCancel[key] = func() {
		close(cancelDelete)
	}
	r.mu.Unlock()

	go func() {
		select {
		case <-time.After(time.Duration(expirationInMs) * time.Millisecond):
			r.mu.Lock()
			_, ok := r.storage[key]
			r.mu.Unlock()

			if ok {
				r.mu.Lock()
				delete(r.storage, key)
				delete(r.storageCancel, key)
				r.mu.Unlock()
			}
			return
		case <-cancelDelete:
			return
		case <-ctx.Done():
			return
		}
	}()
}

func (r *Redis) Delete(key string) {
	r.mu.Lock()
	_, ok := r.storage[key]
	r.mu.Unlock()
	if ok {
		r.mu.Lock()
		delete(r.storage, key)
		r.storageCancel[key]()
		delete(r.storageCancel, key)
		r.mu.Unlock()
	}
}

func (r *Redis) Get(key string) (any, error) {
	definition, ok := r.storage[key]
	if !ok {
		return nil, errors.New("could not find the key you were looking for")
	}
	return definition, nil
}

func (r *Redis) ShowAll() map[string]any {
	return r.storage
}
