Go-Core-xconfig
================================

![Go Test Coverage](https://raw.githubusercontent.com/github.com/AeonDigital/Go-Core-xconfig/badges/.badges/main/coverage.svg)

> [Aeon Digital](http://www.aeondigital.com.br)  
> rianna@aeondigital.com.br

&nbsp;

> Unified configuration loading for Go applications with support for multiple formats and sources.

`xconfig` is a lightweight, concurrent-safe configuration orchestrator designed to
unify and centralize application properties from diverse external environments. By
abstracting configuration providers behind a common interface, the engine effortlessly
aggregates environment variables, specialized files, and structured syntax into a
single normalized memory registry.




## CORE PHILOSOPHY & ARCHITECTURAL PURPOSE

Modern distributed architectures require flexible, multi-tiered configuration strategies
where production deployment parameters dynamically layer over development defaults.
`xconfig` solves this by orchestrating an ordered execution pipeline where subsequent
providers overwrite previously ingested properties with strict, deterministic linear
priority.

To guarantee semantic consistency and zero ambiguity across independent providers,
`xconfig` enforces an opinionated normalization layer:

- **Key Standardization**: Every loaded configuration key is natively converted to
  lowercase and stripped of leading or trailing whitespace.
- **Hierarchical Flattening**: Environmental boundaries and structural keys (such
  as `APP_DB_PORT`) are normalized using standard dot-notation layout (converting
  into `app.db.port`).
- **Thread Safety**: Backed internally by a robust `sync.RWMutex` coordinator, the
  entire data state supports ultra-fast, concurrent-safe lookups across independent
  application goroutines.




## SUPPORTED FORMATS & SOURCES

The module includes production-ready driver implementations out of the box for:

- **`VARENV`**: Host OS system environment variables filtered by a mandatory application
  prefix.
- **`DOTENV`**: Standard POSIX shell `.env` key-value layout files.
- **`JSON`**: Structured JavaScript Object Notation files.
- **`YAML`**: Structured Human-friendly YAML (`.yaml`, `.yml`) configuration files.




## EXPOSED ARCHITECTURAL COMPONENTS

The package exposes three atomic entry points for application bootstrapping:

- **`xconfig.Parser`** — The universal interface defining the operational contract
  for target data providers.
- **`xconfig.Config`** — The concurrent-safe state engine managing the internal variable
  map and provider execution queue.
- **`xconfig.InitAppConfig`** — The bootstrap orchestrator that pairs parsers with
  configuration parameters in a single sequential sweep.




&nbsp;
________________________________________________________________________________

## 1. INSTALLATION

This section describes how to install the isolated package module and map its internal
driver dependencies into your local development workspace.



### 1.1. Prerequisites

Before importing this package, ensure your local environment satisfies the following
development boundaries:

- **Go Version**: Go 1.21 or higher initialized within your target workspace.
- **Upstream Dependencies**: This module relies on the core error management library.
  Ensure `github.com/AeonDigital/Go-Core-xerrors` is available within your dependency
  chain.



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 1.2. Package Module Installation

To fetch and bind the standalone configuration engine into your active project workspace
layout, execute the standard `go get` command from your terminal:

```shell
go get github.com/AeonDigital/Go-Core-xconfig@latest
```



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 1.3. Import Declarations

To leverage the specific features, register the primary engine layer alongside any
specialized driver parsers required by your bootstrap pipeline:

```go
import (
    "github com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
    "github com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/dotenv"
    "github com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/json"
    "github com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/varenv"
    "github com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/yaml"
)
```




&nbsp;
________________________________________________________________________________

## 2. ARCHITECTURAL CONVENTIONS & CONFIGURATION

This section details the unified configuration properties, deterministic structural
safety rules, and priority patterns implemented by the engine.



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 2.1. The Options Parameter Strategy

The generic `xconfig.Options` struct unifies all potential parameter variables required
by the underlying driver ecosystem. Instead of introducing fragmented configuration
interfaces for each provider, a single cohesive layout is passed:

- `Prefix` — Mandatory string header that environmental systems filter against (e.g.,
  `APP_`).
- `Separator` — Delimiter token used to break and translate flat hierarchical keys
  into nested dot-notation layouts (e.g., `_`).
- `FilePath` — Absolute or relative path targeting a single literal configuration
  file.
- `DirPath` — Directory path pointing to a folder containing a collection of sibling
  configuration files.
- `ConfigPath` — Adaptive generic path parameter that dynamically self-resolves whether
  it is pointing to a single file or an active folder directory.
- `ExpandOptions` — Extensible generic `map[string]any` reserved for vendor-specific
  or driver-custom parameters.



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 2.2. Deterministic Pipeline Validations

To prevent ambiguous behavior during multi-source overrides, the option engine enforces
absolute architectural safeguards at runtime:

1. **Mutual Exclusivity Violation (`XERR_MUTUAL_EXCLUSIVITY_VIOLATION`)**: Path indicators
   are strictly exclusive. Passing any overlapping combination of `FilePath`, `DirPath`,
   or `ConfigPath` simultaneously within a single option instance will cause immediate
   validation failure.


2. **Naming and Prefix Safeguards (`XERR_EMPTY_NOT_ALLOWED`)**: String lookup drivers
   (such as `varenv` and `dotenv`) require explicit declaration of both `Prefix`
   and `Separator`. Omitting either of these properties blocks the registration cycle.


3. **Cross-Platform Alphabetical Sorting**: When parsing a directory (`DirPath` or
   folder-based `ConfigPath`), files are recursively filtered against acceptable
   extensions and compiled into a slice that is sorted alphabetically using a strict
   cross-platform sequence. This prevents irregular OS file system streams from modifying
   your configuration priorities.



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 2.3. Linear Override Priority

The registration order of parsers inside the execution workspace establishes a predictable
cascading layout strategy. Whenever `Load()` or `Reload()` is executed, the configuration
engine resets the internal registry and fires the registered drivers sequentially
based on their slice index location.

As properties are harvested, keys derived from subsequent providers seamlessly overwrite
existing values with identical names.

```text
  Registration Sequence : [Index 0] -> [Index 1] -> [Index N]
      Priority Strength : Index N (Highest Override) > Index 0 (Base Layer)
```

For example, initializing a baseline `yaml.NewParser()` followed by an active `varenv.NewParser()`
guarantees that system environment variables take immediate priority and patch over
persistent local disk parameters smoothly.




&nbsp;
________________________________________________________________________________

## 3. BASIC USAGE

This section provides actionable code blueprints demonstrating how to initialize
the configuration workspace, query data safely across concurrent environments, and
populate structured custom objects.



### 3.1. Bootstrapping with InitAppConfig

The typical orchestration pattern is to pair a collection of target drivers with
their respective parameters inside `InitAppConfig`. This pairs inputs by index, binds
them sequentially, and triggers the initial data ingest.

```go
package main

import (
	"fmt"
	"log"

	"github com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/varenv"
	"github com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/yaml"
)

func main() {
	// Pair parsers and options symmetrically (Index 0 to N)
	cfg, err := xconfig.InitAppConfig(
		[]xconfig.Parser{
			yaml.NewParser(),    // Index 0: Base settings layer
			varenv.NewParser(),  // Index 1: Environment variable overrides
		},
		[]xconfig.Options{
			{ConfigPath: "./config/settings.yaml"},
			{Prefix: "APP_", Separator: "_"},
		},
	)
	if err != nil {
		log.Fatalf("Failed to bootstrap configuration: %v", err)
	}

	fmt.Println("Configuration workspace initialized successfully.")
}
```




&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 3.2. Querying Keys Safely

`xconfig.Config` encapsulates all lookups behind thread-safe read locks (`sync.RWMutex`),
making it completely safe to check or extract parameters dynamically inside concurrent
goroutines:

```go
// 1. Concurrent-safe lookup returning raw value and structural existence flag
value, exists := cfg.Get("database.host")
if exists {
	fmt.Printf("Database host is configured to: %v\n", value)
}

// 2. Explicit key validation check
if cfg.Has("server.port") {
	fmt.Println("Application server port has been provisioned.")
}

// 3. Extract a complete snapshot slice of all registered and loaded keys
activeKeys := cfg.Keys()
fmt.Printf("Total loaded parameters: %d\n", len(activeKeys))
```



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 3.3. Structural Decoding via Populate

Instead of executing tedious continuous manual casting on generic `any` values, `xconfig`
allows you to directly marshal and ingest the entire configuration tree into a strongly
typed application struct pointer using `Populate`.

The system validates parameters to ensure strict struct type compliance before executing
standard JSON syntax translation:

```go
type DBConfig struct {
	Host     string `json:"database.host"`
	Port     int    `json:"database.port"`
	Username string `json:"database.user"`
}

type AppSpecification struct {
	Environment string   `json:"app.env"`
	Database    DBConfig `json:",inline"` // Maps flattened keys gracefully
}

func MapConfiguration(cfg *xconfig.Config) {
	var appSettings AppSpecification

	// Populate requires a non-nil pointer targeting a valid struct object
	err := cfg.Populate(&appSettings)
	if err != nil {
		log.Fatalf("Failed to bind structural fields: %v", err)
	}

	fmt.Printf("Application bound to environment: %s\n", appSettings.Environment)
}
```




&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 3.4. Runtime Live Re-loading

If configuration updates are modified on the underlying disk or execution platform,
you can invoke `Reload()` to completely flush the workspace data cache and re-execute
the exact order of the previously registered configuration stream without rebuilding
the original bootstrap objects:

```go
func HandleConfigurationRefresh(cfg *xconfig.Config) {
	// Flushes existing data and triggers a clean, cascading reread of all drivers
	err := cfg.Reload()
	if err != nil {
		log.Printf("Emergency configuration live-reload failed: %v", err)
		return
	}

	fmt.Println("Configuration infrastructure cache refreshed successfully.")
}
```




&nbsp;
________________________________________________________________________________

## 4. ADDITIONAL INFORMATION

This project uses the [Semantic Versioning](https://semver.org/) system proposed
by **Tom Preston-Werner**.




&nbsp;
________________________________________________________________________________

## 5. LICENSE

This project is offered under the [MIT license](LICENSE.md).