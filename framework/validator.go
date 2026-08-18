package framework

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// DefaultValidatorTransMap 是默认的中文错误消息模板，key 为校验 tag。
var DefaultValidatorTransMap = map[string]string{
	"min":      "{0} 长度必须至少为 {1} 个字符",
	"max":      "{0} 长度不能超过 {1} 个字符",
	"len":      "{0} 长度应该为 {1} 个字符",
	"required": "{0} 不能为空",
}

// Validator 封装验证器与中文翻译器。
// 惰性初始化、线程安全，可创建多个相互独立的实例。
type Validator struct {
	once     sync.Once
	validate *validator.Validate
	trans    ut.Translator
	transMap map[string]string
}

// NewValidator 创建验证器实例。transMap 用于覆盖默认翻译或新增翻译，key 为校验 tag。
func NewValidator(transMap map[string]string) *Validator {
	merged := make(map[string]string, len(DefaultValidatorTransMap)+len(transMap))
	for tag, tr := range DefaultValidatorTransMap {
		merged[tag] = tr
	}
	for tag, tr := range transMap {
		merged[tag] = tr
	}
	return &Validator{transMap: merged}
}

// defaultValidator 是框架默认验证器实例。
var defaultValidator = NewValidator(nil)

// DefaultValidator 返回框架默认的验证器实例。
func DefaultValidator() *Validator {
	return defaultValidator
}

// init 惰性初始化验证器与翻译器，仅执行一次。
func (v *Validator) init() {
	v.once.Do(func() {
		v.validate = validator.New()
		uni := ut.New(zh.New(), zh.New())
		v.trans, _ = uni.GetTranslator("zh")

		// 字段名优先使用 json tag，使错误信息对调用方更友好。
		v.validate.RegisterTagNameFunc(func(field reflect.StructField) string {
			name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
			if name == "" || name == "-" {
				return field.Name
			}
			return name
		})

		// 统一注册合并后的翻译，避免重复注册导致 panic。
		for tag, tr := range v.transMap {
			v.registerTranslation(tag, tr)
		}
	})
}

func (v *Validator) registerTranslation(tag, tr string) {
	_ = v.validate.RegisterTranslation(tag, v.trans, func(ut ut.Translator) error {
		return ut.Add(tag, tr, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field(), fe.Param())
		return t
	})
}

// Struct 返回原始校验错误，供需要自定义处理的场景使用。
func (v *Validator) Struct(obj interface{}) error {
	v.init()
	return v.validate.Struct(obj)
}

// Validate 校验结构体，返回翻译后的错误消息（通过校验时返回 nil）。
func (v *Validator) Validate(obj interface{}) []string {
	if err := v.Struct(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := make([]string, 0, len(validationErrors))
			for _, fe := range validationErrors {
				messages = append(messages, fe.Translate(v.trans))
			}
			return messages
		}
		return []string{err.Error()}
	}
	return nil
}

// BindJSONAndValidate 绑定 JSON 到 obj 并校验，返回翻译后的错误消息（通过时为 nil）。
// 请求体解析失败时返回解析错误信息。
func BindJSONAndValidate(ctx *requestcontext.Context, obj interface{}) []string {
	if err := ctx.BindJSON(obj); err != nil {
		return []string{"请求体解析失败：" + err.Error()}
	}
	return DefaultValidator().Validate(obj)
}
