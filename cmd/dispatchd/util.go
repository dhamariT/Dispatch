package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

var slugRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])?$`)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, out any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func validSlug(s string) bool {
	return slugRe.MatchString(s)
}
