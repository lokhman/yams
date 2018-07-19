package yams

import (
	"database/sql"
	"net"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/elgris/sqrl"
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
		v.validate.RegisterTagNameFunc(func(field reflect.StructField) string {
			if v, ok := field.Tag.Lookup("form"); ok {
				return v
			}
			return field.Name
		})
		v.validate.RegisterValidation("trim", func(fl validator.FieldLevel) bool {
			if f := fl.Field(); f.Kind() == reflect.String {
				return f.String() == strings.TrimSpace(f.String())
			}
			return true
		})
		v.validate.RegisterValidation("prefix", func(fl validator.FieldLevel) bool {
			return strings.HasPrefix(fl.Field().String(), fl.Param())
		})
		v.validate.RegisterValidation("suffix", func(fl validator.FieldLevel) bool {
			return strings.HasSuffix(fl.Field().String(), fl.Param())
		})
		v.validate.RegisterValidation("host", func(fl validator.FieldLevel) bool {
			host := fl.Field().String()
			if !strings.Contains(host, ":") {
				host += ":80"
			}
			if _, err := net.ResolveTCPAddr("tcp", host); err != nil {
				return false
			}

			// allow HTTP and HTTPS by default
			host = strings.TrimSuffix(host, ":80")
			host = strings.TrimSuffix(host, ":443")

			if idField := reflect.Indirect(fl.Top()).FieldByName("id"); idField.IsValid() {
				q := QB.Select("TRUE").From("profiles").Where("? = ANY(hosts)", host)
				if id := idField.Int(); id != 0 {
					q.Where("id <> ?", id)
				}
				if err := q.Scan(new(bool)); err == nil {
					return false
				} else if err != sql.ErrNoRows {
					panic(err)
				}
			}
			return true
		})
		v.validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
			return regexp.MustCompile(`^[a-z][a-z0-9.]{2,31}$`).MatchString(fl.Field().String())
		})
		v.validate.RegisterValidation("role", func(fl validator.FieldLevel) bool {
			return InStringSlice(AnyRole, fl.Field().String())
		})
		v.validate.RegisterValidation("acl", func(fl validator.FieldLevel) bool {
			field := fl.Field()
			switch field.Kind() {
			case reflect.Slice, reflect.Array:
				ln := field.Len()
				if ln == 0 {
					return true
				}
				acl := make([]int64, ln)
				for i := 0; i < ln; i++ {
					v := field.Index(i).Int()
					if v == 0 {
						return false
					}
					acl[i] = v
				}
				var lnf int
				if err := QB.Select("COUNT(*)").From("profiles").Where(sqrl.Eq{"id": acl}).Scan(&lnf); err != nil {
					panic(err)
				}
				return ln == lnf
			default:
				return false
			}
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
