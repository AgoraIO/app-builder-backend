package routes

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// DesktopOAuthHandler is a REST route that is called when the oauth provider redirects to here and provides the code
func (o *Router) DesktopOAuthHandler(w http.ResponseWriter, r *http.Request) {
	redirect, token, err := Handler(w, r, o.DB, "desktop")
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, err)
		return
	}

	newURL, err := url.Parse(*redirect)
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, err)
		return
	}

	query := newURL.Query()
	query.Set("token", *token)
	newURL.RawQuery = query.Encode()

	http.Redirect(w, r, newURL.String(), http.StatusSeeOther)
}
