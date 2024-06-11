package www

import (
	"context"
)

type ctxKey string

const (
	pageTitleKey            ctxKey = "pageTitleKey"
	pageLanguageKey                = "pageLanguageKey"
	preferredColorSchemaKey        = "preferredColorSchemaKey"
	usernameKey                    = "usernameKey"
)

const DefaultPageTitle = "Chat Room"

func GetPageTitle(ctx context.Context) string {
	title, ok := ctx.Value(pageTitleKey).(string)
	if !ok {
		return DefaultPageTitle
	}
	return title
}

func SetPageTitle(ctx context.Context, title string) context.Context {
	return context.WithValue(ctx, pageTitleKey, title)
}

func GetPageLanguage(ctx context.Context) string {
	language, _ := ctx.Value(pageLanguageKey).(string)
	return language
}

func SetPageLanguage(ctx context.Context, language string) context.Context {
	return context.WithValue(ctx, pageLanguageKey, language)
}

func GetUserName(ctx context.Context) (string, bool) {
	userName, ok := ctx.Value(usernameKey).(string)
	return userName, ok
}

func SetUserName(ctx context.Context, userName string) context.Context {
	return context.WithValue(ctx, usernameKey, userName)
}

func GetPreferredColorSchema(ctx context.Context) string {
	colorSchema, _ := ctx.Value(preferredColorSchemaKey).(string)
	return colorSchema
}

func SetPreferredColorSchema(ctx context.Context, colorSchema string) context.Context {
	return context.WithValue(ctx, preferredColorSchemaKey, colorSchema)
}

func IsCurrentUser(ctx context.Context, username string) bool {
	currentUsername, ok := GetUserName(ctx)
	return ok && currentUsername == username
}
