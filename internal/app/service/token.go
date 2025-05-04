package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/xsyro/goapi/config"
	"github.com/xsyro/goapi/internal/app/api/auth"
	"github.com/xsyro/goapi/internal/app/api/presenter"
	model "github.com/xsyro/goapi/internal/app/repo/sqlc"
	"github.com/xsyro/goapi/internal/utils"
)

const (
	ExpireAccessMinutes  = time.Minute * 30
	ExpireRefreshMinutes = time.Minute * 2 * 60
	AutoLogoffMinutes    = time.Minute * 15
)

type JwtCustomClaims struct {
	ID       uuid.UUID `json:"id"`
	UID      string    `json:"uid"`
	Username string    `json:"username"`
	Claims   map[string]interface{}
}

type CachedTokens struct {
	AccessUID  string `json:"access_token"`
	RefreshUID string `json:"refresh_token"`
}

type tokenService struct {
	redis     *redis.Client
	cfg       *config.Conf
	jwtAuth   *auth.JWTAuth
	presenter presenter.Presenter
}

func NewTokenService(redis *redis.Client, cfg *config.Conf,
	jwtAuth *auth.JWTAuth, presenter presenter.Presenter) *tokenService {
	return &tokenService{
		redis:     redis,
		cfg:       cfg,
		jwtAuth:   jwtAuth,
		presenter: presenter,
	}
}

func (s *tokenService) GenerateTokenPair(ctx context.Context, user *model.User) (
	accessToken,
	refreshToken string,
	exp time.Time,
	err error,
) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	var accessUID, refreshUID string
	if accessToken, accessUID, exp, err = s.createToken(user, ExpireAccessMinutes); err != nil {
		return
	}

	if refreshToken, refreshUID, _, err = s.createToken(user, ExpireRefreshMinutes); err != nil {
		return
	}

	cacheJSON, err := json.Marshal(CachedTokens{
		AccessUID:  accessUID,
		RefreshUID: refreshUID,
	})
	s.redis.Set(ctx, fmt.Sprintf("token-%s", user.ID), string(cacheJSON), AutoLogoffMinutes)
	return
}

func (s *tokenService) ParseToken(ctx context.Context, tokenString string) (*JwtCustomClaims, error) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	token, err := s.jwtAuth.Decode(tokenString)
	if err != nil {
		return nil, err
	}

	claims, err := token.AsMap(ctx)
	if err != nil {
		return nil, err
	}

	customClaims := convertMapToCustomClaims(claims)
	return &customClaims, nil
}

func (s *tokenService) ValidateToken(ctx context.Context, claims map[string]interface{}, isRefresh bool) error {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	cusClaims := convertMapToCustomClaims(claims)
	cacheJSON, _ := s.redis.Get(ctx, fmt.Sprintf("token-%s", cusClaims.ID)).Result()
	cachedTokens := new(CachedTokens)
	err := json.Unmarshal([]byte(cacheJSON), cachedTokens)

	var tokenUID string
	if isRefresh {
		tokenUID = cachedTokens.RefreshUID
	} else {
		tokenUID = cachedTokens.AccessUID
	}

	if err != nil || tokenUID != cusClaims.UID {
		return auth.ErrNoTokenFound
	}
	return nil
}

func (s *tokenService) createToken(user *model.User, expireMinutes time.Duration) (
	token string,
	uid string,
	exp time.Time,
	err error,
) {
	exp = time.Now().Add(expireMinutes)
	uid = uuid.New().String()
	claims := map[string]interface{}{
		"id":       user.ID.String(),
		"username": user.Email,
		"uid":      uid,
		"exp":      auth.ExpireIn(expireMinutes),
	}

	_, token, err = s.jwtAuth.Encode(claims)
	return
}

func (s *tokenService) ProlongToken(ctx context.Context, userId string) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	s.redis.Expire(ctx, fmt.Sprintf("token-%s", userId), AutoLogoffMinutes)
}

func (s *tokenService) RemoveToken(ctx context.Context, userId string) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	s.redis.Del(ctx, fmt.Sprintf("token-%s", userId))
}

func (s *tokenService) Verifier() func(http.Handler) http.Handler {
	return auth.Verifier(s.jwtAuth)
}

func (s *tokenService) Validator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		_, claims, err := auth.FromContext(ctx)
		// Check for token expiration or invalidity
		if err != nil {
			if errors.Is(err, auth.ErrExpired) {
				s.presenter.Error(w, r, utils.ErrorAuth(err))
				return
			}
			s.presenter.Error(w, r, utils.ErrorAuth(err))
			return
		}

		// Check for long inactivity even if the access token is not expired
		if err := s.ValidateToken(r.Context(), claims, false); err != nil {
			if errors.Is(err, auth.ErrNoTokenFound) {
				// Auto logout due to prolonged inactivity
				s.presenter.Error(w, r, utils.ErrorAuth(err))
				return
			}
			// Handle other token validation errors
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Prolong the Redis TTL of the current token pair
		go s.ProlongToken(r.Context(), claims["id"].(string))
		next.ServeHTTP(w, r)
	})
}

func convertMapToCustomClaims(claims map[string]interface{}) JwtCustomClaims {
	return JwtCustomClaims{
		ID:       uuid.MustParse(claims["id"].(string)),
		UID:      claims["uid"].(string),
		Username: claims["username"].(string),
		Claims:   claims,
	}
}
