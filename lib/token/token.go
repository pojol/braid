package token

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

var (
	kTokenSignKey = "token"
)

type Claims struct {
	EntityID string
}

func (c Claims) Valid() error {
	// entityID
	if c.EntityID == "" {
		return fmt.Errorf("openID is empty")
	}
	return nil
}

func makeTokenKey() []byte {
	return []byte(kTokenSignKey)
}

// TokenCreate - 通过 entityID 构建出一个 包含了 entityID 的 token
func Create(entityID string) (string, error) {
	if len(entityID) <= 0 {
		return "", fmt.Errorf("createToken error entityID is empty")
	}

	claims := Claims{
		EntityID: entityID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(makeTokenKey())
	if err != nil {
		return "", fmt.Errorf("createToken error:%v", err)
	}
	return s, nil
}

// TokenParse - 从被包装过的 token 中获取 entityID
func Parse(token string) (string, error) {
	c := &Claims{}
	_, err := jwt.ParseWithClaims(token, c, func(*jwt.Token) (interface{}, error) {
		return makeTokenKey(), nil
	})

	if err != nil {
		return "", fmt.Errorf("jwt.Parse %v error:%v", token, err)
	}
	return c.EntityID, nil
}
