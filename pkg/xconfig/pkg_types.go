package xconfig

// Parser defines the universal contract for modules capable of reading,
// parsing, and converting external configuration files or key-value stores
// into a standardized Go map.
type Parser interface {
	// SetOptions injects the necessary configuration properties into the parser instance.
	SetOptions(opts Options) error
	// Read executes the reading strategy and returns a generic key-value map.
	Read() (map[string]any, error)
}
