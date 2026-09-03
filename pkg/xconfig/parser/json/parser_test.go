package json_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/json"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
)

func TestParser_SetOptions_Validation(t *testing.T) {
	p := json.NewParser()

	// Conflito de caminhos exclusivos para forçar o erro de validação
	optsConflict := xconfig.Options{
		FilePath:   "/path/config.json",
		DirPath:    "/path/dir",
		ConfigPath: "",
	}

	err := p.SetOptions(optsConflict)
	if err == nil {
		t.Errorf("expected error due to exclusive path validation failure, got nil")
	}
}

func TestParser_Read_SuccessAndMerge(t *testing.T) {
	tmpDir := t.TempDir()

	// Criamos dois arquivos JSON válidos para testar a leitura em lote e a mesclagem de chaves
	json1 := `{"app_name": "go-core-api", "port": 8080}`
	json2 := `{"port": 9000, "debug": true}`

	if err := os.WriteFile(filepath.Join(tmpDir, "01_base.json"), []byte(json1), 0o644); err != nil {
		t.Fatalf("failed to setup test json file 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "02_override.json"), []byte(json2), 0o644); err != nil {
		t.Fatalf("failed to setup test json file 2: %v", err)
	}

	p := json.NewParser()
	opts := xconfig.Options{
		DirPath: tmpDir,
	}
	_ = p.SetOptions(opts)

	data, err := p.Read()
	if err != nil {
		t.Fatalf("expected JSON parsing to complete smoothly, got error: %v", err)
	}

	// Valida se capturou a chave única do primeiro arquivo
	if data["app_name"] != "go-core-api" {
		t.Errorf("expected app_name to be 'go-core-api', got %v", data["app_name"])
	}

	// Valida se o segundo arquivo sobrescreveu deterministicamente a porta (9000 ganha de 8080)
	// Nota: json.Unmarshal converte números genéricos para float64 ao mapear em map[string]any
	if data["port"] != float64(9000) {
		t.Errorf("expected port to be overridden by 9000, got %v", data["port"])
	}

	// Valida se capturou a chave única do segundo arquivo
	if data["debug"] != true {
		t.Errorf("expected debug to be true, got %v", data["debug"])
	}
}

func TestParser_Read_EmptyFilesAndErrors(t *testing.T) {
	tmpDir := t.TempDir()

	// Cenário 1: Arquivo JSON completamente vazio (0 bytes) para testar a branch do 'continue'
	emptyFilePath := filepath.Join(tmpDir, "01_empty.json")
	if err := os.WriteFile(emptyFilePath, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to setup empty json file: %v", err)
	}

	// Cenário 2: Arquivo JSON com erro de sintaxe estrutural
	badFilePath := filepath.Join(tmpDir, "02_corrupted.json")
	if err := os.WriteFile(badFilePath, []byte("{bad-json-syntax}"), 0o644); err != nil {
		t.Fatalf("failed to setup corrupted json file: %v", err)
	}

	// Primeiro, testamos apenas o arquivo vazio para garantir que o continue funciona e retorna um mapa limpo
	pEmpty := json.NewParser()
	optsEmpty := xconfig.Options{
		FilePath: emptyFilePath,
	}
	_ = pEmpty.SetOptions(optsEmpty)

	dataEmpty, errEmpty := pEmpty.Read()
	if errEmpty != nil {
		t.Fatalf("expected empty file to be safely bypassed via continue, got error: %v", errEmpty)
	}
	if len(dataEmpty) != 0 {
		t.Errorf("expected data map from empty file pipeline to be empty, got size %d", len(dataEmpty))
	}

	// Agora, testamos o arquivo corrompido para garantir que ele cai na branch de erro de Unmarshal
	pBad := json.NewParser()
	optsBad := xconfig.Options{
		FilePath: badFilePath,
	}
	_ = pBad.SetOptions(optsBad)

	_, errBad := pBad.Read()
	if errBad == nil {
		t.Fatalf("expected syntax error due to malformed JSON body, got nil")
	}
	if !strings.Contains(errBad.Error(), "CTX: XCONFIG.PARSER.JSON") {
		t.Errorf("unexpected error payload message: %q", errBad.Error())
	}
}

func TestParser_Read_FileSystemFailure(t *testing.T) {
	p := json.NewParser()
	opts := xconfig.Options{
		DirPath: "/system/folder/that/does/not/exist/at/all/json",
	}
	_ = p.SetOptions(opts)

	_, err := p.Read()
	if err == nil {
		t.Fatalf("expected error from path resolution pipeline, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "CTX: ERR_XCONFIG") {
		t.Errorf("expected corporate error prefix tracking, got: %q", errStr)
	}
	if !strings.Contains(errStr, "[ERR: E1008]") {
		t.Errorf("expected directory scanning error code, got: %q", errStr)
	}
}

func TestParser_Read_ReadFilePermissionFailure(t *testing.T) {
	xerrors.EnableDebugMode()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "secure.json")

	// 1. Cria o arquivo para que o RetrieveConfigFilePaths o encontre com sucesso
	if err := os.WriteFile(filePath, []byte(`{"secure": true}`), 0o644); err != nil {
		t.Fatalf("failed to setup temporary json file: %v", err)
	}

	// 2. Remove todas as permissões de leitura/escrita do arquivo (0000)
	if err := os.Chmod(filePath, 0o000); err != nil {
		t.Fatalf("failed to change file permissions for test: %v", err)
	}

	p := json.NewParser()
	opts := xconfig.Options{
		FilePath: filePath,
	}
	_ = p.SetOptions(opts)

	_, err := p.Read()
	if err == nil {
		// Se o Chmod falhou em travar o arquivo devido ao privilégio do usuário,
		// tentamos o Plano B (deletar o arquivo antes do Read) para garantir que funcione em qualquer OS:
		_ = os.Remove(filePath)
		_, err = p.Read()
		if err == nil {
			t.Fatalf("expected error from os.ReadFile pipeline, got nil")
		}
	}

	errStr := err.Error()
	// Technical Note: Validates the structured layout and metadata blocks generated by NewError500 and FormatPairsColon
	if !strings.Contains(errStr, "[CTX: XCONFIG.PARSER.JSON]") {
		t.Errorf("expected context block 'XCONFIG.PARSER.JSON', got: %q", errStr)
	}
	if !strings.Contains(errStr, "[ERR: E4002]") {
		t.Errorf("expected error code 'E4002', got: %q", errStr)
	}
	if !strings.Contains(errStr, "[MSG: failed to read configuration file]") {
		t.Errorf("expected configuration read message block, got: %q", errStr)
	}
	if !strings.Contains(errStr, "'filePath':") {
		t.Errorf("expected field targeting parameter inside info block, got: %q", errStr)
	}

	// Technical Note: Windows bypasses Chmod 0000 triggering Plan B (file deletion),
	// so we accept either OS boundary contract error: permission denied or file not found.
	hasValidOSError := strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "The system cannot find the file specified") ||
		strings.Contains(errStr, "no such file or directory")

	if !hasValidOSError {
		t.Errorf("expected underlying OS boundary error contract to report permission denied or file missing, got: %q", errStr)
	}
}
