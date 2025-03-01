package rails

import (
	"context"
	"crypto/sha3"
	"encoding/base64"
	"hash"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/blake2b"

	"github.com/snowmerak/keycl/lib/store"
	"github.com/snowmerak/keycl/lib/store/queries"
	"github.com/snowmerak/keycl/model/gen/rails"
)

func HashPassword(step1, step2 hash.Hash, salt string, password string) string {
	saltBytes := []byte(salt)
	value := []byte(password)
	step := 0
	for step < 12 {
		step1.Reset()
		step1.Write(saltBytes)
		step1.Write(value)
		value = step1.Sum(nil)
	}
	step = 0
	for step < 12 {
		step2.Reset()
		step2.Write(saltBytes)
		step2.Write(value)
		value = step2.Sum(nil)
	}
	return base64.URLEncoding.EncodeToString(saltBytes)
}

func RegisterDefaultHandlers(h *Handler, st *store.Store) error {
	passwordHash1, passwordHash2 := func() hash.Hash {
		return sha3.New512()
	}, func() hash.Hash {
		h, _ := blake2b.New384(nil)
		return h
	}

	h.RegisterCallback(func(ctx context.Context, state *SessionState, request *rails.Message, send func(*rails.Message)) error {
		if !state.lock.TryLock() {
			send(CommonResponse(false, "Already operating by another request"))
			return nil
		}

		switch req := request.Request.(type) {
		case *rails.Message_LoginRequest:
			rs := RequestSession{
				passwordHash1: passwordHash1,
				passwordHash2: passwordHash2,
				store:         st,
				send:          send,
			}
			defaultLoginRequest(ctx, &rs, req.LoginRequest)
		}
		return nil
	})

	return nil
}

type RequestSession struct {
	passwordHash1 func() hash.Hash
	passwordHash2 func() hash.Hash
	store         *store.Store
	send          func(*rails.Message)
}

func defaultLoginRequest(ctx context.Context, rs *RequestSession, request *rails.LoginRequest) {
	email := request.GetEmail()
	password := request.GetPassword()
	salt, registeredHash := "", ""
	if err := rs.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		result, err := q.GetUserPassword(ctx, email)
		if err != nil {
			return err
		}

		salt = result.Salt
		registeredHash = result.Hash

		return nil
	}); err != nil {
		log.Error().Err(err).Str("email", email).Msg("Failed to get user password")
		rs.send(CommonResponse(false, "Invalid email or password"))
		return
	}

	hashed := HashPassword(rs.passwordHash1(), rs.passwordHash2(), salt, password)
	if hashed != registeredHash {
		log.Error().Str("email", email).Str("hashed", hashed).Msg("Invalid email password")
		rs.send(CommonResponse(false, "Invalid email or password"))
		return
	}

	return
}
