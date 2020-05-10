package database

import (
	"fmt"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/foundation"
	"log"
	"testing"
	"time"
)

/*
CREATE TABLE `user` (
`user_id` BIGINT NOT NULL COMMENT '用户ID',
`user_name` char(255) DEFAULT NULL COMMENT '用户名',
`password` char(255) DEFAULT NULL COMMENT '密码',
`alias_name` char(255) DEFAULT NULL COMMENT '用户别名，昵称',
`enabled` INT DEFAULT 1 COMMENT '账户是否启用：1，启用；0，禁用；',
`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
`update_user` char(255) DEFAULT 'system' COMMENT '修改这条记录的管理员用户名',
`delete_time` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间',
PRIMARY KEY (`user_id`),
UNIQUE KEY `user_name`(`user_name`,`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
*/


type User struct {
	Model
	UserId int 			`db:"user_id" json:"userId"`
	UserName string 	`db:"user_name" json:"userName"`
	Password string 	`db:"password" json:"-"`
	AliasName string 	`db:"alias_name" json:"aliasName"`
	Enabled int			`db:"enabled" json:"enabled"`
}

const (
	// insert
	sql_user_insert = "insert into user(user_id, user_name, password, alias_name)values(?, ?, ?, ?)"
	// select to check exist
	sql_user_check_username_exist = "select user_id from user where user_name=? and delete_time=0"
	// select to login
	sql_user_select_login = "select * from user where user_name=? and password=? and delete_time=0 limit 1"
	// select by id
	sql_user_select_by_id = "select * from user where user_id=? and delete_time=0 limit 1"
	// select all
	sql_user_select_all = "select * from user where delete_time=0"
	// select all
	sql_user_select_all_with_deleted = "select * from user"
	// update by id
	sql_user_update_by_id = "update user set alias_name=?, password=? where user_id=? and delete_time=0"
	// update to delete by id
	sql_user_delete_by_id = "update user set delete_time=? where user_id=? and delete_time=0"
	// update to enable by id
	sql_user_enabled_by_id = "update user set enabled=1 where user_id=? and delete_time=0"
	// update to disenable by id
	sql_user_disenabled_by_id = "update user set enabled=0 where user_id=? and delete_time=0"
)

var idWorker foundation.SnowId
var mysqlManager IMysqlManager
func mysqlTestStart() {
	var err error
	idWorker, err = foundation.NewWorker(0)
	if err != nil {
		panic(err)
	}

	config := MysqlConfig{
		Type:        "mysql",
		User:        "root",
		Password:    "admin2018",
		Hosts: 		 map[int]string{0:"127.0.0.1:3306"},
		Name:        "etcd_center",
		MaxIdle:     30,
		MaxOpen:     30,
		MaxLifetime: 200 * time.Second,
	}
	mysqlManager = NewMysqlManager(config)
	mysqlManager.Open()
}

var (
	username = "zhagnsan"
	password = encrypt.GetMd5().Encrypt([]byte("123456"))
	userId = 99349769469558784
)

func TestUserInsert(t *testing.T) {
	mysqlTestStart()
	if user, err := userDao.userAdd(username, password); err != nil {
		t.Error(err)
	} else {
		fmt.Println("user insert success : ", user)
	}
}

func TestUserSelect(t *testing.T) {
	mysqlTestStart()
	if user, err := userDao.userGetByUserNameAndPassword(username, password); err != nil {
		t.Error(err)
	} else {
		fmt.Println("user select success : ", user)
	}
}

func TestUserDisEnabled(t *testing.T) {
	mysqlTestStart()
	if err := userDao.userEnabled(userId, username, false); err != nil {
		t.Error(err)
	}
	if user, err := userDao.userGetById(userId, username); err != nil {
		t.Error(err)
	} else {
		fmt.Println("user disenabled success : ", user)
	}
}

func TestUserDelete(t *testing.T) {
	mysqlTestStart()
	if err := userDao.userDelete(userId, username); err != nil {
		t.Error(err)
	} else {
		fmt.Print("user delete success : ")
	}
	if users, err := userDao.userGetAll(false); err != nil {
		t.Error(err)
	} else {
		fmt.Printf("%+v", users)
	}
}



type userDaoImpl struct {}
var userDao = &userDaoImpl{}

/**
 * 用户名检查
 */
func (this *userDaoImpl) userExistCheck(userName string) (bool, error) {
	// db get
	db, err := mysqlManager.GetDbByUserName(userName)
	if err != nil {
		return false, err
	}
	// userName check
	var users []*User
	if err := db.Select(&users, sql_user_check_username_exist, userName); err != nil {
		return false, err
	}
	if len(users) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

/**
 * 用户数据新增入库
 */
func (this *userDaoImpl) userAdd(userName, password string) (*User, error) {
	// db get
	db, err := mysqlManager.GetDbByUserName(userName)
	if err != nil {
		return nil, err
	}
	// userName check
	exist, err := this.userExistCheck(userName)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, foundation.NewError(error_code_user_exist, error_msg_user_exist)
	}
	// userId generate
	userId := idWorker.GetId()
	if err != nil {
		return nil, err
	}
	// user insert
	_, err = db.Exec(sql_user_insert, userId, userName, password, userName)
	if err != nil {
		log.Print("exec failed, ", err)
		return nil, err
	}
	return this.userGetById(int(userId), userName)
}

/**
 * 根据用户名密码查询用户信息
 */
func (this *userDaoImpl) userGetByUserNameAndPassword(userName, password string) (*User, error) {
	// db get
	db, err := mysqlManager.GetDbByUserName(userName)
	if err != nil {
		return nil, err
	}
	// user select
	var users []*User
	err = db.Select(&users, sql_user_select_login, userName, password)
	if err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, foundation.NewError(error_code_notfilter_username_password, error_msg_notfilter_username_password)
	}
	user := users[0]
	if user.Enabled == 0 {
		return nil, foundation.NewError(error_code_user_disenabled, error_msg_user_disenabled)
	}
	return user, nil
}

/**
 * 根据用户Id查询用户信息
 */
func (this *userDaoImpl) userGetById(userId int, userName string) (*User, error) {
	// db get
	db, err := mysqlManager.GetDbByUserName(userName)
	if err != nil {
		return nil, err
	}
	// user select
	var users []*User
	err = db.Select(&users, sql_user_select_by_id, userId)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, foundation.NewError(error_code_user_not_exist, error_msg_user_not_exist)
	}
	user := users[0]
	return user, nil
}

/**
 * 修改用户密码
 */
func (this *userDaoImpl) userUpdate(userId int, userName, password, aliasName string) error {
	// db get
	db, err := mysqlManager.GetDbByUserName(userName)
	if err != nil {
		return err
	}
	user, err := this.userGetById(userId, userName)
	if err != nil {
		return err
	}
	if password == "" {
		password = user.Password
	}
	if aliasName == "" {
		aliasName = user.AliasName
	}
	if user.Password == password && user.AliasName == aliasName {
		return nil
	}

	// update
	res, err := db.Exec(sql_user_update_by_id, aliasName, password, userId)
	if err != nil {
		return err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if row == 0 {
		return foundation.NewError(error_code_update_failed, error_msg_update_failed)
	}
	return nil
}

/**
 * 删除用户（软删除）
 */
func (this *userDaoImpl) userDelete(userId int, userName string) error {
	// db get
	db, err := mysqlManager.GetDbByUserName(userName)
	if err != nil {
		return err
	}
	// userName check
	exist, err := this.userExistCheck(userName)
	if err != nil {
		return err
	}
	if !exist {
		return foundation.NewError(error_code_user_not_exist, error_msg_user_not_exist)
	}
	// user delete update
	res, err := db.Exec(sql_user_delete_by_id, idWorker.GetProjTimeStamp(), userId)
	if err != nil {
		return err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if row == 0 {
		return foundation.NewError(error_code_delete_failed, error_msg_delete_failed)
	}
	return nil
}

func (this *userDaoImpl) userEnabled(userId int, userName string, enabled bool) error {
	// db get
	db, err := mysqlManager.GetDbByUserName(userName)
	if err != nil {
		return err
	}
	// userName check
	exist, err := this.userExistCheck(userName)
	if err != nil {
		return err
	}
	if !exist {
		return foundation.NewError(error_code_user_not_exist, error_msg_user_not_exist)
	}
	// user enable update
	var sql string; var code int; var msg string
	if enabled {
		sql = sql_user_enabled_by_id
		code = error_code_enable_failed
		msg = error_msg_enabled_failed
	} else {
		sql = sql_user_disenabled_by_id
		code = error_code_disenabled_failed
		msg = error_msg_disenabled_failed
	}
	res, err := db.Exec(sql, userId)
	if err != nil {
		return err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if row == 0 {
		return foundation.NewError(code, msg)
	}
	return nil
}

func (this *userDaoImpl) userGetAll(containsDeletedData bool) ([]*User, error) {
	var users []*User
	var errs []error
	for _, db := range mysqlManager.GetAllDbs() {
		// user select
		var tmp_users []*User
		var sql string
		if containsDeletedData {
			sql = sql_user_select_all_with_deleted
		} else {
			sql = sql_user_select_all
		}
		if err := db.Select(&tmp_users, sql); err != nil {
			errs = append(errs, err)
			continue
		}

		users = append(users, tmp_users...)
	}
	if len(errs) != 0 {
		return users, errs[0]
	}
	return users, nil
}



const (
	_                                     int = 0
	error_code_user_exist                                    = 11001
	error_code_user_not_exist                                = 11002
	error_code_notfilter_username_password                   = 11003
	error_code_user_disenabled								 = 11009
	error_code_update_failed			                     = 11004
	error_code_delete_failed			                     = 11005
	error_code_enable_failed			                     = 11006
	error_code_disenabled_failed			                 = 11007
)

const (
	_                                     string = ""
	error_msg_user_exist                                    = "用户已存在"
	error_msg_user_not_exist                                = "用户不存在"
	error_msg_notfilter_username_password                   = "用户名或密码不正确"
	error_msg_user_disenabled			                    = "用户已被禁用"
	error_msg_update_failed			                        = "修改失败"
	error_msg_delete_failed			                        = "删除失败"
	error_msg_enabled_failed			                    = "启用失败"
	error_msg_disenabled_failed			                    = "禁用失败"
)