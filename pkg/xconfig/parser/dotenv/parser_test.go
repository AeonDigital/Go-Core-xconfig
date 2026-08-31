package dotenv_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/dotenv"
)

func TestParser_Read_RetrievePathsFailure(t *testing.T) {
	p := dotenv.NewParser()

	// Configurar um diretório que comprovadamente não existe no sistema
	// Isso fará com que o RetrieveConfigFilePaths falhe imediatamente no os.ReadDir interno
	opts := xconfig.Options{
		Prefix:    "APP",
		Separator: "_",
		DirPath:   "/system/path/that/does/not/exist/at/all",
	}

	if err := p.SetOptions(opts); err != nil {
		t.Fatalf("unexpected failure during SetOptions: %v", err)
	}

	_, err := p.Read()
	if err == nil {
		t.Fatalf("expected error from RetrieveConfigFilePaths due to missing directory, got nil")
	}

	// Verifica se a string do erro contém o identificador padrão do provedor DotEnv
	if !strings.Contains(err.Error(), "XCONFIG") {
		t.Errorf("expected error message to contain provider identifier, got: %q", err.Error())
	}
}

func TestParser_SetOptions_Validation(t *testing.T) {
	// Cenário 1: Falha por falta de prefixo/separador obrigatórios no dotenv
	p := dotenv.NewParser()
	optsInvalid := xconfig.Options{
		Prefix:    "", // vazio causa falha
		Separator: "_",
		FilePath:  "/any/path.env",
	}

	err := p.SetOptions(optsInvalid)
	if err == nil {
		t.Errorf("expected error during options validation due to empty prefix, got nil")
	}

	// Cenário 2: Falha por conflito de caminhos exclusivos
	optsConflict := xconfig.Options{
		Prefix:    "APP",
		Separator: "_",
		FilePath:  "/path.env",
		DirPath:   "/path/dir",
	}

	err = p.SetOptions(optsConflict)
	if err == nil {
		t.Errorf("expected error due to exclusive path contract violation, got nil")
	}
}

func TestParser_Read_SuccessPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "app.env")

	// Conteúdo mockado cobrindo todas as regras de negócio de parsing de string
	envContent := `
# This is a comment line and must be ignored
  # Another comment with spaces

APP_SERVER_PORT=8080
export APP_DB_HOST=127.0.0.1
APP_SERVICE_NAME="core-api"
APP_SECRET_KEY='xyz123'
INVALID_LINE_WITHOUT_EQUALS
EMPTY_VALUE_KEY=

# Key with lowercase mix that should be flattened
APP_MONITOR_URL=http://localhost
`
	if err := os.WriteFile(envFile, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to create temporary dotenv file: %v", err)
	}

	p := dotenv.NewParser()
	opts := xconfig.Options{
		Prefix:    "APP",
		Separator: "_",
		FilePath:  envFile,
	}

	if err := p.SetOptions(opts); err != nil {
		t.Fatalf("unexpected failure during SetOptions: %v", err)
	}

	data, err := p.Read()
	if err != nil {
		t.Fatalf("expected parsing to complete smoothly, got error: %v", err)
	}

	// Como o prefixo "APP" é removido no motor geral ou tratado no parser?
	// Nota: No seu parser de dotenv original, você NÃO remove o prefixo (diferente do parser do ENV).
	// O seu parser apenas substitui o Separador por "." e joga para caixa baixa.

	// Valida se removeu aspas e tratou o Separador "_" para "."
	if data["app.server.port"] != "8080" {
		t.Errorf("expected 8080, got %v", data["app.server.port"])
	}
	if data["app.db.host"] != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %v", data["app.db.host"])
	}
	if data["app.service.name"] != "core-api" {
		t.Errorf("expected core-api without double quotes, got %v", data["app.service.name"])
	}
	if data["app.secret.key"] != "xyz123" {
		t.Errorf("expected xyz123 without single quotes, got %v", data["app.secret.key"])
	}
	if data["app.monitor.url"] != "http://localhost" {
		t.Errorf("expected http://localhost, got %v", data["app.monitor.url"])
	}

	// Chaves sem o sinal de "=" ou fora da regra devem ser completamente ignoradas
	if _, exists := data["invalid_line_without_equals"]; exists {
		t.Errorf("expected invalid lines without token mapping to be discarded")
	}
}

func TestParser_Read_FileErrors(t *testing.T) {
	// Cenário 1: Erro de arquivo não encontrado
	p := dotenv.NewParser()
	opts := xconfig.Options{
		Prefix:    "APP",
		Separator: "_",
		FilePath:  "/non/existent/location/file.env",
	}

	_ = p.SetOptions(opts)
	_, err := p.Read()
	if err == nil {
		t.Errorf("expected file missing read error, got nil")
	}

	// Cenário 2: Erro do scanner interno (causado por linha excessivamente longa que estoura o buffer nativo do bufio)
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "huge.env")

	// Cria uma linha gigante com mais de 64k (limite padrão do bufio.MaxScanTokenSize) sem quebras de linha
	hugeLine := strings.Repeat("A", 70000) + "\n"
	_ = os.WriteFile(badFile, []byte(hugeLine), 0o644)

	optsBad := xconfig.Options{
		Prefix:    "APP",
		Separator: "_",
		FilePath:  badFile,
	}

	pBad := dotenv.NewParser()
	_ = pBad.SetOptions(optsBad)

	_, err = pBad.Read()
	if err == nil {
		t.Errorf("expected buffer token scan error due to long line length, got nil")
	}
}
