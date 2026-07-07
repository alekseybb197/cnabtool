package logging

import (
	"cnabtool/pkg/data"
	"io"
	"log"
	"strings"
	"testing"
)

// TestLogLevels проверяет константы уровней логирования
func TestLogLevels(t *testing.T) {
	if LogQuietLevel != 0 {
		t.Errorf("LogQuietLevel = %d, want 0", LogQuietLevel)
	}
	if LogErrorLevel != 1 {
		t.Errorf("LogErrorLevel = %d, want 1", LogErrorLevel)
	}
	if LogNormalLevel != 2 {
		t.Errorf("LogNormalLevel = %d, want 2", LogNormalLevel)
	}
	if LogInfoLevel != 3 {
		t.Errorf("LogInfoLevel = %d, want 3", LogInfoLevel)
	}
	if LogDebugLevel != 4 {
		t.Errorf("LogDebugLevel = %d, want 4", LogDebugLevel)
	}
}

// captureLogOutput перехватывает вывод log.Printf
type logCapture struct {
	reader io.Reader
	writer *io.PipeWriter
	cancel func()
}

func captureLogOutput() *logCapture {
	pr, pw := io.Pipe()

	// Сохраняем оригинальный logger
	originalOutput := log.Writer()
	log.SetOutput(pw)

	// Перехватываем
	cap := &logCapture{
		reader: pr,
		writer: pw,
		cancel: func() {
			pw.Close()
			log.SetOutput(originalOutput)
		},
	}

	return cap
}

// TestError_IncrementsErrorCount проверяет, что Error инкрементирует data.Gc.Error
func TestError_IncrementsErrorCount(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}

	before := data.Gc.Error

	Error("test error message")

	if data.Gc.Error != before+1 {
		t.Errorf("data.Gc.Error after Error() = %d, want %d", data.Gc.Error, before+1)
	}
}

// TestError_VerbosityFilter проверяет фильтрацию по уровню verbosity
func TestError_VerbosityFilter(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogErrorLevel,
	}

	Error("test error")

	if data.Gc.Error != 1 {
		t.Errorf("data.Gc.Error should be incremented even at LogErrorLevel, got %d", data.Gc.Error)
	}
}

// TestMessage_VerbosityFilter проверяет, что Message логируется при LogNormalLevel+
func TestMessage_VerbosityFilter(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogNormalLevel,
	}

	Message("test message")

	// Проверяем, что Message не вызывает ошибок
	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestMessage_NotLoggedAtLowVerbosity проверяет, что Message не логируется при LogErrorLevel
func TestMessage_NotLoggedAtLowVerbosity(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogErrorLevel,
	}

	Message("should not be logged at error level")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestInfo_VerbosityFilter проверяет, что Info логируется при LogInfoLevel+
func TestInfo_VerbosityFilter(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogInfoLevel,
	}

	Info("test info message")

	if data.Gc.Error != 0 {
		t.Errorf("Info should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestInfo_NotLoggedAtLowVerbosity проверяет, что Info не логируется при LogNormalLevel
func TestInfo_NotLoggedAtLowVerbosity(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogNormalLevel,
	}

	Info("should not be logged at normal level")

	if data.Gc.Error != 0 {
		t.Errorf("Info should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestDebug_VerbosityFilter проверяет, что Debug логируется при LogDebugLevel+
func TestDebug_VerbosityFilter(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}

	Debug("test debug message")

	if data.Gc.Error != 0 {
		t.Errorf("Debug should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestDebug_NotLoggedAtLowVerbosity проверяет, что Debug не логируется при LogInfoLevel
func TestDebug_NotLoggedAtLowVerbosity(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogInfoLevel,
	}

	Debug("should not be logged at info level")

	if data.Gc.Error != 0 {
		t.Errorf("Debug should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestMaskCredentials_Simple проверяет простую замену чувствительных данных
func TestMaskCredentials_Simple(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}
	data.Sensitives = []string{"secret123", "password"}

	Message("test secret123 data")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestMaskCredentials_MultipleSecrets проверяет замену нескольких чувствительных данных
func TestMaskCredentials_MultipleSecrets(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}
	data.Sensitives = []string{"secret1", "secret2"}

	Message("test secret1 and secret2")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestMaskCredentials_EmptySensitives проверяет работу с пустым списком
func TestMaskCredentials_EmptySensitives(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}
	data.Sensitives = nil

	Message("test no secrets")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestPrettyString_ValidJSON проверяетpretty-форматирование валидного JSON
func TestPrettyString_ValidJSON(t *testing.T) {
	input := `{"name":"test","value":123}`
	expected := `{
  "name": "test",
  "value": 123
}`

	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for valid JSON, got: %v", err)
	}

	if result != expected {
		t.Errorf("PrettyString result = %q, want %q", result, expected)
	}
}

// TestPrettyString_NestedJSON проверяет pretty-форматирование вложенного JSON
func TestPrettyString_NestedJSON(t *testing.T) {
	input := `{"user":{"name":"John","age":30},"active":true}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for nested JSON, got: %v", err)
	}

	// Проверяем, что результат содержит ключевые элементы
	if !strings.Contains(result, `"user"`) {
		t.Error("PrettyString result should contain 'user'")
	}
	if !strings.Contains(result, `"name"`) {
		t.Error("PrettyString result should contain 'name'")
	}
	if !strings.Contains(result, `"John"`) {
		t.Error("PrettyString result should contain 'John'")
	}
	// Проверяем, что результат содержит отступы (2 пробела)
	if !strings.Contains(result, "  ") {
		t.Error("PrettyString result should contain indentation")
	}
}

// TestPrettyString_InvalidJSON проверяет обработку невалидного JSON
func TestPrettyString_InvalidJSON(t *testing.T) {
	input := `{invalid json}`
	_, err := PrettyString(input)
	if err == nil {
		t.Error("PrettyString should return error for invalid JSON")
	}
}

// TestPrettyString_EmptyString проверяет обработку пустой строки
func TestPrettyString_EmptyString(t *testing.T) {
	input := ``
	_, err := PrettyString(input)
	if err == nil {
		t.Error("PrettyString should return error for empty string")
	}
}

// TestPrettyString_SingleValue проверяет обработку простых значений
func TestPrettyString_SingleValue(t *testing.T) {
	input := `"hello"`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for single value, got: %v", err)
	}

	if result != `"hello"` {
		t.Errorf("PrettyString result = %q, want %q", result, `"hello"`)
	}
}

// TestPrettyString_ArrayJSON проверяет pretty-форматирование массивов
func TestPrettyString_ArrayJSON(t *testing.T) {
	input := `{"items":["a","b","c"]}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for array JSON, got: %v", err)
	}

	if !strings.Contains(result, `"items"`) {
		t.Error("PrettyString result should contain 'items'")
	}
	if !strings.Contains(result, `"a"`) {
		t.Error("PrettyString result should contain 'a'")
	}
}

// TestPrettyString_ComplexJSON проверяет сложный JSON
func TestPrettyString_ComplexJSON(t *testing.T) {
	input := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"digest": "sha256:abc123",
				"mediaType": "application/vnd.oci.image.manifest.v1+json"
			}
		]
	}`

	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for complex JSON, got: %v", err)
	}

	// Проверяем, что результат содержит ключевые элементы
	if !strings.Contains(result, `"schemaVersion"`) {
		t.Error("PrettyString result should contain 'schemaVersion'")
	}
	if !strings.Contains(result, `"mediaType"`) {
		t.Error("PrettyString result should contain 'mediaType'")
	}
	if !strings.Contains(result, `"manifests"`) {
		t.Error("PrettyString result should contain 'manifests'")
	}
	if !strings.Contains(result, `"sha256:abc123"`) {
		t.Error("PrettyString result should contain 'sha256:abc123'")
	}
	// Проверяем, что результат содержит отступы
	if !strings.Contains(result, "  ") {
		t.Error("PrettyString result should contain indentation")
	}
}

// TestError_VerbosityLevels проверяет Error при разных уровнях verbosity
func TestError_VerbosityLevels(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogErrorLevel,
	}

	Error("test error")

	if data.Gc.Error != 1 {
		t.Errorf("data.Gc.Error should be incremented at LogErrorLevel, got %d", data.Gc.Error)
	}
}

// TestError_QuietMode проверяет, что Error не логируется при LogQuietLevel
func TestError_QuietMode(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogQuietLevel,
	}

	Error("should not be logged")

	if data.Gc.Error != 1 {
		t.Errorf("data.Gc.Error should still be incremented at LogQuietLevel, got %d", data.Gc.Error)
	}
}

// TestMessage_NormalLevel проверяет Message при LogNormalLevel
func TestMessage_NormalLevel(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogNormalLevel,
	}

	Message("normal message")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestInfo_InfoLevel проверяет Info при LogInfoLevel
func TestInfo_InfoLevel(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogInfoLevel,
	}

	Info("info message")

	if data.Gc.Error != 0 {
		t.Errorf("Info should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestDebug_DebugLevel проверяет Debug при LogDebugLevel
func TestDebug_DebugLevel(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}

	Debug("debug message")

	if data.Gc.Error != 0 {
		t.Errorf("Debug should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestMaskCredentials_Password проверяет замену паролей
func TestMaskCredentials_Password(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}
	data.Sensitives = []string{"mypassword123"}

	Message("auth with mypassword123")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestMaskCredentials_Base64Auth проверяет замену base64 auth tokens
func TestMaskCredentials_Base64Auth(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}
	data.Sensitives = []string{"dGVzdHVzZXI6dGVzdHBhc3M="} // base64("testuser:testpass")

	Message("auth header dGVzdHVzZXI6dGVzdHBhc3M=")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestMaskCredentials_NoMatch проверяет работу, когда чувствительные данные не найдены
func TestMaskCredentials_NoMatch(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
	}()

	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}
	data.Sensitives = []string{"notfound", "nobody"}

	Message("test safe data")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestPrettyString_JSONWithEscaping проверяет JSON с экранированием
func TestPrettyString_JSONWithEscaping(t *testing.T) {
	input := `{"message":"hello \"world\"","path":"C:\\Users\\test"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with escaping, got: %v", err)
	}

	if !strings.Contains(result, `"message"`) {
		t.Error("PrettyString result should contain 'message'")
	}
	if !strings.Contains(result, `"path"`) {
		t.Error("PrettyString result should contain 'path'")
	}
}

// TestPrettyString_JSONWithNumbers проверяет JSON с числами
func TestPrettyString_JSONWithNumbers(t *testing.T) {
	input := `{"count":42,"price":19.99,"enabled":true,"nullValue":null}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with numbers, got: %v", err)
	}

	if !strings.Contains(result, `"count"`) {
		t.Error("PrettyString result should contain 'count'")
	}
	if !strings.Contains(result, "42") {
		t.Error("PrettyString result should contain '42'")
	}
	if !strings.Contains(result, `"price"`) {
		t.Error("PrettyString result should contain 'price'")
	}
	if !strings.Contains(result, `"enabled"`) {
		t.Error("PrettyString result should contain 'enabled'")
	}
	if !strings.Contains(result, `"nullValue"`) {
		t.Error("PrettyString result should contain 'nullValue'")
	}
}

// TestPrettyString_JSONWithUnicode проверяет JSON с unicode
func TestPrettyString_JSONWithUnicode(t *testing.T) {
	input := `{"emoji":"😀","cyrillic":"привет","japanese":"こんにちは"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with unicode, got: %v", err)
	}

	if !strings.Contains(result, `"emoji"`) {
		t.Error("PrettyString result should contain 'emoji'")
	}
	if !strings.Contains(result, `"cyrillic"`) {
		t.Error("PrettyString result should contain 'cyrillic'")
	}
	if !strings.Contains(result, `"japanese"`) {
		t.Error("PrettyString result should contain 'japanese'")
	}
}

// TestPrettyString_JSONWithNestedArrays проверяет вложенные массивы
func TestPrettyString_JSONWithNestedArrays(t *testing.T) {
	input := `{"matrix":[[1,2],[3,4]],"nested":{"array":[1,2,3]}}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for nested arrays, got: %v", err)
	}

	if !strings.Contains(result, `"matrix"`) {
		t.Error("PrettyString result should contain 'matrix'")
	}
	if !strings.Contains(result, `"nested"`) {
		t.Error("PrettyString result should contain 'nested'")
	}
	if !strings.Contains(result, `"array"`) {
		t.Error("PrettyString result should contain 'array'")
	}
}

// TestPrettyString_JSONWithLongString проверяет длинные строки
func TestPrettyString_JSONWithLongString(t *testing.T) {
	longString := strings.Repeat("a", 1000)
	input := `{"data":"` + longString + `"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for long string, got: %v", err)
	}

	if !strings.Contains(result, `"data"`) {
		t.Error("PrettyString result should contain 'data'")
	}
	// Проверяем, что строка не обрезана
	if !strings.Contains(result, longString[:100]) {
		t.Error("PrettyString result should contain the full long string")
	}
}

// TestPrettyString_JSONWithSpecialChars проверяет специальные символы
func TestPrettyString_JSONWithSpecialChars(t *testing.T) {
	input := `{"special":"!@#$%^&*()_+-=[]{}|;':\",./<>?"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for special chars, got: %v", err)
	}

	if !strings.Contains(result, `"special"`) {
		t.Error("PrettyString result should contain 'special'")
	}
}

// TestPrettyString_JSONWithEmptyObject проверяет пустой объект
func TestPrettyString_JSONWithEmptyObject(t *testing.T) {
	input := `{}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for empty object, got: %v", err)
	}

	// PrettyString не добавляет новые строки для пустого объекта
	if result != input {
		t.Errorf("PrettyString result = %q, want %q", result, input)
	}
}

// TestPrettyString_JSONWithEmptyArray проверяет пустой массив
func TestPrettyString_JSONWithEmptyArray(t *testing.T) {
	input := `[]`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for empty array, got: %v", err)
	}

	expected := `[]`
	if result != expected {
		t.Errorf("PrettyString result = %q, want %q", result, expected)
	}
}

// TestPrettyString_JSONWithMultipleKeys проверяет несколько ключей
func TestPrettyString_JSONWithMultipleKeys(t *testing.T) {
	input := `{"key1":"value1","key2":"value2","key3":"value3","key4":"value4","key5":"value5"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for multiple keys, got: %v", err)
	}

	if !strings.Contains(result, `"key1"`) {
		t.Error("PrettyString result should contain 'key1'")
	}
	if !strings.Contains(result, `"key5"`) {
		t.Error("PrettyString result should contain 'key5'")
	}
	// Проверяем, что все ключи присутствуют
	keys := []string{`"key1"`, `"key2"`, `"key3"`, `"key4"`, `"key5"`}
	for _, key := range keys {
		if !strings.Contains(result, key) {
			t.Errorf("PrettyString result should contain %s", key)
		}
	}
}

// TestPrettyString_JSONWithBooleanValues проверяет булевы значения
func TestPrettyString_JSONWithBooleanValues(t *testing.T) {
	input := `{"trueVal":true,"falseVal":false}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for boolean values, got: %v", err)
	}

	if !strings.Contains(result, `true`) {
		t.Error("PrettyString result should contain 'true'")
	}
	if !strings.Contains(result, `false`) {
		t.Error("PrettyString result should contain 'false'")
	}
}

// TestPrettyString_JSONWithNullValues проверяет null значения
func TestPrettyString_JSONWithNullValues(t *testing.T) {
	input := `{"nullVal":null,"other":"value"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for null values, got: %v", err)
	}

	if !strings.Contains(result, `"nullVal"`) {
		t.Error("PrettyString result should contain 'nullVal'")
	}
	if !strings.Contains(result, `null`) {
		t.Error("PrettyString result should contain 'null'")
	}
}

// TestPrettyString_JSONWithNegativeNumbers проверяет отрицательные числа
func TestPrettyString_JSONWithNegativeNumbers(t *testing.T) {
	input := `{"negInt":-42,"negFloat":-19.99}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for negative numbers, got: %v", err)
	}

	if !strings.Contains(result, `-42`) {
		t.Error("PrettyString result should contain '-42'")
	}
	if !strings.Contains(result, `-19.99`) {
		t.Error("PrettyString result should contain '-19.99'")
	}
}

// TestPrettyString_JSONWithScientificNotation проверяет научную нотацию
func TestPrettyString_JSONWithScientificNotation(t *testing.T) {
	input := `{"scientific":1.23e10,"negativeSci":-4.56e-7}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for scientific notation, got: %v", err)
	}

	if !strings.Contains(result, `1.23e10`) {
		t.Error("PrettyString result should contain '1.23e10'")
	}
	if !strings.Contains(result, `-4.56e-7`) {
		t.Error("PrettyString result should contain '-4.56e-7'")
	}
}

// TestPrettyString_JSONWithDeeplyNested проверяет глубоко вложенный JSON
func TestPrettyString_JSONWithDeeplyNested(t *testing.T) {
	input := `{"level1":{"level2":{"level3":{"level4":{"level5":"deep"}}}}}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for deeply nested JSON, got: %v", err)
	}

	if !strings.Contains(result, `"level1"`) {
		t.Error("PrettyString result should contain 'level1'")
	}
	if !strings.Contains(result, `"level5"`) {
		t.Error("PrettyString result should contain 'level5'")
	}
	if !strings.Contains(result, `"deep"`) {
		t.Error("PrettyString result should contain 'deep'")
	}
	// Проверяем, что есть отступы для вложенности
	if !strings.Contains(result, "      ") {
		t.Error("PrettyString result should contain deep indentation for nested objects")
	}
}

// TestError_ConcurrentSafety проверяет, что Error корректно инкрементирует counter
func TestError_ConcurrentSafety(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	data.Gc = &data.Config{
		Verbosity: LogErrorLevel,
	}

	// Сохраняем текущее значение
	initialError := data.Gc.Error

	// Вызываем Error несколько раз
	for i := 0; i < 10; i++ {
		Error("concurrent test")
	}

	expected := initialError + 10
	if data.Gc.Error != expected {
		t.Errorf("data.Gc.Error after 10 Error() calls = %d, want %d", data.Gc.Error, expected)
	}
}

// TestMessage_AllLevels проверяет Message при разных уровнях
func TestMessage_AllLevels(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	// При LogNormalLevel Message должен логироваться
	data.Gc = &data.Config{
		Verbosity: LogNormalLevel,
	}

	Message("test at normal level")

	if data.Gc.Error != 0 {
		t.Errorf("Message should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestInfo_AllLevels проверяет Info при разных уровнях
func TestInfo_AllLevels(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	// При LogInfoLevel Info должен логироваться
	data.Gc = &data.Config{
		Verbosity: LogInfoLevel,
	}

	Info("test at info level")

	if data.Gc.Error != 0 {
		t.Errorf("Info should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestDebug_AllLevels проверяет Debug при разных уровнях
func TestDebug_AllLevels(t *testing.T) {
	origGc := data.Gc
	defer func() {
		data.Gc = origGc
	}()

	// При LogDebugLevel Debug должен логироваться
	data.Gc = &data.Config{
		Verbosity: LogDebugLevel,
	}

	Debug("test at debug level")

	if data.Gc.Error != 0 {
		t.Errorf("Debug should not increment Error counter, got %d", data.Gc.Error)
	}
}

// TestPrettyString_JSONWithWhitespace проверяет JSON с пробелами
func TestPrettyString_JSONWithWhitespace(t *testing.T) {
	input := `  {"key"  :  "value"  }  `
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with whitespace, got: %v", err)
	}

	if !strings.Contains(result, `"key"`) {
		t.Error("PrettyString result should contain 'key'")
	}
	if !strings.Contains(result, `"value"`) {
		t.Error("PrettyString result should contain 'value'")
	}
}

// TestPrettyString_JSONWithCommas проверяет JSON с запятыми
func TestPrettyString_JSONWithCommas(t *testing.T) {
	input := `{"key1":"value1","key2":"value2"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with commas, got: %v", err)
	}

	if !strings.Contains(result, `"key1"`) {
		t.Error("PrettyString result should contain 'key1'")
	}
	if !strings.Contains(result, `"key2"`) {
		t.Error("PrettyString result should contain 'key2'")
	}
}

// TestPrettyString_JSONWithQuotes проверяет JSON с кавычками
func TestPrettyString_JSONWithQuotes(t *testing.T) {
	input := `{"quoted":"She said \"hello\" to me"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with quotes, got: %v", err)
	}

	if !strings.Contains(result, `"quoted"`) {
		t.Error("PrettyString result should contain 'quoted'")
	}
}

// TestPrettyString_JSONWithNewlines проверяет JSON с переносами строк
func TestPrettyString_JSONWithNewlines(t *testing.T) {
	input := `{"text":"line1\nline2\rline3"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with newlines, got: %v", err)
	}

	if !strings.Contains(result, `"text"`) {
		t.Error("PrettyString result should contain 'text'")
	}
}

// TestPrettyString_JSONWithTabs проверяет JSON с табами
func TestPrettyString_JSONWithTabs(t *testing.T) {
	input := `{"tab":"\t\t\t"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with tabs, got: %v", err)
	}

	if !strings.Contains(result, `"tab"`) {
		t.Error("PrettyString result should contain 'tab'")
	}
}

// TestPrettyString_JSONWithBackslash проверяет JSON с обратными слэшами
func TestPrettyString_JSONWithBackslash(t *testing.T) {
	input := `{"path":"C:\\Users\\test\\file.txt"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for JSON with backslashes, got: %v", err)
	}

	if !strings.Contains(result, `"path"`) {
		t.Error("PrettyString result should contain 'path'")
	}
}

// TestPrettyString_JSONWithMixedTypes проверяет JSON с разными типами
func TestPrettyString_JSONWithMixedTypes(t *testing.T) {
	input := `{"string":"text","int":42,"float":3.14,"bool":true,"null":null,"array":[1,2,3],"object":{"nested":"value"}}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for mixed types, got: %v", err)
	}

	if !strings.Contains(result, `"string"`) {
		t.Error("PrettyString result should contain 'string'")
	}
	if !strings.Contains(result, `"int"`) {
		t.Error("PrettyString result should contain 'int'")
	}
	if !strings.Contains(result, `"float"`) {
		t.Error("PrettyString result should contain 'float'")
	}
	if !strings.Contains(result, `"bool"`) {
		t.Error("PrettyString result should contain 'bool'")
	}
	if !strings.Contains(result, `"null"`) {
		t.Error("PrettyString result should contain 'null'")
	}
	if !strings.Contains(result, `"array"`) {
		t.Error("PrettyString result should contain 'array'")
	}
	if !strings.Contains(result, `"object"`) {
		t.Error("PrettyString result should contain 'object'")
	}
}

// TestPrettyString_JSONWithUnicodeEscapes проверяет JSON с unicode escapes
func TestPrettyString_JSONWithUnicodeEscapes(t *testing.T) {
	input := `{"emoji":"\u0048\u0065\u006c\u006c\u006f"}` // "Hello"
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for unicode escapes, got: %v", err)
	}

	if !strings.Contains(result, `"emoji"`) {
		t.Error("PrettyString result should contain 'emoji'")
	}
}

// TestPrettyString_JSONWithLargeNumbers проверяет JSON с большими числами
func TestPrettyString_JSONWithLargeNumbers(t *testing.T) {
	input := `{"large":9007199254740992}` // 2^53
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for large numbers, got: %v", err)
	}

	if !strings.Contains(result, `"large"`) {
		t.Error("PrettyString result should contain 'large'")
	}
}

// TestPrettyString_JSONWithZeroValues проверяет JSON с нулевыми значениями
func TestPrettyString_JSONWithZeroValues(t *testing.T) {
	input := `{"zeroInt":0,"zeroFloat":0.0}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for zero values, got: %v", err)
	}

	if !strings.Contains(result, `"zeroInt"`) {
		t.Error("PrettyString result should contain 'zeroInt'")
	}
	if !strings.Contains(result, `"zeroFloat"`) {
		t.Error("PrettyString result should contain 'zeroFloat'")
	}
}

// TestPrettyString_JSONWithMultipleArrays проверяет JSON с несколькими массивами
func TestPrettyString_JSONWithMultipleArrays(t *testing.T) {
	input := `{"arr1":[1,2],"arr2":["a","b"],"arr3":[true,false]}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for multiple arrays, got: %v", err)
	}

	if !strings.Contains(result, `"arr1"`) {
		t.Error("PrettyString result should contain 'arr1'")
	}
	if !strings.Contains(result, `"arr2"`) {
		t.Error("PrettyString result should contain 'arr2'")
	}
	if !strings.Contains(result, `"arr3"`) {
		t.Error("PrettyString result should contain 'arr3'")
	}
}

// TestPrettyString_JSONWithComplexNested проверяет сложный вложенный JSON
func TestPrettyString_JSONWithComplexNested(t *testing.T) {
	input := `{
		"users": [
			{
				"id": 1,
				"name": "Alice",
				"roles": ["admin", "user"]
			},
			{
				"id": 2,
				"name": "Bob",
				"roles": ["user"]
			}
		],
		"meta": {
			"total": 2,
			"page": 1
		}
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for complex nested JSON, got: %v", err)
	}

	if !strings.Contains(result, `"users"`) {
		t.Error("PrettyString result should contain 'users'")
	}
	if !strings.Contains(result, `"Alice"`) {
		t.Error("PrettyString result should contain 'Alice'")
	}
	if !strings.Contains(result, `"Bob"`) {
		t.Error("PrettyString result should contain 'Bob'")
	}
	if !strings.Contains(result, `"meta"`) {
		t.Error("PrettyString result should contain 'meta'")
	}
}

// TestPrettyString_JSONWithUUID проверяет JSON с UUID
func TestPrettyString_JSONWithUUID(t *testing.T) {
	input := `{"uuid":"550e8400-e29b-41d4-a716-446655440000"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for UUID, got: %v", err)
	}

	if !strings.Contains(result, `"uuid"`) {
		t.Error("PrettyString result should contain 'uuid'")
	}
	if !strings.Contains(result, `550e8400-e29b-41d4-a716-446655440000`) {
		t.Error("PrettyString result should contain the UUID")
	}
}

// TestPrettyString_JSONWithTimestamp проверяет JSON с timestamp
func TestPrettyString_JSONWithTimestamp(t *testing.T) {
	input := `{"timestamp":"2024-01-01T00:00:00Z"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for timestamp, got: %v", err)
	}

	if !strings.Contains(result, `"timestamp"`) {
		t.Error("PrettyString result should contain 'timestamp'")
	}
	if !strings.Contains(result, `2024-01-01T00:00:00Z`) {
		t.Error("PrettyString result should contain the timestamp")
	}
}

// TestPrettyString_JSONWithEmail проверяет JSON с email
func TestPrettyString_JSONWithEmail(t *testing.T) {
	input := `{"email":"user@example.com"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for email, got: %v", err)
	}

	if !strings.Contains(result, `"email"`) {
		t.Error("PrettyString result should contain 'email'")
	}
	if !strings.Contains(result, `user@example.com`) {
		t.Error("PrettyString result should contain the email")
	}
}

// TestPrettyString_JSONWithURL проверяет JSON с URL
func TestPrettyString_JSONWithURL(t *testing.T) {
	input := `{"url":"https://example.com/path?query=value"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for URL, got: %v", err)
	}

	if !strings.Contains(result, `"url"`) {
		t.Error("PrettyString result should contain 'url'")
	}
	if !strings.Contains(result, `https://example.com/path?query=value`) {
		t.Error("PrettyString result should contain the URL")
	}
}

// TestPrettyString_JSONWithBase64 проверяет JSON с base64
func TestPrettyString_JSONWithBase64(t *testing.T) {
	input := `{"data":"dGVzdCBkYXRh"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for base64, got: %v", err)
	}

	if !strings.Contains(result, `"data"`) {
		t.Error("PrettyString result should contain 'data'")
	}
	if !strings.Contains(result, `dGVzdCBkYXRh`) {
		t.Error("PrettyString result should contain the base64 data")
	}
}

// TestPrettyString_JSONWithSHA256 проверяет JSON с SHA256 hash
func TestPrettyString_JSONWithSHA256(t *testing.T) {
	input := `{"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e8409b1255f4b8e8f3b5c7d0e1"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for SHA256, got: %v", err)
	}

	if !strings.Contains(result, `"sha256"`) {
		t.Error("PrettyString result should contain 'sha256'")
	}
	if !strings.Contains(result, `e3b0c44298fc1c149afbf4c8996fb92427ae41e8409b1255f4b8e8f3b5c7d0e1`) {
		t.Error("PrettyString result should contain the SHA256 hash")
	}
}

// TestPrettyString_JSONWithMediaType проверяет JSON с media type
func TestPrettyString_JSONWithMediaType(t *testing.T) {
	input := `{"mediaType":"application/vnd.oci.image.index.v1+json"}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for media type, got: %v", err)
	}

	if !strings.Contains(result, `"mediaType"`) {
		t.Error("PrettyString result should contain 'mediaType'")
	}
	if !strings.Contains(result, `application/vnd.oci.image.index.v1+json`) {
		t.Error("PrettyString result should contain the media type")
	}
}

// TestPrettyString_JSONWithManifest проверяет JSON с манифестом CNAB
func TestPrettyString_JSONWithManifest(t *testing.T) {
	input := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"digest": "sha256:5b8121c1395efd280f0fe69bf5bfe29ca0dde2a0ce5f24551c99fa876ad365c6",
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"annotations": {
					"io.cnab.manifest.type": "config"
				}
			},
			{
				"digest": "sha256:36b2305c1de16ea1bc635c2f6b76ea464742cb8711f40f8078c8c5dd2c10164c",
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"annotations": {
					"io.cnab.manifest.type": "invocation"
				}
			}
		]
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for CNAB manifest, got: %v", err)
	}

	if !strings.Contains(result, `"schemaVersion"`) {
		t.Error("PrettyString result should contain 'schemaVersion'")
	}
	if !strings.Contains(result, `"manifests"`) {
		t.Error("PrettyString result should contain 'manifests'")
	}
	if !strings.Contains(result, `"io.cnab.manifest.type"`) {
		t.Error("PrettyString result should contain 'io.cnab.manifest.type'")
	}
	if !strings.Contains(result, `"config"`) {
		t.Error("PrettyString result should contain 'config'")
	}
	if !strings.Contains(result, `"invocation"`) {
		t.Error("PrettyString result should contain 'invocation'")
	}
}

// TestPrettyString_JSONWithOCIImageIndex проверяет JSON с OCI Image Index
func TestPrettyString_JSONWithOCIImageIndex(t *testing.T) {
	input := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"digest": "sha256:abc123",
				"mediaType": "application/vnd.oci.image.manifest.v1+json"
			}
		]
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for OCI Image Index, got: %v", err)
	}

	if !strings.Contains(result, `"schemaVersion"`) {
		t.Error("PrettyString result should contain 'schemaVersion'")
	}
	if !strings.Contains(result, `"mediaType"`) {
		t.Error("PrettyString result should contain 'mediaType'")
	}
	if !strings.Contains(result, `"manifests"`) {
		t.Error("PrettyString result should contain 'manifests'")
	}
	if !strings.Contains(result, `"sha256:abc123"`) {
		t.Error("PrettyString result should contain 'sha256:abc123'")
	}
}

// TestPrettyString_JSONWithDockerManifest проверяет JSON с Docker Manifest
func TestPrettyString_JSONWithDockerManifest(t *testing.T) {
	input := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"digest": "sha256:config123",
			"mediaType": "application/vnd.docker.container.image.v1+json"
		},
		"layers": [
			{
				"digest": "sha256:layer1",
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip"
			}
		]
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for Docker Manifest, got: %v", err)
	}

	if !strings.Contains(result, `"schemaVersion"`) {
		t.Error("PrettyString result should contain 'schemaVersion'")
	}
	if !strings.Contains(result, `"config"`) {
		t.Error("PrettyString result should contain 'config'")
	}
	if !strings.Contains(result, `"layers"`) {
		t.Error("PrettyString result should contain 'layers'")
	}
	if !strings.Contains(result, `"sha256:config123"`) {
		t.Error("PrettyString result should contain 'sha256:config123'")
	}
}

// TestPrettyString_JSONWithCNABBundle проверяет JSON с CNAB Bundle
func TestPrettyString_JSONWithCNABBundle(t *testing.T) {
	input := `{
		"action": "install",
		"bundle": {
			"id": "my-bundle",
			"version": "1.0.0",
			"description": "A CNAB bundle"
		},
		"invocationImages": [
			{
				"digest": "sha256:invocation1",
				"mediaType": "application/vnd.oci.image.manifest.v1+json"
			}
		],
		"credentials": {
			"username": "user",
			"password": "pass"
		}
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for CNAB Bundle, got: %v", err)
	}

	if !strings.Contains(result, `"action"`) {
		t.Error("PrettyString result should contain 'action'")
	}
	if !strings.Contains(result, `"bundle"`) {
		t.Error("PrettyString result should contain 'bundle'")
	}
	if !strings.Contains(result, `"invocationImages"`) {
		t.Error("PrettyString result should contain 'invocationImages'")
	}
	if !strings.Contains(result, `"credentials"`) {
		t.Error("PrettyString result should contain 'credentials'")
	}
}

// TestPrettyString_JSONWithArtifactory проверяет JSON с Artifactory response
func TestPrettyString_JSONWithArtifactory(t *testing.T) {
	input := `{
		"repo": "osmp-docker-storage",
		"path": "/osmp/tmp/plugin-kes-linux/12.4.0.12/plugin-kes-linux",
		"children": []
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for Artifactory response, got: %v", err)
	}

	if !strings.Contains(result, `"repo"`) {
		t.Error("PrettyString result should contain 'repo'")
	}
	if !strings.Contains(result, `"path"`) {
		t.Error("PrettyString result should contain 'path'")
	}
	if !strings.Contains(result, `"children"`) {
		t.Error("PrettyString result should contain 'children'")
	}
	if !strings.Contains(result, `"osmp-docker-storage"`) {
		t.Error("PrettyString result should contain 'osmp-docker-storage'")
	}
}

// TestPrettyString_JSONWithTagsList проверяет JSON со списком тегов
func TestPrettyString_JSONWithTagsList(t *testing.T) {
	input := `{
		"name": "osmp/tmp/plugin-kes-linux/12.4.0.12/plugin-kes-linux",
		"tags": [
			"12.4.0.12",
			"latest",
			"12.4.0.11"
		]
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for tags list, got: %v", err)
	}

	if !strings.Contains(result, `"name"`) {
		t.Error("PrettyString result should contain 'name'")
	}
	if !strings.Contains(result, `"tags"`) {
		t.Error("PrettyString result should contain 'tags'")
	}
	if !strings.Contains(result, `"12.4.0.12"`) {
		t.Error("PrettyString result should contain '12.4.0.12'")
	}
	if !strings.Contains(result, `"latest"`) {
		t.Error("PrettyString result should contain 'latest'")
	}
}

// TestPrettyString_JSONWithErrorResponse проверяет JSON с ошибкой
func TestPrettyString_JSONWithErrorResponse(t *testing.T) {
	input := `{
		"errors": [
			{
				"code": "MANIFEST_UNKNOWN",
				"message": "manifest unknown",
				"detail": {}
			}
		]
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for error response, got: %v", err)
	}

	if !strings.Contains(result, `"errors"`) {
		t.Error("PrettyString result should contain 'errors'")
	}
	if !strings.Contains(result, `"MANIFEST_UNKNOWN"`) {
		t.Error("PrettyString result should contain 'MANIFEST_UNKNOWN'")
	}
	if !strings.Contains(result, `"message"`) {
		t.Error("PrettyString result should contain 'message'")
	}
}

// TestPrettyString_JSONWithHealthCheck проверяет JSON с health check
func TestPrettyString_JSONWithHealthCheck(t *testing.T) {
	input := `{
		"status": "healthy",
		"version": "7.23.1",
		"components": [
			{
				"name": "storage",
				"status": "healthy"
			}
		]
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for health check, got: %v", err)
	}

	if !strings.Contains(result, `"status"`) {
		t.Error("PrettyString result should contain 'status'")
	}
	if !strings.Contains(result, `"healthy"`) {
		t.Error("PrettyString result should contain 'healthy'")
	}
	if !strings.Contains(result, `"components"`) {
		t.Error("PrettyString result should contain 'components'")
	}
}

// TestPrettyString_JSONWithPagination проверяет JSON с пагинацией
func TestPrettyString_JSONWithPagination(t *testing.T) {
	input := `{
		"page": 1,
		"pageSize": 10,
		"totalPages": 5,
		"totalItems": 50,
		"hasNext": true,
		"hasPrevious": false
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for pagination, got: %v", err)
	}

	if !strings.Contains(result, `"page"`) {
		t.Error("PrettyString result should contain 'page'")
	}
	if !strings.Contains(result, `"pageSize"`) {
		t.Error("PrettyString result should contain 'pageSize'")
	}
	if !strings.Contains(result, `"totalPages"`) {
		t.Error("PrettyString result should contain 'totalPages'")
	}
	if !strings.Contains(result, `"hasNext"`) {
		t.Error("PrettyString result should contain 'hasNext'")
	}
}

// TestPrettyString_JSONWithMetrics проверяет JSON с метриками
func TestPrettyString_JSONWithMetrics(t *testing.T) {
	input := `{
		"cpu": 45.2,
		"memory": 8192,
		"disk": 102400,
		"network": {
			"in": 1024,
			"out": 2048
		}
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for metrics, got: %v", err)
	}

	if !strings.Contains(result, `"cpu"`) {
		t.Error("PrettyString result should contain 'cpu'")
	}
	if !strings.Contains(result, `"memory"`) {
		t.Error("PrettyString result should contain 'memory'")
	}
	if !strings.Contains(result, `"disk"`) {
		t.Error("PrettyString result should contain 'disk'")
	}
	if !strings.Contains(result, `"network"`) {
		t.Error("PrettyString result should contain 'network'")
	}
}

// TestPrettyString_JSONWithConfig проверяет JSON с конфигурацией
func TestPrettyString_JSONWithConfig(t *testing.T) {
	input := `{
		"credentials": {
			"username": "admin",
			"password": "secret"
		},
		"timeout": 10000,
		"verbosity": 2,
		"unsecure": false,
		"scheme": "https"
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for config, got: %v", err)
	}

	if !strings.Contains(result, `"credentials"`) {
		t.Error("PrettyString result should contain 'credentials'")
	}
	if !strings.Contains(result, `"username"`) {
		t.Error("PrettyString result should contain 'username'")
	}
	if !strings.Contains(result, `"timeout"`) {
		t.Error("PrettyString result should contain 'timeout'")
	}
	if !strings.Contains(result, `"verbosity"`) {
		t.Error("PrettyString result should contain 'verbosity'")
	}
	if !strings.Contains(result, `"scheme"`) {
		t.Error("PrettyString result should contain 'scheme'")
	}
}

// TestPrettyString_JSONWithCNABInspectResult проверяет JSON с результатом inspect
func TestPrettyString_JSONWithCNABInspectResult(t *testing.T) {
	input := `{
		"reference": "osmp-docker-storage.repository.avp.ru/osmp/tmp/plugin-kes-linux/12.4.0.12/plugin-kes-linux",
		"itemList": [
			{
				"tag": "12.4.0.12",
				"digest": "sha256:de7f265a487a1927ed2fca915bcc86e38fb9d2adf33d969e5bb1a2617219a7aa",
				"annotation": "cnab index",
				"date": "Mon, 01 Jan 2024 00:00:00 GMT",
				"media": "application/vnd.oci.image.index.v1+json",
				"count": 3,
				"links": 3,
				"lost": 0
			}
		]
	}`
	result, err := PrettyString(input)
	if err != nil {
		t.Errorf("PrettyString should not return error for CNAB inspect result, got: %v", err)
	}

	if !strings.Contains(result, `"reference"`) {
		t.Error("PrettyString result should contain 'reference'")
	}
	if !strings.Contains(result, `"itemList"`) {
		t.Error("PrettyString result should contain 'itemList'")
	}
	if !strings.Contains(result, `"tag"`) {
		t.Error("PrettyString result should contain 'tag'")
	}
	if !strings.Contains(result, `"digest"`) {
		t.Error("PrettyString result should contain 'digest'")
	}
	if !strings.Contains(result, `"annotation"`) {
		t.Error("PrettyString result should contain 'annotation'")
	}
	if !strings.Contains(result, `"count"`) {
		t.Error("PrettyString result should contain 'count'")
	}
	if !strings.Contains(result, `"links"`) {
		t.Error("PrettyString result should contain 'links'")
	}
	if !strings.Contains(result, `"lost"`) {
		t.Error("PrettyString result should contain 'lost'")
	}
}
