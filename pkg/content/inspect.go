package content

import (
	"cnabtool/pkg/client"
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"strconv"
)

// AddCnab add cnab to ItemByDigest collection

func AddCnab(regres *client.RegResponse, tag string) error {

	ri, err := AddIndex(regres, tag)
	js := ri.Content
	// drop old list, if exists
	ri.DownLinks = nil
	ri.UpLinks = nil
	ri.Lost = 0

	// at first check if manifests is exists
	manifests, keytype, _, err := jsonparser.Get(([]byte)(js), "manifests")
	if err != nil {
		errLine := fmt.Sprintf("json isn't contain manifests key, %+v", err.Error())
		logging.Error(errLine)
		return errors.New(errLine)
	}
	//logging.Info(fmt.Sprintf("manifests %+v, %+v", string(manifests), keytype))
	if keytype.String() != "array" {
		errLine := fmt.Sprintf("manifests key must contain array %+v", string(manifests))
		logging.Error(errLine)
		return errors.New(errLine)
	}

	// parse manifests
	jsonparser.ArrayEach(([]byte)(js), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		digest, _, _, err := jsonparser.Get(value, "digest")
		if err != nil {
			errLine := fmt.Sprintf("manifests digest is invalid, %+v", err.Error())
			logging.Error(errLine)
		}
		media, _, _, err := jsonparser.Get(value, "mediaType")
		if err != nil {
			errLine := fmt.Sprintf("manifests mediaType is invalid, %+v", err.Error())
			logging.Error(errLine)
		}
		annotation, _, _, err := jsonparser.Get(value, "annotations", "io.cnab.manifest.type")
		if err != nil {
			errLine := fmt.Sprintf("manifests annotations is invalid, %+v", err.Error())
			logging.Error(errLine)
		}

		// for component item continue decode
		realAnnotation := string(annotation)
		if realAnnotation == "component" {
			annotation, _, _, err := jsonparser.Get(value, "annotations", "io.cnab.component.name")
			if err != nil {
				errLine := fmt.Sprintf("manifests annotations is invalid, %+v", err.Error())
				logging.Error(errLine)
			}
			realAnnotation = string(annotation)
		}

		logging.Debug(fmt.Sprintf("Found media %s, annotation %s, digest %s", media, realAnnotation, digest))

		if string(media) == client.MediaTypeOciManifest || string(media) == client.MediaTypeV2Manifest {
			data.ItemsQueue = append(data.ItemsQueue, string(digest))
			ri.DownLinks = append(ri.DownLinks, data.CnabItem{Digest: string(digest), Annotation: realAnnotation})
		}

	}, "manifests")

	logging.Debug(fmt.Sprintf("new registry index %+v", ri))

	return nil
}

func AddIndex(regres *client.RegResponse, tag string) (*data.RegIndex, error) {
	ri, ok := data.ItemByDigest[regres.Digest]
	if ok {
		logging.Debug(fmt.Sprintf("already has %+v", ri))
	} else {
		// otherwise make new
		ri = &data.RegIndex{
			Reference: regres.Reference,
			Tag:       tag,
			Media:     regres.Media,
			Date:      regres.Date,
			Digest:    regres.Digest,
			Lost:      0,
		}

		switch regres.Media {
		case client.MediaTypeV1Pretty:
			ri.Annotation = data.ItemTypeImage
		case client.MediaTypeOciManifest:
			ri.Annotation = data.ItemTypeConfig
		case client.MediaTypeOciIndex:
			ri.Annotation = data.ItemTypeCnab
		default:
			ri.Annotation = data.ItemTypeStuff
		}

		data.ItemByDigest[regres.Digest] = ri
		data.ProjectList = append(data.ProjectList, ri)

		// scan context and push digests to queue

		// check entry call - it must be correct cnab index
		js, err := logging.PrettyString(regres.Content)
		if err != nil {
			errLine := fmt.Sprintf("invalid context, %+v", err.Error())
			logging.Error(errLine)
			return nil, errors.New(errLine)
		}
		ri.Content = js
		logging.Debug(fmt.Sprintf("new index %+v", ri))
	}
	data.ItemByTag[tag] = ri
	return ri, nil

}

func (cc *Config) InspectCnab(cl *client.RegClient) {

	// do request and get current tags list of cnab project
	regres, err := cl.GetTagList()
	if err != nil {
		errLine := fmt.Sprintf("failed to fetch tag list %s", err.Error())
		logging.Fatal(errLine)
	}
	logging.Debug(fmt.Sprintf("Response with Tag List %+v", regres))

	// check entry call - it must be correct cnab index
	js, err := logging.PrettyString(regres.Content)
	if err != nil {
		errLine := fmt.Sprintf("invalid context, %+v", err.Error())
		logging.Fatal(errLine)
	}

	// at first check if manifests is exists
	tags, keytype, _, err := jsonparser.Get(([]byte)(js), "tags")
	if err != nil {
		errLine := fmt.Sprintf("json isn't contain tags key, %+v", err.Error())
		logging.Fatal(errLine)
	}
	if data.Gc.Verbosity <= logging.LogInfoLevel {
		// avoid double logging
		logging.Info(fmt.Sprintf("Project tags list %+v", string(tags)))
	}
	if keytype.String() != "array" {
		errLine := fmt.Sprintf("manifests key must contain array %+v", string(tags))
		logging.Fatal(errLine)
	}

	// parse tags
	i := 0
	for true {
		val, err := jsonparser.GetString(([]byte)(js), "tags", "["+strconv.Itoa(i)+"]")
		if err != nil {
			break
		}
		logging.Debug(fmt.Sprintf("Get item %d with tag %+v\n", i, val))

		i++
		cl.Digest = ""
		cl.Tag = val
		// get index by tag
		regres, err := cl.GetRegIndex()
		if err != nil {
			errLine := fmt.Sprintf("can't fetch index for tar %s, %+v", val, err.Error())
			logging.Error(errLine)
		} else {
			logging.Debug(fmt.Sprintf("Content %+v", regres))
			switch regres.Media {
			case client.MediaTypeOciIndex:
				// cnab
				AddCnab(regres, val)
			default:
				AddIndex(regres, val)
			}
		}
	}

	// scan cnab indexes and mark used resources
	for _, item := range data.ItemByTag { // for all tags
		if item.Annotation == data.ItemTypeCnab { // chose cnab only
			for _, link := range item.DownLinks { // for all down links from selected cnab
				cri, ok := data.ItemByDigest[link.Digest] // try to get item by digest from down link
				if ok {
					// if the uplink has already been registered, it does not need to be re-registered
					needed := true
					for _, uplink := range cri.UpLinks {
						if uplink.Digest == item.Digest {
							needed = false
						}
					}
					if needed {
						//logging.Info(fmt.Sprintf("For cnab %s component %s add uplink", item.Tag, link.Digest))
						cri.UpLinks = append(cri.UpLinks, data.CnabItem{Digest: item.Digest, Annotation: item.Annotation})
						cri.Annotation = link.Annotation
					}
				} else {
					logging.Error(fmt.Sprintf("For cnab %s component %s was not found", item.Tag, link.Digest))
					item.Lost++
				}
			}
		}
	}
	// here data.ProjectList made completely!
}

func (cc *Config) ShowCnabReport(cl *client.RegClient) {

	if data.Gc.Verbosity >= logging.LogNormalLevel {

		if cc.Raw { // very long output
			i := 1
			for tag, item := range data.ItemByTag {
				fmt.Printf("Project item %d: %s ---- %+v\n\n\n", i, tag, item)
				i++
			}
		} else {
			// translate ProjectList to shortOutput
			type shortListItem struct {
				Tag        string `json:"tag"`
				Digest     string `json:"digest"`
				Annotation string `json:"annotation"`
				Date       string `json:"date"`
				Media      string `json:"media"`
				Count      int    `json:"count"`
				Links      int    `json:"links"`
				Lost       int    `json:"lost"`
			}
			type shortOutput struct {
				Reference string          `json:"reference"`
				Shortlist []shortListItem `json:"itemList"`
			}
			i := 1
			sl := shortOutput{
				Reference: data.ProjectList[0].Reference,
				Shortlist: nil,
			}
			for tag, item := range data.ItemByTag {
				i++
				sl.Shortlist = append(sl.Shortlist, shortListItem{
					Tag:        tag,
					Digest:     item.Digest,
					Annotation: item.Annotation,
					Date:       item.Date,
					Media:      item.Media,
					Count:      len(item.UpLinks),
					Links:      len(item.DownLinks),
					Lost:       item.Lost,
				})
			}
			// print shortOutput
			slout, err := json.Marshal(sl)
			if err == nil {
				if jsonres, err := logging.PrettyString(string(slout)); err == nil {
					fmt.Println(jsonres)
				} else {
					// print as is
					fmt.Printf("%+v\n", sl)
				}
			} else {
				logging.Fatal(fmt.Sprintf("can not convert list item - %+v\n", err.Error()))
			}
		} //format
	} //logging
}
