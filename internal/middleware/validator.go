package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	validatorOnce sync.Once
	v             *validator.Validate
)

// Validator returns the lazily-initialised singleton validator instance.
// It is safe to call concurrently. The phone_cn custom validator is
// registered on first call.
func Validator() *validator.Validate {
	validatorOnce.Do(func() {
		v = validator.New()
		if err := v.RegisterValidation("phone_cn", validatePhoneCN); err != nil {
			panic("register phone_cn validator: " + err.Error())
		}
	})
	return v
}

// BindAndValidate binds the JSON request body to req, then validates it
// using the validator package. On failure it aborts with 400 (code=10001)
// and a description of the validation error. Returns true on success.
func BindAndValidate(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":     apperrors.ErrInvalidParam,
			"msg":      "invalid JSON: " + err.Error(),
			"trace_id": log.TraceID(c.Request.Context()),
		})
		return false
	}
	if err := Validator().Struct(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":     apperrors.ErrInvalidParam,
			"msg":      formatValidationErrors(err),
			"trace_id": log.TraceID(c.Request.Context()),
		})
		return false
	}
	return true
}

// BindAndValidateQuery binds query string params to req, then validates it.
func BindAndValidateQuery(c *gin.Context, req any) bool {
	if err := c.ShouldBindQuery(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":     apperrors.ErrInvalidParam,
			"msg":      "invalid query: " + err.Error(),
			"trace_id": log.TraceID(c.Request.Context()),
		})
		return false
	}
	if err := Validator().Struct(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":     apperrors.ErrInvalidParam,
			"msg":      formatValidationErrors(err),
			"trace_id": log.TraceID(c.Request.Context()),
		})
		return false
	}
	return true
}

// formatValidationErrors flattens validator.ValidationErrors into a single
// human-readable string like "phone: required, code: len".
func formatValidationErrors(err error) string {
	if verrs, ok := err.(validator.ValidationErrors); ok && len(verrs) > 0 {
		parts := make([]string, 0, len(verrs))
		for _, fe := range verrs {
			parts = append(parts, fe.Field()+": "+fe.Tag())
		}
		return strings.Join(parts, ", ")
	}
	return err.Error()
}

// phoneCNRegex matches Chinese mainland mobile numbers: 11 digits starting
// with 1 and a non-zero second digit (1[3-9]xxxxxxxxx).
var phoneCNRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// validatePhoneCN is the registered validator func for the "phone_cn" tag.
func validatePhoneCN(fl validator.FieldLevel) bool {
	phone, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return phoneCNRegex.MatchString(phone)
}
