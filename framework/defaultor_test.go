package framework_test

import (
	"testing"

	"github.com/jcdomt/yuhuo/framework"
)

type config struct {
	Name string  `json:"name" validate:"max=10" default:"guest"`
	Age  int     `json:"age" validate:"min=1" default:"18"`
	VIP  bool    `json:"vip" default:"true"`
	Rate float64 `json:"rate" default:"1.5"`
	Nick string  `json:"nick" validate:"required" default:"should-not-apply"`
}

func TestSetDefaultsFillsMissingFields(t *testing.T) {
	var c config
	if err := framework.SetDefaults(&c); err != nil {
		t.Fatal(err)
	}

	if c.Name != "guest" || c.Age != 18 || c.VIP != true || c.Rate != 1.5 {
		t.Fatalf("默认值未注入: %+v", c)
	}
	if c.Nick != "" {
		t.Fatalf("required 字段不应注入默认值: %+v", c)
	}
}

func TestSetDefaultsDoesNotOverrideProvidedValues(t *testing.T) {
	c := config{Name: "alice", Age: 30, Nick: "n"}
	if err := framework.SetDefaults(&c); err != nil {
		t.Fatal(err)
	}

	if c.Name != "alice" || c.Age != 30 {
		t.Fatalf("已输入的值不应被覆盖: %+v", c)
	}
}

func TestSetDefaultsPointerField(t *testing.T) {
	type ptrConfig struct {
		Count *int `json:"count" default:"7"`
	}
	var p ptrConfig
	if err := framework.SetDefaults(&p); err != nil {
		t.Fatal(err)
	}

	if p.Count == nil || *p.Count != 7 {
		t.Fatalf("指针字段默认值未注入: %+v", p.Count)
	}
}

func TestSetDefaultsNestedStruct(t *testing.T) {
	type inner struct {
		Value int `json:"value" default:"9"`
	}
	type outer struct {
		Inner inner `json:"inner"`
	}

	var o outer
	if err := framework.SetDefaults(&o); err != nil {
		t.Fatal(err)
	}
	if o.Inner.Value != 9 {
		t.Fatalf("嵌套结构体默认值未注入: %+v", o.Inner)
	}
}

func TestSetDefaultsInvalidValueReturnsError(t *testing.T) {
	type badConfig struct {
		Age int `json:"age" default:"not-a-number"`
	}
	var b badConfig
	if err := framework.SetDefaults(&b); err == nil {
		t.Fatal("非法默认值应返回错误")
	}
}

func TestSetDefaultsRejectsNonPointer(t *testing.T) {
	if err := framework.SetDefaults(config{}); err == nil {
		t.Fatal("非指针目标应返回错误")
	}
}
