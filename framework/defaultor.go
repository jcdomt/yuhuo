package framework

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// SetDefaults 为结构体中带 default 标签的字段注入默认值。
// 注入条件：validate 标签中不含 required、字段当前为零值（用户未输入）、default 值非空。
// 支持基本类型及其指针，并递归处理嵌套结构体。
//
// 示例：
//
//	type Request struct {
//		Name string `json:"name" validate:"max=10" default:"guest"`
//		Age  int    `json:"age" validate:"min=1" default:"18"`
//	}
func SetDefaults(obj interface{}) error {
	value := reflect.ValueOf(obj)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("framework: SetDefaults 目标必须是非空结构体指针")
	}
	return setDefaults(value.Elem())
}

func setDefaults(value reflect.Value) error {
	if value.Kind() != reflect.Struct {
		return nil
	}

	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typ.Field(i)

		if fieldType.PkgPath != "" { // 跳过未导出字段
			continue
		}

		// 匿名嵌入结构体：递归处理其字段
		if fieldType.Anonymous {
			if err := setDefaultsField(field); err != nil {
				return err
			}
			continue
		}

		defaultValue, hasDefault := fieldType.Tag.Lookup("default")
		if hasDefault && defaultValue != "" &&
			!isRequired(fieldType.Tag.Get("validate")) &&
			field.IsZero() {
			if err := setField(field, defaultValue); err != nil {
				return fmt.Errorf("framework: 字段 %s 默认值注入失败: %w", fieldType.Name, err)
			}
			continue
		}

		// 未注入默认值，仍递归处理嵌套结构体
		if err := setDefaultsField(field); err != nil {
			return err
		}
	}
	return nil
}

// setDefaultsField 递归进入嵌套结构体字段（含指针）。
func setDefaultsField(field reflect.Value) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil
		}
		return setDefaults(field.Elem())
	}
	return setDefaults(field)
}

// isRequired 判断 validate 标签中是否包含 required 规则。
func isRequired(validateTag string) bool {
	for _, rule := range strings.Split(validateTag, ",") {
		if strings.TrimSpace(rule) == "required" {
			return true
		}
	}
	return false
}

// setField 将字符串默认值写入字段，支持指针类型。
func setField(field reflect.Value, value string) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setBasic(field.Elem(), value)
	}
	return setBasic(field, value)
}

// setBasic 将字符串默认值按字段基本类型转换并写入。
func setBasic(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("字段不可设置")
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("不支持的默认值字段类型 %s", field.Kind())
	}
	return nil
}
