package content

import (
	"cnabtool/pkg/client"
	"cnabtool/pkg/logging"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// PurgeEmptyFolders удаляет пустые родительские папки через Artifactory Storage API.
// Работает и в --dry-run (показывает, что было бы удалено).
// Перехватывает все ошибки без panic — даже при таймаутах или сетевых сбоях.
func (c *Config) PurgeEmptyFolders(cl *client.RegClient) {
	if !c.Purge {
		return
	}

	repoKey := c.deriveRepoKey(cl)
	if repoKey == "" {
		logging.Error("purge: cannot derive repo-key, use --repo-key")
		return
	}

	// Начинаем с пути репозитория (без тега)
	currentPath := cl.Repository
	if currentPath == "" || currentPath == "/" {
		logging.Debug("purge: empty repository path, nothing to purge")
		return
	}

	// Artifactory DELETE папки может занимать >60 секунд.
	// Отдельный клиент с таймаутом 180 секунд только для DELETE-операций purge.
	purgeClient := &http.Client{Timeout: 180 * time.Second}

	// Собираем длительности успешных DELETE для адаптивной остановки.
	var deletionTimes []time.Duration

	for {
		// 1. Проверить содержимое папки через Storage API
		storageURL := fmt.Sprintf("%s://%s/artifactory/api/storage/%s/%s?list",
			cl.Scheme, cl.Registry, repoKey, currentPath)

		if c.DryRun {
			logging.Normal(fmt.Sprintf("[dry-run] Purge: check folder %s", storageURL))
		} else {
			logging.Debug(fmt.Sprintf(">> Purge: check folder %s for emptiness", currentPath))
		}

		resp, err := cl.WebRequestEx("GET", storageURL)
		if err != nil {
			logging.Error(fmt.Sprintf("purge: cannot list folder %s: %v", currentPath, err))
			break
		}
		if resp == nil {
			logging.Error(fmt.Sprintf("purge: nil response for folder %s", currentPath))
			break
		}
		if resp.StatusCode != 200 {
			logging.Error(fmt.Sprintf("purge: unexpected status %d for folder %s", resp.StatusCode, currentPath))
			resp.Body.Close()
			break
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, client.MaxBodySize))
		resp.Body.Close()
		if err != nil {
			logging.Error(fmt.Sprintf("purge: cannot read response for %s: %v", currentPath, err))
			break
		}

		var info struct {
			Children []struct {
				URI    string `json:"uri"`
				Folder bool   `json:"folder"`
			} `json:"children"`
		}
		if err := json.Unmarshal(body, &info); err != nil {
			logging.Error(fmt.Sprintf("purge: cannot parse list for %s: %v", currentPath, err))
			break
		}

		// Если есть дочерние элементы — папка не пустая, останавливаемся
		if len(info.Children) > 0 {
			logging.Debug(fmt.Sprintf(">> Purge: folder %s is not empty, stopping", currentPath))
			break
		}

		// 2. Папка пустая — удаляем (или показываем, что удалили бы)
		deleteURL := fmt.Sprintf("%s://%s/artifactory/%s/%s",
			cl.Scheme, cl.Registry, repoKey, currentPath)

		if c.DryRun {
			logging.Normal(fmt.Sprintf("[dry-run] Purge: folder %s is empty, would delete %s", currentPath, deleteURL))
		} else {
			logging.Normal(fmt.Sprintf("Purge: delete empty folder %s", currentPath))

			req, err := http.NewRequest("DELETE", deleteURL, nil)
			if err != nil {
				logging.Error(fmt.Sprintf("purge: failed to build request for %s: %v", currentPath, err))
				break
			}
			if cl.Credentials.Username != "" && cl.Credentials.Password != "" {
				req.SetBasicAuth(cl.Credentials.Username, cl.Credentials.Password)
			}

			start := time.Now()
			delResp, err := purgeClient.Do(req)
			elapsed := time.Since(start)

			if err != nil {
				// Штатная обработка таймаута: Artifactory может дообработать удаление в фоне.
				if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
					logging.Normal(fmt.Sprintf("[warning] Purge: delete folder %s timed out after 180s. Artifactory may still process it in background. Stopping purge.", currentPath))
				} else {
					logging.Error(fmt.Sprintf("purge: failed to delete folder %s: %v", currentPath, err))
				}
				break
			}
			if delResp != nil {
				delResp.Body.Close()
				if delResp.StatusCode >= 300 {
					logging.Error(fmt.Sprintf("purge: failed to delete folder %s: HTTP %d", currentPath, delResp.StatusCode))
					break
				}
			}

			logging.Normal(fmt.Sprintf("Purge: folder %s deleted in %v", currentPath, elapsed))
			deletionTimes = append(deletionTimes, elapsed)

			// Адаптивная остановка: если текущее удаление в 5+ раз медленнее среднего предыдущего.
			if len(deletionTimes) >= 2 {
				var sum time.Duration
				for _, d := range deletionTimes[:len(deletionTimes)-1] {
					sum += d
				}
				avg := sum / time.Duration(len(deletionTimes)-1)
				if avg > 0 && elapsed > 5*avg {
					logging.Normal(fmt.Sprintf("[warning] Purge: delete time %v is >5× average %v. Parent folder likely too heavy. Stopping purge.", elapsed, avg))
					break
				}
			}
		}

		// 3. Подняться к родителю
		parent := path.Dir(currentPath)
		if parent == "." || parent == "/" || parent == currentPath {
			break
		}
		currentPath = parent
	}

	logging.Normal("Purge: completed")
}

// deriveRepoKey вычисляет repo-key Artifactory из hostname registry.
// Если задан --repo-key, использует его. Иначе берёт первую часть hostname.
func (c *Config) deriveRepoKey(cl *client.RegClient) string {
	if c.RepoKey != "" {
		return c.RepoKey
	}
	host := cl.Registry
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx] // убрать порт, если есть
	}
	parts := strings.Split(host, ".")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}
