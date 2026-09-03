package dotenv

import (
	"bufio"
	"os"
	"strings"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github.com/AeonDigital/Go-Core-xconfig/pkg/xstrings"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
)

const XERR_PKGCTX_PARSER xerrors.ErrorCode = "XCONFIG.PARSER.DOTENV"

// ParserFormat defines the standard identifier and file extension target for this provider.
const ParserFormat string = "ENV"

// Parser implements the xconfig.Parser interface to read, tokenize, and ingest
// standard key-value ".env" configuration files.
type Parser struct {
	options xconfig.Options
}

// NewParser instantiates an unconfigured pointer to a Parser.
func NewParser() *Parser {
	return &Parser{}
}

// SetOptions executes strict path and naming validations against the incoming criteria.
// It configuration targets must satisfy single-path exclusivity (File vs Dir vs ConfigPath)
// and must provide valid Prefix and Separator tokens required for environment variables mapping.
func (o *Parser) SetOptions(opts xconfig.Options) error {
	err := opts.ValidateExclusivePaths(ParserFormat)
	if err != nil {
		return err
	}

	err = opts.ValidatePrefixAndSeparator(ParserFormat)
	if err != nil {
		return err
	}

	o.options = opts
	return nil
}

// Read processes all matched configuration files, parses their text lines sequentially,
// converts valid assignments into a generic data map, and normalizes key hierarchies.
//
// Parsing features:
//   - Skips empty lines and comment lines starting with '#'.
//   - Strips optional POSIX shell "export " statements.
//   - Trims enclosing single (') or double (") quotes from text values.
//   - Transforms structural delimiters (e.g., "APP_DB_PORT" with Separator "_" becomes "app.db.port").
//
// Returns a merged generic map of strings, or an error if any file read or scanning operation fails.
func (o *Parser) Read() (map[string]any, error) {
	configFilePaths, err := o.options.RetrieveConfigFilePaths([]string{".env"}, ParserFormat)
	if err != nil {
		return nil, err
	}

	finalMap := make(map[string]any)

	for _, filePath := range configFilePaths {
		if err := o.parseSingleFile(filePath, finalMap); err != nil {
			return nil, err
		}
	}

	return finalMap, nil
}

// parseSingleFile encapsulates file streaming and safety tasks, ensuring file
// descriptors are predictably closed via defer, avoiding descriptor leaks.
func (o *Parser) parseSingleFile(filePath string, targetMap map[string]any) error {
	file, err := os.Open(filePath)
	if err != nil {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"filePath", filePath,
			},
			nil,
		)

		return xerrors.NewError500(
			XERR_PKGCTX_PARSER,
			xerrors.XERR_RESOURCE_NOT_FOUND,
			err,
			"",
			errInfo.String(),
		)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore blank lines and shell comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle standard POSIX export prefix declarations
		if strings.HasPrefix(strings.ToLower(line), "export ") {
			line = line[7:]
			line = strings.TrimSpace(line)
		}

		// Enforce strict key=value pairing semantics
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip surrounding literal string encapsulations
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		// Flatten hierarchical key definitions if a separator is defined
		if o.options.Separator != "" {
			key = strings.ReplaceAll(key, o.options.Separator, ".")
		}
		key = strings.ToLower(key)

		targetMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"filePath", filePath,
			},
			nil,
		)

		return xerrors.NewError500(
			XERR_PKGCTX_PARSER,
			xerrors.XERR_RESOURCE_UNAVAILABLE,
			err,
			"",
			errInfo.String(),
		)
	}

	return nil
}
