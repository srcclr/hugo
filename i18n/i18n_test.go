// Copyright 2017 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package i18n

import (
	"testing"

	"io/ioutil"
	"os"

	"log"

	"github.com/nicksnyder/go-i18n/i18n/bundle"
	"github.com/srcclr/hugo/config"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

var logger = jww.NewNotepad(jww.LevelError, jww.LevelError, os.Stdout, ioutil.Discard, "", log.Ldate|log.Ltime)

type i18nTest struct {
	data                             map[string][]byte
	args                             interface{}
	lang, id, expected, expectedFlag string
}

var i18nTests = []i18nTest{
	// All translations present
	{
		data: map[string][]byte{
			"en.yaml": []byte("- id: \"hello\"\n  translation: \"Hello, World!\""),
			"es.yaml": []byte("- id: \"hello\"\n  translation: \"¡Hola, Mundo!\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "¡Hola, Mundo!",
		expectedFlag: "¡Hola, Mundo!",
	},
	// Translation missing in current language but present in default
	{
		data: map[string][]byte{
			"en.yaml": []byte("- id: \"hello\"\n  translation: \"Hello, World!\""),
			"es.yaml": []byte("- id: \"goodbye\"\n  translation: \"¡Adiós, Mundo!\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "Hello, World!",
		expectedFlag: "[i18n] hello",
	},
	// Translation missing in default language but present in current
	{
		data: map[string][]byte{
			"en.yaml": []byte("- id: \"goodybe\"\n  translation: \"Goodbye, World!\""),
			"es.yaml": []byte("- id: \"hello\"\n  translation: \"¡Hola, Mundo!\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "¡Hola, Mundo!",
		expectedFlag: "¡Hola, Mundo!",
	},
	// Translation missing in both default and current language
	{
		data: map[string][]byte{
			"en.yaml": []byte("- id: \"goodbye\"\n  translation: \"Goodbye, World!\""),
			"es.yaml": []byte("- id: \"goodbye\"\n  translation: \"¡Adiós, Mundo!\""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "",
		expectedFlag: "[i18n] hello",
	},
	// Default translation file missing or empty
	{
		data: map[string][]byte{
			"en.yaml": []byte(""),
		},
		args:         nil,
		lang:         "es",
		id:           "hello",
		expected:     "",
		expectedFlag: "[i18n] hello",
	},
	// Context provided
	{
		data: map[string][]byte{
			"en.yaml": []byte("- id: \"wordCount\"\n  translation: \"Hello, {{.WordCount}} people!\""),
			"es.yaml": []byte("- id: \"wordCount\"\n  translation: \"¡Hola, {{.WordCount}} gente!\""),
		},
		args: struct {
			WordCount int
		}{
			50,
		},
		lang:         "es",
		id:           "wordCount",
		expected:     "¡Hola, 50 gente!",
		expectedFlag: "¡Hola, 50 gente!",
	},
}

func doTestI18nTranslate(t *testing.T, data map[string][]byte, lang, id string, args interface{}, cfg config.Provider) string {
	i18nBundle := bundle.New()

	for file, content := range data {
		err := i18nBundle.ParseTranslationFileBytes(file, content)
		if err != nil {
			t.Errorf("Error parsing translation file: %s", err)
		}
	}

	translator := NewTranslator(i18nBundle, cfg, logger)

	f := translator.Func(lang)

	translated := f(id, args)

	return translated
}

func TestI18nTranslate(t *testing.T) {
	var actual, expected string
	v := viper.New()
	v.SetDefault("defaultContentLanguage", "en")

	// Test without and with placeholders
	for _, enablePlaceholders := range []bool{false, true} {
		v.Set("enableMissingTranslationPlaceholders", enablePlaceholders)

		for _, test := range i18nTests {
			if enablePlaceholders {
				expected = test.expectedFlag
			} else {
				expected = test.expected
			}
			actual = doTestI18nTranslate(t, test.data, test.lang, test.id, test.args, v)
			require.Equal(t, expected, actual)
		}
	}
}
