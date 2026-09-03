package xconfig

import (
	"github.com/AeonDigital/Go-Core-xconfig/pkg/xstrings"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
)

// InitAppConfig acts as a bootstrap orchestrator. It receives a collection of
// 'Parser' and 'Options' objects, pairs them together, registers them to a new
// Config instance, and executes the initial data load in a single sequential flow.
//
// Constraints:
//   - The length of both slices must be identical. If not, it fails immediately.
//   - Registration and loading order follows the exact slice position (index 0 to N).
//
// Returns a fully initialized, populated pointer to Config, or nil and an error if it fails.
func InitAppConfig(parsers []Parser, options []Options) (*Config, error) {
	if len(parsers) != len(options) {
		parsersLength := xstrings.ConvertAllToStrings(len(parsers), len(options))

		errInfo := xstrings.FormatPairsColon(
			[]string{
				"parsers size", parsersLength[0],
				"options size", parsersLength[1],
			},
			nil,
		)

		return nil, xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_ASYMMETRIC_SIZES,
			nil,
			"",
			errInfo.String(),
		)
	}

	cfg := NewConfig()

	if len(parsers) == 0 {
		return &cfg, nil
	}

	for i := range parsers {
		err := cfg.Register(
			parsers[i],
			options[i],
		)
		if err != nil {
			return nil, err
		}
	}

	err := cfg.Load()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
