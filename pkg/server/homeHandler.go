package server

import (
	"fmt"
	"net/http"
)

//HomeHandler handles /.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "")
}
