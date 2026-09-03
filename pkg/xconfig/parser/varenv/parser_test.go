package varenv_test

import (
	"testing"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/varenv"
)

func TestParser_SetOptions_Validation(t *testing.T) {
	p := varenv.NewParser()
	optsInvalid := xconfig.Options{
		Prefix:    "",
		Separator: "_",
	}
	err := p.SetOptions(optsInvalid)
	if err == nil {
		t.Errorf("expected error due to missing prefix, got nil")
	}
}

func TestParser_Read_NewLogicAndEdgeCases(t *testing.T) {
	p := varenv.NewParser()
	opts := xconfig.Options{
		Prefix:    "MYAPP",
		Separator: "_",
	}
	_ = p.SetOptions(opts)

	// Usamos a ponte pública de testes para injetar o mock no campo privado
	p.SetEnvironFuncForTest(func() []string {
		return []string{
			"MYAPP_SERVER_PORT=4000",
			"MYAPP_EMPTY_CONFIG=",
			"MYAPP_ONLY_NAME",
			"MYAPP_CONNECTION=url=postgres://user:pass@host/db===",
			"SYSTEM_VARIABLE=true",
		}
	})

	data, err := p.Read()
	if err != nil {
		t.Fatalf("expected parsing to complete smoothly, got: %v", err)
	}

	// Valida cenário 1: Valor normal
	if val, exists := data["server.port"]; !exists || val != "4000" {
		t.Errorf("expected 'server.port' to be '4000', got %v", val)
	}

	// Valida cenário 2: Nome + '=' resulta em string vazia
	if val, exists := data["empty.config"]; !exists || val != "" {
		t.Errorf("expected 'empty.config' to exist as empty string, got %v", val)
	}

	// Valida cenário 3: Apenas nome (sem '=') resulta em string vazia (Cobre o 'else' do len(parts) > 1)
	if val, exists := data["only.name"]; !exists || val != "" {
		t.Errorf("expected 'only.name' to exist as empty string, got %v", val)
	}

	// Valida cenário 4: Múltiplos '=' preservados corretamente
	expectedValue := "url=postgres://user:pass@host/db==="
	if val, exists := data["connection"]; !exists || val != expectedValue {
		t.Errorf("expected 'connection' to preserve extra equals tokens, got %v", val)
	}

	// Valida cenário 5: Ignorado sem o prefixo
	if _, exists := data["system.variable"]; exists {
		t.Errorf("expected keys without app prefix to be discarded")
	}
}
