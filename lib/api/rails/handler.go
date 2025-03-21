package rails

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/snowmerak/keycl/model/gen/rails"
)

type SessionState struct {
	remoteAddr string
	validated  bool
	email      string

	lock *sync.RWMutex
}

func (s *SessionState) SetRemoteAddr(remoteAddr string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.remoteAddr = remoteAddr
}

func (s *SessionState) RemoteAddr() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.remoteAddr
}

func (s *SessionState) SetValidated(validated bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.validated = validated
}

func (s *SessionState) Validated() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.validated
}

func (s *SessionState) SetEmail(email string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.email = email
}

func (s *SessionState) Email() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.email
}

type Callback func(ctx context.Context, state *SessionState, request *rails.Message, send func(message *rails.Message)) error

type Handler struct {
	sessions     map[string]net.Conn
	sessionsLock sync.RWMutex

	callbacks     []Callback
	callbacksLock sync.RWMutex
}

func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		w.Write([]byte("Bad Protocol"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	remoteAddr := r.RemoteAddr

	h.sessionsLock.Lock()
	h.sessions[remoteAddr] = conn
	h.sessionsLock.Unlock()

	context.AfterFunc(ctx, func() {
		h.sessionsLock.Lock()
		delete(h.sessions, remoteAddr)
		h.sessionsLock.Unlock()
	})

	ss := &SessionState{
		remoteAddr: remoteAddr,
	}

	for {
		data, err := wsutil.ReadClientBinary(conn)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read message")
			break
		}

		message := new(rails.Message)
		if err := proto.Unmarshal(data, message); err != nil {
			log.Error().Err(err).Bytes("data", data).Msg("Failed to unmarshal message")
			continue
		}

		h.callbacksLock.RLock()
		for _, callback := range h.callbacks {
			go callback(ctx, ss, message, func(response *rails.Message) {
				defer func() {
					if err := recover(); err != nil {
						log.Error().Interface("err", err).Msg("Failed to send response")
					}
				}()

				data, err := proto.Marshal(response)
				if err != nil {
					log.Error().Err(err).Msg("Failed to marshal response")
					return
				}

				if err := wsutil.WriteServerBinary(conn, data); err != nil {
					log.Error().Err(err).Msg("Failed to write response")
				}
			})
		}
		h.callbacksLock.RUnlock()
	}
}

func (h *Handler) Broadcast(message *rails.Message) {
	data, err := proto.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal message")
		return
	}

	h.sessionsLock.RLock()
	defer h.sessionsLock.RUnlock()

	for _, conn := range h.sessions {
		go func() {
			if err := wsutil.WriteServerBinary(conn, data); err != nil {
				log.Error().Err(err).Msg("Failed to write message")
			}
		}()
	}
}

func (h *Handler) RegisterCallback(callback Callback) {
	h.callbacksLock.Lock()
	h.callbacks = append(h.callbacks, callback)
	h.callbacksLock.Unlock()
}
