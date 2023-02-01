// Package middleware hold all middleware related such as HTTP middleware, etc
package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/model"
)

// RequestID generate ID and set it to context and response header
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := helper.GenerateID()
			ctx := setRequestIDToContext(c.Request().Context(), id)
			c.SetRequest(c.Request().WithContext(ctx))

			req := c.Request()
			res := c.Response()
			rid := req.Header.Get(echo.HeaderXRequestID)
			if rid == "" {
				rid = id
			}
			res.Header().Set(echo.HeaderXRequestID, rid)

			return next(c)
		}
	}
}

// can only set request ID from this middleware. Other can only read
func setRequestIDToContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, model.ReqIDCtxKey, id)
}
