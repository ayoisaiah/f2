package localize

import (
	"embed"
	"fmt"
	"os"

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

	if langEnv := os.Getenv("LANG"); langEnv != "" {
		lang = langEnv
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
