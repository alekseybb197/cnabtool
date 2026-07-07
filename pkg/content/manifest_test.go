package content

import (
	"cnabtool/pkg/client"
	"cnabtool/pkg/data"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetManifest_ValidReference проверяет успешное получение манифеста через httptest-сервер
func TestGetManifest_ValidReference(t *testing.T) {
	// Сохраняем состояние для восстановления
	origGc := data.Gc
	origSensitives := data.Sensitives
	origScheme := data.Scheme
	origRegistry := data.Registry
	origRepository := data.Repository
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
		data.Scheme = origScheme
		data.Registry = origRegistry
		data.Repository = origRepository
	}()

	// Инициализируем Gc
	data.Gc = &data.Config{
		Verbosity: 4,
		Timeout:   10000,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}
	defer func() { data.Gc = nil }()

	// Создаём тестовый сервер
	manifestJSON := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"digest": "sha256:abc123",
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"size": 1234
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.oci.image.index.v1+json")
		w.Header().Set("Docker-Content-Digest", "sha256:abc123")
		w.Header().Set("Last-Modified", "Mon, 01 Jan 2024 00:00:00 GMT")
		w.WriteHeader(200)
		w.Write([]byte(manifestJSON))
	}))
	defer server.Close()

	// Извлекаем хост из URL сервера
	serverHost := strings.TrimPrefix(server.URL, "http://")

	// Создаём Config
	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	// Вызываем GetManifest
	cnf := (*Config)(cfg)
	regres, cl, err := cnf.GetManifest(serverHost + "/repo/image:v1@sha256:abc123")

	if err != nil {
		t.Fatalf("GetManifest should not return error, got: %v", err)
	}
	if regres == nil {
		t.Fatal("GetManifest returned nil response")
	}
	if cl == nil {
		t.Fatal("GetManifest returned nil client")
	}

	// Проверяем распарсенный клиент
	if cl.Registry != serverHost {
		t.Errorf("cl.Registry = %q, want %q", cl.Registry, serverHost)
	}
	if cl.Repository != "repo/image" {
		t.Errorf("cl.Repository = %q, want %q", cl.Repository, "repo/image")
	}
	if cl.Tag != "v1" {
		t.Errorf("cl.Tag = %q, want %q", cl.Tag, "v1")
	}
	if cl.Digest != "sha256:abc123" {
		t.Errorf("cl.Digest = %q, want %q", cl.Digest, "sha256:abc123")
	}
	if cl.Scheme != "http" {
		t.Errorf("cl.Scheme = %q, want %q", cl.Scheme, "http")
	}

	// Проверяем сохранённое глобальное состояние
	if data.Registry != serverHost {
		t.Errorf("data.Registry = %q, want %q", data.Registry, serverHost)
	}
	if data.Repository != "repo/image" {
		t.Errorf("data.Repository = %q, want %q", data.Repository, "repo/image")
	}
	if data.Scheme != "http" {
		t.Errorf("data.Scheme = %q, want %q", data.Scheme, "http")
	}

	// Проверяем ответ
	if regres.Media != "application/vnd.oci.image.index.v1+json" {
		t.Errorf("regres.Media = %q, want %q", regres.Media, "application/vnd.oci.image.index.v1+json")
	}
	if regres.Digest != "sha256:abc123" {
		t.Errorf("regres.Digest = %q, want %q", regres.Digest, "sha256:abc123")
	}
	if regres.Status != 200 {
		t.Errorf("regres.Status = %d, want 200", regres.Status)
	}
}

// TestGetManifest_InvalidReference проверяет обработку невалидной ссылки
func TestGetManifest_InvalidReference(t *testing.T) {
	origGc := data.Gc
	defer func() { data.Gc = origGc }()

	data.Gc = &data.Config{
		Verbosity: 1,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}

	testcases := []struct {
		name string
		ref  string
	}{
		{
			name: "no slash",
			ref:  "noslash",
		},
		{
			name: "registry without dot",
			ref:  "localhost/repo/image:v1",
		},
		{
			name: "empty reference",
			ref:  "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &data.Config{
				Scheme: "https",
				Credentials: data.Credentials{
					Username: "testuser",
					Password: "testpass",
				},
				Client:  "cnabtool/0.1.1",
				Timeout: 10000,
			}
			cnf := (*Config)(cfg)
			_, _, err := cnf.GetManifest(tc.ref)
			if err == nil {
				t.Errorf("GetManifest(%q) should return error, got nil", tc.ref)
			}
		})
	}
}

// TestGetManifest_TagOnly проверяет ссылку только с тегом (без digest)
func TestGetManifest_TagOnly(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	origScheme := data.Scheme
	origRegistry := data.Registry
	origRepository := data.Repository
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
		data.Scheme = origScheme
		data.Registry = origRegistry
		data.Repository = origRepository
	}()

	data.Gc = &data.Config{
		Verbosity: 4,
		Timeout:   10000,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}
	defer func() { data.Gc = nil }()

	manifestJSON := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": []
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.oci.image.index.v1+json")
		w.Header().Set("Docker-Content-Digest", "sha256:tagonly123")
		w.WriteHeader(200)
		w.Write([]byte(manifestJSON))
	}))
	defer server.Close()

	serverHost := strings.TrimPrefix(server.URL, "http://")

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cnf := (*Config)(cfg)
	_, cl, err := cnf.GetManifest(serverHost + "/repo/image:v2.0")

	if err != nil {
		t.Fatalf("GetManifest should not return error, got: %v", err)
	}

	if cl.Digest != "" {
		t.Errorf("cl.Digest should be empty for tag-only reference, got %q", cl.Digest)
	}
	if cl.Tag != "v2.0" {
		t.Errorf("cl.Tag = %q, want %q", cl.Tag, "v2.0")
	}
}

// TestResponsePrettyPrint_ValidJSON проверяет pretty-печать валидного ответа
func TestResponsePrettyPrint_ValidJSON(t *testing.T) {
	// Сохраняем вывод для проверки
	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Status:    200,
		Media:     "application/vnd.oci.image.index.v1+json",
		Digest:    "sha256:abc123",
		Date:      "Mon, 01 Jan 2024 00:00:00 GMT",
		Content:   `{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json"}`,
	}

	// Просто проверяем, что функция не паникует
	ResponsePrettyPrint(regres)
}

// TestResponsePrettyPrint_EmptyContent проверяет печатание пустого ответа
func TestResponsePrettyPrint_EmptyContent(t *testing.T) {
	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Status:    404,
		Media:     "",
		Digest:    "",
		Content:   "",
	}

	// Проверяем, что функция не паникует
	ResponsePrettyPrint(regres)
}

// TestResponsePrettyPrint_NonJSONContent проверяет печатание не-JSON контента
func TestResponsePrettyPrint_NonJSONContent(t *testing.T) {
	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Status:    200,
		Media:     "application/octet-stream",
		Digest:    "sha256:nonjson",
		Content:   `not a json content`,
	}

	// Проверяем, что функция не паникует
	ResponsePrettyPrint(regres)
}

// TestResponsePrettyPrint_NestedJSON проверяет pretty-печать вложенного JSON
func TestResponsePrettyPrint_NestedJSON(t *testing.T) {
	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Status:    200,
		Media:     "application/vnd.oci.image.index.v1+json",
		Digest:    "sha256:nested",
		Content: `{
			"schemaVersion": 2,
			"manifests": [
				{
					"digest": "sha256:abc",
					"mediaType": "application/vnd.oci.image.manifest.v1+json"
				}
			]
		}`,
	}

	// Проверяем, что функция не паникует
	ResponsePrettyPrint(regres)
}

// TestResponsePrettyPrint_JSONWithSpecialChars проверяет JSON со спецсимволами
func TestResponsePrettyPrint_JSONWithSpecialChars(t *testing.T) {
	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Status:    200,
		Media:     "application/vnd.oci.image.index.v1+json",
		Digest:    "sha256:special",
		Content:   `{"path":"C:\\Users\\test","quote":"She said \"hello\""}`,
	}

	// Проверяем, что функция не паникует
	ResponsePrettyPrint(regres)
}

// TestResponsePrettyPrint_LargeContent проверяет печатание большого контента
func TestResponsePrettyPrint_LargeContent(t *testing.T) {
	// Генерируем большой JSON
	manifests := []map[string]interface{}{}
	for i := 0; i < 50; i++ {
		manifests = append(manifests, map[string]interface{}{
			"digest":    "sha256:" + string(rune(i)),
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"size":      1234 + i,
			"platform":  map[string]string{"os": "linux", "arch": "amd64"},
		})
	}
	content, _ := json.Marshal(map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests":     manifests,
	})

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Status:    200,
		Media:     "application/vnd.oci.image.index.v1+json",
		Digest:    "sha256:large",
		Content:   string(content),
	}

	// Проверяем, что функция не паникует на большом контенте
	ResponsePrettyPrint(regres)
}

// TestConfig_TypeAlias проверяет, что Config является тип-алиасом data.Config
func TestConfig_TypeAlias(t *testing.T) {
	origGc := data.Gc
	defer func() { data.Gc = origGc }()

	data.Gc = &data.Config{
		Verbosity: 2,
		Timeout:   10000,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}

	cfg := &data.Config{
		Scheme:  "https",
		Raw:     true,
		DryRun:  true,
		Purge:   true,
		RepoKey: "my-repo",
		Client:  "cnabtool/0.1.1",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}

	// Создаём Config из data.Config
	cnf := (*Config)(cfg)

	// Проверяем, что поля доступны
	if cnf.Scheme != "https" {
		t.Errorf("cnf.Scheme = %q, want %q", cnf.Scheme, "https")
	}
	if cnf.Raw != true {
		t.Errorf("cnf.Raw = %v, want true", cnf.Raw)
	}
	if cnf.DryRun != true {
		t.Errorf("cnf.DryRun = %v, want true", cnf.DryRun)
	}
	if cnf.Purge != true {
		t.Errorf("cnf.Purge = %v, want true", cnf.Purge)
	}
	if cnf.RepoKey != "my-repo" {
		t.Errorf("cnf.RepoKey = %q, want %q", cnf.RepoKey, "my-repo")
	}
}

// TestGetManifest_DockerManifest проверяет получение docker manifest v2
func TestGetManifest_DockerManifest(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	origScheme := data.Scheme
	origRegistry := data.Registry
	origRepository := data.Repository
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
		data.Scheme = origScheme
		data.Registry = origRegistry
		data.Repository = origRepository
	}()

	data.Gc = &data.Config{
		Verbosity: 4,
		Timeout:   10000,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}
	defer func() { data.Gc = nil }()

	dockerManifest := `{
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
		w.Header().Set("Docker-Content-Digest", "sha256:docker123")
		w.WriteHeader(200)
		w.Write([]byte(dockerManifest))
	}))
	defer server.Close()

	serverHost := strings.TrimPrefix(server.URL, "http://")

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cnf := (*Config)(cfg)
	regres, _, err := cnf.GetManifest(serverHost + "/repo/image:v1@sha256:docker123")

	if err != nil {
		t.Fatalf("GetManifest should not return error, got: %v", err)
	}

	if regres.Media != "application/vnd.docker.distribution.manifest.v2+json" {
		t.Errorf("regres.Media = %q, want %q", regres.Media, "application/vnd.docker.distribution.manifest.v2+json")
	}
	if regres.Digest != "sha256:docker123" {
		t.Errorf("regres.Digest = %q, want %q", regres.Digest, "sha256:docker123")
	}
}

// TestGetManifest_ErrorResponse проверяет обработку HTTP-ошибки от сервера
func TestGetManifest_ErrorResponse(t *testing.T) {
	origGc := data.Gc
	defer func() { data.Gc = origGc }()

	data.Gc = &data.Config{
		Verbosity: 1,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}

	// Сервер возвращает 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	serverHost := strings.TrimPrefix(server.URL, "http://")

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cnf := (*Config)(cfg)
	_, _, err := cnf.GetManifest(serverHost + "/repo/image:v1@sha256:abc")

	if err == nil {
		t.Error("GetManifest should return error for 500 response")
	}
}

// TestGetManifest_Unauthorized проверяет обработку 401
func TestGetManifest_Unauthorized(t *testing.T) {
	origGc := data.Gc
	defer func() { data.Gc = origGc }()

	data.Gc = &data.Config{
		Verbosity: 1,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	serverHost := strings.TrimPrefix(server.URL, "http://")

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cnf := (*Config)(cfg)
	_, _, err := cnf.GetManifest(serverHost + "/repo/image:v1@sha256:abc")

	if err == nil {
		t.Error("GetManifest should return error for 401 response")
	}
}

// TestGetManifest_DigestOnlyReference проверяет ссылку только с digest (без тега)
func TestGetManifest_DigestOnlyReference(t *testing.T) {
	origGc := data.Gc
	origSensitives := data.Sensitives
	origScheme := data.Scheme
	origRegistry := data.Registry
	origRepository := data.Repository
	defer func() {
		data.Gc = origGc
		data.Sensitives = origSensitives
		data.Scheme = origScheme
		data.Registry = origRegistry
		data.Repository = origRepository
	}()

	data.Gc = &data.Config{
		Verbosity: 4,
		Timeout:   10000,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}
	defer func() { data.Gc = nil }()

	manifestJSON := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": []
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.oci.image.index.v1+json")
		w.Header().Set("Docker-Content-Digest", "sha256:digestonly")
		w.WriteHeader(200)
		w.Write([]byte(manifestJSON))
	}))
	defer server.Close()

	serverHost := strings.TrimPrefix(server.URL, "http://")

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cnf := (*Config)(cfg)
	_, cl, err := cnf.GetManifest(serverHost + "/repo/image@sha256:digestonly")

	if err != nil {
		t.Fatalf("GetManifest should not return error, got: %v", err)
	}

	if cl.Tag != "" {
		t.Errorf("cl.Tag should be empty for digest-only reference, got %q", cl.Tag)
	}
	if cl.Digest != "sha256:digestonly" {
		t.Errorf("cl.Digest = %q, want %q", cl.Digest, "sha256:digestonly")
	}
}
