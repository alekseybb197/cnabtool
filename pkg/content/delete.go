package content

import (
	"cnabtool/pkg/client"
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"encoding/json"
	"fmt"
	"io"
)

func (cc *Config) DeleteCnab(cl *client.RegClient) {

	// parse the first reference to get the project metadata
	if err := cl.ParseReference(); err != nil {
		logging.Fatal(fmt.Sprintf("can not parse reference %+v", err.Error()))
	}

	// Collect all items to delete: for every CNAB index, delete its DownLinks first, then the index itself.
	// This ensures we delete by digest (not by tag), which is critical for untagged manifests
	// (config, invocation, etc.) that were fetched directly by digest during InspectCnab.
	type deleteEntry struct {
		annotation string
		digest     string
		url        string
	}

	var toDelete []deleteEntry
	deletedDigests := make(map[string]bool) // avoid duplicate deletions

	for _, item := range data.ItemByTag {
		if item.Annotation != data.ItemTypeCnab {
			continue
		}

		// Delete all DownLinks (child components) first
		for _, link := range item.DownLinks {
			_, ok := data.ItemByDigest[link.Digest]
			if !ok {
				// This link was not found during inspection — skip it
				continue
			}
			// Only delete leaf nodes (referenced by exactly one parent) or all if no uplink info
			ri := data.ItemByDigest[link.Digest]
			if len(ri.UpLinks) > 1 {
				// Referenced by multiple parents — skip to avoid deleting shared components
				continue
			}
			if deletedDigests[link.Digest] {
				continue
			}
			deletedDigests[link.Digest] = true

			url := data.Scheme + "://" + data.Registry + "/v2/" + data.Repository + "/manifests/" + link.Digest
			toDelete = append(toDelete, deleteEntry{
				annotation: link.Annotation,
				digest:     link.Digest,
				url:        url,
			})
		}

		// Delete the CNAB index itself
		if deletedDigests[item.Digest] {
			continue
		}
		deletedDigests[item.Digest] = true
		url := data.Scheme + "://" + data.Registry + "/v2/" + data.Repository + "/manifests/" + item.Digest
		toDelete = append(toDelete, deleteEntry{
			annotation: item.Annotation,
			digest:     item.Digest,
			url:        url,
		})
	}

	logging.Info(fmt.Sprintf("Items to delete: %d", len(toDelete)))

	// Perform deletions
	for _, entry := range toDelete {
		logging.Message(fmt.Sprintf("Delete %s %s", entry.annotation, entry.url))
		if cc.DryRun {
			continue
		}
		res, err := cl.WebDelete(entry.url)
		if err != nil {
			logging.Error(fmt.Sprintf("response is nil, %+v", err.Error()))
			continue
		}
		if res != nil {
			if res.StatusCode == 202 {
				logging.Message(fmt.Sprintf("Item %s was deleted successfully", entry.digest))
			} else {
				logging.Error(fmt.Sprintf("Error %d", res.StatusCode))

				// get body if there was an error
				reader := res.Body
				bytesbody, readErr := io.ReadAll(io.LimitReader(reader, client.MaxBodySize))
				if readErr != nil {
					errLine := fmt.Sprintf("failed to fetch response body %s", readErr)
					logging.Error(errLine)
				}
				res.Body.Close()

				// body must be json
				if !json.Valid(bytesbody) {
					errLine := fmt.Sprintf("response body is not valid json, status %d, headers %+v", res.StatusCode, res.Header)
					logging.Error(errLine)
					logging.Debug(fmt.Sprintf("response body  %+v", string(bytesbody)))
				}
				jsonres, err := logging.PrettyString(string(bytesbody))
				if err != nil {
					errLine := fmt.Sprintf("response body is unvalid json, status %d, headers %+v", res.StatusCode, res.Header)
					logging.Error(errLine)
					logging.Debug(fmt.Sprintf("response body  %+v", string(bytesbody)))
				} else {
					if data.Gc.Verbosity >= logging.LogNormalLevel {
						fmt.Printf("%s\n", jsonres)
					}
				}
			}
		}
	}
}
