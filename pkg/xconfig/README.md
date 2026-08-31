xconfig
================================

> Unified configuration loading for Go applications with support for multiple formats and sources.

&nbsp;

This package provides a configuration loader abstraction based on parsers. Each parser
implements a common interface that converts external sources into a normalized internal
map with lowercased, trimmed keys.




&nbsp;
________________________________________________________________________________

## PURPOSE

`xconfig` is designed to centralize configuration loading from multiple sources (environment
variables, `.env` files, JSON, YAML) and to compose multiple parsers with deterministic
linear override priority.

It exposes:

- `xconfig.Parser` — generic interface for configuration providers.
- `xconfig.Config` — concurrent-safe structure for storing and querying data.
- `xconfig.InitAppConfig` — bootstrap function for registering parsers and loading
  data.




&nbsp;
________________________________________________________________________________

## INSTALLATION

Use `go get` pointing to the package module:

```shell
  go get github.com/AeonDigital/Go-Core/xconfig@latest
```

In your code, import the required dependencies:

```go
import (
    "github.com/AeonDigital/Go-Core/xconfig"
    "github.com/AeonDigital/Go-Core/xconfig/parser/dotenv"
    "github.com/AeonDigital/Go-Core/xconfig/parser/json"
    "github.com/AeonDigital/Go-Core/xconfig/parser/varenv"
    "github.com/AeonDigital/Go-Core/xconfig/parser/yaml"
)
```




&nbsp;
________________________________________________________________________________

## BASIC USAGE

The typical usage pattern is to instantiate one or more parsers and pass them to
`InitAppConfig`:

```go
cfg, err := xconfig.InitAppConfig(
    []xconfig.Parser{
        varenv.NewParser(),
        yaml.NewParser(),
    },
    []xconfig.Options{
        {Prefix: "APP_", Separator: "_"},
        {ConfigPath: "config"},
    },
)
if err != nil {
    // handle error
}

value, ok := cfg.Get("database.host")
if ok {
    fmt.Println(value)
}
```



The package also provides:

- `cfg.Has(key)` — check if a key exists.
- `cfg.Keys()` — list all loaded keys.
- `cfg.Reload()` — reload all registered parsers.
- `cfg.Populate(target)` — populate a struct from the loaded data.




&nbsp;
________________________________________________________________________________

## SUPPORTED PARSERS

The package includes drivers for:

- `xconfig/parser/varenv` — system environment variables.
- `xconfig/parser/dotenv` — `.env` files.
- `xconfig/parser/json` — JSON files.
- `xconfig/parser/yaml` — YAML files.




&nbsp;
________________________________________________________________________________

## EXTERNAL DEPENDENCIES

`xconfig` uses only the Go standard library and one external YAML parser:

- `gopkg.in/yaml.v3`

It also depends on the internal repository package `github.com/AeonDigital/Go-Core/xerrors`.