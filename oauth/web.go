package oauth

import (
	"fmt"
	"log"
	"net/http"
)

// WebOAuthHandler is a REST route that is called when the oauth provider redirects to here and provides the code
func (o *Router) WebOAuthHandler(w http.ResponseWriter, r *http.Request) {
	redirect, err := Handler(w, r, o.DB)
	if err != nil {
		log.Panic(err)
		fmt.Fprint(w, err)
		return
	}

	http.Redirect(w, r, *redirect, http.StatusSeeOther)
}
