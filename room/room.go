package room

import (
	"context"
	"log"
	"slices"
	"sync"
	"time"

	"chat-room/data"
	. "chat-room/internal"
	"chat-room/user"
	"github.com/google/uuid"
)

type Options struct {
	BatchSize           uint
	BatchInterval       time.Duration
	ListenerSendTimeout time.Duration
}

const DefaultBatchSize = 1000
const DefaultBatchInterval = 100 * time.Millisecond
const DefaultListenerSendTimeout = 100 * time.Microsecond

func (o *Options) withDefaults() Options {
	var opts Options
	if o != nil {
		opts = *o
	}
	if opts.BatchSize == 0 {
		opts.BatchSize = DefaultBatchSize
	}
	if opts.BatchInterval == 0 {
		opts.BatchInterval = DefaultBatchInterval
	}
	if opts.ListenerSendTimeout == 0 {
		opts.ListenerSendTimeout = DefaultListenerSendTimeout
	}
	return opts
}

type Room struct {
	// Options is used to configure the [Room]
	Options

	// isProper is used to assert that [Room] is created via the [New] or [NewWithContext] functions.
	isProper  bool
	incoming  chan data.UserMessage
	listeners map[uuid.UUID]chan []data.UserMessage
	ctx       context.Context
	closeCtx  context.CancelCauseFunc
	mu        sync.RWMutex
}

func (r *Room) SendContext(ctx context.Context, message data.UserMessage) error {
	select {
	case <-r.ctx.Done():
		return ClosedError
	case <-ctx.Done():
		return ctx.Err()
	case r.incoming <- message:
		return nil
	}
}

func (r *Room) SendTimeout(message data.UserMessage, timeout time.Duration) error {
	ctx, closeCtx := context.WithTimeoutCause(context.Background(), timeout, SendTimeoutError)
	defer closeCtx()
	return r.SendContext(ctx, message)
}

func (r *Room) assertProper() {
	if !r.isProper {
		panic("room was not created using room.New() function")
	}
}

func (r *Room) IsOpen() bool {
	return !r.IsClosed()
}

func (r *Room) IsClosed() bool {
	r.assertProper()
	select {
	case <-r.ctx.Done():
		return true
	default:
		return false
	}
}

func (r *Room) Join(user *user.User) error {
	r.assertProper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.IsClosed() {
		return ClosedError
	}
	r.listeners[user.UUID] = user.Incoming
	return nil
}

func (r *Room) Leave(user *user.User) {
	r.assertProper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.IsClosed() {
		return
	}
	delete(r.listeners, user.UUID)
}

func (r *Room) collectMessages() []data.UserMessage {
	batch := make([]data.UserMessage, 0, r.BatchSize)
	collectTimeout := time.NewTimer(r.BatchInterval)
	for {
		select {
		case <-r.ctx.Done():
			return batch
		case <-collectTimeout.C:
			return batch
		case message := <-r.incoming:
			batch = append(batch, message)
			if len(batch) == cap(batch) {
				return batch
			}
		}
	}
}

func (r *Room) broadcast(batch []data.UserMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sendTimeout := time.NewTimer(0)
	for listenerUUID, listener := range r.listeners {
		sendTimeout.Reset(r.ListenerSendTimeout)
		select {
		case <-r.ctx.Done():
			log.Printf("stop listening: %q", r.ctx.Err())
			return
		case listener <- slices.Clone(batch):
		case <-sendTimeout.C:
			log.Printf("%s listener timeout", listenerUUID)
		}
	}
}

func (r *Room) start() {
	t := time.NewTicker(r.BatchInterval)
	defer t.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-t.C:
			batch := r.collectMessages()
			if len(batch) > 0 {
				r.broadcast(batch)
			}
		}
	}
}

func (r *Room) Close() error {
	r.assertProper()
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.IsClosed() {
		return nil
	}
	r.closeCtx(ClosedError)
	close(r.incoming)
	for _, listener := range r.listeners {
		close(listener)
	}
	return nil
}

func New(opts *Options) *Room {
	return NewWithContext(context.Background(), opts)
}

func NewWithContext(ctx context.Context, optsPtr *Options) *Room {
	ctx, closeCtx := context.WithCancelCause(ctx)
	opts := optsPtr.withDefaults()
	r := Room{
		Options:   opts,
		isProper:  true,
		incoming:  make(chan data.UserMessage, opts.BatchSize),
		listeners: make(map[uuid.UUID]chan []data.UserMessage),
		ctx:       ctx,
		closeCtx:  closeCtx,
	}
	go r.start()
	return &r
}

const ClosedError StringError = "room is closed"
const SendTimeoutError StringError = "send timeout"
