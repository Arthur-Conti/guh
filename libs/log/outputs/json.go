package outputs

import (
	"encoding/json"
	"fmt"
	"os"

	loglevels "github.com/Arthur-Conti/guh/libs/log/log_levels"
)

type JsonMessage struct {
	LogLevel loglevels.LogLevel
	Message  string
}

type JsonOutput struct {
	file string
}

func NewJsonOutput(filePath, fileName string) *JsonOutput {
	return &JsonOutput{
		file: filePath + "/" + fileName,
	}
}

func (jo *JsonOutput) Log(applicationPackage string, level loglevels.LogLevel, message string) {
	jsonMessage := JsonMessage{
		LogLevel: level,
		Message:  applicationPackage + message,
	}
	if err := jsonEncoder(jsonMessage, jo.file); err != nil {
		panic(err)
	}
}

func (jo *JsonOutput) Logf(applicationPackage string, level loglevels.LogLevel, message string, vals ...any) {
	jsonMessage := JsonMessage{
		LogLevel: level,
		Message:  fmt.Sprintf(applicationPackage+message, vals...),
	}
	if err := jsonEncoder(jsonMessage, jo.file); err != nil {
		panic(err)
	}
}

func jsonEncoder(message JsonMessage, file string) error {
	var logs []JsonMessage

	data, err := os.ReadFile(file)
	if err == nil {
		if len(data) > 0 {
			var log JsonMessage
			err = json.Unmarshal(data, &logs)
			if err != nil {
				err = json.Unmarshal(data, &log)
				if err != nil {
					return err
				}
				logs = append(logs, log)
			}
		}
	}
	logs = append(logs, message)

	newFile, err := os.Create(file)
	if err != nil {
		return err
	}
	defer newFile.Close()

	encoder := json.NewEncoder(newFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(logs); err != nil {
		return err
	}
	return nil
}
