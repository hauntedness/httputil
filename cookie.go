package httputil

import "net/http"

func ParseSetCookie(line string) (*http.Cookie, error) {
	return http.ParseSetCookie(line)
}
