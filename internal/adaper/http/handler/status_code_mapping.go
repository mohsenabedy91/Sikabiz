package handler

import (
	"net/http"
	"github.com/mohsenabedy91/Sikabiz/pkg/serviceerror"
)

// StatusCodeMapping maps error to http status code
var StatusCodeMapping = map[serviceerror.ErrorMessage]int{
	// General
	serviceerror.ServerError:        http.StatusInternalServerError,
	serviceerror.ServiceUnavailable: http.StatusServiceUnavailable,
	serviceerror.RecordNotFound:     http.StatusNotFound,
	serviceerror.PermissionDenied:   http.StatusForbidden,
	serviceerror.Unauthorized:       http.StatusUnauthorized,
	serviceerror.IsNotDeletable:     http.StatusForbidden,
	serviceerror.NoRowsEffected:     http.StatusNotFound,
	serviceerror.FailedSendEmail:    http.StatusInternalServerError,
	// User
	serviceerror.UserIsBanned:      http.StatusForbidden,
	serviceerror.UserInActive:      http.StatusForbidden,
	serviceerror.UserUnVerified:    http.StatusForbidden,
	serviceerror.EmailRegistered:   http.StatusConflict,
	serviceerror.CredentialInvalid: http.StatusUnauthorized,
	serviceerror.UserLogout:        http.StatusUnauthorized,
	// OTP
	serviceerror.InvalidOTP: http.StatusBadRequest,
	serviceerror.OTPExpired: http.StatusUnauthorized,
	// Token
	serviceerror.InvalidToken: http.StatusUnauthorized,
	serviceerror.TokenExpired: http.StatusUnauthorized,
	// Validation
	serviceerror.InvalidRequestBody: http.StatusBadRequest,
	// Role
	serviceerror.RoleExisted: http.StatusConflict,
	// TOTP
	serviceerror.InvalidTOTPCode: http.StatusUnauthorized,
	serviceerror.TOTPNotEnrolled: http.StatusBadRequest,
}
