package gohttp

type StatusCode int
type StatusMsg string

const (
	_ StatusCode = 0
	// base
	STATUS_CODE_SUCCESS        = 200
	STATUS_CODE_INVALID_PARAMS = 400
	STATUS_CODE_ERROR          = 500
	STATUS_CODE_FAILED         = 800

	// jwt token
	STATUS_CODE_AUTH_CHECK_TOKEN_EMPTY    = 801
	STATUS_CODE_AUTH_CHECK_TOKEN_FAILED    = 802
	STATUS_CODE_AUTH_CHECK_TOKEN_TIMEOUT   = 803
	STATUS_CODE_AUTH_TOKEN_GENERATE_FAILED = 804
	STATUS_CODE_AUTH_TYPE_ERROR                = 805
	// secret
	STATUS_CODE_SECRET_CHECK_FAILED    = 809

	// upload
	STATUS_CODE_UPLOAD_FILE_SAVE_FAILED        = 811
	STATUS_CODE_UPLOAD_FILE_CHECK_FAILED       = 812
	STATUS_CODE_UPLOAD_FILE_CHECK_FORMAT_WRONG = 813
)

const (
	_ StatusMsg = ""
	// base
	STATUS_MSG_SUCCESS        = "ok"
	STATUS_MSG_INVALID_PARAMS = "参数校验失败"
	STATUS_MSG_ERROR          = "system error"
	STATUS_MSG_FAILED         = "操作失败"

	// jwt token
	STATUS_MSG_AUTH_CHECK_TOKEN_EMPTY	  = "当前未登录，请先登录"
	STATUS_MSG_AUTH_CHECK_TOKEN_FAILED    = "用户信息校验失败"
	STATUS_MSG_AUTH_CHECK_TOKEN_TIMEOUT   = "用户信息已过期"
	STATUS_MSG_AUTH_TOKEN_GENERATE_FAILED = "Token生成失败"
	STATUS_MSG_AUTH_TYPE_ERROR            = "Token校验类型错误"
	// secret
	STATUS_MSG_SECRET_CHECK_FAILED    = "安全校验失败"

	// upload
	STATUS_MSG_UPLOAD_FILE_SAVE_FAILED        = "文件保存失败"
	STATUS_MSG_UPLOAD_FILE_CHECK_FAILED       = "文件检查失败"
	STATUS_MSG_UPLOAD_FILE_CHECK_FORMAT_WRONG = "文件校验错误，文件格式或大小不正确"
)

var StatusCodeMsgMap = map[StatusCode]string {
	// base
	STATUS_CODE_SUCCESS:        STATUS_MSG_SUCCESS,
	STATUS_CODE_INVALID_PARAMS: STATUS_MSG_INVALID_PARAMS,
	STATUS_CODE_ERROR:          STATUS_MSG_ERROR,
	STATUS_CODE_FAILED:         STATUS_MSG_FAILED,

	// jwt token
	STATUS_CODE_AUTH_CHECK_TOKEN_EMPTY     : STATUS_MSG_AUTH_CHECK_TOKEN_EMPTY,
	STATUS_CODE_AUTH_CHECK_TOKEN_FAILED    : STATUS_MSG_AUTH_CHECK_TOKEN_FAILED,
	STATUS_CODE_AUTH_CHECK_TOKEN_TIMEOUT   : STATUS_MSG_AUTH_CHECK_TOKEN_TIMEOUT,
	STATUS_CODE_AUTH_TOKEN_GENERATE_FAILED : STATUS_MSG_AUTH_TOKEN_GENERATE_FAILED,
	STATUS_CODE_AUTH_TYPE_ERROR            : STATUS_MSG_AUTH_TYPE_ERROR,

	// secret
	STATUS_CODE_SECRET_CHECK_FAILED:    STATUS_MSG_SECRET_CHECK_FAILED,

	// upload
	STATUS_CODE_UPLOAD_FILE_SAVE_FAILED:        STATUS_MSG_UPLOAD_FILE_SAVE_FAILED,
	STATUS_CODE_UPLOAD_FILE_CHECK_FAILED:       STATUS_MSG_UPLOAD_FILE_CHECK_FAILED,
	STATUS_CODE_UPLOAD_FILE_CHECK_FORMAT_WRONG: STATUS_MSG_UPLOAD_FILE_CHECK_FORMAT_WRONG,
}

func (code StatusCode) Msg() string {
	msg, ok := StatusCodeMsgMap[code]
	if ok {
		return msg
	}
	return StatusCodeMsgMap[STATUS_CODE_ERROR]
}