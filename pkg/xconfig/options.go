package xconfig

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xstrings"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
)

// Options centralizes all possible configuration parameters required by
// the various configuration providers (e.g., Environment, DotEnv, JSON).
type Options struct {
	// Prefix defines the string that environment variables must start with.
	Prefix string
	// Separator defines the delimiter used to split hierarchical keys (e.g., "_").
	Separator string
	// FilePath points directly to a single configuration file.
	FilePath string
	// DirPath points to a directory containing multiple configuration files.
	DirPath string
	// ConfigPath is a generic path that can accept either a single file or a directory.
	ConfigPath string
	// ExpandOptions holds dynamic, vendor-specific extra options.
	ExpandOptions map[string]any
}

// ValidateExclusivePaths ensures that path options are strictly exclusive.
// It counts how many path strategies are defined (FilePath, DirPath, ConfigPath).
// It returns an error if more than one path property is set at the same time,
// preventing ambiguous behavior during configuration loading.
func (o *Options) ValidateExclusivePaths(parserType string) error {
	count := 0

	if o.FilePath != "" {
		count++
	}
	if o.DirPath != "" {
		count++
	}
	if o.ConfigPath != "" {
		count++
	}

	if count > 1 {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"FilePath", o.FilePath,
				"DirPath", o.DirPath,
				"ConfigPath", o.ConfigPath,
			},
			nil,
		)

		return xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_MUTUAL_EXCLUSIVITY_VIOLATION,
			nil,
			"",
			errInfo.String(),
		)
	}

	return nil
}

// ValidatePrefixAndSeparator ensures that both 'Prefix' and 'Separator' properties are filled.
// This validation is essential for providers that rely on structured string lookups,
// such as Environment Variables and DotEnv files.
// Returns an error specifying which field is missing if either is empty.
func (o *Options) ValidatePrefixAndSeparator(parserType string) error {
	if o.Prefix == "" {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"field", "Prefix",
			},
			nil,
		)

		return xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_EMPTY_NOT_ALLOWED,
			nil,
			"",
			errInfo.String(),
		)
	}
	if o.Separator == "" {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"field", "Separator",
			},
			nil,
		)

		return xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_EMPTY_NOT_ALLOWED,
			nil,
			"",
			errInfo.String(),
		)
	}

	return nil
}

// RetrieveConfigFilePaths inspects the configured path properties and resolves them
// into a slice of valid file paths matching any of the requested extensions.
// The resulting file path slice is explicitly sorted in alphabetical order to guarantee
// a deterministic and predictable execution pipeline where subsequent files overwrite prior ones.
//
// Key behaviors:
//   - If 'ConfigPath' is provided, it auto-detects whether it is a file or directory
//     and normalizes the lookup strategy accordingly.
//   - If 'FilePath' is provided, it adds the file directly to the checklist.
//   - If 'DirPath' is provided, it scans the directory and retrieves all files
//     matching any of the requested extensions (e.g., []string{".yaml", ".yml"}).
//
// Parameters:
//   - extensions: A list of target file extensions to match (e.g., []string{"yaml", ".yml"}).
//     Leading dots are automatically added if missing. Pass an empty slice to match all files.
//   - parserType: Used exclusively to generate descriptive, context-aware error messages.
//
// Returns a sorted slice of string paths if successful, or an error if paths cannot be read,
// resolved, or if zero files match the criteria.
func (o Options) RetrieveConfigFilePaths(extensions []string, parserType string) ([]string, error) {
	// Normalize extensions to lowercase and ensure they contain the leading dot separator
	normalizedExts := make([]string, 0, len(extensions))
	for _, ext := range extensions {
		extLower := strings.ToLower(strings.TrimSpace(ext))
		if extLower != "" && !strings.HasPrefix(extLower, ".") {
			extLower = "." + extLower
		}
		if extLower != "" {
			normalizedExts = append(normalizedExts, extLower)
		}
	}

	targetFile := o.FilePath
	targetDir := o.DirPath

	if o.ConfigPath != "" {
		info, err := os.Stat(o.ConfigPath)
		if err != nil {
			errInfo := xstrings.FormatPairsColon(
				[]string{
					"ConfigPath", o.ConfigPath,
				},
				nil,
			)

			return nil, xerrors.NewError500(
				XERR_PKGCTX,
				xerrors.XERR_OPERATION_FAILED,
				err,
				"failed to stat the specified configuration path",
				errInfo.String(),
			)
		}

		if info.IsDir() {
			targetFile = ""
			targetDir = o.ConfigPath
		} else {
			targetFile = o.ConfigPath
			targetDir = ""
		}
	}

	var configFiles []string
	var err error

	if targetFile != "" {
		configFiles = append(configFiles, targetFile)
	}

	if targetDir != "" {
		configFiles, err = o.scanDirectory(targetDir, normalizedExts)
		if err != nil {
			return nil, err
		}
	}

	if len(configFiles) == 0 {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"FilePath", o.FilePath,
				"DirPath", o.DirPath,
				"ConfigPath", o.ConfigPath,
			},
			nil,
		)

		return nil, xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_RESOURCE_NOT_FOUND,
			nil,
			"",
			errInfo.String(),
		)
	}

	// Enforce a strict, cross-platform deterministic alphabetical file sorting sequence
	slices.Sort(configFiles)

	return configFiles, nil
}

// scanDirectory reads the contents of the given directory path and filters out
// any subdirectories, keeping only files that match at least one of the specified extensions.
// If the extensions slice is empty, all files within the directory are accepted.
// Returns a slice of fully qualified file paths found within the directory,
// or an error if the directory cannot be read.
func (o Options) scanDirectory(
	dirPath string,
	extensions []string,
) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"dirPath", dirPath,
			},
			nil,
		)

		return nil, xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_RESOURCE_UNAVAILABLE,
			err,
			"failed to scan the targeted directory structure",
			errInfo.String(),
		)
	}

	var resolvedFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()

		// If specific extensions are requested, validate the file matches at least one
		if len(extensions) > 0 {
			fileExt := strings.ToLower(filepath.Ext(fileName))
			matched := false
			for _, targetExt := range extensions {
				if fileExt == targetExt {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		fullPath := filepath.Join(dirPath, fileName)
		resolvedFiles = append(resolvedFiles, fullPath)
	}

	return resolvedFiles, nil
}
