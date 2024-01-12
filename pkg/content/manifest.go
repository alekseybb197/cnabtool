package content

import (
	"cnabtool/pkg/client"
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// type tricks

type Config data.Config

// get manifest by reference

func (cc *Config) GetManifest(reference string) (*client.RegResponse, *client.RegClient, error) {

	cl := client.NewRegClient((*client.Config)(cc), reference)

	// parse argument
	err := cl.ParseReference()
	if err != nil {
		err_line := fmt.Sprintf("invalid reference %+v", err)
		logging.Error(err_line)
		return nil, nil, errors.New(err_line)
	}
	logging.Debug(fmt.Sprintf("Client %+v", cl))

	// save current project root
	data.Repository = cl.Repository
	data.Registry = cl.Registry
	data.Scheme = cl.Scheme

	// do request
	regres, err := cl.GetRegIndex()
	//if regres != nil {
	//	logging.Debug(fmt.Sprintf("response content %+v", regres))
	//}
	return regres, cl, err
}

// pretty print RegResponse

func ResponsePrettyPrint(regres *client.RegResponse) {

	rout, err := json.Marshal(regres)
	if err == nil {
		// try to pretty print
		sout := string(rout)
		sout0 := strings.ReplaceAll(sout, "\\\\", "")
		sout1 := strings.ReplaceAll(sout0, "\\\"", "\"")
		sout2 := strings.ReplaceAll(sout1, "\"{", "{")
		sout3 := strings.ReplaceAll(sout2, "}\"", "}")
		sout4 := strings.ReplaceAll(sout3, "\\n", "")
		sout5 := strings.ReplaceAll(sout4, "\\r", "")
		dropunicode := regexp.MustCompile(`\\u....`)
		cleanout := dropunicode.ReplaceAllString(sout5, "")

		//logging.Debug(fmt.Sprintf("content %+v", cleanout))
		if jsonres, err := logging.PrettyString(cleanout); err == nil {
			fmt.Println(jsonres)
		} else {
			// try to make json by hand
			raw0 := regres.Content
			pretty, err := logging.PrettyString(raw0)
			if err != nil {
				pretty = regres.Content
			}
			raw1 := strings.Split(pretty, "\n")
			raw2 := strings.Join(raw1, "\n  ")
			//fmt.Println(raw2)
			fmt.Printf("{\n  \"Reference\": \"%s\",\n", regres.Reference)
			fmt.Printf("  \"Status\": %d,\n", regres.Status)
			fmt.Printf("  \"Filename\": \"%s\",\n", regres.Filename)
			fmt.Printf("  \"Length\": %d,", regres.Length)
			fmt.Printf("  \"Media\": \"%s\",\n", regres.Media)
			fmt.Printf("  \"Date\": \"%s\",\n", regres.Date)
			fmt.Printf("  \"Digest\": \"%s\",\n", regres.Digest)
			fmt.Printf("  \"Content\": %s\n}\n", raw2)
		}

	} else {
		// print as is
		fmt.Printf("%+v\n", regres)
	}
}
