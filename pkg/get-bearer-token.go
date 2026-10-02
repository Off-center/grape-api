package pkg

import "net/http"

func GetBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")

	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	return ""
}
