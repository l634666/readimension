package route

import (
	"encoding/gob"
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packr"
	mw "github.com/kyicy/readimension/middleware"
	"github.com/kyicy/readimension/model"
	"github.com/kyicy/readimension/utility/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	validator "gopkg.in/go-playground/validator.v9"
)

type userData map[string]string

var validate *validator.Validate

func init() {
	gob.Register(&userData{})
	validate = validator.New()
}

// Register registers all handler to a url path
func Register(e *echo.Echo) {
	render := getRender()
	e.Renderer = render

	e.GET("/", getExplorerRoot)
	e.GET("/u/explorer", getExplorerRoot)
	e.GET("/u/explorer/:list_id", getExplorer)

	e.GET("/sign-up", getSignUp)
	e.POST("/sign-up", postSignUp)

	e.GET("/sign-in", getSignIn)
	e.POST("/sign-in", postSignIn)
	e.GET("/sign-out", getSignOut)

	conf := config.Get()
	if conf.ServeStatic {
		e.Static("/covers", "covers")
		e.Static("/books", "books")
	}

	userGroup := e.Group("/u", mw.UserAuth)
	userGroup.DELETE("/explorer/:list_id", deleteExplorer)
	userGroup.POST("/:list_id/books/new", postBooksNew)
	userGroup.POST("/:list_id/books/new/chunksdone", postChunksDone)
	userGroup.POST("/lists/:id/child/new", postListChildNew)

	box := packr.NewBox("../bib")
	box.Walk(func(path string, f packr.File) error {
		extName := filepath.Ext(path)
		mt := mime.TypeByExtension(extName)

		e.GET("/u/"+path, func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "max-age=3600")
			r := strings.NewReader(box.String(path))
			return c.Stream(http.StatusOK, mt, r)
		})
		return nil
	})
	e.GET("/u/i", func(c echo.Context) error {
		tc := newTemplateCommon(c, "")
		return c.Render(http.StatusOK, "bibi", tc)
	})

	e.GET("/u/i/", func(c echo.Context) error {
		tc := newTemplateCommon(c, "")
		return c.Render(http.StatusOK, "bibi", tc)
	})
}

func getSessionUser(c echo.Context) (*model.User, error) {
	userID, err := getSessionUserID(c)
	if err != nil {
		return nil, err
	}
	var userRecord model.User
	model.DB.Where("id = ?", userID).First(&userRecord)
	return &userRecord, nil
}

func getSessionUserID(c echo.Context) (string, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return "", err
	}
	if ud, flag := sess.Values["userData"]; flag {
		userDataPointer := ud.(*userData)
		userID := (*userDataPointer)["id"]
		return userID, nil
	}
	return "", errors.New("session not found")

}
