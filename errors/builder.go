package errors

import "fmt"

type errBuilder struct {
	code string
	msg  string
	err  error
	op   string
	typ  string
}

// Create a new error builder
// Example:
//
//	err := errors.B().
//		Code("code").
//		Msg("msg").
//		Err(err).
//		Op("op").
//		Build()
func B() *errBuilder {
	return &errBuilder{}
}

func Bf(format string, args ...interface{}) *errBuilder {
	return &errBuilder{msg: fmt.Sprintf(format, args...)}
}

func (b *errBuilder) Code(code string) *errBuilder {
	b.code = code
	return b
}

func (b *errBuilder) Msg(msg string) *errBuilder {
	b.msg = msg
	return b
}

func (b *errBuilder) Msgf(format string, args ...interface{}) *errBuilder {
	b.msg = fmt.Sprintf(format, args...)
	return b
}

func (b *errBuilder) Err(err error) *errBuilder {
	b.err = err
	return b
}

func (b *errBuilder) Op(op string) *errBuilder {
	b.op = op
	return b
}

func (b *errBuilder) Type(t string) *errBuilder {
	b.typ = t
	return b
}

func (b *errBuilder) Build() error {
	return &appError{
		Code: b.code,
		Msg:  b.msg,
		Err:  b.err,
		Op:   b.op,
		Type: b.typ,
	}
}
