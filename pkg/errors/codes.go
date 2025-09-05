package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

//	{
//		"type":   "/errors/schema-validation",
//		"title":  "Schema validation failed",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func ErrorSchemaValidation(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/schema-validation",
		"title":  "Schema validation failed",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/json-decode",
//		"title":  "Failed to decode JSON",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func ErrorJsonUnmarshal(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/json-decode",
		"title":  "Failed to decode JSON",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/yaml-decode",
//		"title":  "Failed to decode YAML",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func ErrorYamlUnmarshal(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/json-decode",
		"title":  "Failed to decode JSON",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/base64",
//		"title":  "Failed to decode/encode base64",
//		"status": http.StatusInternalServerError,
//		"error":  err.Error(),
//	}
func ErrorBase64(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/base64",
		"title":  "Failed to decode/encode base64",
		"status": http.StatusInternalServerError,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusInternalServerError, body
}

//	{
//		"type":   "/errors/database",
//		"title":  "Failed to insert into database",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func ErrorDatabaseInsert(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/database",
		"title":  "Failed to insert into database",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/database",
//		"title":  "Failed to insert into database",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func ErrorDatabaseUpdate(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/database",
		"title":  "Failed to update database",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/database",
//		"title":  "Failed to query the database",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func ErrorDatabaseSelect(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/database",
		"title":  "Failed to query the database",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/database",
//		"title":  "Record not found",
//		"status": http.StatusBadRequest,
//	}
func ErrorDatabaseRecordNotFound() (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/database",
		"title":  "Record not found",
		"status": http.StatusBadRequest,
	}
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/tracing",
//		"title":  "Failure creating trace",
//		"status": http.StatusInternalServerError,
//		"error":  err.Error(),
//	}
func ErrorTracing(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/tracing",
		"title":  "Failure creating trace",
		"status": http.StatusInternalServerError,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusInternalServerError, body
}

//	{
//		"type":   "/errors/teleport-api",
//		"title":  "Failure Teleport API",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func ErrorTeleportApi(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/teleport-api",
		"title":  "Failure Teleport API",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/middleware-auth",
//		"title":  "Required headers are missing",
//		"status": http.StatusUnauthorized,
//		"error":  err.Error(),
//	}
func ErrorAuthMissingAuthHeader(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/middleware-auth",
		"title":  "Required headers are missing",
		"status": http.StatusUnauthorized,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusUnauthorized, body
}

//	{
//		"type":   "/errors/middleware-auth",
//		"title":  "Failed to init OIDC provider",
//		"status": http.StatusUnauthorized,
//		"error":  err.Error(),
//	}
func ErrorAuthOidcInit(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/middleware-auth",
		"title":  "Failed to init OIDC provider",
		"status": http.StatusUnauthorized,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusUnauthorized, body
}

//	{
//		"type":   "/errors/middleware-auth",
//		"title":  "Invalid token",
//		"scope": scope,
//		"status": http.StatusUnauthorized,
//		"error":  err.Error(),
//	}
func ErrorAuthInvalidToken(err error, scope string) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/middleware-auth",
		"title":  "Invalid token",
		"scope":  scope,
		"status": http.StatusUnauthorized,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusUnauthorized, body
}

//	{
//		"type":   "/errors/middleware-auth",
//		"title":  "Token verification failed",
//		"scope":  scope,
//		"status": http.StatusUnauthorized,
//		"error":  err.Error(),
//	}
func ErrorAuthTokenVerificationFailed(err error, scope string) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/middleware-auth",
		"title":  "Token verification failed",
		"scope":  scope,
		"status": http.StatusUnauthorized,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusUnauthorized, body
}

//	{
//		"type":   "/errors/user-profile",
//		"title":  "Invalid user profile configuration",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func ErrorInvalidUserProfile(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/user-profile",
		"title":  "Invalid user profile configuration",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/providers",
//		"title":  "Failure calling access provider",
//		"status": http.StatusBadRequest,
//		"error":  err.Error(),
//	}
func AccessProviderCallFailed(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/providers",
		"title":  "Failure calling access provider",
		"status": http.StatusBadRequest,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusBadRequest, body
}

//	{
//		"type":   "/errors/providers",
//		"title":  "Partial failure calling access provider",
//		"status": http.StatusMultiStatus,
//		"error":  err.Error(),
//	}
func AccessProviderCallPartiallyFailed(err error) (code int, body gin.H) {
	body = gin.H{
		"type":   "/errors/providers",
		"title":  "Partial failure calling access provider",
		"status": http.StatusMultiStatus,
		"error":  err.Error(),
	}
	log.Error().Msg(fmt.Sprintf("%+v", body))
	return http.StatusMultiStatus, body
}

//	{
//		"type":   "/status/denied",
//		"title":  "You are not authorised to perform this action",
//		"status": http.StatusForbidden,
//	}
func StatusDenied() (code int, body gin.H) {
	body = gin.H{
		"type":   "/status/denied",
		"title":  "You are not authorised to perform this action",
		"status": http.StatusForbidden,
	}
	return http.StatusCreated, body
}

//	{
//		"type":   "/status/success",
//		"title":  "Record successfully created",
//		"status": http.StatusCreated,
//	}
func StatusCreated() (code int, body gin.H) {
	body = gin.H{
		"type":   "/status/success",
		"title":  "Record successfully created",
		"status": http.StatusCreated,
	}
	return http.StatusCreated, body
}

//	{
//		"type":   "/status/success",
//		"title":  "Record successfully updated",
//		"status": http.StatusCreated,
//	}
func StatusUpdated() (code int, body gin.H) {
	body = gin.H{
		"type":   "/status/success",
		"title":  "Record successfully updated",
		"status": http.StatusCreated,
	}
	return http.StatusCreated, body
}

//	{
//		"type":   "/status/success",
//		"title":  "Record successfully deleted",
//		"status": http.StatusOK,
//	}
func StatusDeleted() (code int, body gin.H) {
	body = gin.H{
		"type":   "/status/success",
		"title":  "Record successfully deleted",
		"status": http.StatusOK,
	}
	return http.StatusCreated, body
}

//	{
//		"type":   "/status/ok",
//		"title":  "Operation was successful",
//		"status": http.StatusOk,
//	}
func StatusOk() (code int, body gin.H) {
	body = gin.H{
		"type":   "/status/success",
		"title":  "Record successfully created",
		"status": http.StatusOK,
	}
	return http.StatusOK, body
}
