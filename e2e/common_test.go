package e2e_test

import (
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

// ctxFunction is a context based evaluation callback
var ctxFunction = func(this memprovider.InMemoryFlag, evalCtx openfeature.FlattenedContext) (
	any, openfeature.ProviderResolutionDetail) {

	defaultValue := this.Variants[this.DefaultVariant]
	defaultResolution := openfeature.ProviderResolutionDetail{
		Reason:  openfeature.DefaultReason,
		Variant: this.DefaultVariant,
	}

	expects := openfeature.FlattenedContext{
		"fn":       "Sulisław",
		"ln":       "Świętopełk",
		"age":      int64(29),
		"customer": false,
	}

	for k, v := range expects {
		if v != evalCtx[k] {
			return defaultValue, defaultResolution
		}
	}

	return this.Variants["internal"], openfeature.ProviderResolutionDetail{
		Reason:  openfeature.TargetingMatchReason,
		Variant: "internal",
	}
}

var memoryFlags = map[string]memprovider.InMemoryFlag{
	"boolean-flag": {
		Key:            "boolean-flag",
		State:          memprovider.Enabled,
		DefaultVariant: "on",
		Variants: map[string]any{
			"on":  true,
			"off": false,
		},
		ContextEvaluator: nil,
	},
	"string-flag": {
		Key:            "string-flag",
		State:          memprovider.Enabled,
		DefaultVariant: "greeting",
		Variants: map[string]any{
			"greeting": "hi",
			"parting":  "bye",
		},
		ContextEvaluator: nil,
	},
	"integer-flag": {
		Key:            "integer-flag",
		State:          memprovider.Enabled,
		DefaultVariant: "ten",
		Variants: map[string]any{
			"one": 1,
			"ten": 10,
		},
		ContextEvaluator: nil,
	},
	"float-flag": {
		Key:            "float-flag",
		State:          memprovider.Enabled,
		DefaultVariant: "half",
		Variants: map[string]any{
			"tenth": 0.1,
			"half":  0.5,
		},
		ContextEvaluator: nil,
	},
	"object-flag": {
		Key:            "object-flag",
		State:          memprovider.Enabled,
		DefaultVariant: "template",
		Variants: map[string]any{
			"empty": map[string]any{},
			"template": map[string]any{
				"showImages":    true,
				"title":         "Check out these pics!",
				"imagesPerPage": 100,
			},
		},
		ContextEvaluator: nil,
	},
	"wrong-flag": {
		Key:            "wrong-flag",
		State:          memprovider.Enabled,
		DefaultVariant: "one",
		Variants: map[string]any{
			"one": "uno",
			"two": "dos",
		},
		ContextEvaluator: nil,
	},
	"context-aware": {
		Key:            "context-aware",
		State:          memprovider.Enabled,
		DefaultVariant: "external",
		Variants: map[string]any{
			"internal": "INTERNAL",
			"external": "EXTERNAL",
		},
		ContextEvaluator: &ctxFunction,
	},
}
