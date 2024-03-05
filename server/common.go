package server

import (
	"context"
	"log"
	"net/http"
	"time"
)

type ContextKey string

const INIT_TIME ContextKey = "init_time"

func injectTime(r *http.Request) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, INIT_TIME, time.Now().UnixNano())
	return r.WithContext(ctx)
}

func printTime(r *http.Request) {
	ctx := r.Context()
	initTime := time.Unix(0, ctx.Value(INIT_TIME).(int64))
	log.Printf("%s - %s - %s", r.Method, r.RequestURI, time.Since(initTime))
}
