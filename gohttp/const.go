package gohttp

import "github.com/sean-tech/gokit/requisition"

const (
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

func init() {
	requisition.SetMsgMap(requisition.LangeageZh, map[int]string{
		STATUS_CODE_SUCCESS : 						"操作成功",
		STATUS_CODE_INVALID_PARAMS : 				"参数校验失败",
		STATUS_CODE_ERROR : 						"系统错误",
		STATUS_CODE_FAILED : 						"操作失败",
		STATUS_CODE_SECRET_CHECK_FAILED : 			"安全校验失败",
		STATUS_CODE_PERMISSION_DENIED : 			"您没有访问权限",
		STATUS_CODE_UPLOAD_FILE_SAVE_FAILED : 		"文件保存失败",
		STATUS_CODE_UPLOAD_FILE_CHECK_FAILED : 		"文件校验失败",
		STATUS_CODE_UPLOAD_FILE_CHECK_FORMAT_WRONG :"文件校验错误，文件格式或大小不正确",
	})
	requisition.SetMsgMap(requisition.LanguageEn, map[int]string{
		STATUS_CODE_SUCCESS : 						"ok",
		STATUS_CODE_INVALID_PARAMS : 				"params validate failed",
		STATUS_CODE_ERROR : 						"system error",
		STATUS_CODE_FAILED : 						"failed",
		STATUS_CODE_SECRET_CHECK_FAILED : 			"secret check failed",
		STATUS_CODE_PERMISSION_DENIED : 			"permission denied",
		STATUS_CODE_UPLOAD_FILE_SAVE_FAILED : 		"file save failed",
		STATUS_CODE_UPLOAD_FILE_CHECK_FAILED : 		"file check failed",
		STATUS_CODE_UPLOAD_FILE_CHECK_FORMAT_WRONG :"file format or size wrong",
	})
}






