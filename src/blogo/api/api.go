package api

import (
	"github.com/fulldump/golax"
	"github.com/fulldump/kip"

	"googleapi"

	"blogo/articles"
	"blogo/constants"
	"blogo/home"
	"blogo/login"
	"blogo/sessions"
	"blogo/users"

	"blogo/audits"

	"blogo/statics"

	"blogo/config"
	"encoding/json"
	"net/http"

	"github.com/fulldump/apidoc"
	"github.com/fulldump/goaudit"
)

func Build(articles_dao, sessions_dao, users_dao, audits_dao *kip.Dao, g *googleapi.GoogleApi, google_analytics, statics_dir string, channel_audits chan *goaudit.Audit, config *config.Config) *golax.Api {

	api := golax.NewApi()

	api.Root.Interceptor(goaudit.InterceptorAudit2Channel(channel_audits))
	api.Root.Interceptor(goaudit.InterceptorAudit(&goaudit.Service{
		Name:    constants.SERVICE,
		Version: constants.VERSION,
		Commit:  constants.COMMIT,
	}))
	api.Root.Interceptor(sessions.NewSessionInterceptor(sessions_dao))
	api.Root.Interceptor(users.NewUserInterceptor(users_dao))
	api.Root.Interceptor(golax.InterceptorError)

	home.Build(api.Root, articles_dao, g, google_analytics)

	// Connect articles API
	articles.Build(api.Root, articles_dao)

	// Connect sessions API
	sessions.Build(api.Root, sessions_dao)

	// Connect users API
	users.Build(api.Root, users_dao)

	// Connect login API
	login.Build(api.Root, users_dao, sessions_dao, g)

	// Conenct audits API
	audits.Build(api.Root, audits_dao)

	// Documentation
	doc := apidoc.Build(api, api.Root)
	doc.Title = "BloGo"
	doc.Subtitle = "API Reference v" + constants.VERSION

	// Configuration
	api.Root.Node("config").Method("GET", func(c *golax.Context) {
		user := users.GetUser(c)
		if nil == user || !user.Scopes.Admin {
			c.Error(http.StatusForbidden, "You are not allowed")
			return
		}

		json.NewEncoder(c.Response).Encode(config)
	})

	// Static files
	statics.Build(api.Root, statics_dir)

	return api
}
