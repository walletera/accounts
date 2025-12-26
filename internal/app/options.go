package app

import "log/slog"

type Option func(app *App)

func WithPublicAPIConfig(config PublicAPIConfig) func(a *App) {
	return func(a *App) {
		a.publicAPIConfig = NewOptional[PublicAPIConfig](config)
	}
}

func WithMongoDBURL(url string) func(a *App) { return func(a *App) { a.mongodbURL = url } }

func WithLogHandler(handler slog.Handler) func(app *App) {
	return func(app *App) { app.logHandler = handler }
}
