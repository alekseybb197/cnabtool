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

// saveGlobalState сохраняет текущее состояние data для восстановления
func saveGlobalState() *globalStateSnapshot {
	return &globalStateSnapshot{
		Gc:           data.Gc,
		Sensitives:   data.Sensitives,
		ItemsQueue:   data.ItemsQueue,
		ProjectList:  data.ProjectList,
		ItemByDigest: data.ItemByDigest,
		ItemByTag:    data.ItemByTag,
		Scheme:       data.Scheme,
		Registry:     data.Registry,
		Repository:   data.Repository,
	}
}

// globalStateSnapshot хранит снимок глобального состояния
type globalStateSnapshot struct {
	Gc           *data.Config
	Sensitives   []string
	ItemsQueue   []string
	ProjectList  []*data.RegIndex
	ItemByDigest map[string]*data.RegIndex
	ItemByTag    map[string]*data.RegIndex
	Scheme       string
	Registry     string
	Repository   string
}

func restoreGlobalState(snap *globalStateSnapshot) {
	data.Gc = snap.Gc
	data.Sensitives = snap.Sensitives
	data.ItemsQueue = snap.ItemsQueue
	data.ProjectList = snap.ProjectList
	data.ItemByDigest = snap.ItemByDigest
	data.ItemByTag = snap.ItemByTag
	data.Scheme = snap.Scheme
	data.Registry = snap.Registry
	data.Repository = snap.Repository
}

// resetGlobalState очищает глобальное состояние data перед тестом
func resetGlobalState(t *testing.T) {
	t.Helper()
	data.Gc = &data.Config{
		Verbosity: 4,
		Timeout:   10000,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}
	data.ItemsQueue = nil
	data.ProjectList = nil
	data.ItemByDigest = make(map[string]*data.RegIndex)
	data.ItemByTag = make(map[string]*data.RegIndex)
	data.Sensitives = []string{"testpass"}
	data.Scheme = "http"
	data.Registry = ""
	data.Repository = ""
}

// TestAddIndex_NewIndex проверяет создание нового RegIndex
func TestAddIndex_NewIndex(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:abc123",
		Date:      "Mon, 01 Jan 2024 00:00:00 GMT",
		Content:   `{"schemaVersion":2,"manifests":[]}`,
	}

	ri, err := AddIndex(regres, "v1")
	if err != nil {
		t.Fatalf("AddIndex should not return error, got: %v", err)
	}
	if ri == nil {
		t.Fatal("AddIndex returned nil RegIndex")
	}

	if ri.Reference != "registry.example.com/repo/image:v1" {
		t.Errorf("ri.Reference = %q, want %q", ri.Reference, "registry.example.com/repo/image:v1")
	}
	if ri.Tag != "v1" {
		t.Errorf("ri.Tag = %q, want %q", ri.Tag, "v1")
	}
	if ri.Media != client.MediaTypeOciIndex {
		t.Errorf("ri.Media = %q, want %q", ri.Media, client.MediaTypeOciIndex)
	}
	if ri.Annotation != data.ItemTypeCnab {
		t.Errorf("ri.Annotation = %q, want %q", ri.Annotation, data.ItemTypeCnab)
	}
	if ri.Digest != "sha256:abc123" {
		t.Errorf("ri.Digest = %q, want %q", ri.Digest, "sha256:abc123")
	}
	if ri.Date != "Mon, 01 Jan 2024 00:00:00 GMT" {
		t.Errorf("ri.Date = %q, want %q", ri.Date, "Mon, 01 Jan 2024 00:00:00 GMT")
	}

	if data.ItemByDigest["sha256:abc123"] != ri {
		t.Error("ItemByDigest['sha256:abc123'] should contain the RegIndex")
	}
	if data.ItemByTag["v1"] != ri {
		t.Error("ItemByTag['v1'] should contain the RegIndex")
	}
	if len(data.ProjectList) != 1 {
		t.Errorf("ProjectList length = %d, want 1", len(data.ProjectList))
	}
}

// TestAddIndex_DuplicateIndex проверяет обработку существующего индекса
func TestAddIndex_DuplicateIndex(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	regres1 := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:abc123",
		Content:   `{"schemaVersion":2}`,
	}

	ri1, err := AddIndex(regres1, "v1")
	if err != nil {
		t.Fatalf("First AddIndex should not return error, got: %v", err)
	}

	regres2 := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:abc123",
		Content:   `{"schemaVersion":2,"different":"content"}`,
	}

	ri2, err := AddIndex(regres2, "v2")
	if err != nil {
		t.Fatalf("Second AddIndex should not return error, got: %v", err)
	}

	if ri1 != ri2 {
		t.Error("AddIndex should return the same RegIndex for duplicate digest")
	}

	if len(data.ProjectList) != 1 {
		t.Errorf("ProjectList length = %d, want 1 (duplicate should not add)", len(data.ProjectList))
	}

	if data.ItemByTag["v2"] != ri1 {
		t.Error("ItemByTag['v2'] should point to the existing RegIndex")
	}
}

// TestAddIndex_MediaTypeAnnotationMapping проверяет маппинг media type → annotation
func TestAddIndex_MediaTypeAnnotationMapping(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	testcases := []struct {
		media    string
		expected string
		itemType string
	}{
		{
			media:    client.MediaTypeOciIndex,
			expected: data.ItemTypeCnab,
			itemType: "cnab index",
		},
		{
			media:    client.MediaTypeOciManifest,
			expected: data.ItemTypeConfig,
			itemType: "cnab config",
		},
		{
			media:    client.MediaTypeV1Pretty,
			expected: data.ItemTypeImage,
			itemType: "docker image",
		},
		{
			media:    "application/octet-stream",
			expected: data.ItemTypeStuff,
			itemType: "stuff",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.itemType, func(t *testing.T) {
			digest := "sha256:" + tc.itemType
			regres := &client.RegResponse{
				Reference: "registry.example.com/repo/image:v1",
				Media:     tc.media,
				Digest:    digest,
				Content:   `{"test":true}`,
			}

			ri, err := AddIndex(regres, "v1")
			if err != nil {
				t.Fatalf("AddIndex should not return error, got: %v", err)
			}

			if ri.Annotation != tc.expected {
				t.Errorf("Media %q → Annotation = %q, want %q", tc.media, ri.Annotation, tc.expected)
			}
		})
	}
}

// TestAddCnab_ValidManifest проверяет разбор валидного OCI index
func TestAddCnab_ValidManifest(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	ociIndex := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"digest": "sha256:config123",
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"annotations": {
					"io.cnab.manifest.type": "config"
				}
			},
			{
				"digest": "sha256:invocation456",
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"annotations": {
					"io.cnab.manifest.type": "invocation"
				}
			},
			{
				"digest": "sha256:component789",
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"annotations": {
					"io.cnab.manifest.type": "component",
					"io.cnab.component.name": "myapp"
				}
			}
		]
	}`

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/cnab:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:parent123",
		Content:   ociIndex,
	}

	err := AddCnab(regres, "v1")
	if err != nil {
		t.Fatalf("AddCnab should not return error, got: %v", err)
	}

	ri := data.ItemByDigest["sha256:parent123"]
	if ri == nil {
		t.Fatal("RegIndex not found in ItemByDigest")
	}

	if len(ri.DownLinks) != 3 {
		t.Errorf("DownLinks length = %d, want 3", len(ri.DownLinks))
	}

	if len(data.ItemsQueue) != 3 {
		t.Errorf("ItemsQueue length = %d, want 3", len(data.ItemsQueue))
	}
}

// TestAddCnab_EmptyManifests проверяет обработку пустого manifests
func TestAddCnab_EmptyManifests(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/cnab:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:empty123",
		Content:   `{"schemaVersion":2,"manifests":[]}`,
	}

	err := AddCnab(regres, "v1")
	if err != nil {
		t.Fatalf("AddCnab should not return error for empty manifests, got: %v", err)
	}

	ri := data.ItemByDigest["sha256:empty123"]
	if ri == nil {
		t.Fatal("RegIndex not found")
	}

	if len(ri.DownLinks) != 0 {
		t.Errorf("DownLinks length = %d, want 0", len(ri.DownLinks))
	}
}

// TestAddCnab_MissingManifestsKey проверяет ошибку при отсутствии manifests
func TestAddCnab_MissingManifestsKey(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/cnab:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:nomani",
		Content:   `{"schemaVersion":2}`,
	}

	err := AddCnab(regres, "v1")
	if err == nil {
		t.Error("AddCnab should return error when manifests key is missing")
	}
}

// TestAddCnab_NotArrayManifests проверяет ошибку, если manifests не массив
func TestAddCnab_NotArrayManifests(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/cnab:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:notarray",
		Content:   `{"schemaVersion":2,"manifests":{"digest":"sha256:x"}}`,
	}

	err := AddCnab(regres, "v1")
	if err == nil {
		t.Error("AddCnab should return error when manifests is not an array")
	}
}

// TestAddCnab_ComponentAnnotation проверяет special handling component annotation
func TestAddCnab_ComponentAnnotation(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/cnab:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:comp123",
		Content: `{
			"schemaVersion": 2,
			"manifests": [
				{
					"digest": "sha256:comp789",
					"mediaType": "application/vnd.oci.image.manifest.v1+json",
					"annotations": {
						"io.cnab.manifest.type": "component",
						"io.cnab.component.name": "mycomponent"
					}
				}
			]
		}`,
	}

	err := AddCnab(regres, "v1")
	if err != nil {
		t.Fatalf("AddCnab should not return error, got: %v", err)
	}

	ri := data.ItemByDigest["sha256:comp123"]
	if ri == nil {
		t.Fatal("RegIndex not found")
	}

	if len(ri.DownLinks) != 1 {
		t.Fatalf("DownLinks length = %d, want 1", len(ri.DownLinks))
	}

	if ri.DownLinks[0].Annotation != "mycomponent" {
		t.Errorf("DownLinks[0].Annotation = %q, want %q", ri.DownLinks[0].Annotation, "mycomponent")
	}
}

// TestInspectCnab_FullFlow проверяет полный цикл InspectCnab через httptest
func TestInspectCnab_FullFlow(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	tagsList := `{
		"name": "repo/cnab",
		"tags": ["v1.0", "v2.0"]
	}`

	ociIndexV1 := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"digest": "sha256:config_v1",
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"annotations": {
					"io.cnab.manifest.type": "config"
				}
			}
		]
	}`

	ociIndexV2 := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"digest": "sha256:config_v2",
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"annotations": {
					"io.cnab.manifest.type": "config"
				}
			}
		]
	}`

	configManifestV1 := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.manifest.v1+json",
		"config": {
			"digest": "sha256:config_v1"
		}
	}`

	configManifestV2 := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.manifest.v1+json",
		"config": {
			"digest": "sha256:config_v2"
		}
	}`

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		path := r.URL.Path

		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(path, "/tags/list/"):
			w.WriteHeader(200)
			w.Write([]byte(tagsList))
		case strings.HasSuffix(path, "/manifests/v1.0"):
			w.Header().Set("Docker-Content-Digest", "sha256:oci_v1")
			w.WriteHeader(200)
			w.Write([]byte(ociIndexV1))
		case strings.HasSuffix(path, "/manifests/v2.0"):
			w.Header().Set("Docker-Content-Digest", "sha256:oci_v2")
			w.WriteHeader(200)
			w.Write([]byte(ociIndexV2))
		case strings.HasSuffix(path, "/manifests/sha256:config_v1"):
			w.Header().Set("Docker-Content-Digest", "sha256:config_v1")
			w.WriteHeader(200)
			w.Write([]byte(configManifestV1))
		case strings.HasSuffix(path, "/manifests/sha256:config_v2"):
			w.Header().Set("Docker-Content-Digest", "sha256:config_v2")
			w.WriteHeader(200)
			w.Write([]byte(configManifestV2))
		default:
			w.WriteHeader(404)
			w.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer server.Close()

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cl := client.NewRegClient((*client.Config)(cfg), "test")
	cl.Registry = strings.TrimPrefix(server.URL, "http://")
	cl.Repository = "repo/cnab"
	cl.Tag = "v1.0"
	cl.Digest = ""

	cnf := (*Config)(cfg)
	cnf.InspectCnab(cl)

	if len(data.ItemByTag) < 2 {
		t.Errorf("ItemByTag length = %d, want at least 2", len(data.ItemByTag))
	}

	if len(data.ProjectList) < 2 {
		t.Errorf("ProjectList length = %d, want at least 2", len(data.ProjectList))
	}

	if requestCount < 3 {
		t.Errorf("Request count = %d, want at least 3 (tags list + 2 manifests)", requestCount)
	}
}

// TestInspectCnab_EmptyTagsList проверяет обработку пустого списка тегов
func TestInspectCnab_EmptyTagsList(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	tagsList := `{
		"name": "repo/cnab",
		"tags": []
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(tagsList))
	}))
	defer server.Close()

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cl := client.NewRegClient((*client.Config)(cfg), "test")
	cl.Registry = strings.TrimPrefix(server.URL, "http://")
	cl.Repository = "repo/cnab"
	cl.Tag = "v1"

	cnf := (*Config)(cfg)
	cnf.InspectCnab(cl)

	// Пустой список тегов → ничего не добавляется в ItemByTag
	if len(data.ItemByTag) != 0 {
		t.Errorf("ItemByTag length = %d, want 0", len(data.ItemByTag))
	}
}

// TestShowCnabReport_ValidData проверяет генерацию JSON-отчёта
func TestShowCnabReport_ValidData(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	ri1 := &data.RegIndex{
		Reference:  "registry.example.com/repo/cnab:v1",
		Tag:        "v1",
		Digest:     "sha256:parent123",
		Media:      client.MediaTypeOciIndex,
		Annotation: data.ItemTypeCnab,
		Date:       "Mon, 01 Jan 2024 00:00:00 GMT",
		Content:    `{"schemaVersion":2}`,
		DownLinks: []data.CnabItem{
			{Digest: "sha256:config123", Annotation: "config"},
			{Digest: "sha256:invocation456", Annotation: "invocation"},
		},
		UpLinks: []data.CnabItem{
			{Digest: "sha256:grandparent", Annotation: "cnab index"},
		},
	}

	ri2 := &data.RegIndex{
		Reference:  "registry.example.com/repo/cnab:v2",
		Tag:        "v2",
		Digest:     "sha256:parent456",
		Media:      client.MediaTypeOciIndex,
		Annotation: data.ItemTypeCnab,
		Date:       "Tue, 01 Jan 2024 00:00:00 GMT",
		Content:    `{"schemaVersion":2}`,
		DownLinks: []data.CnabItem{
			{Digest: "sha256:config456", Annotation: "config"},
		},
	}

	data.ProjectList = []*data.RegIndex{ri1, ri2}
	data.ItemByDigest["sha256:parent123"] = ri1
	data.ItemByDigest["sha256:parent456"] = ri2
	data.ItemByTag["v1"] = ri1
	data.ItemByTag["v2"] = ri2

	cfg := &data.Config{
		Verbosity: 3,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}
	data.Gc = cfg

	cl := &client.RegClient{
		Reference:  "registry.example.com/repo/cnab:v1",
		Registry:   "registry.example.com",
		Repository: "repo/cnab",
		Tag:        "v1",
	}

	cnf := (*Config)(cfg)
	cnf.ShowCnabReport(cl)
}

// TestShowCnabReport_RawMode проверяет raw-режим вывода
func TestShowCnabReport_RawMode(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	ri := &data.RegIndex{
		Reference:  "registry.example.com/repo/cnab:v1",
		Tag:        "v1",
		Digest:     "sha256:parent123",
		Media:      client.MediaTypeOciIndex,
		Annotation: data.ItemTypeCnab,
		Content:    `{"test":true}`,
	}

	data.ItemByTag["v1"] = ri
	data.ProjectList = []*data.RegIndex{ri}

	cfg := &data.Config{
		Verbosity: 3,
		Raw:       true,
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}
	data.Gc = cfg

	cl := &client.RegClient{
		Reference:  "registry.example.com/repo/cnab:v1",
		Registry:   "registry.example.com",
		Repository: "repo/cnab",
	}

	cnf := (*Config)(cfg)
	cnf.ShowCnabReport(cl)
}

// TestShowCnabReport_LowVerbosity проверяет отсутствие вывода при низком verbosity
func TestShowCnabReport_LowVerbosity(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	ri := &data.RegIndex{
		Reference:  "registry.example.com/repo/cnab:v1",
		Tag:        "v1",
		Digest:     "sha256:parent123",
		Media:      client.MediaTypeOciIndex,
		Annotation: data.ItemTypeCnab,
	}

	data.ItemByTag["v1"] = ri
	data.ProjectList = []*data.RegIndex{ri}

	cfg := &data.Config{
		Verbosity: 1, // LogErrorLevel - меньше LogNormalLevel
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}
	data.Gc = cfg

	cl := &client.RegClient{
		Reference:  "registry.example.com/repo/cnab:v1",
		Registry:   "registry.example.com",
		Repository: "repo/cnab",
	}

	cnf := (*Config)(cfg)
	cnf.ShowCnabReport(cl)
}

// TestAddIndex_InvalidContent проверяет обработку невалидного JSON в Content
func TestAddIndex_InvalidContent(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:invalid",
		Content:   "{invalid json content!!!",
	}

	_, err := AddIndex(regres, "v1")
	if err == nil {
		t.Error("AddIndex should return error for invalid JSON content")
	}
}

// TestAddIndex_PrettyContent проверяет, что Content сохраняется в pretty-формате
func TestAddIndex_PrettyContent(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	originalContent := `{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[]}`

	regres := &client.RegResponse{
		Reference: "registry.example.com/repo/image:v1",
		Media:     client.MediaTypeOciIndex,
		Digest:    "sha256:pretty123",
		Content:   originalContent,
	}

	ri, err := AddIndex(regres, "v1")
	if err != nil {
		t.Fatalf("AddIndex should not return error, got: %v", err)
	}

	if ri.Content == "" {
		t.Error("RegIndex.Content should not be empty")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(ri.Content), &parsed); err != nil {
		t.Errorf("RegIndex.Content should be valid JSON, got error: %v", err)
	}
}

// TestInspectCnab_TagNotFound проверяет обработку ошибки при получении манифеста тега
func TestInspectCnab_TagNotFound(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	tagsList := `{
		"name": "repo/cnab",
		"tags": ["v1.0", "v2.0"]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(tagsList))
	}))
	defer server.Close()

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cl := client.NewRegClient((*client.Config)(cfg), "test")
	cl.Registry = strings.TrimPrefix(server.URL, "http://")
	cl.Repository = "repo/cnab"
	cl.Tag = "v1.0"

	cnf := (*Config)(cfg)
	cnf.InspectCnab(cl)

	if len(data.ItemByTag) < 1 {
		t.Errorf("ItemByTag length = %d, want at least 1", len(data.ItemByTag))
	}
}

// TestInspectCnab_MultipleCNABIndexes проверяет обработку нескольких CNAB индексов
func TestInspectCnab_MultipleCNABIndexes(t *testing.T) {
	snap := saveGlobalState()
	defer restoreGlobalState(snap)
	resetGlobalState(t)

	tagsList := `{
		"name": "repo/cnab",
		"tags": ["v1.0", "v2.0", "v3.0"]
	}`

	ociIndex := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"digest": "sha256:config_%s",
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"annotations": {
					"io.cnab.manifest.type": "config"
				}
			}
		]
	}`

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		path := r.URL.Path

		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(path, "/tags/list/"):
			w.WriteHeader(200)
			w.Write([]byte(tagsList))
		case strings.HasSuffix(path, "/manifests/v1.0"):
			w.Header().Set("Docker-Content-Digest", "sha256:oci_v1")
			w.WriteHeader(200)
			w.Write([]byte(strings.Replace(ociIndex, "%s", "v1", 1)))
		case strings.HasSuffix(path, "/manifests/v2.0"):
			w.Header().Set("Docker-Content-Digest", "sha256:oci_v2")
			w.WriteHeader(200)
			w.Write([]byte(strings.Replace(ociIndex, "%s", "v2", 1)))
		case strings.HasSuffix(path, "/manifests/v3.0"):
			w.Header().Set("Docker-Content-Digest", "sha256:oci_v3")
			w.WriteHeader(200)
			w.Write([]byte(strings.Replace(ociIndex, "%s", "v3", 1)))
		default:
			w.WriteHeader(404)
			w.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer server.Close()

	cfg := &data.Config{
		Scheme: "http",
		Credentials: data.Credentials{
			Username: "testuser",
			Password: "testpass",
		},
		Client:  "cnabtool/0.1.1",
		Timeout: 10000,
	}

	cl := client.NewRegClient((*client.Config)(cfg), "test")
	cl.Registry = strings.TrimPrefix(server.URL, "http://")
	cl.Repository = "repo/cnab"
	cl.Tag = "v1.0"

	cnf := (*Config)(cfg)
	cnf.InspectCnab(cl)

	if len(data.ItemByTag) < 3 {
		t.Errorf("ItemByTag length = %d, want at least 3", len(data.ItemByTag))
	}

	if len(data.ProjectList) < 3 {
		t.Errorf("ProjectList length = %d, want at least 3", len(data.ProjectList))
	}

	if requestCount < 4 {
		t.Errorf("Request count = %d, want at least 4", requestCount)
	}
}
