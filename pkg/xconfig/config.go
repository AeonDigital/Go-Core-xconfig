package xconfig

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xstrings"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
)

// Config manages the internal global state of loaded configuration variables.
// It maintains a thread-safe registry of parsers and a consolidated map of data.
type Config struct {
	mu      sync.RWMutex
	data    map[string]any
	parsers []Parser
}

// NewConfig initializes an empty Config instance with prepared maps and slices,
// ensuring it is fully ready to safely register parsers or look up keys.
func NewConfig() Config {
	return Config{
		data:    make(map[string]any),
		parsers: make([]Parser, 0),
	}
}

// Register adds a custom provider/parser and its respective Options to the
// internal execution queue. This function does not trigger the data read operation.
// Returns an error if the underlying parser fails to ingest the provided options.
func (c *Config) Register(p Parser, opts Options) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := p.SetOptions(opts)
	if err != nil {
		return err
	}

	c.parsers = append(c.parsers, p)

	return nil
}

// Load clears the current state and triggers all registered parsers sequentially.
// Keys extracted from subsequent parsers will overwrite existing values with the
// same name, establishing a strict linear override priority matching the registration order.
// All keys are lowercased and stripped of leading/trailing spaces during ingestion to avoid ambiguity.
func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]any)

	for _, rp := range c.parsers {
		extracted, err := rp.Read()
		if err != nil {
			return err
		}

		// Merge data respecting linear priority and standardizing layout
		for k, v := range extracted {
			cleanKey := strings.ToLower(strings.TrimSpace(k))
			c.data[cleanKey] = v
		}
	}
	return nil
}

// Reload completely clears the internal data map and re-executes the ordered
// queue of registered parsers. This is particularly useful for live-reload strategies.
func (c *Config) Reload() error {
	return c.Load()
}

// Keys returns a thread-safe snapshot slice containing all keys currently loaded.
func (c *Config) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}

	return keys
}

// Has verifies the presence of a specific configuration key in a thread-safe manner.
// The search key parameter is normalized (lowercased and trimmed) to guarantee matching consistency.
func (c *Config) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cleanKey := strings.ToLower(strings.TrimSpace(key))
	_, exists := c.data[cleanKey]
	return exists
}

// Get performs a concurrent-safe lookup for a specific key.
// The target key parameter is automatically normalized (lowercased and trimmed).
// Returns the raw value and a boolean indicating whether the key was found.
func (c *Config) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cleanKey := strings.ToLower(strings.TrimSpace(key))
	val, exists := c.data[cleanKey]
	return val, exists
}

// Populate maps the consolidated internal configuration data into a custom struct pointer.
// It leverages standard JSON marshaling/unmarshaling workflows to map generic string/interface
// structures into typed application structures.
//
// Constraints:
//   - 'target' must be a non-nil pointer.
//   - 'target' must point strictly to a struct category object.
//
// Returns an error if validation fails or if the structural decoding cannot be achieved.
func (c *Config) Populate(target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Pointer {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"target kind", val.Kind().String(),
				"expected", reflect.Pointer.String(),
			},
			nil,
		)

		return xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_INVALID_TYPE,
			nil,
			"the target parameter must be a pointer to a struct",
			errInfo.String(),
		)
	}
	if val.IsNil() {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"field", "target",
			},
			nil,
		)

		return xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_NIL_NOT_ALLOWED,
			nil,
			"",
			errInfo.String(),
		)
	}
	if val.Elem().Kind() != reflect.Struct {
		errInfo := xstrings.FormatPairsColon(
			[]string{
				"target element kind", val.Elem().Kind().String(),
				"expected", reflect.Struct.String(),
			},
			nil,
		)

		return xerrors.NewError500(
			XERR_PKGCTX,
			xerrors.XERR_INVALID_TYPE,
			nil,
			"",
			errInfo.String(),
		)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	jsonData, err := json.Marshal(c.data)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))

	return decoder.Decode(target)
}
