package minjson

import (
	"unsafe"
	"reflect"
	"strings"
	"unicode"
)

var bytesType = reflect.TypeOf([]byte(nil))

type Encoder interface {
	Encode(space []byte, ptr unsafe.Pointer) []byte
}

func EncoderOf(valType reflect.Type) Encoder {
	return encoderOf("", valType)
}

func encoderOf(prefix string, valType reflect.Type) Encoder {
	if bytesType == valType {
		return &bytesEncoder{}
	}
	switch valType.Kind() {
	case reflect.Int8:
		return &int8Encoder{}
	case reflect.Uint8:
		return &uint8Encoder{}
	case reflect.Int16:
		return &int16Encoder{}
	case reflect.Uint16:
		return &uint16Encoder{}
	case reflect.Int32:
		return &int32Encoder{}
	case reflect.Uint32:
		return &uint32Encoder{}
	case reflect.Int64, reflect.Int:
		return &int64Encoder{}
	case reflect.Uint64, reflect.Uint:
		return &uint64Encoder{}
	case reflect.Float64:
		return &lossyFloat64Encoder{}
	case reflect.Float32:
		return &lossyFloat32Encoder{}
	case reflect.String:
		return &stringEncoder{}
	case reflect.Ptr:
		elemEncoder := encoderOf(prefix + " [ptrElem]", valType.Elem())
		return &pointerEncoder{elemEncoder:elemEncoder}
	case reflect.Slice:
		elemEncoder := encoderOf(prefix + " [sliceElem]", valType.Elem())
		return &sliceEncoder{
			elemEncoder: elemEncoder,
			elemSize: valType.Elem().Size(),
		}
	case reflect.Array:
		elemEncoder := encoderOf(prefix + " [sliceElem]", valType.Elem())
		return &arrayEncoder{
			elemEncoder: elemEncoder,
			elemSize: valType.Elem().Size(),
			length: valType.Len(),
		}
	case reflect.Struct:
		var fields []structEncoderField
		for i := 0; i < valType.NumField(); i++ {
			field := valType.Field(i)
			name := getFieldName(field)
			if name == "" {
				continue
			}
			prefix := ""
			if len(fields) != 0 {
				prefix += ","
			}
			prefix += `"`
			prefix += name
			prefix += `":`
			fields = append(fields, structEncoderField{
				offset: field.Offset,
				prefix: prefix,
				encoder: encoderOf(prefix + " ." + name, field.Type),
			})
		}
		return &structEncoder{
			fields: fields,
		}
	}
	return nil
}

func getFieldName(field reflect.StructField) string {
	if !unicode.IsUpper(rune(field.Name[0])) {
		return ""
	}
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return field.Name
	}
	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return ""
	}
	if parts[0] == "" {
		return field.Name
	}
	return parts[0]
}

func PtrOf(val interface{}) unsafe.Pointer {
	return (*emptyInterface)(unsafe.Pointer(&val)).word
}

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  unsafe.Pointer
	word unsafe.Pointer
}