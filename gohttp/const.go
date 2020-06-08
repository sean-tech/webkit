package gohttp

type StatusCode int
type StatusMsg string

const (
	_ StatusCode = 0
	// base
	STATUS_CODE_SUCCESS        					= 200
	STATUS_CODE_INVALID_PARAMS 					= 400
	STATUS_CODE_ERROR          					= 500
	STATUS_CODE_FAILED         					= 800

	// secret
	STATUS_CODE_SECRET_CHECK_FAILED    			= 843
	STATUS_CODE_PERMISSION_DENIED    			= 844

	// upload
	STATUS_CODE_UPLOAD_FILE_SAVE_FAILED        	= 811
	STATUS_CODE_UPLOAD_FILE_CHECK_FAILED       	= 812
	STATUS_CODE_UPLOAD_FILE_CHECK_FORMAT_WRONG 	= 813
)

const (
	_ StatusMsg = ""
	// base
	STATUS_MSG_SUCCESS        					= "ok"
	STATUS_MSG_INVALID_PARAMS 					= "参数校验失败"
	STATUS_MSG_ERROR          					= "system error"
	STATUS_MSG_FAILED         					= "操作失败"

	// secret
	STATUS_MSG_SECRET_CHECK_FAILED    			= "安全校验失败"
	STATUS_MSG_PERMISSION_DENIED    			= "您没有访问权限"

	// upload
	STATUS_MSG_UPLOAD_FILE_SAVE_FAILED        	= "文件保存失败"
	STATUS_MSG_UPLOAD_FILE_CHECK_FAILED       	= "文件检查失败"
	STATUS_MSG_UPLOAD_FILE_CHECK_FORMAT_WRONG 	= "文件校验错误，文件格式或大小不正确"
)

var StatusCodeMsgMap = map[StatusCode]string {
	// base
	STATUS_CODE_SUCCESS:        STATUS_MSG_SUCCESS,
	STATUS_CODE_INVALID_PARAMS: STATUS_MSG_INVALID_PARAMS,
	STATUS_CODE_ERROR:          STATUS_MSG_ERROR,
	STATUS_CODE_FAILED:         STATUS_MSG_FAILED,

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