package util

import "strings"

type Option struct {
	options map[string]string
}

func (op Option) Get(key string) string {
	return op.options[key]
}

func (op Option) GetWithDefault(key string, defaultValue string) string {
	v := op.options[key]
	if v == "" {
		return defaultValue
	}

	return v
}

func (op Option) Set(key, value string) {
	op.options[key] = value
}

func ParseOption(fields string) Option {
	ss := strings.Split(fields, " ")

	var op = Option{
		options: make(map[string]string),
	}
	for _, ss := range ss {
		kv := strings.SplitN(ss, "=", 2)
		if len(kv) == 2 {
			op.options[kv[0]] = kv[1]
		}
	}

	return op
}
