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

	var tagList []string // list tags for delete

	// parse the first reference to get the first tag
	if err := cl.ParseReference(); err != nil {
		logging.Fatal(fmt.Sprintf("can not parse reference %+v", err.Error()))
	}
	//fmt.Printf("RegClient %+v\n", cl)
	//fmt.Printf("Tag %+v\n", cl.Tag)
	logging.Debug(fmt.Sprintf("Registry item %+v\n", data.ItemByTag[cl.Tag]))

	for _, childItem := range data.ItemByTag[cl.Tag].DownLinks { // for all child items
		_, ok := data.ItemByDigest[childItem.Digest]
		if ok {
			linksCount := len(data.ItemByDigest[childItem.Digest].UpLinks)
			logging.Debug(fmt.Sprintf("Down link %+v, count %d, tag %s\n", childItem, linksCount, data.ItemByDigest[childItem.Digest].Tag))
			if linksCount == 1 {
				tagList = append(tagList, data.ItemByDigest[childItem.Digest].Tag)
			}
		}
	}
	tagList = append(tagList, cl.Tag) // delete parent manifest at finally only!

	logging.Info(fmt.Sprintf("Tags for delete %+v", tagList))

	for _, tag := range tagList {

		url := data.Scheme + "://" + data.Registry + "/v2/" + data.Repository + "/manifests/" + data.ItemByTag[tag].Digest
		if data.Gc.Verbosity >= logging.LogNormalLevel {
			fmt.Printf("Delete %s %s\n", data.ItemByTag[tag].Annotation, url)
		}
		if cc.DryRun { // do nothing if dry run mode
			continue
		}
		res, err := cl.WebDelete(url)
		if err != nil { // unrecoverable error
			logging.Error(fmt.Sprintf("response is nil, %+v", err.Error()))
		} else if res != nil {
			//fmt.Printf("Delete %s %+v\n\n", tag, res)
			if res.StatusCode == 202 {
				logging.Info(fmt.Sprintf("Item %s was deleted successfully", tag))
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
