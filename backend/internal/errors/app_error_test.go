package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew_DefaultMsg(t *testing.T) {
	e := New(ErrInvalidParam, "")
	if e.Code != ErrInvalidParam {
		t.Errorf("Code = %d, want %d", e.Code, ErrInvalidParam)
	}
	if e.Msg != Msg(ErrInvalidParam) {
		t.Errorf("Msg = %q, want default %q", e.Msg, Msg(ErrInvalidParam))
	}
	if e.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %d, want %d", e.HTTPStatus, http.StatusBadRequest)
	}
	if e.Cause != nil {
		t.Errorf("Cause = %v, want nil", e.Cause)
	}
}

func TestNew_CustomMsg(t *testing.T) {
	e := New(ErrNotFound, "parcel not found")
	if e.Msg != "parcel not found" {
		t.Errorf("Msg = %q, want %q", e.Msg, "parcel not found")
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("db connection refused")
	e := Wrap(cause, ErrInternal, "query failed")
	if e.Code != ErrInternal {
		t.Errorf("Code = %d, want %d", e.Code, ErrInternal)
	}
	if e.Msg != "query failed" {
		t.Errorf("Msg = %q, want %q", e.Msg, "query failed")
	}
	if e.Cause != cause {
		t.Errorf("Cause = %v, want %v", e.Cause, cause)
	}
}

func TestWrap_DefaultMsg(t *testing.T) {
	cause := errors.New("boom")
	e := Wrap(cause, ErrInternal, "")
	if e.Msg != Msg(ErrInternal) {
		t.Errorf("Msg = %q, want default %q", e.Msg, Msg(ErrInternal))
	}
}

func TestError_String(t *testing.T) {
	e := New(ErrInvalidParam, "bad input")
	got := e.Error()
	want := "[code=10001] bad input"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_String_WithCause(t *testing.T) {
	cause := errors.New("connection reset")
	e := Wrap(cause, ErrInternal, "query failed")
	got := e.Error()
	want := "[code=10009] query failed: connection reset"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("io error")
	e := Wrap(cause, ErrInternal, "x")
	if !errors.Is(e, cause) {
		t.Errorf("errors.Is returned false; cause should be detectable")
	}

	var target *AppError
	if !errors.As(e, &target) {
		t.Errorf("errors.As returned false; should detect *AppError")
	}
	if target != e {
		t.Errorf("target = %p, want %p", target, e)
	}
}

func TestConstructors(t *testing.T) {
	cases := []struct {
		name   string
		err    *AppError
		code   int
		httpSt int
	}{
		{"InvalidParam", InvalidParam(""), ErrInvalidParam, http.StatusBadRequest},
		{"Unauthenticated", Unauthenticated(""), ErrUnauthenticated, http.StatusUnauthorized},
		{"Forbidden", Forbidden(""), ErrForbidden, http.StatusForbidden},
		{"NotFound", NotFound(""), ErrNotFound, http.StatusNotFound},
		{"Conflict", Conflict(""), ErrConflict, http.StatusConflict},
		{"Internal", Internal(""), ErrInternal, http.StatusInternalServerError},
		{"RateLimited", RateLimited(""), ErrRateLimited, http.StatusTooManyRequests},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.err.Code != c.code {
				t.Errorf("Code = %d, want %d", c.err.Code, c.code)
			}
			if c.err.HTTPStatus != c.httpSt {
				t.Errorf("HTTPStatus = %d, want %d", c.err.HTTPStatus, c.httpSt)
			}
			if c.err.Msg != Msg(c.code) {
				t.Errorf("Msg = %q, want default %q", c.err.Msg, Msg(c.code))
			}
		})
	}
}

func TestConstructors_CustomMsg(t *testing.T) {
	if got := InvalidParam("missing field").Msg; got != "missing field" {
		t.Errorf("InvalidParam custom msg = %q", got)
	}
}
