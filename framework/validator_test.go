package framework_test

import (
	"strings"
	"testing"

	"github.com/jcdomt/yuhuo/framework"
)

type sample struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"min=1"`
}

func TestDefaultValidatorValidate(t *testing.T) {
	errs := framework.DefaultValidator().Validate(sample{})

	joined := strings.Join(errs, "|")
	if !strings.Contains(joined, "name") || !strings.Contains(joined, "age") {
		t.Fatalf("错误信息应包含 json tag 字段名, got %q", errs)
	}
}

func TestValidatePasses(t *testing.T) {
	if errs := framework.DefaultValidator().Validate(sample{Name: "alice", Age: 18}); errs != nil {
		t.Fatalf("应通过校验, got %q", errs)
	}
}

func TestNewValidatorOverrideTranslation(t *testing.T) {
	v := framework.NewValidator(map[string]string{"required": "{0} 必填"})

	errs := v.Validate(sample{})
	joined := strings.Join(errs, "|")
	if !strings.Contains(joined, "name 必填") {
		t.Fatalf("自定义翻译未生效, got %q", errs)
	}
}

func TestValidatorsAreIndependent(t *testing.T) {
	custom := framework.NewValidator(map[string]string{"required": "{0} 必填"})

	customErrs := strings.Join(custom.Validate(sample{}), "|")
	defaultErrs := strings.Join(framework.DefaultValidator().Validate(sample{}), "|")

	if customErrs == defaultErrs {
		t.Fatal("自定义验证器与默认验证器应相互独立")
	}
}

func TestValidatorIsSafeForConcurrentUse(t *testing.T) {
	v := framework.NewValidator(nil)
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			_ = v.Validate(sample{Name: "x", Age: 1})
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
