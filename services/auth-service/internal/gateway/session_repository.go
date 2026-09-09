package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"kaffein/auth-service/internal/interfaces/IRepository"
	"kaffein/auth-service/utils/constants"

	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"
)

type sessionRepository struct{ client *redislib.Client }

var rotateSession = redislib.NewScript(`
local current = redis.call('GET', KEYS[1])
if not current or cjson.decode(current).refresh_token ~= ARGV[1] then return 0 end
redis.call('DEL', KEYS[1])
redis.call('SREM', KEYS[3], ARGV[3])
if tonumber(ARGV[5]) > 0 then redis.call('PSETEX', KEYS[2], ARGV[5], '1') end
redis.call('PSETEX', KEYS[4], ARGV[6], ARGV[2])
redis.call('SADD', KEYS[5], ARGV[4])
redis.call('PEXPIRE', KEYS[5], ARGV[6])
return 1
`)

func NewSessionRepository(client *redislib.Client) IRepository.SessionRepository {
	return &sessionRepository{client: client}
}
func refreshKey(sid uuid.UUID) string     { return "auth:refresh:" + sid.String() }
func revokedKey(sid uuid.UUID) string     { return "auth:revoked:" + sid.String() }
func userSessionsKey(id uuid.UUID) string { return "auth:user-sessions:" + id.String() }

func (r *sessionRepository) Save(ctx context.Context, sid uuid.UUID, session IRepository.Session, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, refreshKey(sid), data, ttl)
	pipe.SAdd(ctx, userSessionsKey(session.UserID), sid.String())
	pipe.Expire(ctx, userSessionsKey(session.UserID), ttl)

	_, err = pipe.Exec(ctx)

	return err
}

func (r *sessionRepository) Get(ctx context.Context, sid uuid.UUID) (*IRepository.Session, error) {
	data, err := r.client.Get(ctx, refreshKey(sid)).Bytes()
	if errors.Is(err, redislib.Nil) {
		return nil, errors.New(constants.ErrSessionNotFound)
	}

	if err != nil {
		return nil, err
	}

	var session IRepository.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *sessionRepository) Rotate(ctx context.Context, oldSID uuid.UUID, oldSession IRepository.Session, newSID uuid.UUID, newSession IRepository.Session, ttl time.Duration) (bool, error) {
	newData, err := json.Marshal(newSession)
	if err != nil {
		return false, err
	}

	result, err := rotateSession.Run(ctx, r.client, []string{
		refreshKey(oldSID),
		revokedKey(oldSID),
		userSessionsKey(oldSession.UserID),
		refreshKey(newSID),
		userSessionsKey(newSession.UserID),
	},
		oldSession.RefreshToken,
		newData,
		oldSID.String(),
		newSID.String(),
		time.Until(oldSession.AccessExpiresAt).Milliseconds(),
		ttl.Milliseconds(),
	).Int()

	return result == 1, err
}

func (r *sessionRepository) Revoke(ctx context.Context, sid uuid.UUID) error {
	session, err := r.Get(ctx, sid)
	if err != nil && err.Error() != constants.ErrSessionNotFound {
		return err
	}

	if session == nil {
		return nil
	}

	ttl := time.Until(session.AccessExpiresAt)
	return r.revoke(ctx, sid, session.UserID, ttl)
}

func (r *sessionRepository) RevokeWithTTL(ctx context.Context, sid uuid.UUID, ttl time.Duration) error {
	session, err := r.Get(ctx, sid)
	if err != nil && err.Error() != constants.ErrSessionNotFound {
		return err
	}

	var userID uuid.UUID
	if session != nil {
		userID = session.UserID
	}

	return r.revoke(ctx, sid, userID, ttl)
}

func (r *sessionRepository) revoke(ctx context.Context, sid, userID uuid.UUID, ttl time.Duration) error {
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, refreshKey(sid))
	if userID != uuid.Nil {
		pipe.SRem(ctx, userSessionsKey(userID), sid.String())
	}

	if ttl > 0 {
		pipe.Set(ctx, revokedKey(sid), "1", ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (r *sessionRepository) IsRevoked(ctx context.Context, sid uuid.UUID) (bool, error) {
	n, err := r.client.Exists(ctx, revokedKey(sid)).Result()
	return n > 0, err
}

func (r *sessionRepository) RevokeUser(ctx context.Context, userID uuid.UUID) error {
	sids, err := r.client.SMembers(ctx, userSessionsKey(userID)).Result()
	if err != nil {
		return err
	}

	for _, value := range sids {
		sid, parseErr := uuid.Parse(value)
		if parseErr != nil {
			continue
		}

		if err := r.Revoke(ctx, sid); err != nil {
			return err
		}
	}
	return r.client.Del(ctx, userSessionsKey(userID)).Err()
}
