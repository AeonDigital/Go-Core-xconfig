package json

import (
	"encoding/json"
	"os"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github.com/AeonDigital/Go-Core-xconfig/pkg/xstrings"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
)

const XERR_PKGCTX_PARSER xerrors.ErrorCode = "XCONFIG.PARSER.JSON"

// ParserFormat defines the standard identifier and file extension target for this provider.
const ParserFormat string = "JSON"

// Parser implements the xconfig.Parser interface to read, parse, and merge
// data from structured JSON configuration files.
type Parser struct {
	options xconfig.Options
}

// NewParser instantiates an unconfigured pointer to a JSON Parser.
func NewParser() *Parser {
	return &Parser{}
}

// SetOptions executes strict path validation against the incoming criteria.
// It configuration targets must satisfy single-path exclusivity (File vs Dir vs ConfigPath)
// ensuring unambiguous resolution of JSON files.
func (o *Parser) SetOptions(opts xconfig.Options) error {
	err := opts.ValidateExclusivePaths(ParserFormat)
	if err != nil {
		return err
	}

	o.options = opts
	return nil
}

// Read processes all matched JSON configuration files, loads their byte content into memory,
// and unmarshals the syntax structures into a single consolidated generic map.
//
// Key behaviors:
//   - Loops through every identified path and aggregates properties into a shared map workspace.
//   - Safely bypasses empty file documents instead of failing or erasing previous updates.
//   - Note: Complex nested object trees inside JSON fields are captured as map[string]any. For
//     consistent flat-key overrides (e.g. key.subkey), input files should adopt flattened properties.
//
// Returns a merged generic map of configurations, or an error if any file read or syntax decoding fails.
func (o *Parser) Read() (map[string]any, error) {
	configFilePaths, err := o.options.RetrieveConfigFilePaths([]string{".json"}, ParserFormat)
	if err != nil {
		return nil, err
	}

	finalMap := make(map[string]any)

	for _, filePath := range configFilePaths {
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			errInfo := xstrings.FormatPairsColon(
				[]string{
					"filePath", filePath,
				},
				nil,
			)

			return nil, xerrors.NewError500(
				XERR_PKGCTX_PARSER,
				xerrors.XERR_OPERATION_FAILED,
				err,
				"failed to read configuration file",
				errInfo.String(),
			)
		}

		// Skip empty configuration records safely to continue evaluating sibling file pipelines
		if len(fileBytes) == 0 {
			continue
		}

		var tempMap map[string]any
		err = json.Unmarshal(fileBytes, &tempMap)
		if err != nil {
			errInfo := xstrings.FormatPairsColon(
				[]string{
					"filePath", "json",
					"content", string(fileBytes),
				},
				nil,
			)

			return nil, xerrors.NewError500(
				XERR_PKGCTX_PARSER,
				xerrors.XERR_INVALID_FORMAT_UNMARSHAL,
				err,
				"",
				errInfo.String(),
			)
		}

		// Merge keys into the persistent result matrix
		for k, v := range tempMap {
			finalMap[k] = v
		}
	}

	return finalMap, nil
}
