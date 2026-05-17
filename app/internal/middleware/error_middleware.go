package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
)

const (
	CodeBadRequest   = "BAD_REQUEST"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeInternal     = "INTERNAL_SERVER_ERROR"

	MessageInternalServerError = "internal server error"
)

type AppError struct {
	Code    string
	Message string
	Status  int
	Err     error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		Code:    CodeBadRequest,
		Message: message,
		Status:  http.StatusBadRequest,
		Err:     err,
	}
}

func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		Code:    CodeUnauthorized,
		Message: message,
		Status:  http.StatusUnauthorized,
		Err:     err,
	}
}

func NewForbiddenError(message string, err error) *AppError {
	return &AppError{
		Code:    CodeForbidden,
		Message: message,
		Status:  http.StatusForbidden,
		Err:     err,
	}
}

func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: message,
		Status:  http.StatusNotFound,
		Err:     err,
	}
}

func NewConflictError(message string, err error) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: message,
		Status:  http.StatusConflict,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: message,
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}

func WrapErrorHandler(next ErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			WriteAppError(w, r, err)
		}
	}
}

func ErrorMiddleware(next http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("panic recovered: %v", rec)

				logger.ErrorContext(
					r.Context(),
					"panic recovered",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", r.URL.RawQuery),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("request_id", requestIDFrom(r)),
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)

				WriteAppErrorWithLogger(rw, r, NewInternalError(MessageInternalServerError, err), logger)
				return
			}

			if rw.status >= http.StatusBadRequest {
				logger.ErrorContext(
					r.Context(),
					"http request completed with error status",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", r.URL.RawQuery),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("request_id", requestIDFrom(r)),
					slog.Int("status", rw.status),
				)
			}
		}()

		next.ServeHTTP(rw, r)
	})
}

func WriteAppError(w http.ResponseWriter, r *http.Request, err error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	WriteAppErrorWithLogger(w, r, err, logger)
}

func WriteAppErrorWithLogger(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	appErr := normalizeError(err)

	logger.ErrorContext(
		contextOrBackground(r),
		"request failed",
		slog.String("method", requestMethod(r)),
		slog.String("path", requestPath(r)),
		slog.String("query", requestQuery(r)),
		slog.String("remote_addr", requestRemoteAddr(r)),
		slog.String("request_id", requestIDFrom(r)),
		slog.Int("status", appErr.Status),
		slog.String("error_code", appErr.Code),
		slog.String("message", appErr.Message),
		slog.Any("cause", appErr.Err),
	)

	writeJSONError(w, appErr.Status, appErr.Code, appErr.Message)
}

func normalizeError(err error) *AppError {
	if err == nil {
		return NewInternalError(MessageInternalServerError, nil)
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return NewInternalError(MessageInternalServerError, err)
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	payload := errorEnvelope{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(
			w,
			fmt.Sprintf(`{"error":{"code":"%s","message":"%s"}}`, CodeInternal, MessageInternalServerError),
			http.StatusInternalServerError,
		)
	}
}

func requestIDFrom(r *http.Request) string {
	if r == nil {
		return ""
	}

	return r.Header.Get("X-Request-Id")
}

func requestMethod(r *http.Request) string {
	if r == nil {
		return ""
	}

	return r.Method
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}

	return r.URL.Path
}

func requestQuery(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}

	return r.URL.RawQuery
}

func requestRemoteAddr(r *http.Request) string {
	if r == nil {
		return ""
	}

	return r.RemoteAddr
}

func contextOrBackground(r *http.Request) context.Context {
	if r == nil {
		return context.Background()
	}

	return r.Context()
}
