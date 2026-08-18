package context

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// defaultMultipartMemory 是解析多部分表单时驻留内存的最大字节数。
const defaultMultipartMemory = 32 << 20 // 32 MB

// ---------- 查询参数 ----------

// QueryMap 返回全部查询参数。
func (ctx *Context) QueryMap() map[string][]string {
	return ctx.queryParams()
}

// Query 获取指定名称的第一个查询参数。
func (ctx *Context) Query(name string) string {
	return ctx.queryParams().Get(name)
}

// GetQuery 获取查询参数，并返回该参数是否存在。
func (ctx *Context) GetQuery(name string) (string, bool) {
	values, ok := ctx.queryParams()[name]
	if !ok || len(values) == 0 {
		return "", false
	}
	return values[0], true
}

// DefaultQuery 获取查询参数，不存在或为空时返回默认值。
func (ctx *Context) DefaultQuery(name, defaultValue string) string {
	if value, ok := ctx.GetQuery(name); ok && value != "" {
		return value
	}
	return defaultValue
}

// QueryArray 获取指定名称查询参数的全部值。
func (ctx *Context) QueryArray(name string) []string {
	values, ok := ctx.queryParams()[name]
	if !ok {
		return nil
	}
	return values
}

// queryParams 惰性解析并缓存查询参数。
func (ctx *Context) queryParams() url.Values {
	if ctx.query == nil {
		ctx.query = ctx.request.URL.Query()
	}
	return ctx.query
}

// ---------- 表单参数 ----------

// FormMap 返回全部 POST 表单参数。
func (ctx *Context) FormMap() map[string][]string {
	ctx.initPostForm()
	return ctx.request.PostForm
}

// Form 获取表单参数，优先取 POST body 中的值，其次查询参数。
func (ctx *Context) Form(name string) string {
	ctx.initPostForm()
	return ctx.request.Form.Get(name)
}

// PostForm 获取 POST body 中的表单参数。
func (ctx *Context) PostForm(name string) string {
	ctx.initPostForm()
	return ctx.request.PostForm.Get(name)
}

// GetPostForm 获取 POST body 中的表单参数，并返回该参数是否存在。
func (ctx *Context) GetPostForm(name string) (string, bool) {
	ctx.initPostForm()
	values, ok := ctx.request.PostForm[name]
	if !ok || len(values) == 0 {
		return "", false
	}
	return values[0], true
}

// DefaultPostForm 获取 POST 表单参数，不存在或为空时返回默认值。
func (ctx *Context) DefaultPostForm(name, defaultValue string) string {
	if value, ok := ctx.GetPostForm(name); ok && value != "" {
		return value
	}
	return defaultValue
}

// PostFormArray 获取指定名称 POST 表单参数的全部值。
func (ctx *Context) PostFormArray(name string) []string {
	ctx.initPostForm()
	values, ok := ctx.request.PostForm[name]
	if !ok {
		return nil
	}
	return values
}

// initPostForm 确保表单数据已解析。
// ParseMultipartForm 会同时填充 urlencoded 与 multipart 两种表单数据。
func (ctx *Context) initPostForm() {
	if ctx.formParsed {
		return
	}
	ctx.formParsed = true
	_ = ctx.request.ParseMultipartForm(defaultMultipartMemory)
}

// ---------- 文件上传 ----------

// MultipartForm 返回已解析的多部分表单。
func (ctx *Context) MultipartForm() (*multipart.Form, error) {
	if err := ctx.request.ParseMultipartForm(defaultMultipartMemory); err != nil {
		return nil, err
	}
	return ctx.request.MultipartForm, nil
}

// FormFile 获取指定名称的上传文件。
func (ctx *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.request.FormFile(name)
}

// SaveUploadedFile 将上传文件保存到目标路径。
func (ctx *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	if file == nil {
		return http.ErrMissingFile
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// ---------- 原始 Body 与绑定 ----------

// Body 返回原始请求体内容，多次调用返回相同结果。
func (ctx *Context) Body() []byte {
	return ctx.readBody()
}

// BodyString 以字符串形式返回原始请求体内容。
func (ctx *Context) BodyString() string {
	return string(ctx.readBody())
}

// BindJSON 将 JSON 请求体解码到目标结构体。
func (ctx *Context) BindJSON(obj interface{}) error {
	return json.NewDecoder(bytes.NewReader(ctx.readBody())).Decode(obj)
}

// readBody 读取并缓存请求体，同时恢复 Body 供表单解析等后续读取使用。
func (ctx *Context) readBody() []byte {
	if ctx.body != nil {
		return ctx.body
	}
	data, err := io.ReadAll(ctx.request.Body)
	if err != nil {
		return nil
	}
	ctx.body = data
	ctx.request.Body = io.NopCloser(bytes.NewReader(data))
	return data
}

// BindQuery 将 URL 查询参数绑定到目标结构体。
// 字段通过 form 或 query 标签指定参数名，缺省使用字段名；标签值为 "-" 时跳过该字段。
// 支持基本类型、指针以及基本类型切片（多值参数，如 ?tag=a&tag=b）。
func (ctx *Context) BindQuery(obj interface{}) error {
	value := reflect.ValueOf(obj)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("context: BindQuery 目标必须是非空结构体指针")
	}
	return bindValues(value.Elem(), ctx.queryParams())
}

// bindValues 将 url.Values 绑定到结构体。
func bindValues(v reflect.Value, values url.Values) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("context: 绑定目标必须是结构体")
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := typ.Field(i)
		if fieldType.PkgPath != "" { // 跳过未导出字段
			continue
		}

		name := bindFieldName(fieldType)
		if name == "-" {
			continue
		}

		raw, ok := values[name]
		if !ok || len(raw) == 0 {
			continue
		}
		if err := setBindField(field, raw); err != nil {
			return fmt.Errorf("context: 参数 %s 绑定失败: %w", name, err)
		}
	}
	return nil
}

// bindFieldName 解析字段对应的参数名，依次尝试 form、query 标签，缺省用字段名。
func bindFieldName(field reflect.StructField) string {
	for _, tag := range []string{"form", "query"} {
		if name := field.Tag.Get(tag); name != "" {
			return strings.SplitN(name, ",", 2)[0]
		}
	}
	return field.Name
}

// setBindField 将参数值写入字段，切片类型按多值处理，其余取第一个值。
func setBindField(field reflect.Value, raw []string) error {
	if field.Kind() == reflect.Slice {
		return setBindSlice(field, raw)
	}
	return setBindScalar(field, raw[0])
}

// setBindSlice 将多值参数写入基本类型切片。
func setBindSlice(field reflect.Value, raw []string) error {
	elemType := field.Type().Elem()
	slice := reflect.MakeSlice(field.Type(), 0, len(raw))
	for _, item := range raw {
		elem := reflect.New(elemType).Elem()
		if err := setBindScalar(elem, item); err != nil {
			return err
		}
		slice = reflect.Append(slice, elem)
	}
	field.Set(slice)
	return nil
}

// setBindScalar 将字符串按字段基本类型转换并写入，支持指针类型。
func setBindScalar(field reflect.Value, s string) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setBindScalar(field.Elem(), s)
	}
	if !field.CanSet() {
		return fmt.Errorf("字段不可设置")
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("不支持的绑定字段类型 %s", field.Kind())
	}
	return nil
}
