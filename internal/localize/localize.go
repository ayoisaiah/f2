package localize

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	//go:embed all:i18n
	i18nFS    embed.FS
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
)

// getMessagesLocale determines the effective locale for program messages
// by checking standard environment variables in the correct order.
func getMessagesLocale() string {
	if locale := os.Getenv("LC_ALL"); locale != "" {
		return locale
	}

	if locale := os.Getenv("LC_MESSAGES"); locale != "" {
		return locale
	}

	if locale := os.Getenv("LANG"); locale != "" {
		return locale
	}

	return ""
}

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	fs, err := i18nFS.ReadDir("i18n")
	if err != nil {
		panic(err)
	}

	for _, f := range fs {
		path := fmt.Sprintf("i18n/%s", f.Name())

		_, err = bundle.LoadMessageFileFS(i18nFS, path)
		if err != nil {
			panic(err)
		}
	}

	lang := language.English.String()

	langEnv := getMessagesLocale()

	if langEnv != "" {
		lang = langEnv
	}

	before, _, found := strings.Cut(langEnv, "_")
	if found {
		lang = before
	}

	localizer = i18n.NewLocalizer(bundle, lang)
}

func T(id string) string {
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: id,
	})
}

func TWithOpts(lc *i18n.LocalizeConfig) string {
	return localizer.MustLocalize(lc)
}
