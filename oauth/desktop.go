package oauth

import (
	"fmt"
	"html/template"
	"net/http"
)

// DesktopOAuthHandler is a REST route that is called when the oauth provider redirects to here and provides the code
func (o *Router) DesktopOAuthHandler(w http.ResponseWriter, r *http.Request) {
	_, token, err := o.Handler(w, r, "desktop")
	if err != nil {
		fmt.Fprint(w, "Internal Server Error")
		return
	}

	t, err := template.ParseFiles("web/desktop.html")
	if err != nil {
		fmt.Fprint(w, "Internal Server Error")
		return
	}

	t.Execute(w, TokenTemplate{
		Token: *token,
	})
}
