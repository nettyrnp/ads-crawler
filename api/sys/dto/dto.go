package dto

import (
	"net/http"
)

type ServiceResponse struct {
	Body   interface{} `json:"body"`
	Status struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"status" `
}

func NewServiceResponse() *ServiceResponse {
	s := &ServiceResponse{}
	s.Status.Code = http.StatusOK
	return s
}

type ResetPasswordRequest struct {
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}
