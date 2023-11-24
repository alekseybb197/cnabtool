/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package logging

import (
	"bytes"
	"cnabtool/pkg/data"
	"encoding/json"
	"log"
)

const (
	LogQuietLevel  = 0 // no messages, no errors, only return code
	LogErrorLevel  = 1 // no messages, only errors
	LogNormalLevel = 2 // messages + errors
	LogInfoLevel   = 3 // info messages + errors
	LogDebugLevel  = 4 // debug messages + errors
)

func Error(point, mess string) {
	if data.Gc.Verbosity >= LogErrorLevel {
		log.Printf("%s >> %s\n", point, mess)
	}
}

func Fatal(point, mess string) {
	if data.Gc.Verbosity >= LogErrorLevel {
		log.Fatalf("%s >> %s\n", point, mess)
	}
}

func Message(point, mess string) {
	if data.Gc.Verbosity >= LogNormalLevel {
		log.Printf("%s >> %s\n", point, mess)
	}
}

func Info(point, mess string) {
	if data.Gc.Verbosity >= LogInfoLevel {
		log.Printf("%s >> %s\n", point, mess)
	}
}

func Debug(point, mess string) {
	if data.Gc.Verbosity >= LogDebugLevel {
		log.Printf("%s >> %s\n", point, mess)
	}
}

//

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "  "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
