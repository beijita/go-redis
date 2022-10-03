package config

import (
	"bufio"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type ServerProperties struct {
	Bind           string   `config:"bind"`
	Port           int      `config:"port"`
	AppendOnly     bool     `config:"appendOnly"`
	AppendFilename string   `config:"appendFilename"`
	MaxClients     int      `config:"maxClients"`
	RequirePass    string   `config:"requirePass"`
	DataBases      int      `config:"dataBases"`
	RDBFilename    string   `config:"dbFilename"`
	Peers          []string `config:"peers"`
	Self           string   `config:"self"`
}

var Properties *ServerProperties

func init() {
	Properties = &ServerProperties{
		Bind:       "127.0.0.1",
		Port:       6379,
		AppendOnly: false,
	}
}

func SetupConfig(configFilename string) {
	configFile, err := os.Open(configFilename)
	if err != nil {
		panic(err)
	}
	defer configFile.Close()
	Properties = parseConfig(configFile)
}

func parseConfig(src io.Reader) *ServerProperties {
	config := &ServerProperties{}

	rawMap := make(map[string]string)
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		pivot := strings.IndexAny(line, " ")
		if pivot > 0 && pivot < len(line)-1 {
			key := line[0:pivot]
			value := strings.Trim(line[pivot+1:], " ")
			rawMap[strings.ToLower(key)] = value
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln(" scanner.Err ", err)
	}

	t := reflect.TypeOf(config)
	v := reflect.ValueOf(config)
	n := t.Elem().NumField()
	for i := 0; i < n; i++ {
		fieldType := t.Elem().Field(i)
		fieldValue := v.Elem().Field(i)
		key, ok := fieldType.Tag.Lookup("config")
		if !ok {
			key = fieldType.Name
		}
		value, ok := rawMap[strings.ToLower(key)]
		if !ok {
			continue
		}
		switch fieldType.Type.Kind() {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int:
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				fieldValue.SetInt(intValue)
			}
		case reflect.Bool:
			fieldValue.SetBool("yes" == value)
		case reflect.Slice:
			if fieldType.Type.Elem().Kind() == reflect.String {
				slice := strings.Split(value, ",")
				fieldValue.Set(reflect.ValueOf(slice))
			}
		}
	}
	return config
}
