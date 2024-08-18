package api

import (
	"net/http"
	"strconv"
)

func IntPathValue(r *http.Request, name string) (int, error) {
	value := r.PathValue(name)

	result, err := strconv.Atoi(value)
	if err != nil {
		return result, err
	}

	return result, err
}
