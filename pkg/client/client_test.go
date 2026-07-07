package client

import (
	"cnabtool/pkg/data"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestParseReference_HappyPath проверяет корректный разбор ссылок
func TestParseReference_HappyPath(t *testing.T) {
	testlines := []struct {
		ref        string
		registry   string
		repository string
		tag        string
		digest     string
	}{
		{
			ref:        "registry.example.com/repository/image:v1.2.3@sha256:456",
			registry:   "registry.example.com",
			repository: "repository/image",
			tag:        "v1.2.3",
			digest:     "sha256:456",
		},
		{
			ref:        "registry.example.com/repository/image@sha256:456",
			registry:   "registry.example.com",
			repository: "repository/image",
			tag:        "",
			digest:     "sha256:456",
		},
		{
			ref:        "registry.example.com/repository/image:v1.2.3",
			registry:   "registry.example.com",
			repository: "repository/image",
			tag:        "v1.2.3",
			digest:     "",
		},
		{
			ref:        "osmp-docker-storage.repository.avp.ru/osmp/tmp/plugin-kes-linux/12.5.0.97/plugin-kes-linux:12.5.0.97",
			registry:   "osmp-docker-storage.repository.avp.ru",
			repository: "osmp/tmp/plugin-kes-linux/12.5.0.97/plugin-kes-linux",
			tag:        "12.5.0.97",
			digest:     "",
		},
	}

	for _, tc := range testlines {
		cl := &RegClient{
			Reference: tc.ref,
		}

		if err := cl.ParseReference(); err != nil {
			t.Errorf("ParseReference(%q) should not produce an error, got: %v", tc.ref, err)
			continue
		}

		if cl.Registry != tc.registry {
			t.Errorf("ParseReference(%q) Registry = %q, want %q", tc.ref, cl.Registry, tc.registry)
		}
		if cl.Repository != tc.repository {
			t.Errorf("ParseReference(%q) Repository = %q, want %q", tc.ref, cl.Repository, tc.repository)
		}
		if cl.Tag != tc.tag {
			t.Errorf("ParseReference(%q) Tag = %q, want %q", tc.ref, cl.Tag, tc.tag)
		}
		if cl.Digest != tc.digest {
			t.Errorf("ParseReference(%q) Digest = %q, want %q", tc.ref, cl.Digest, tc.digest)
		}
	}
}

// TestParseReference_ErrorCases проверяет невалидные ссылки
func TestParseReference_ErrorCases(t *testing.T) {
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
			ref:  "localhost/repository/image:v1",
		},
		{
			name: "no tag or digest",
			ref:  "registry.example.com/repository/image",
		},
		{
			name: "empty tag",
			ref:  "registry.example.com/repository/image:",
		},
		{
			name: "empty digest",
			ref:  "registry.example.com/repository/image@",
		},
		{
			name: "empty reference",
			ref:  "",
		},
		{
			name: "only registry",
			ref:  "registry.example.com",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cl := &RegClient{
				Reference: tc.ref,
			}
			err := cl.ParseReference()
			if err == nil {
				t.Errorf("ParseReference(%q) should return error, got nil", tc.ref)
			}
		})
	}
}

// TestParseReference_PortHandling проверяет работу с портами
// Примечание: текущая реализация ParseReference не поддерживает порты без доменной точки
// (например, localhost:5000 не проходит проверку на наличие точки в hostname)
func TestParseReference_PortHandling(t *testing.T) {
	testcases := []struct {
		name     string
		ref      string
		registry string
		wantErr  bool
	}{
		{
			name:     "with port and domain",
			ref:      "registry.example.com:5000/repo/image@sha256:abc",
			registry: "registry.example.com:5000",
			wantErr:  false,
		},
		{
			name:     "localhost without domain (known limitation)",
			ref:      "localhost:5000/repo/image:v1",
			registry: "",
			wantErr:  true, // текущая реализация требует точку в hostname
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cl := &RegClient{
				Reference: tc.ref,
			}
			err := cl.ParseReference()
			if tc.wantErr && err == nil {
				t.Errorf("ParseReference(%q) should return error, got nil", tc.ref)
				return
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ParseReference(%q) should not produce an error, got: %v", tc.ref, err)
				return
			}
			if !tc.wantErr && cl.Registry != tc.registry {
				t.Errorf("ParseReference(%q) Registry = %q, want %q", tc.ref, cl.Registry, tc.registry)
			}
		})
	}
}

// TestNewRegClient проверяет корректную инициализацию клиента
func TestNewRegClient(t *testing.T) {
	cfg := &Config{
		Scheme: "https",
		Credentials: struct {
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		}{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cl := NewRegClient(cfg, "registry.example.com/repo/image:v1")

	if cl.Reference != "registry.example.com/repo/image:v1" {
		t.Errorf("NewRegClient Reference = %q, want %q", cl.Reference, "registry.example.com/repo/image:v1")
	}
	if cl.Scheme != "https" {
		t.Errorf("NewRegClient Scheme = %q, want %q", cl.Scheme, "https")
	}
	if cl.Credentials.Username != "testuser" {
		t.Errorf("NewRegClient Credentials.Username = %q, want %q", cl.Credentials.Username, "testuser")
	}
	if cl.Credentials.Password != "testpass" {
		t.Errorf("NewRegClient Credentials.Password = %q, want %q", cl.Credentials.Password, "testpass")
	}
	if cl.Client != "cnabtool/0.1.1" {
		t.Errorf("NewRegClient Client = %q, want %q", cl.Client, "cnabtool/0.1.1")
	}
}

// TestFillResponse проверяет декодирование HTTP-ответа
func TestFillResponse(t *testing.T) {
	// Инициализируем Gc, чтобы logging.Debug не паниковал
	data.Gc = &data.Config{
		Verbosity: 4,
	}
	defer func() { data.Gc = nil }()

	testcases := []struct {
		name       string
		statusCode int
		content    string
		media      string
		digest     string
		date       string
		wantErr    bool
	}{
		{
			name:       "valid json response",
			statusCode: 200,
			content:    `{"schemaVersion": 2, "mediaType": "application/vnd.oci.image.index.v1+json"}`,
			media:      "application/vnd.oci.image.index.v1+json",
			digest:     "sha256:abc123",
			date:       "Mon, 01 Jan 2024 00:00:00 GMT",
			wantErr:    false,
		},
		{
			name:       "invalid json",
			statusCode: 200,
			content:    `{invalid json}`,
			wantErr:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tc.statusCode,
				Header: http.Header{
					"Content-Type":          {tc.media},
					"Last-Modified":         {tc.date},
					"Content-Length":        {"100"},
					"Docker-Content-Digest": {tc.digest},
				},
				Body: http.MaxBytesReader(nil, io.NopCloser(strings.NewReader(tc.content)), MaxBodySize+1),
			}

			regres := &RegResponse{}
			err := regres.FillResponse(resp)

			if tc.wantErr && err == nil {
				t.Errorf("FillResponse should return error for invalid JSON")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("FillResponse should not return error, got: %v", err)
				return
			}

			if !tc.wantErr {
				if regres.Media != tc.media {
					t.Errorf("FillResponse Media = %q, want %q", regres.Media, tc.media)
				}
				if regres.Digest != tc.digest {
					t.Errorf("FillResponse Digest = %q, want %q", regres.Digest, tc.digest)
				}
				if regres.Date != tc.date {
					t.Errorf("FillResponse Date = %q, want %q", regres.Date, tc.date)
				}
				if regres.Content != tc.content {
					t.Errorf("FillResponse Content = %q, want %q", regres.Content, tc.content)
				}
			}
		})
	}
}

// TestFillResponse_EmptyBody проверяет обработку пустого тела
func TestFillResponse_EmptyBody(t *testing.T) {
	// Инициализируем Gc, чтобы logging не паниковал
	data.Gc = &data.Config{
		Verbosity: 4,
	}
	defer func() { data.Gc = nil }()
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{},
		Body:       nil,
	}

	regres := &RegResponse{}
	err := regres.FillResponse(resp)

	if err == nil {
		t.Error("FillResponse should return error for nil body")
	}
}

// TestWebRequestEx проверяет заголовок авторизации
func TestWebRequestEx(t *testing.T) {
	data.Gc = &data.Config{
		Verbosity: 4,
	}
	defer func() { data.Gc = nil }()
	var receivedAuth string
	var receivedUserAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"repo":"test","path":"/test","children":[]}`))
	}))
	defer server.Close()

	cfg := &Config{
		Scheme: "http",
		Credentials: struct {
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		}{
			Username: "testuser",
			Password: "testpass",
		},
		Client: "cnabtool/0.1.1",
	}

	cl := NewRegClient(cfg, "test")
	cl.Registry = strings.TrimPrefix(server.URL, "http://")

	resp, err := cl.WebRequestEx("GET", server.URL+"/artifactory/api/storage/test/repo?list")
	if err != nil {
		t.Fatalf("WebRequestEx should not return error, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("WebRequestEx status = %d, want 200", resp.StatusCode)
	}

	// Проверяем, что Authorization отправлен
	if receivedAuth == "" {
		t.Error("WebRequestEx should send Authorization header")
	}

	// Проверяем базовую авторизацию
	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
	if receivedAuth != expectedAuth {
		t.Errorf("WebRequestEx Authorization = %q, want %q", receivedAuth, expectedAuth)
	}

	// Проверяем User-Agent
	if receivedUserAgent != "cnabtool/0.1.1" {
		t.Errorf("WebRequestEx User-Agent = %q, want %q", receivedUserAgent, "cnabtool/0.1.1")
	}
}

// TestGetTagList проверяет получение списка тегов
func TestGetTagList(t *testing.T) {
	data.Gc = &data.Config{
		Verbosity: 4,
	}
	defer func() { data.Gc = nil }()
	expectedTags := `{"name":"test/repo","tags":["v1.0","v2.0","latest"]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(expectedTags))
	}))
	defer server.Close()

	cfg := &Config{
		Scheme: "http",
		Credentials: struct {
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		}{
			Username: "testuser",
			Password: "testpass",
		},
		Client: "cnabtool/0.1.1",
	}

	cl := NewRegClient(cfg, "test")
	cl.Registry = strings.TrimPrefix(server.URL, "http://")
	cl.Repository = "test/repo"

	resp, err := cl.GetTagList()
	if err != nil {
		t.Fatalf("GetTagList should not return error, got: %v", err)
	}

	if resp.Status != 200 {
		t.Errorf("GetTagList status = %d, want 200", resp.Status)
	}

	if resp.Content != expectedTags {
		t.Errorf("GetTagList Content = %q, want %q", resp.Content, expectedTags)
	}
}

// TestGetTagList_Error проверяет обработку ошибки
func TestGetTagList_Error(t *testing.T) {
	data.Gc = &data.Config{
		Verbosity: 4,
	}
	defer func() { data.Gc = nil }()
	// Сервер, который возвращает 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	cfg := &Config{
		Scheme: "http",
		Credentials: struct {
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		}{
			Username: "testuser",
			Password: "testpass",
		},
		Client: "cnabtool/0.1.1",
	}

	cl := NewRegClient(cfg, "test")
	cl.Registry = strings.TrimPrefix(server.URL, "http://")
	cl.Repository = "test/repo"

	_, err := cl.GetTagList()
	if err == nil {
		t.Error("GetTagList should return error for 404 response")
	}
}

// TestMediaTypeConstants проверяет константы media types
func TestMediaTypeConstants(t *testing.T) {
	if MediaTypeOciIndex != "application/vnd.oci.image.index.v1+json" {
		t.Errorf("MediaTypeOciIndex = %q, want %q", MediaTypeOciIndex, "application/vnd.oci.image.index.v1+json")
	}
	if MediaTypeOciManifest != "application/vnd.oci.image.manifest.v1+json" {
		t.Errorf("MediaTypeOciManifest = %q, want %q", MediaTypeOciManifest, "application/vnd.oci.image.manifest.v1+json")
	}
	if MediaTypeV2Manifest != "application/vnd.docker.distribution.manifest.v2+json" {
		t.Errorf("MediaTypeV2Manifest = %q, want %q", MediaTypeV2Manifest, "application/vnd.docker.distribution.manifest.v2+json")
	}
	if MediaTypeJson != "application/json" {
		t.Errorf("MediaTypeJson = %q, want %q", MediaTypeJson, "application/json")
	}
}

// TestManifestDiscovery проверяет порядок media types в discovery
func TestManifestDiscovery(t *testing.T) {
	if len(manifest_discovery) == 0 {
		t.Error("manifest_discovery should not be empty")
	}
	if manifest_discovery[0] != MediaTypeOciIndex {
		t.Errorf("manifest_discovery[0] = %q, want %q", manifest_discovery[0], MediaTypeOciIndex)
	}
}
