package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Validator 自定义验证器结构体
type Validator struct {
	validator *validator.Validate
}

// NewValidator 创建新的验证器实例
func NewValidator() *Validator {
	v := validator.New()

	// 注册自定义验证函数
	v.RegisterValidation("custom", validateCustom)

	// 自定义字段名获取函数
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validator: v}
}

// Validate 实现 echo.Validator 接口
func (cv *Validator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return formatValidationError(err)
	}
	return nil
}

// ValidateStruct 验证结构体
func (cv *Validator) ValidateStruct(s interface{}) error {
	if err := cv.validator.Struct(s); err != nil {
		return formatValidationError(err)
	}
	return nil
}

// formatValidationError 格式化验证错误信息
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errorMsgs []string
		for _, e := range validationErrors {
			errorMsgs = append(errorMsgs, formatFieldError(e))
		}
		return fmt.Errorf("%s", strings.Join(errorMsgs, "; "))
	}
	return err
}

// formatFieldError 格式化单个字段错误
func formatFieldError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, e.Param())
	case "custom":
		return fmt.Sprintf("%s format is invalid", field)
	default:
		return fmt.Sprintf("%s validation failed", field)
	}
}

// validateCustom 自定义验证函数
func validateCustom(fl validator.FieldLevel) bool {
	// 这里可以添加自定义验证逻辑
	// 例如：密码强度验证、用户名格式验证等
	return true
}

// GetValidator 从 Echo Context 中获取验证器
func GetValidator(c echo.Context) *Validator {
	if validator, ok := c.Get("validator").(*Validator); ok {
		return validator
	}
	return nil
}
