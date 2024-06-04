package generic_config

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"src/util"
)

/*
	This config loader doesn't use a specific fancy format like yaml or json. They annoy me.
	It's just my own simple format. The file extension is .whatever for that reason.
	It's basically this:

	<key> : <value>
	# comment

	any whitespace on either side of <key> and <value> is ignored.
	Empty lines are ignored.
	You can also have comments with #, but not on the same line as key-value pairs.

	The type that is associated with a value is based on the type defined in the struct that you use for the config.
	For example, the line "foo : 1" doesn't strictly hold an integer, your config struct could be { foo string } or { foo int }, and the file would be parsed accordingly.
*/

func LoadConfig[T any](config_filename string, config_ptr *T) error {
	config_bytes, err := os.ReadFile(config_filename)
	if err != nil {
		return err
	}

	var errs []error

	config_value := reflect.ValueOf(config_ptr).Elem()
	config_type  := config_value.Type()
	num_fields   := config_value.NumField()

	type ConfigValueInfo struct{
		index    int // index of the struct member
		line_num int // line in the config file where this value was given
	}
	name_info := make(map[string]*ConfigValueInfo)

	for i := 0; i < num_fields; i++ {
		field := config_type.Field(i)
		if field.IsExported() == false {
			return util.Err_fmt("Config contains unexported field: \"%v\", all fields of the config must be exported.", field.Name)
		}
		name_info[field.Name] = &ConfigValueInfo{i, 0}
	}

	iter := string(config_bytes)

	for line_num, lastline := 1, false; !lastline; line_num++ {

		var line string
		if line_end := strings.IndexByte(iter, '\n'); line_end != -1 {
			line = iter[:line_end]
			iter = iter[line_end+1:]
		} else {
			lastline = true
			line = iter
		}

		line = strings.TrimSpace(line)

		if line == "" { continue }
		if strings.HasPrefix(line, "#") { continue }

		colon_index := strings.IndexByte(line, ':')
		if colon_index == -1 {
			// Do not print values/names because malformed lines could  contain sensitive data
			return util.Err_fmt("%v:%v: Malformed file, missing ':'", config_filename, line_num)
		}

		name  := strings.TrimSpace(line[:colon_index])
		value := strings.TrimSpace(line[colon_index+1:])

		if value == "" {
			// Do not print values/names because malformed lines could  contain sensitive data
			return util.Err_fmt("%v:%v: Malformed file, key is missing a value", config_filename, line_num)
		}


		info, ok := name_info[name]

		// 1) all keys listed in config file must be members of the struct
		if !ok {
			util.Err_add(&errs, "Unknown key \"%v\" on line %v", name, line_num)
			continue
		}

		// 2) config file may only contain each key once
		if info.line_num != 0 {
			util.Err_add(&errs, "%v:%v: Duplicate key \"%v\", already on line %v", config_filename, line_num, name, info.line_num)
			continue
		}

		field := config_value.Field(info.index)
		info.line_num = line_num

		switch field.Kind() {
			case reflect.String:
				field.SetString(strings.Clone(value)) // cloned so that no slices of the complete config string are retained, allowing the GC to free it.

			case reflect.Int:
				if v, err := strconv.ParseInt(value, 0, 0); err != nil {
					util.Err_add(&errs, "%v:%v: Error parsing integer \"%v\"", config_filename, line_num, name)
				} else {
					field.SetInt(v)
				}

			case reflect.Float64:
				if v, err := strconv.ParseFloat(value, 64); err != nil {
					util.Err_add(&errs, "%v:%v: Error parsing float \"%v\"", config_filename, line_num, name)
				} else {
					field.SetFloat(v)
				}

			case reflect.Bool:
				if v, err := strconv.ParseBool(value); err != nil {
					util.Err_add(&errs, "%v:%v: Error parsing bool \"%v\"", config_filename, line_num, name)
				} else {
					field.SetBool(v)
				}

			default:
				util.Err_add(&errs, "Config: Cannot set \"%v\" because we do not handle type %v in configs yet", name, field.Kind())
		}
	}

	// 3) All members of the struct must be present in config file
	for name, info := range name_info {
		if info.line_num == 0 {
			util.Err_add(&errs, "Missing \"%v\"", name)
		}
	}

	if len(errs) > 0 {
		return util.Err_join(errs)
	} else {
		return nil
	}
}
