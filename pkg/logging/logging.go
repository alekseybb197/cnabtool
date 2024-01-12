/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package logging

import (
	"bytes"
	"cnabtool/pkg/data"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"runtime"
	"strings"
)

const (
	LogQuietLevel  = 0 // no messages, no errors, only return code
	LogErrorLevel  = 1 // no messages, only errors
	LogNormalLevel = 2 // messages + errors
	LogInfoLevel   = 3 // info messages + errors
	LogDebugLevel  = 4 // debug messages + errors
)

func maskcredentials(mess string) string {
	res := mess
	for _, secret := range data.Sensitives {
		s := strings.ReplaceAll(res, secret, "*******")
		res = s
	}
	return res
}

func Error(mess string) {
	data.Gc.Error++
	pc, file, lineNo, ok := runtime.Caller(1)
	point := "n/a"
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		fileName := path.Base(file)
		point = fmt.Sprintf("%s - %s - %d", fileName, funcName, lineNo)
	}
	if data.Gc.Verbosity >= LogErrorLevel {
		log.Printf("%s Error >> %s\n", point, maskcredentials(mess))
	}
}

func Fatal(mess string) {
	pc, file, lineNo, ok := runtime.Caller(1)
	point := "n/a"
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		fileName := path.Base(file)
		point = fmt.Sprintf("%s - %s - %d", fileName, funcName, lineNo)
	}
	if data.Gc.Verbosity >= LogErrorLevel {
		log.Fatalf("%s Error >> %s\n", point, maskcredentials(mess))
	}
}

func Message(mess string) {
	if data.Gc.Verbosity >= LogNormalLevel {
		log.Printf(">> %s\n", maskcredentials(mess))
	}
}

func Info(mess string) {
	if data.Gc.Verbosity >= LogInfoLevel {
		log.Printf("Info >> %s\n", maskcredentials(mess))
	}
}

func Debug(mess string) {
	pc, file, lineNo, ok := runtime.Caller(1)
	point := "n/a"
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		fileName := path.Base(file)
		point = fmt.Sprintf("%s - %s - %d", fileName, funcName, lineNo)
	}

	if data.Gc.Verbosity >= LogDebugLevel {
		log.Printf("Debug: %s >> %s\n", point, maskcredentials(mess))
	}
}

// PrettyString convert json in string to pretty format

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "  "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
