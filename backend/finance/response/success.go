package response

import "net/http"

const (
	RES_SUCC_OK_CODE = 100000
	RES_SUCC_OK_KEY  = "res_succ_ok"
)

var RES_SUCC_OK = ResponseTemplate{
	Success: true, StatusCode: http.StatusOK, Code: RES_SUCC_OK_CODE, Key: RES_SUCC_OK_KEY,
}
