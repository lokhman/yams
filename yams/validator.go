package yams

import (
	"net"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v9"
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &defaultValidator{}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyinit()
		if err := v.validate.Struct(obj); err != nil {
			return error(err)
		}
	}
	return nil
}

func (v *defaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")
		v.validate.RegisterValidation("prefix", func(fl validator.FieldLevel) bool {
			return strings.HasPrefix(fl.Field().String(), fl.Param())
		})
		v.validate.RegisterValidation("suffix", func(fl validator.FieldLevel) bool {
			return strings.HasSuffix(fl.Field().String(), fl.Param())
		})
		v.validate.RegisterValidation("host", func(fl validator.FieldLevel) bool {
			if _, err := net.LookupHost(fl.Field().String()); err != nil {
				return false
			}
			return true
		})
		v.validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
			return regexp.MustCompile(`^[a-z][a-z0-9.]{2,31}$`).MatchString(fl.Field().String())
		})
		v.validate.RegisterValidation("acl", func(fl validator.FieldLevel) bool {
			return regexp.MustCompile(`^[a-z]+(?::\d+)?$`).MatchString(fl.Field().String())
		})
	})
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func init() {
	binding.Validator = &defaultValidator{}
}
