package xconfig_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
)

// MockParser implementa a interface Parser para simular retornos nos testes de motor
type MockParser struct {
	opts    xconfig.Options
	readFn  func() (map[string]any, error)
	errOpts error
}

func (m *MockParser) SetOptions(opts xconfig.Options) error {
	m.opts = opts
	return m.errOpts
}

func (m *MockParser) Read() (map[string]any, error) {
	if m.readFn != nil {
		return m.readFn()
	}
	return make(map[string]any), nil
}

func TestInitAppConfig_AsymmetricSizes(t *testing.T) {
	parsers := []xconfig.Parser{&MockParser{}}
	options := []xconfig.Options{} // Zero size, will fail validation

	cfg, err := xconfig.InitAppConfig(parsers, options)
	if err == nil {
		t.Fatalf("expected error due to asymmetric input sizes, got nil")
	}
	if cfg != nil {
		t.Errorf("expected return config pointer to be nil on critical validation failure")
	}

	// Technical Note: Asserts compliance with the structured corporate layout format
	expectedSubStr := "[CTX: ERR_XCONFIG][ERR: E3008]"
	if !strings.Contains(err.Error(), expectedSubStr) {
		t.Errorf("unexpected corporate error payload shape: %q", err.Error())
	}
}

func TestInitAppConfig_EmptySlices(t *testing.T) {
	var parsers []xconfig.Parser
	var options []xconfig.Options

	cfg, err := xconfig.InitAppConfig(parsers, options)
	if err != nil {
		t.Fatalf("expected no error on empty configuration bootstrapping, got: %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected a valid initialized pointer to Config, got nil")
	}
	if len(cfg.Keys()) != 0 {
		t.Errorf("expected a clean configuration state with 0 keys, got %d", len(cfg.Keys()))
	}
}

func TestInitAppConfig_FullSuccessPipeline(t *testing.T) {
	mock := &MockParser{
		readFn: func() (map[string]any, error) {
			return map[string]any{"app.port": 8080}, nil
		},
	}

	parsers := []xconfig.Parser{mock}
	options := []xconfig.Options{{Prefix: "TEST"}}

	cfg, err := xconfig.InitAppConfig(parsers, options)
	if err != nil {
		t.Fatalf("expected bootstrap chain to run seamlessly, got error: %v", err)
	}
	if cfg == nil {
		t.Fatalf("returned config instance pointer is nil")
	}

	val, exists := cfg.Get("app.port")
	if !exists || val != 8080 {
		t.Errorf("expected key 'app.port' to be populated with 8080, got %v (exists: %t)", val, exists)
	}
}

func TestInitAppConfig_RegisterFailure(t *testing.T) {
	mockErr := errors.New("mock register error")
	mock := &MockParser{
		errOpts: mockErr,
	}

	parsers := []xconfig.Parser{mock}
	options := []xconfig.Options{{Prefix: "FAIL"}}

	cfg, err := xconfig.InitAppConfig(parsers, options)
	if err == nil {
		t.Fatalf("expected error during registration phase, got nil")
	}
	if cfg != nil {
		t.Errorf("expected config pointer to be nil when registration fails")
	}
	if !errors.Is(err, mockErr) {
		t.Errorf("expected root cause error %v, got %v", mockErr, err)
	}
}

func TestInitAppConfig_LoadFailure(t *testing.T) {
	mockErr := errors.New("mock read error")
	mock := &MockParser{
		readFn: func() (map[string]any, error) {
			return nil, mockErr
		},
	}

	parsers := []xconfig.Parser{mock}
	options := []xconfig.Options{{Prefix: "OK"}}

	cfg, err := xconfig.InitAppConfig(parsers, options)
	if err == nil {
		t.Fatalf("expected error during loading phase, got nil")
	}
	if cfg != nil {
		t.Errorf("expected config pointer to be nil when loading fails")
	}
	if !errors.Is(err, mockErr) {
		t.Errorf("expected root cause error %v, got %v", mockErr, err)
	}
}
