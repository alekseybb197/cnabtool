package client

import (
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const MaxBodySize = 32384 // max body response for index request

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
	MediaTypeV1Pretty    = "application/vnd.docker.distribution.manifest.v1+prettyjws"
	MediaTypeBlobTar     = "application/vnd.docker.image.rootfs.diff.tar.gzip"         // layer
	MediaTypeForeignBlob = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip" // foreign layer
	MediaTypePlugin      = "application/vnd.docker.plugin.v1+json"                     // plugin config

	// https://github.com/cnabio/cnab-spec/blob/main/201-representing-CNAB-in-OCI.md
	MediaTypeCnabManifest = "application/vnd.cnab.manifest.v1+json"      // cnab manifest
	MediaTypeCnabConfig   = "application/vnd.cnab.config.v1+json"        // bundle.json
	MediaTypeCnabBConfig  = "application/vnd.cnab.bundle.config.v1+json" // bundle.json

	MediaTypeJson = "application/json" // json response
)

// type tricks

type Config data.Config

// human form

var mediatype = map[string]string{
	MediaTypeOciDescriptor:   "oci content descriptor",
	MediaTypeOciHeader:       "oci layout header",
	MediaTypeOciIndex:        "oci image index",
	MediaTypeOciManifest:     "oci manifest",
	MediaTypeOciConfig:       "oci image config",
	MediaTypeOciBlobTar:      "oci layer tar",
	MediaTypeOciBlobTarGz:    "oci layer tar gz",
	MediaTypeOciBlobZstd:     "oci layer tar zstd",
	MediaTypeOciEmpty:        "oci unused",
	MediaTypeOciBlobNonTar:   "oci nondistributable layer tar",
	MediaTypeOciBlobNonTarGz: "oci nondistributable layer tar gz",
	MediaTypeOciBlobNonZstd:  "oci nondistributable layer tar zstd",

	MediaTypeV2Manifest:  "docker index v2",
	MediaTypeV2List:      "docker list v2",
	MediaTypeV1Manifest:  "docker manifest v1",
	MediaTypeV1Pretty:    "docker manifest v1 pretty",
	MediaTypeBlobTar:     "docker layer tar gz",
	MediaTypeForeignBlob: "docker foreign layer tar gz",
	MediaTypePlugin:      "docker plugin",

	MediaTypeCnabManifest: "cnab manifest",
	MediaTypeCnabConfig:   "cnab config",
	MediaTypeCnabBConfig:  "cnab bundle config",

	MediaTypeJson: "json",
}

/*
https://docs.docker.com/registry/
Docker Hub supports the following image manifest formats for pulling images:

OCI image manifest
Docker image manifest version 2, schema 2
Docker image manifest version 2, schema 1
Docker image manifest version 1
*/

/*var manifest_discovery = []string{
	MediaTypeV1Pretty,
	MediaTypeV1Manifest,
	MediaTypeV2Manifest,
	MediaTypeOciIndex,
	MediaTypeJson,
}*/

var manifest_discovery = []string{
	MediaTypeOciIndex,
	MediaTypeV2Manifest, // it's very important! this media return correct sha256
	MediaTypeV1Manifest,
	MediaTypeV1Pretty,
	MediaTypeJson,
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

func NewRegClient(cc *Config, reference string) *RegClient {
	// make client
	cl := &RegClient{
		Reference: reference,
		Scheme:    cc.Scheme,
		Credentials: data.Credentials{
			Username: cc.Credentials.Username,
			Password: cc.Credentials.Password,
		},
		Client: cc.Client,
		WebClient: http.Client{
			Timeout: time.Millisecond * time.Duration(cc.Timeout),
		},
	}
	return cl
}

// split reference to registry, repository, tag and direst

func (cl *RegClient) ParseReference() error {
	if !strings.Contains(cl.Reference, StringSlash) {
		return errors.New(fmt.Sprintf("reference does not contain registry part in %s", cl.Reference))
	}
	// first part before slash is registry address
	cl.Registry = (strings.Split(cl.Reference, StringSlash))[0]
	if !strings.Contains(cl.Registry, StringDot) {
		return errors.New(fmt.Sprintf("reference does not contain repository part in %s", cl.Reference))
	}

	// reference without registry consist repository:tag@digest
	drop_registry := strings.TrimPrefix(cl.Reference, cl.Registry+StringSlash)
	if !strings.Contains(drop_registry, StringColon) && !strings.Contains(drop_registry, StringAt) {
		return errors.New(fmt.Sprintf("reference does not contain tag or digest part in %s", cl.Reference))
	}

	// at first split digest suffix
	drop_digest := ""
	if strings.Contains(drop_registry, StringSha) {
		cl.Digest = (strings.Split(drop_registry, StringAt))[1]
		drop_digest = strings.TrimSuffix(drop_registry, StringAt+cl.Digest)
	} else {
		cl.Digest = ""
		drop_digest = drop_registry
	}
	if strings.Contains(drop_digest, StringColon) {
		repo_tag := strings.Split(drop_digest, StringColon)
		cl.Repository = repo_tag[0]
		cl.Tag = repo_tag[1]
	} else {
		cl.Repository = drop_digest
		cl.Tag = ""
	}

	// check validity - registry and repository mustn't void
	// if tag and digest is empty , reference point to tags list
	if len(cl.Registry) == 0 || len(cl.Repository) == 0 || (len(cl.Tag) == 0 && len(cl.Digest) == 0) {
		//logging.Debug(fmt.Sprintf("ref %+v", ref))
		return errors.New(fmt.Sprintf("reference has non valid format - %s", cl.Reference))
	}

	return nil
}

// WebRequest - provide get request to registry

func (cl *RegClient) WebRequest(url, media string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err.Error()))
		return nil, err
	}
	req.Header.Set("User-Agent", cl.Client)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept", media)
	req.SetBasicAuth(cl.Credentials.Username, cl.Credentials.Password)

	logging.Debug(fmt.Sprintf("request %+v", req))

	res, err := cl.WebClient.Do(req)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err.Error()))
		return nil, err
	}
	return res, nil
}

// WebDelete - provide delete request for destroy resources

func (cl *RegClient) WebDelete(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err.Error()))
		return nil, err
	}
	req.Header.Set("User-Agent", cl.Client)
	req.SetBasicAuth(cl.Credentials.Username, cl.Credentials.Password)

	logging.Debug(fmt.Sprintf("request %+v", req))

	res, err := cl.WebClient.Do(req)
	if err != nil {
		logging.Error(fmt.Sprintf("%+v", err.Error()))
		return nil, err
	}
	return res, nil
}

// FillResponse - do decode response

func (regres *RegResponse) FillResponse(res *http.Response) error {

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
		errLine := fmt.Sprintf("body is nil, status %d, headers %+v", res.StatusCode, res.Header)
		logging.Error(errLine)
		return errors.New(errLine)
	}

	// get body
	reader := res.Body
	bytesbody, readErr := io.ReadAll(io.LimitReader(reader, MaxBodySize))
	if readErr != nil {
		errLine := fmt.Sprintf("failed to fetch response body %s", readErr)
		logging.Error(errLine)
		return errors.New(errLine)
	}
	res.Body.Close()

	// body must be json
	if !json.Valid(bytesbody) {
		errLine := fmt.Sprintf("response body is not valid json, status %d, headers %+v", res.StatusCode, res.Header)
		logging.Error(errLine)
		logging.Debug(fmt.Sprintf("response body  %+v", string(bytesbody)))
		return errors.New(errLine)
	}
	regres.Content = string(bytesbody)

	jsonres, err := logging.PrettyString(string(bytesbody))
	if err != nil {
		errLine := fmt.Sprintf("response body is unvalid json, status %d, headers %+v", res.StatusCode, res.Header)
		logging.Error(errLine)
		logging.Debug(fmt.Sprintf("response body  %+v", string(bytesbody)))
		return errors.New(errLine)
	}

	//logging.Debug(fmt.Sprintf("registry response %+v", regres))
	logging.Debug(fmt.Sprintf("pretty json %+v", jsonres))
	return nil
}

func (cl *RegClient) GetTagList() (*RegResponse, error) {

	//logging.Info(fmt.Sprintf("%+v", cl))

	// tune url
	url := cl.Scheme + "://" + cl.Registry + "/v2/" + cl.Repository + "/tags/list/"

	regres := &RegResponse{
		Reference: strings.Replace(url, cl.Scheme+"://", "", 1),
		Media:     MediaTypeJson,
	}

	res, err := cl.WebRequest(url, MediaTypeJson)

	if err != nil { // unrecoverable error
		logging.Error(fmt.Sprintf("response is nil, %+v", err.Error()))
		return regres, err
	}
	logging.Debug(fmt.Sprintf("response %+v", res))

	if err := regres.FillResponse(res); err != nil {
		err_line := fmt.Sprintf("failed to decode response %s", err.Error())
		logging.Error(err_line)
		return regres, errors.New(err_line)
	}
	//logging.Info(fmt.Sprintf("regres %+v", regres))

	if res.StatusCode != 200 {
		err_line := fmt.Sprintf("failed to fetch data %+v", regres)
		logging.Error(err_line)
		return regres, errors.New(err_line)
	}
	//logging.Info(fmt.Sprintf("regres content %+v", regres.Content))

	return regres, nil
}

func (cl *RegClient) GetRegIndex() (*RegResponse, error) {

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
		res, err := cl.WebRequest(url, ml)

		// unrecoverable error
		if err != nil {
			logging.Error(fmt.Sprintf("response is nil, %+v", err.Error()))
			return regres, err
		}

		logging.Debug(fmt.Sprintf("status %d, response headers %+v", res.StatusCode, res.Header))

		if err := regres.FillResponse(res); err != nil {
			err_line := fmt.Sprintf("failed to decode response %s", err.Error())
			logging.Error(err_line)
			return regres, errors.New(err_line)
		}

		if res.StatusCode != 200 && res.StatusCode != 400 && res.StatusCode != 404 && res.StatusCode != 401 {
			err_line := fmt.Sprintf("failed to fetch data %s", res.Status)
			logging.Error(err_line)
			return regres, errors.New(err_line)
		}
		if res.StatusCode != 200 {
			logging.Debug(fmt.Sprintf("failed to fetch data: %s", res.Status))
			continue
		}

		break
	}
	return regres, nil
}
