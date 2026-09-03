package varenv

import (
	"os"
	"strings"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
)

const XERR_PKGCTX_PARSER xerrors.ErrorCode = "XCONFIG.PARSER.VARENV"

// ParserFormat defines the standard identifier used to identify system environment variables.
const ParserFormat string = "VARENV"

// Parser implements the xconfig.Parser interface to read, filter, and map
// system environment variables into a consolidated application state.
type Parser struct {
	options     xconfig.Options
	environFunc func() []string
}

// NewParser instantiates an unconfigured pointer to an environment variables Parser.
func NewParser() *Parser {
	return &Parser{
		environFunc: os.Environ, // Padrão de produção
	}
}

// SetOptions executes validation against the incoming criteria.
// It requires both Prefix and Separator properties to be explicitly defined,
// establishing a precise lookup filter for environmental keys.
func (o *Parser) SetOptions(opts xconfig.Options) error {
	err := opts.ValidatePrefixAndSeparator(ParserFormat)
	if err != nil {
		return err
	}

	o.options = opts
	return nil
}

// Read intercepts all active system environment variables, filters out keys
// that do not match the configured application prefix, strips the prefix,
// and normalizes the hierarchical separators into standard dot-notation format.
//
// Processing features:
//   - Trims the mandatory application prefix to clean up the configuration map keys.
//   - Automatically handles boundary separators to avoid leading dots (e.g., converts APP_DB to db).
//   - Replaces custom layout delimiters with dot notation (e.g., DB_PORT becomes db.port).
//
// Returns a generic string-interface map containing all parsed and matched environment data.
func (o *Parser) Read() (map[string]any, error) {
	envVars := o.environFunc()
	extractedData := make(map[string]any)

	prefix := strings.ToUpper(o.options.Prefix)

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)

		key := parts[0]
		value := ""

		if len(parts) > 1 {
			value = parts[1]
		}

		// Filter variables using the strict mandatory application prefix
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		key = strings.TrimPrefix(key, prefix)

		// Prevent leading dots if the prefix removal leaves a leading separator (e.g., "_DB" with Separator "_")
		if o.options.Separator != "" {
			key = strings.TrimPrefix(key, o.options.Separator)
			key = strings.ReplaceAll(key, o.options.Separator, ".")
		}
		key = strings.ToLower(key)

		extractedData[key] = value
	}

	return extractedData, nil
}
