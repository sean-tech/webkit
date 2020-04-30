package serving

//import (
//	"fmt"
//	"github.com/sean-tech/webservice/config"
//	"github.com/sean-tech/webservice/fileutils"
//	"log"
//	"mime/multipart"
//	"os"
//	"path"
//	"strings"
//)
//
//func GetUploadFileFullUrl(name string) string {
//	return config.Upload.FilePrefixUrl + "/" + config.Upload.FileSavePath + name
//}
//
//func GetUploadFileName(name string) string {
//	ext := path.Ext(name)
//	fileName := strings.TrimSuffix(name, ext)
//	//fileName = encrypt.GetMd5().EncryptWithTimestamp([]byte(fileName), 0)
//	return fileName + ext
//}
//
//func GetUploadFilePath() string {
//	return config.Upload.FileSavePath
//}
//
//func GetUploadFileFullPath() string {
//	return config.App.RuntimeRootPath + config.Upload.FileSavePath
//}
//
//func CheckUploadFileExt(fileName string) bool {
//	ext := fileutils.GetExt(fileName)
//	for _, allowExt := range config.Upload.FileAllowExts {
//		if strings.ToUpper(allowExt) == strings.ToUpper(ext) {
//			return true
//		}
//	}
//	return false
//}
//
//func CheckUploadFileSize(f multipart.File) bool {
//	size, err := fileutils.GetSize(f)
//	if err != nil {
//		log.Println(err)
//		return false
//	}
//
//	return size <= config.Upload.FileMaxSize
//}
//
//func CheckUploadFile(src string) error {
//	dir, err := os.Getwd()
//	if err != nil {
//		return fmt.Errorf("os.Getwd err: %v", err)
//	}
//
//	err = fileutils.MKDirIfNotExist(dir + "/" + src)
//	if err != nil {
//		return fmt.Errorf("file.IsNotExistMkDir err: %v", err)
//	}
//
//	perm := fileutils.CheckPermission(src)
//	if perm == true {
//		return fmt.Errorf("file.CheckPermission Permission denied src: %s", src)
//	}
//
//	return nil
//}









/**
 * 文件上传处理函数
 */
//func (g *Gin) UploadFile() (fileUrl, filePath string, ok bool) {
//
//	data := make(map[string]string)
//
//	file, fileHeader, err := g.Ctx.Request.FormFile("file")
//	if err != nil {
//		logging.Warning(err)
//		g.Response(STATUS_CODE_ERROR, err.Error(), data)
//		return "", "", false
//	}
//	if fileHeader == nil {
//		g.ResponseCode(STATUS_CODE_INVALID_PARAMS, data)
//		return "", "", false
//	}
//
//	fileName := GetUploadFileName(fileHeader.Filename)
//	fullPath := GetUploadFileFullPath()
//	savePath := GetUploadFilePath()
//	src := fullPath + fileName
//	if !CheckUploadFileExt(src) || !CheckUploadFileSize(file) {
//		g.ResponseCode(STATUS_CODE_UPLOAD_FILE_CHECK_FORMAT_WRONG, nil)
//		return "", "", false
//	}
//	if err := CheckUploadFile(fullPath); err != nil {
//		logging.Warning(err)
//		g.ResponseCode(STATUS_CODE_UPLOAD_FILE_CHECK_FAILED, nil)
//		return "", "", false
//	}
//	if err := g.Ctx.SaveUploadedFile(fileHeader, src); err != nil {
//		logging.Warning(err)
//		g.ResponseCode(STATUS_CODE_UPLOAD_FILE_SAVE_FAILED, nil)
//		return "", "", false
//	}
//	fileUrl = GetUploadFileFullUrl(fileName)
//	filePath = savePath + fileName
//	return fileUrl, filePath, true
//}
