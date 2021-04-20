package vibe

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

const (
	EXT_YAML = ".yml"
)

var v *Vibe

func init() {
	v = New()
}

type Vibe struct {
	keyDelimiter string
	// A set of paths to look for the config file in
	configFiles []string

	config map[string]interface{}
}

//New returns an initialized Vibe instance.
func New() *Vibe {

	v := new(Vibe)
	v.keyDelimiter = "."
	v.config = make(map[string]interface{})

	return v
}

//AddConfigFiles add config files.
func AddConfigFiles(in ...string) {
	v.AddConfigFiles(in...)
}

func (v *Vibe) AddConfigFiles(in ...string) {
	v.configFiles = append(v.configFiles, in...)
}

func (v *Vibe) ReadInConfig() error {
	for _, file := range v.configFiles {
		config := make(map[string]interface{})
		err := v.readInConfig(file, &config)
		if nil != err {
			panic(err)
		}
		desensitizeMap(config)
		mergeMaps(v.config, config)
	}
	return nil
}

func (v *Vibe) readInConfig(fileName string, config *map[string]interface{}) error {

	ext := filepath.Ext(fileName)

	if EXT_YAML != ext {
		return UnsupportedConfigError(ext)
	}

	file, err := os.ReadFile(fileName)
	if nil != err {
		return ConfigParseError{err}
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return ConfigParseError{err}
	}

	return nil
}

//Unmarshal unmarshal the config into a Struct.
//Make sure that the tags on the fields of the structure are properly set.
func Unmarshal(rawVal interface{}) error {
	return v.Unmarshal(rawVal)
}

func (v *Vibe) Unmarshal(rawVal interface{}) error {
	if err := decode(v.config, rawVal); err != nil {
		panic(err)
	}
	return nil
}

//AllSettings merges all settings and returns them as a map[string]interface{}.
func AllSettings() map[string]interface{} { return v.AllSettings() }

func (v *Vibe) AllSettings() map[string]interface{} {
	m := map[string]interface{}{}
	// start from the list of keys, and construct the map one value at a time
	for _, k := range v.AllKeys() {
		value := v.Get(k)
		if value == nil {
			// should not happen, since AllKeys() returns only keys holding a value,
			// check just in case anything changes
			continue
		}
		path := strings.Split(k, v.keyDelimiter)
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := deepSearch(m, path[0:len(path)-1])
		// set innermost value
		deepestMap[lastKey] = value
	}
	return m
}

//AllKeys all keys of config
func AllKeys() []string { return v.AllKeys() }

func (v *Vibe) AllKeys() []string {
	m := map[string]bool{}

	m = mergeKeys(m, v.config, "", v.keyDelimiter)

	a := make([]string, 0, len(m))
	for x := range m {
		a = append(a, x)
	}
	return a
}

//find Given a key, find the value.
func (v *Vibe) find(lowerKey string) interface{} {
	var (
		val  interface{}
		path = strings.Split(lowerKey, v.keyDelimiter)
	)

	path = strings.Split(lowerKey, v.keyDelimiter)

	// Set() override first
	val = searchMap(v.config, path)
	if val != nil {
		return val
	}

	val = searchMapWithPathPrefixes(v.config, path)
	if val != nil {
		return val
	}

	return nil
}

//Get can retrieve any value given the key to use.
func Get(key string) interface{} { return v.Get(key) }

func (v *Vibe) Get(key string) interface{} {
	lowerKey := strings.ToLower(key)
	value := v.find(lowerKey)
	if value == nil {
		return nil
	}
	return value
}
