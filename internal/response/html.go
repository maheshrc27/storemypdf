package response

import (
	"context"
	"net/http"

	"github.com/maheshrc27/storemypdf/templates/components"
)

func HTML(w http.ResponseWriter, msg string) {
	html := components.Error(msg)
	html.Render(context.Background(), w)
}
