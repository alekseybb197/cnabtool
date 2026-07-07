package data

import (
	"testing"
)

// TestItemTypeConstants проверяет константы типов
func TestItemTypeConstants(t *testing.T) {
	if ItemTypeCnab != "cnab index" {
		t.Errorf("ItemTypeCnab = %q, want %q", ItemTypeCnab, "cnab index")
	}
	if ItemTypeImage != "docker image" {
		t.Errorf("ItemTypeImage = %q, want %q", ItemTypeImage, "docker image")
	}
	if ItemTypeConfig != "cnab config" {
		t.Errorf("ItemTypeConfig = %q, want %q", ItemTypeConfig, "cnab config")
	}
	if ItemTypeStuff != "stuff" {
		t.Errorf("ItemTypeStuff = %q, want %q", ItemTypeStuff, "stuff")
	}
}

// TestConfig_Struct проверяет структуру Config
func TestConfig_Struct(t *testing.T) {
	cfg := &Config{
		Verbosity: 3,
		Timeout:   15000,
		Unsecure:  true,
		Client:    "cnabtool/1.0",
		Scheme:    "http",
		Raw:       true,
		DryRun:    true,
		Purge:     true,
		RepoKey:   "my-repo",
		Error:     2,
		Credentials: Credentials{
			Username: "testuser",
			Password: "testpass",
		},
	}

	if cfg.Verbosity != 3 {
		t.Errorf("Config.Verbosity = %d, want 3", cfg.Verbosity)
	}
	if cfg.Timeout != 15000 {
		t.Errorf("Config.Timeout = %d, want 15000", cfg.Timeout)
	}
	if cfg.Unsecure != true {
		t.Errorf("Config.Unsecure = %v, want true", cfg.Unsecure)
	}
	if cfg.Client != "cnabtool/1.0" {
		t.Errorf("Config.Client = %q, want %q", cfg.Client, "cnabtool/1.0")
	}
	if cfg.Scheme != "http" {
		t.Errorf("Config.Scheme = %q, want %q", cfg.Scheme, "http")
	}
	if cfg.Raw != true {
		t.Errorf("Config.Raw = %v, want true", cfg.Raw)
	}
	if cfg.DryRun != true {
		t.Errorf("Config.DryRun = %v, want true", cfg.DryRun)
	}
	if cfg.Purge != true {
		t.Errorf("Config.Purge = %v, want true", cfg.Purge)
	}
	if cfg.RepoKey != "my-repo" {
		t.Errorf("Config.RepoKey = %q, want %q", cfg.RepoKey, "my-repo")
	}
	if cfg.Error != 2 {
		t.Errorf("Config.Error = %d, want 2", cfg.Error)
	}
	if cfg.Credentials.Username != "testuser" {
		t.Errorf("Config.Credentials.Username = %q, want %q", cfg.Credentials.Username, "testuser")
	}
	if cfg.Credentials.Password != "testpass" {
		t.Errorf("Config.Credentials.Password = %q, want %q", cfg.Credentials.Password, "testpass")
	}
}

// TestConfig_DefaultValues проверяет значения по умолчанию
func TestConfig_DefaultValues(t *testing.T) {
	cfg := &Config{}

	if cfg.Verbosity != 0 {
		t.Errorf("Config.Verbosity default = %d, want 0", cfg.Verbosity)
	}
	if cfg.Timeout != 0 {
		t.Errorf("Config.Timeout default = %d, want 0", cfg.Timeout)
	}
	if cfg.Unsecure != false {
		t.Errorf("Config.Unsecure default = %v, want false", cfg.Unsecure)
	}
	if cfg.Client != "" {
		t.Errorf("Config.Client default = %q, want empty", cfg.Client)
	}
	if cfg.Scheme != "" {
		t.Errorf("Config.Scheme default = %q, want empty", cfg.Scheme)
	}
	if cfg.Raw != false {
		t.Errorf("Config.Raw default = %v, want false", cfg.Raw)
	}
	if cfg.DryRun != false {
		t.Errorf("Config.DryRun default = %v, want false", cfg.DryRun)
	}
	if cfg.Purge != false {
		t.Errorf("Config.Purge default = %v, want false", cfg.Purge)
	}
	if cfg.RepoKey != "" {
		t.Errorf("Config.RepoKey default = %q, want empty", cfg.RepoKey)
	}
	if cfg.Error != 0 {
		t.Errorf("Config.Error default = %d, want 0", cfg.Error)
	}
}

// TestCredentials_Struct проверяет структуру Credentials
func TestCredentials_Struct(t *testing.T) {
	creds := Credentials{
		Username: "admin",
		Password: "secret",
	}

	if creds.Username != "admin" {
		t.Errorf("Credentials.Username = %q, want %q", creds.Username, "admin")
	}
	if creds.Password != "secret" {
		t.Errorf("Credentials.Password = %q, want %q", creds.Password, "secret")
	}
}

// TestCredentials_Default проверяет значения по умолчанию
func TestCredentials_Default(t *testing.T) {
	creds := Credentials{}

	if creds.Username != "" {
		t.Errorf("Credentials.Username default = %q, want empty", creds.Username)
	}
	if creds.Password != "" {
		t.Errorf("Credentials.Password default = %q, want empty", creds.Password)
	}
}

// TestCnabItem_Struct проверяет структуру CnabItem
func TestCnabItem_Struct(t *testing.T) {
	item := CnabItem{
		Digest:     "sha256:abc123",
		Annotation: "component",
	}

	if item.Digest != "sha256:abc123" {
		t.Errorf("CnabItem.Digest = %q, want %q", item.Digest, "sha256:abc123")
	}
	if item.Annotation != "component" {
		t.Errorf("CnabItem.Annotation = %q, want %q", item.Annotation, "component")
	}
}

// TestCnabItem_Default проверяет значения по умолчанию
func TestCnabItem_Default(t *testing.T) {
	item := CnabItem{}

	if item.Digest != "" {
		t.Errorf("CnabItem.Digest default = %q, want empty", item.Digest)
	}
	if item.Annotation != "" {
		t.Errorf("CnabItem.Annotation default = %q, want empty", item.Annotation)
	}
}

// TestRegIndex_Struct проверяет структуру RegIndex
func TestRegIndex_Struct(t *testing.T) {
	ri := &RegIndex{
		Reference:  "registry.example.com/repo/image:v1",
		Tag:        "v1",
		Media:      "application/vnd.oci.image.index.v1+json",
		Annotation: "cnab index",
		Date:       "Mon, 01 Jan 2024 00:00:00 GMT",
		Digest:     "sha256:def456",
		Lost:       0,
		Content:    `{"schemaVersion": 2}`,
		DownLinks: []CnabItem{
			{Digest: "sha256:111", Annotation: "config"},
			{Digest: "sha256:222", Annotation: "invocation"},
		},
		UpLinks: []CnabItem{
			{Digest: "sha256:parent", Annotation: "parent"},
		},
	}

	if ri.Reference != "registry.example.com/repo/image:v1" {
		t.Errorf("RegIndex.Reference = %q, want %q", ri.Reference, "registry.example.com/repo/image:v1")
	}
	if ri.Tag != "v1" {
		t.Errorf("RegIndex.Tag = %q, want %q", ri.Tag, "v1")
	}
	if ri.Media != "application/vnd.oci.image.index.v1+json" {
		t.Errorf("RegIndex.Media = %q, want %q", ri.Media, "application/vnd.oci.image.index.v1+json")
	}
	if ri.Annotation != "cnab index" {
		t.Errorf("RegIndex.Annotation = %q, want %q", ri.Annotation, "cnab index")
	}
	if ri.Date != "Mon, 01 Jan 2024 00:00:00 GMT" {
		t.Errorf("RegIndex.Date = %q, want %q", ri.Date, "Mon, 01 Jan 2024 00:00:00 GMT")
	}
	if ri.Digest != "sha256:def456" {
		t.Errorf("RegIndex.Digest = %q, want %q", ri.Digest, "sha256:def456")
	}
	if ri.Lost != 0 {
		t.Errorf("RegIndex.Lost = %d, want 0", ri.Lost)
	}
	if ri.Content != `{"schemaVersion": 2}` {
		t.Errorf("RegIndex.Content = %q, want %q", ri.Content, `{"schemaVersion": 2}`)
	}
	if len(ri.DownLinks) != 2 {
		t.Errorf("RegIndex.DownLinks length = %d, want 2", len(ri.DownLinks))
	}
	if len(ri.UpLinks) != 1 {
		t.Errorf("RegIndex.UpLinks length = %d, want 1", len(ri.UpLinks))
	}
	if ri.DownLinks[0].Digest != "sha256:111" {
		t.Errorf("RegIndex.DownLinks[0].Digest = %q, want %q", ri.DownLinks[0].Digest, "sha256:111")
	}
	if ri.UpLinks[0].Digest != "sha256:parent" {
		t.Errorf("RegIndex.UpLinks[0].Digest = %q, want %q", ri.UpLinks[0].Digest, "sha256:parent")
	}
}

// TestRegIndex_Default проверяет значения по умолчанию
func TestRegIndex_Default(t *testing.T) {
	ri := &RegIndex{}

	if ri.Reference != "" {
		t.Errorf("RegIndex.Reference default = %q, want empty", ri.Reference)
	}
	if ri.Tag != "" {
		t.Errorf("RegIndex.Tag default = %q, want empty", ri.Tag)
	}
	if ri.Media != "" {
		t.Errorf("RegIndex.Media default = %q, want empty", ri.Media)
	}
	if ri.Annotation != "" {
		t.Errorf("RegIndex.Annotation default = %q, want empty", ri.Annotation)
	}
	if ri.Digest != "" {
		t.Errorf("RegIndex.Digest default = %q, want empty", ri.Digest)
	}
	if ri.Lost != 0 {
		t.Errorf("RegIndex.Lost default = %d, want 0", ri.Lost)
	}
	if ri.Content != "" {
		t.Errorf("RegIndex.Content default = %q, want empty", ri.Content)
	}
	if ri.DownLinks != nil && len(ri.DownLinks) != 0 {
		t.Errorf("RegIndex.DownLinks default should be empty slice, got %d items", len(ri.DownLinks))
	}
	if ri.UpLinks != nil && len(ri.UpLinks) != 0 {
		t.Errorf("RegIndex.UpLinks default should be empty slice, got %d items", len(ri.UpLinks))
	}
}

// TestGlobalState_Default проверяет начальное состояние глобальных переменных
func TestGlobalState_Default(t *testing.T) {
	// Сохраняем текущее состояние для восстановления
	origGc := Gc
	origSensitives := Sensitives
	origScheme := Scheme
	origRegistry := Registry
	origRepository := Repository
	origItemsQueue := ItemsQueue
	origProjectList := ProjectList
	origItemByDigest := ItemByDigest
	origItemByTag := ItemByTag

	defer func() {
		// Восстанавливаем состояние
		Gc = origGc
		Sensitives = origSensitives
		Scheme = origScheme
		Registry = origRegistry
		Repository = origRepository
		ItemsQueue = origItemsQueue
		ProjectList = origProjectList
		ItemByDigest = origItemByDigest
		ItemByTag = origItemByTag
	}()

	// Очищаем глобальное состояние
	Gc = nil
	Sensitives = nil
	Scheme = ""
	Registry = ""
	Repository = ""
	ItemsQueue = nil
	ProjectList = nil
	ItemByDigest = make(map[string]*RegIndex)
	ItemByTag = make(map[string]*RegIndex)

	// Проверяем начальное состояние
	if Gc != nil {
		t.Errorf("Gc default = %v, want nil", Gc)
	}
	if Sensitives != nil {
		t.Errorf("Sensitives default = %v, want nil", Sensitives)
	}
	if Scheme != "" {
		t.Errorf("Scheme default = %q, want empty", Scheme)
	}
	if Registry != "" {
		t.Errorf("Registry default = %q, want empty", Registry)
	}
	if Repository != "" {
		t.Errorf("Repository default = %q, want empty", Repository)
	}
	if ItemsQueue != nil {
		t.Errorf("ItemsQueue default = %v, want nil", ItemsQueue)
	}
	if ProjectList != nil {
		t.Errorf("ProjectList default = %v, want nil", ProjectList)
	}
	if len(ItemByDigest) != 0 {
		t.Errorf("ItemByDigest default length = %d, want 0", len(ItemByDigest))
	}
	if len(ItemByTag) != 0 {
		t.Errorf("ItemByTag default length = %d, want 0", len(ItemByTag))
	}
}

// TestGlobalState_Modification проверяет модификацию глобальных переменных
func TestGlobalState_Modification(t *testing.T) {
	// Сохраняем текущее состояние для восстановления
	origGc := Gc
	origSensitives := Sensitives
	origScheme := Scheme
	origRegistry := Registry
	origRepository := Repository
	origItemsQueue := ItemsQueue
	origProjectList := ProjectList
	origItemByDigest := ItemByDigest
	origItemByTag := ItemByTag

	defer func() {
		// Восстанавливаем состояние
		Gc = origGc
		Sensitives = origSensitives
		Scheme = origScheme
		Registry = origRegistry
		Repository = origRepository
		ItemsQueue = origItemsQueue
		ProjectList = origProjectList
		ItemByDigest = origItemByDigest
		ItemByTag = origItemByTag
	}()

	// Очищаем
	Gc = nil
	Sensitives = nil
	Scheme = ""
	Registry = ""
	Repository = ""
	ItemsQueue = nil
	ProjectList = nil
	ItemByDigest = make(map[string]*RegIndex)
	ItemByTag = make(map[string]*RegIndex)

	// Модифицируем
	Sensitives = []string{"password123", "secret-token"}
	Scheme = "https"
	Registry = "registry.example.com"
	Repository = "repo/image"
	ItemsQueue = []string{"sha256:abc", "sha256:def"}

	ri := &RegIndex{
		Reference:  "registry.example.com/repo/image:v1",
		Tag:        "v1",
		Digest:     "sha256:v1",
		Annotation: "cnab index",
	}

	ProjectList = []*RegIndex{ri}
	ItemByDigest["sha256:v1"] = ri
	ItemByTag["v1"] = ri

	// Проверяем
	if len(Sensitives) != 2 {
		t.Errorf("Sensitives length = %d, want 2", len(Sensitives))
	}
	if Sensitives[0] != "password123" {
		t.Errorf("Sensitives[0] = %q, want %q", Sensitives[0], "password123")
	}
	if Scheme != "https" {
		t.Errorf("Scheme = %q, want %q", Scheme, "https")
	}
	if Registry != "registry.example.com" {
		t.Errorf("Registry = %q, want %q", Registry, "registry.example.com")
	}
	if Repository != "repo/image" {
		t.Errorf("Repository = %q, want %q", Repository, "repo/image")
	}
	if len(ItemsQueue) != 2 {
		t.Errorf("ItemsQueue length = %d, want 2", len(ItemsQueue))
	}
	if len(ProjectList) != 1 {
		t.Errorf("ProjectList length = %d, want 1", len(ProjectList))
	}
	if ItemByDigest["sha256:v1"] != ri {
		t.Errorf("ItemByDigest['sha256:v1'] = %v, want %v", ItemByDigest["sha256:v1"], ri)
	}
	if ItemByTag["v1"] != ri {
		t.Errorf("ItemByTag['v1'] = %v, want %v", ItemByTag["v1"], ri)
	}
}

// TestRegIndex_LinkOperations проверяет операции с ссылками
func TestRegIndex_LinkOperations(t *testing.T) {
	ri := &RegIndex{
		Reference: "registry.example.com/repo/image:v1",
		Tag:       "v1",
		Digest:    "sha256:v1",
		Annotation: "cnab index",
	}

	// Добавляем DownLinks
	ri.DownLinks = append(ri.DownLinks, CnabItem{
		Digest:     "sha256:config",
		Annotation: "cnab config",
	})
	ri.DownLinks = append(ri.DownLinks, CnabItem{
		Digest:     "sha256:invocation",
		Annotation: "invocation",
	})

	// Добавляем UpLinks
	ri.UpLinks = append(ri.UpLinks, CnabItem{
		Digest:     "sha256:parent1",
		Annotation: "parent",
	})

	if len(ri.DownLinks) != 2 {
		t.Errorf("DownLinks length = %d, want 2", len(ri.DownLinks))
	}
	if len(ri.UpLinks) != 1 {
		t.Errorf("UpLinks length = %d, want 1", len(ri.UpLinks))
	}

	// Очищаем DownLinks
	ri.DownLinks = nil
	if ri.DownLinks != nil {
		t.Errorf("After clearing DownLinks, expected nil, got %v", ri.DownLinks)
	}

	// Очищаем UpLinks
	ri.UpLinks = nil
	if ri.UpLinks != nil {
		t.Errorf("After clearing UpLinks, expected nil, got %v", ri.UpLinks)
	}
}

// TestRegIndex_LostIncrement проверяет инкремент Lost
func TestRegIndex_LostIncrement(t *testing.T) {
	ri := &RegIndex{
		Reference: "registry.example.com/repo/image:v1",
		Tag:       "v1",
		Digest:    "sha256:v1",
		Annotation: "cnab index",
	}

	if ri.Lost != 0 {
		t.Errorf("Initial Lost = %d, want 0", ri.Lost)
	}

	ri.Lost++
	if ri.Lost != 1 {
		t.Errorf("After Lost++, Lost = %d, want 1", ri.Lost)
	}

	ri.Lost += 2
	if ri.Lost != 3 {
		t.Errorf("After Lost += 2, Lost = %d, want 3", ri.Lost)
	}

	ri.Lost = 0
	if ri.Lost != 0 {
		t.Errorf("After Lost = 0, Lost = %d, want 0", ri.Lost)
	}
}

// TestRegIndex_ContentPrettyJSON проверяет работу с pretty JSON
func TestRegIndex_ContentPrettyJSON(t *testing.T) {
	ri := &RegIndex{
		Reference: "registry.example.com/repo/image:v1",
		Tag:       "v1",
		Digest:    "sha256:v1",
		Annotation: "cnab index",
		Content: `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "manifests": [
    {
      "digest": "sha256:abc",
      "mediaType": "application/vnd.oci.image.manifest.v1+json"
    }
  ]
}`,
	}

	if ri.Content == "" {
		t.Error("RegIndex.Content should not be empty")
	}

	// Проверяем, что Content содержит ключевые поля
	expectedKeys := []string{"schemaVersion", "mediaType", "manifests", "sha256:abc"}
	for _, key := range expectedKeys {
		if !containsString(ri.Content, key) {
			t.Errorf("RegIndex.Content should contain %q", key)
		}
	}
}

// containsString проверяет, содержит ли строку подстроку
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestRegIndex_MediaTypes проверяет различные media types
func TestRegIndex_MediaTypes(t *testing.T) {
	testCases := []struct {
		media    string
		expected string
	}{
		{
			media:    "application/vnd.oci.image.index.v1+json",
			expected: "application/vnd.oci.image.index.v1+json",
		},
		{
			media:    "application/vnd.oci.image.manifest.v1+json",
			expected: "application/vnd.oci.image.manifest.v1+json",
		},
		{
			media:    "application/vnd.docker.distribution.manifest.v2+json",
			expected: "application/vnd.docker.distribution.manifest.v2+json",
		},
		{
			media:    "application/vnd.cnab.manifest.v1+json",
			expected: "application/vnd.cnab.manifest.v1+json",
		},
		{
			media:    "application/json",
			expected: "application/json",
		},
	}

	for _, tc := range testCases {
		ri := &RegIndex{
			Reference: "registry.example.com/repo/image:v1",
			Tag:       "v1",
			Digest:    "sha256:test",
			Media:     tc.media,
			Annotation: "test",
		}

		if ri.Media != tc.expected {
			t.Errorf("RegIndex.Media = %q, want %q", ri.Media, tc.expected)
		}
	}
}

// TestRegIndex_AnnotationMapping проверяет маппинг annotation по media type
func TestRegIndex_AnnotationMapping(t *testing.T) {
	testCases := []struct {
		media      string
		annotation string
	}{
		{
			media:      "application/vnd.oci.image.manifest.v1+json",
			annotation: "cnab config",
		},
		{
			media:      "application/vnd.oci.image.index.v1+json",
			annotation: "cnab index",
		},
		{
			media:      "application/vnd.docker.distribution.manifest.v1+prettyjws",
			annotation: "docker image",
		},
		{
			media:      "application/octet-stream",
			annotation: "stuff",
		},
	}

	for _, tc := range testCases {
		ri := &RegIndex{
			Reference:  "registry.example.com/repo/image:v1",
			Tag:        "v1",
			Digest:     "sha256:test",
			Media:      tc.media,
			Annotation: tc.annotation,
		}

		if ri.Annotation != tc.annotation {
			t.Errorf("RegIndex.Annotation for media %q = %q, want %q", tc.media, ri.Annotation, tc.annotation)
		}
	}
}

// TestRegIndex_ReferenceParsing проверяет разбор reference
func TestRegIndex_ReferenceParsing(t *testing.T) {
	testCases := []struct {
		ref        string
		registry   string
		repository string
		tag        string
	}{
		{
			ref:        "registry.example.com/repo/image:v1",
			registry:   "registry.example.com",
			repository: "repo/image",
			tag:        "v1",
		},
		{
			ref:        "osmp-docker-storage.repository.avp.ru/osmp/tmp/plugin-kes-linux/12.5.0.97/plugin-kes-linux:12.5.0.97",
			registry:   "osmp-docker-storage.repository.avp.ru",
			repository: "osmp/tmp/plugin-kes-linux/12.5.0.97/plugin-kes-linux",
			tag:        "12.5.0.97",
		},
		{
			ref:        "registry.example.com/repo/image@sha256:abc",
			registry:   "registry.example.com",
			repository: "repo/image",
			tag:        "",
		},
	}

	for _, tc := range testCases {
		ri := &RegIndex{
			Reference: tc.ref,
			Tag:       tc.tag,
			Digest:    "sha256:test",
			Annotation: "cnab index",
		}

		if ri.Reference != tc.ref {
			t.Errorf("RegIndex.Reference = %q, want %q", ri.Reference, tc.ref)
		}
		if ri.Tag != tc.tag {
			t.Errorf("RegIndex.Tag = %q, want %q", ri.Tag, tc.tag)
		}
	}
}

// TestRegIndex_Compound проверяет составные операции
func TestRegIndex_Compound(t *testing.T) {
	// Создаём несколько RegIndex и добавляем в ProjectList
	ri1 := &RegIndex{
		Reference:  "registry.example.com/repo/image:v1",
		Tag:        "v1",
		Digest:     "sha256:v1",
		Annotation: "cnab index",
	}

	ri2 := &RegIndex{
		Reference:  "registry.example.com/repo/image:v2",
		Tag:        "v2",
		Digest:     "sha256:v2",
		Annotation: "docker image",
	}

	// Добавляем в ProjectList
	ProjectList = append(ProjectList, ri1, ri2)

	if len(ProjectList) != 2 {
		t.Errorf("ProjectList length = %d, want 2", len(ProjectList))
	}

	// Добавляем в ItemByDigest
	ItemByDigest["sha256:v1"] = ri1
	ItemByDigest["sha256:v2"] = ri2

	if len(ItemByDigest) != 2 {
		t.Errorf("ItemByDigest length = %d, want 2", len(ItemByDigest))
	}

	// Добавляем в ItemByTag
	ItemByTag["v1"] = ri1
	ItemByTag["v2"] = ri2

	if len(ItemByTag) != 2 {
		t.Errorf("ItemByTag length = %d, want 2", len(ItemByTag))
	}

	// Проверяем связь
	if ItemByDigest["sha256:v1"] != ri1 {
		t.Error("ItemByDigest['sha256:v1'] should point to ri1")
	}
	if ItemByTag["v1"] != ri1 {
		t.Error("ItemByTag['v1'] should point to ri1")
	}
	if ItemByDigest["sha256:v2"] != ri2 {
		t.Error("ItemByDigest['sha256:v2'] should point to ri2")
	}
	if ItemByTag["v2"] != ri2 {
		t.Error("ItemByTag['v2'] should point to ri2")
	}

	// Обновляем существующий элемент
	ri1.Lost = 3
	ItemByDigest["sha256:v1"] = ri1

	// Проверяем, что обновление отразилось
	if ItemByDigest["sha256:v1"].Lost != 3 {
		t.Errorf("After update, Lost = %d, want 3", ItemByDigest["sha256:v1"].Lost)
	}
}
