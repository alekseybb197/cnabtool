package content

import (
	"cnabtool/pkg/client"
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Config data.Config

// get manifest by reference

func (cc *Config) GetManifest(reference string) {

	config := (*client.Config)(cc)
	//config.Init()

	// make client
	cl := client.RegClient{
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

	// parse argument
	err := client.ParseReference(&cl)
	if err != nil {
		logging.Fatal("get content", fmt.Sprintf("%+v", err))
	}
	logging.Debug("GetManifest", fmt.Sprintf("client %+v", cl))

	// do request
	regres := config.GetRegIndex(&cl)
	if regres != nil {
		logging.Debug("GetManifest", fmt.Sprintf("response content %+v", regres))
	}
	ResponsePrettyPrint(regres)
}

// pretty print RegResponse

func ResponsePrettyPrint(regres *client.RegResponse) {

	//logging.Info("ResponsePrettyPrint", fmt.Sprintf("content1 %+v", regres))

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

		logging.Info("ResponsePrettyPrint", fmt.Sprintf("content %+v", cleanout))
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
