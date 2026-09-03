package yaml

import (
	"os"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github.com/AeonDigital/Go-Core-xconfig/pkg/xstrings"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
	"gopkg.in/yaml.v3"
)

const XERR_PKGCTX_PARSER xerrors.ErrorCode = "XCONFIG.PARSER.YAML"

// ParserFormat defines the standard identifier and file extension target for this provider.
const ParserFormat string = "YAML"

// Parser implements the xconfig.Parser interface to read, parse, and merge
// data from structured YAML configuration files.
type Parser struct {
	options xconfig.Options
}

// NewParser instantiates an unconfigured pointer to a YAML Parser.
func NewParser() *Parser {
	return &Parser{}
}

// SetOptions executes strict path validation against the incoming criteria.
// It configuration targets must satisfy single-path exclusivity (File vs Dir vs ConfigPath)
// ensuring unambiguous resolution of YAML files.
func (o *Parser) SetOptions(opts xconfig.Options) error {
	err := opts.ValidateExclusivePaths(ParserFormat)
	if err != nil {
		return err
	}

	o.options = opts
	return nil
}

// Read processes all matched YAML configuration files, loads their byte content into memory,
// and unmarshals the content syntax into a single consolidated generic map workspace.
//
// Key behaviors:
//   - Loops through every identified path and aggregates properties into a shared map object.
//   - Bypasses empty text streams securely using a continue directive to evaluate adjacent file configurations.
//   - Note: The file scanner relies strictly on the "YAML" token extension layout. Ensure target
//     files match this literal specification exactly (e.g., config.yaml).
//
// Returns a merged generic map of configurations, or an error if any file read or decoding routine fails.
func (o *Parser) Read() (map[string]any, error) {
	configFilePaths, err := o.options.RetrieveConfigFilePaths([]string{".yaml", ".yml"}, ParserFormat)
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

		// Safely skip empty text bodies to preserve earlier pipeline ingestions
		if len(fileBytes) == 0 {
			continue
		}

		currentMap := make(map[string]any)
		err = yaml.Unmarshal(fileBytes, &currentMap)
		if err != nil {
			errInfo := xstrings.FormatPairsColon(
				[]string{
					"filePath", "yaml",
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

		// Linear merge of configuration parameters
		for k, v := range currentMap {
			finalMap[k] = v
		}
	}

	return finalMap, nil
}
