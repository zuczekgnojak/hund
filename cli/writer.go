package cli

import (
	"hund/logger"
	"hund/util"
	"strings"
)

type CliWriter interface {
	Write(name string, value ...string) error
}

type DummyWriter struct{}

func NewDummyWriter() *DummyWriter {
	return &DummyWriter{}
}

func (self *DummyWriter) Write(name string, value ...string) error {
	logger.Debugf("writing %s as %v", name, value)
	return nil
}

type MapWriter struct {
	values    map[string]string
	flagValue string
}

func NewMapWriter(values map[string]string, flagValue string) *MapWriter {
	return &MapWriter{values, flagValue}
}

func (self *MapWriter) Write(name string, value ...string) error {
	logger.Debugf("writing param \"%s\" %v\n", name, value)
	if len(value) == 0 {
		self.values[name] = self.flagValue
		return nil
	}

	joinedValue := strings.Join(value, " ")
	self.values[name] = joinedValue
	return nil
}

type PointerWriter struct {
	flags  map[string]*bool
	values map[string]*string
}

func NewPointerWriter() *PointerWriter {
	flags := make(map[string]*bool)
	values := make(map[string]*string)
	return &PointerWriter{flags, values}
}

func (self *PointerWriter) Write(name string, value ...string) error {
	logger.Debugf("writing option \"%s\" %v\n", name, value)
	if len(value) == 0 {
		p, ok := self.flags[name]
		if !ok {
			return util.NewError("missing \"%s\" flag", name)
		}
		*p = true
		return nil
	}

	p, ok := self.values[name]
	if !ok {
		return util.NewError("missing \"%s\" value", name)
	}
	*p = strings.Join(value, " ")
	return nil
}

func (self *PointerWriter) AddFlag(name string, target *bool) error {
	_, ok := self.flags[name]
	if ok {
		return util.NewError("key \"%s\" already added", name)
	}

	self.flags[name] = target
	return nil
}

func (self *PointerWriter) AddValue(name string, target *string) error {
	_, ok := self.values[name]
	if ok {
		return util.NewError("key \"%s\" already added", name)
	}
	self.values[name] = target
	return nil
}
