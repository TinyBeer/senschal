package ilog

import (
	"context"
	"crypto/rand"
	"math/big"
)

const (
	traceIDTag  = "trace_id"
	letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	traceLen    = 16
)

type traceIDKey struct{}

// GenSimpleTraceID 单机使用 简单优先
func GenSimpleTraceID() string {
	b := make([]byte, traceLen)
	max := big.NewInt(int64(len(letterBytes)))
	for i := range b {
		n, _ := rand.Int(rand.Reader, max)
		b[i] = letterBytes[n.Int64()]
	}
	return string(b)
}

// NewTraceCtx 创建带traceId的ctx
func NewTraceCtx(ctx context.Context) context.Context {
	tid := GenSimpleTraceID()
	return context.WithValue(ctx, traceIDKey{}, tid)
}

// GetTraceID 从ctx读取traceId，无则返回空
func GetTraceID(ctx context.Context) string {
	val, ok := ctx.Value(traceIDKey{}).(string)
	if !ok {
		return ""
	}
	return val
}
