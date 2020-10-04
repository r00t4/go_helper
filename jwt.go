package helper

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"time"
)

const (
	jwtPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvt813ynxOv9FIg5Kdk8GTdSWa6sWipUU4FwYczNU43ckNOh5
Jk/pA5OPA1weh4ZOM7ZfFCeHyfkRYVAW8UcjtCIljso7/eCuAkmwXM/5vw2rJYYN
smzlgGOrjbGcPPD4PwaCxsTd2cEKvggjzotoss3Ea30g408rZFVwI4PQBnSBMDkf
RihHbl8/3Xy/0enkRKhG31R7pu4tFFZ78rzUceAsd9jFYLDFMf8qN2vHGfRNE8i2
e+x7bK1GV3sh5MBixb2OsORq7D0CCfjTspDB/Jg+Dn3t+b0+/SkiG2a8u6HRso8c
sUmBrb1hA5l+I49mKntcFPqLGB3XY+fKBbCbMwIDAQABAoIBACOl7JnRa4xpQLAr
mxydhb/jhHR3b65SSaPdj3N0ktYo2kpHYNkW854HYR5vhgQpwVFHLlrFR0chjW1v
V9mYP8LU3c7dVncED3u954JuFWbpVp2be9NnIzXnZ5L/KP74wmSDAsm82vJga3Ey
c/2Pa+55H8YziIDruF701gzMAX4y0TG732y+11MeRTOCL/w6C3rLVZ7TrvaymSIA
EPBXnD6GqNgzQco75Ivy+l4Wf10mlA270NrftBvopR2JWBYlnLEt5Rtu1eZP5rA7
RjwOrxDRB/swnfIoL8xufnX3OZQ1w9JQwE4MlOXUpOLF7AkzLo9zhU4aVpwYU48T
J0r42gECgYEA9Fznp2UbrG4J2cvSApEbWfdESFMGTBLe+uCHTtO9NMz4DMQEghzN
/nsJG8xRiSB/c9ozV4cjQNOIkeSDbY9TC4k/5e8Xhb1F6/AVLyk0bDMkkOQ5Y3e0
5LnGn/3fKudkY+K+nH+ca2WpEUy3W0UUoxAg+cwGqlzP0kr+ursCexECgYEAx/Yu
pDS+4j0fA1AXFvhQjeHUFxlUQW6MBUOt/EFJposUJGlJLZDm0UJZT/xRgtWGjy3f
sM2fFeEosvsy0XVEHt/onNl6Ng8NkepUSAW+fJ2cpzh6OiLdz/XIrjHS34+nAeN2
bXqxYC/JILWnQ66kbgyW2ri64/GJ7YSGNSUyigMCgYEAjVKFruvsm0ZwcANOi6l8
FgXI+cL6a1immJTt7ArM7BJ2int61/zsrXZeiDMcHKAs1cWl18MSAlXUL/vmfqBb
ONrBl6s1AWW7YH5S4hmEdecGCL3U6s+6UGWYl8LtJBT6nEHwVvX+cqYypwylJiXH
j56uU4lJeZF/p3Ez7K5m+uECgYEApYVB/IjwzUN87XgZdNkdjSS3NFuyI+uHGkB4
v8unVKXhiXZhrcc5WVTLq2sYae2oUdLOTIMYwbq8vtMysLGaLth3q4ZWJHN3byaC
l4+xq2OoLb+RZZhA9gjlElSJ0qcNvoF0IZGjTBSiL2JOz7a7w6DGKs0pXtAstSCz
G0DsQdMCgYB2FwAu1wn5+IJE8uDFvEclIv/it0cLCrz4qugvAqhaVPRBHDLesu/l
6/O5W3Y1+d3Q9xycJE53n8P0FCJ5ZvrYVSIX9PSkW+AxSI0TcfY1hFYBjPfrbuQc
X9uufsq9KdbdKODxnwmGvpCO79p27EGTZ9Rp4csEi0f1qmfrlCL0mA==
-----END RSA PRIVATE KEY-----`
	jwtPublicKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAvt813ynxOv9FIg5Kdk8GTdSWa6sWipUU4FwYczNU43ckNOh5Jk/p
A5OPA1weh4ZOM7ZfFCeHyfkRYVAW8UcjtCIljso7/eCuAkmwXM/5vw2rJYYNsmzl
gGOrjbGcPPD4PwaCxsTd2cEKvggjzotoss3Ea30g408rZFVwI4PQBnSBMDkfRihH
bl8/3Xy/0enkRKhG31R7pu4tFFZ78rzUceAsd9jFYLDFMf8qN2vHGfRNE8i2e+x7
bK1GV3sh5MBixb2OsORq7D0CCfjTspDB/Jg+Dn3t+b0+/SkiG2a8u6HRso8csUmB
rb1hA5l+I49mKntcFPqLGB3XY+fKBbCbMwIDAQAB
-----END RSA PUBLIC KEY-----`
)

type Claims struct {
	PhoneNumber string `json:"phone"`
	IsManager   bool   `json:"mng"`
	jwt.StandardClaims
}

func Middleware(w http.ResponseWriter, r *http.Request, next func(w http.ResponseWriter, r *http.Request) (*Response, error)) (*Response, error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 {
		return nil, MiddleHttpError{http.StatusUnauthorized, ErrInvalidToken}
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(authHeader[1], claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtPrivateKey), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, MiddleHttpError{http.StatusUnauthorized, err}
		}
		return nil, MiddleHttpError{http.StatusBadRequest, err}
	}
	if !token.Valid {
		return nil, MiddleHttpError{http.StatusUnauthorized, ErrInvalidToken}
	}
	ctx := context.WithValue(r.Context(), "props", claims)
	return next(w, r.WithContext(ctx))
}

func GenerateToken(claims *Claims) (string, error) {
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		return "", MiddleHttpError{400, ErrUnexpiredToken}
	}

	expirationTime := time.Now().Add(time.Hour)
	// Create the JWT claims, which includes the username and expiry time
	claims.ExpiresAt = expirationTime.Unix()

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	return token.SignedString([]byte(jwtPrivateKey))
}
