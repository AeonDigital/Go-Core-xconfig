package yaml_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig"
	"github.com/AeonDigital/Go-Core-xconfig/pkg/xconfig/parser/yaml"
	"github.com/AeonDigital/Go-Core-xerrors/pkg/xerrors"
)

func TestParser_SetOptions_Validation(t *testing.T) {
	p := yaml.NewParser()

	optsConflict := xconfig.Options{
		FilePath:   "/path/config.yaml",
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

	// Dois arquivos YAML válidos para testar a leitura em lote, ordenação alfabética e mesclagem
	yaml1 := `
app_name: go-core-api
port: 8080
`
	yaml2 := `
port: 9000
debug: true
`

	if err := os.WriteFile(filepath.Join(tmpDir, "01_base.yaml"), []byte(yaml1), 0o644); err != nil {
		t.Fatalf("failed to setup test yaml file 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "02_override.yaml"), []byte(yaml2), 0o644); err != nil {
		t.Fatalf("failed to setup test yaml file 2: %v", err)
	}

	p := yaml.NewParser()
	opts := xconfig.Options{
		DirPath: tmpDir,
	}
	_ = p.SetOptions(opts)

	data, err := p.Read()
	if err != nil {
		t.Fatalf("expected YAML parsing to complete smoothly, got error: %v", err)
	}

	if data["app_name"] != "go-core-api" {
		t.Errorf("expected app_name to be 'go-core-api', got %v", data["app_name"])
	}

	// O parser de YAML decodifica inteiros diretamente como int ou int64 (diferente do JSON que joga para float64)
	if data["port"] != 9000 {
		t.Errorf("expected port to be overridden by 9000, got %v", data["port"])
	}

	if data["debug"] != true {
		t.Errorf("expected debug to be true, got %v", data["debug"])
	}
}

func TestParser_Read_EmptyFilesAndErrors(t *testing.T) {
	tmpDir := t.TempDir()

	// Cenário 1: Arquivo YAML completamente vazio (0 bytes) para testar a branch do 'continue'
	emptyFilePath := filepath.Join(tmpDir, "01_empty.yaml")
	if err := os.WriteFile(emptyFilePath, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to setup empty yaml file: %v", err)
	}

	// Cenário 2: Arquivo YAML com erro crasso de sintaxe / indentação
	badFilePath := filepath.Join(tmpDir, "02_corrupted.yaml")
	badYamlContent := `
invalid_yaml:
  - list_item
  indentation_error: true
`
	if err := os.WriteFile(badFilePath, []byte(badYamlContent), 0o644); err != nil {
		t.Fatalf("failed to setup corrupted yaml file: %v", err)
	}

	// Testa se o arquivo vazio é ignorado e retorna um mapa limpo
	pEmpty := yaml.NewParser()
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

	// Testa se o arquivo corrompido quebra no yaml.Unmarshal
	pBad := yaml.NewParser()
	optsBad := xconfig.Options{
		FilePath: badFilePath,
	}
	_ = pBad.SetOptions(optsBad)

	_, errBad := pBad.Read()
	if errBad == nil {
		t.Fatalf("expected syntax error due to malformed YAML body, got nil")
	}

	errStr := errBad.Error()
	// Technical Note: Validates the structured corporate metadata for unmarshal format failure
	if !strings.Contains(errStr, "[CTX: XCONFIG.PARSER.YAML]") {
		t.Errorf("expected context block 'XCONFIG.PARSER.YAML', got: %q", errStr)
	}
	if !strings.Contains(errStr, "[ERR: E3003]") {
		t.Errorf("expected error code 'E3003', got: %q", errStr)
	}
	if !strings.Contains(errStr, "[MSG: deserialization failed; unmarshal]") {
		t.Errorf("expected semantic unmarshal failure message, got: %q", errStr)
	}
	if !strings.Contains(errStr, "'filePath': 'yaml'") {
		t.Errorf("expected targeted layout engine inside info block to be 'yaml', got: %q", errStr)
	}
	if !strings.Contains(errStr, "'content':") {
		t.Errorf("expected file content to be detailed inside info block, got: %q", errStr)
	}
}

func TestParser_Read_ReadFileFailure(t *testing.T) {
	xerrors.EnableDebugMode()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "disappearing.yaml")

	// 1. Cria o arquivo para passar pelo RetrieveConfigFilePaths com sucesso
	if err := os.WriteFile(filePath, []byte("temp: true"), 0o644); err != nil {
		t.Fatalf("failed to setup file: %v", err)
	}

	p := yaml.NewParser()
	opts := xconfig.Options{
		FilePath: filePath,
	}
	_ = p.SetOptions(opts)

	// 2. Deleta o arquivo imediatamente antes do Read para forçar o erro no os.ReadFile
	_ = os.Remove(filePath)

	_, err := p.Read()
	if err == nil {
		t.Fatalf("expected error from os.ReadFile pipeline, got nil")
	}

	errStr := err.Error()
	// Technical Note: Validates the structured corporate layout for missing resource targets
	if !strings.Contains(errStr, "[CTX: XCONFIG.PARSER.YAML]") {
		t.Errorf("expected context block 'XCONFIG.PARSER.YAML', got: %q", errStr)
	}
	if !strings.Contains(errStr, "[MSG: failed to read configuration file]") {
		t.Errorf("expected default resource message block, got: %q", errStr)
	}
	if !strings.Contains(errStr, "'filePath':") {
		t.Errorf("expected field identifier inside info block, got: %q", errStr)
	}

	// Ativa o modo debug se precisar inspecionar a causa raiz nativa anexada no final do erro
	// Como a validação final busca a mensagem do OS, adicione xerrors.EnableDebugMode() no início do TestParser_Read_ReadFileFailure se ela falhar.
	hasMissingFileErr := strings.Contains(errStr, "no such file or directory") ||
		strings.Contains(errStr, "The system cannot find the file specified")

	if !hasMissingFileErr {
		t.Errorf("expected underlying OS boundary error contract to report missing file, got: %q", errStr)
	}
}

func TestParser_Read_FileSystemFailure(t *testing.T) {
	p := yaml.NewParser()
	opts := xconfig.Options{
		DirPath: "/system/folder/that/does/not/exist/at/all/yaml",
	}
	_ = p.SetOptions(opts)

	_, err := p.Read()
	if err == nil {
		t.Fatalf("expected error from path resolution pipeline, got nil")
	}
}
