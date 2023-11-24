package client

import (
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const MaxBodySize = 8096 // max body response for index request

const (
	StringSlash = "/"
	StringDot   = "."
	StringColon = ":"
	StringAt    = "@"
	StringSha   = "@sha256:"
)

type RegClient struct {
	Reference string

	// registry item
	Scheme     string
	Registry   string
	Repository string
	Tag        string
	Digest     string

	Credentials data.Credentials
	Client      string

	WebClient http.Client // web client
}

const (
	// https://github.com/opencontainers/image-spec/blob/main/media-types.md
	MediaTypeOciDescriptor   = "application/vnd.oci.descriptor.v1+json"                       // Content Descriptor
	MediaTypeOciHeader       = "application/vnd.oci.layout.header.v1+json"                    // OCI Layout
	MediaTypeOciIndex        = "application/vnd.oci.image.index.v1+json"                      // Image Index
	MediaTypeOciManifest     = "application/vnd.oci.image.manifest.v1+json"                   // Image manifest
	MediaTypeOciConfig       = "application/vnd.oci.image.config.v1+json"                     // Image config
	MediaTypeOciBlobTar      = "application/vnd.oci.image.layer.v1.tar"                       // Layer, as a tar archive
	MediaTypeOciBlobTarGz    = "application/vnd.oci.image.layer.v1.tar+gzip"                  // Layer, as a tar archive compressed with gzip
	MediaTypeOciBlobZstd     = "application/vnd.oci.image.layer.v1.tar+zstd"                  // Layer, as a tar archive compressed with zstd
	MediaTypeOciEmpty        = "application/vnd.oci.empty.v1+json"                            // Empty for unused descriptors
	MediaTypeOciBlobNonTar   = "application/vnd.oci.image.layer.nondistributable.v1.tar"      // Layer, as a tar archive
	MediaTypeOciBlobNonTarGz = "application/vnd.oci.image.layer.nondistributable.v1.tar+gzip" // Layer, as a tar archive with distribution restrictions compressed with gzip
	MediaTypeOciBlobNonZstd  = "application/vnd.oci.image.layer.nondistributable.v1.tar+zstd" // Layer, as a tar archive with distribution restrictions compressed with zstd

	// https://github.com/opencontainers/image-spec/blob/main/manifest.md
	// https://distribution.github.io/distribution/spec/manifest-v2-2/
	MediaTypeV2Manifest  = "application/vnd.docker.distribution.manifest.v2+json"      // new image manifest
	MediaTypeV2List      = "application/vnd.docker.distribution.manifest.list.v2+json" // manifest list
	MediaTypeV1Manifest  = "application/vnd.docker.container.image.v1+json"            // container config
	MediaTypeBlobTar     = "application/vnd.docker.image.rootfs.diff.tar.gzip"         // layer
	MediaTypeForeignBlob = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip" // foreign layer
	MediaTypePlugin      = "application/vnd.docker.plugin.v1+json"                     // plugin config

	// https://github.com/cnabio/cnab-spec/blob/main/201-representing-CNAB-in-OCI.md
	MediaTypeCnabManifest = "application/vnd.cnab.manifest.v1+json"      // cnab manifest
	MediaTypeCnabConfig   = "application/vnd.cnab.config.v1+json"        // bundle.json
	MediaTypeCnabBConfig  = "application/vnd.cnab.bundle.config.v1+json" // bundle.json

	MediaTypeJson = "application/json" // json response

)

type Config data.Config

//

const ()

var mediatype = map[string]string{
	MediaTypeOciIndex:    "oci index",
	MediaTypeOciManifest: "oci manifest",
	MediaTypeV2Manifest:  "docker index",
	MediaTypeV2List:      "",
	MediaTypeV1Manifest:  "docker manifest",
}

/*
https://docs.docker.com/registry/
Docker Hub supports the following image manifest formats for pulling images:

OCI image manifest
Docker image manifest version 2, schema 2
Docker image manifest version 2, schema 1
Docker image manifest version 1
*/

var manifest_discovery = []string{
	MediaTypeV1Manifest,
	MediaTypeV2Manifest,
	MediaTypeOciIndex,
	MediaTypeJson,
}

// split reference to registry, repository, tag and direst

func ParseReference(ref *RegClient) error {
	if !strings.Contains(ref.Reference, StringSlash) {
		return errors.New(fmt.Sprintf("reference does not contain registry part in %s", ref.Reference))
	}
	// first part before slash is registry address
	ref.Registry = (strings.Split(ref.Reference, StringSlash))[0]
	if !strings.Contains(ref.Registry, StringDot) {
		return errors.New(fmt.Sprintf("reference does not contain repository part in %s", ref.Reference))
	}

	// reference without registry consist repository:tag@digest
	drop_registry := strings.TrimPrefix(ref.Reference, ref.Registry+StringSlash)
	if !strings.Contains(drop_registry, StringColon) && !strings.Contains(drop_registry, StringAt) {
		return errors.New(fmt.Sprintf("reference does not contain tag or digest part in %s", ref.Reference))
	}

	// at first split digest suffix
	drop_digest := ""
	if strings.Contains(drop_registry, StringSha) {
		ref.Digest = (strings.Split(drop_registry, StringAt))[1]
		drop_digest = strings.TrimSuffix(drop_registry, StringAt+ref.Digest)
	} else {
		ref.Digest = ""
		drop_digest = drop_registry
	}
	if strings.Contains(drop_digest, StringColon) {
		repo_tag := strings.Split(drop_digest, StringColon)
		ref.Repository = repo_tag[0]
		ref.Tag = repo_tag[1]
	} else {
		ref.Repository = drop_digest
		ref.Tag = ""
	}

	// check validity - registry and repository mustn't void
	// if tag and digest is empty , reference point to tags list
	if len(ref.Registry) == 0 || len(ref.Repository) == 0 || (len(ref.Tag) == 0 && len(ref.Digest) == 0) {
		//logging.Debug("ParseReference", fmt.Sprintf("ref %+v", ref))
		return errors.New(fmt.Sprintf("reference has non valid format - %s", ref.Reference))
	}

	return nil
}

func (cl *RegClient) DoWebRequest(url, media string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("User-Agent", cl.Client)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept", media)
	req.SetBasicAuth(cl.Credentials.Username, cl.Credentials.Password)

	logging.Debug("DoWebRequest", fmt.Sprintf("request %+v", req))

	res, err := cl.WebClient.Do(req)
	if err != nil {
		logging.Error("DoWebRequest", fmt.Sprintf("%+v", err))
		return nil
	}
	return res
}

type RegResponse struct {
	Reference string // request
	Status    int    // status
	Filename  string
	Length    int
	Media     string // media
	Date      string
	Digest    string
	Content   string // response json
}

func (cc *Config) GetRegIndex(cl *RegClient) *RegResponse {

	// tune url
	url := cl.Scheme + "://" + cl.Registry + "/v2/" + cl.Repository + "/manifests/"
	if len(cl.Tag) != 0 {
		url = url + cl.Tag
	} else {
		url = url + cl.Digest
	}

	regres := &RegResponse{
		Reference: cl.Reference,
	}

	// loop for success
	for _, ml := range manifest_discovery {

		regres.Media = ml
		res := cl.DoWebRequest(url, ml)

		// unrecoverable error
		if res == nil {
			logging.Error("GetRegIndex", fmt.Sprintf("response is nil, status %d", res.StatusCode))
			return regres
		}

		logging.Debug("GetRegIndex", fmt.Sprintf("status %d, response headers %+v", res.StatusCode, res.Header))
		// fetch response header fields
		regres.Media = res.Header.Get("Content-Type")
		regres.Date = res.Header.Get("Last-Modified")
		if i, err := strconv.Atoi(res.Header.Get("Content-Length")); err == nil {
			regres.Length = i
		}
		regres.Digest = res.Header.Get("Docker-Content-Digest")
		regres.Status = res.StatusCode
		cds := res.Header.Get("Content-Disposition")
		if strings.Contains(cds, "filename=\"") {
			regres.Filename = strings.Split(cds, "\"")[1]
		}

		// response is not valid
		if res.Body == nil {
			logging.Error("GetRegIndex", fmt.Sprintf("body is nil, status %d, headers %+v", res.StatusCode, res.Header))
			return regres
		}

		// get body
		reader := res.Body
		bytesbody, readErr := io.ReadAll(io.LimitReader(reader, MaxBodySize))
		if readErr != nil {
			logging.Fatal("GetRegIndex", fmt.Sprintf("failed to fetch response body %s", readErr))
			return nil
		}
		res.Body.Close()

		// body must be json
		if !json.Valid(bytesbody) {
			logging.Error("GetRegIndex", fmt.Sprintf("response body is not valid json, status %d, headers %+v", res.StatusCode, res.Header))
			logging.Debug("GetRegIndex", fmt.Sprintf("response body  %+v", string(bytesbody)))
			return regres
		}
		regres.Content = string(bytesbody)

		jsonres, err := logging.PrettyString(string(bytesbody))
		if err != nil {
			logging.Error("GetRegIndex", fmt.Sprintf("response body is unvalid json, status %d, headers %+v", res.StatusCode, res.Header))
			logging.Debug("GetRegIndex", fmt.Sprintf("response body  %+v", string(bytesbody)))
			return regres
		}

		logging.Debug("GetRegIndex", fmt.Sprintf("registry response %+v", regres))
		logging.Debug("GetRegIndex", fmt.Sprintf("pretty json %+v", jsonres))

		if res.StatusCode != 200 && res.StatusCode != 400 && res.StatusCode != 404 && res.StatusCode != 401 {
			logging.Error("GetRegIndex", fmt.Sprintf("failed to fetch data %s", res.Status))
			return regres
		}
		if res.StatusCode != 200 {
			logging.Debug("GetRegIndex", fmt.Sprintf("failed to fetch data: %s", res.Status))
			continue
		}

		break
	}
	return regres
}
