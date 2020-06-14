package v1

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"time"

	gin "github.com/gin-gonic/gin"
	"github.com/xunull/etcd-manage/program/common"
	"github.com/xunull/etcd-manage/program/config"
	"github.com/xunull/etcd-manage/program/etcdv3"
	"github.com/xunull/etcd-manage/program/logger"
)

// V1 v1 版接口
func V1(v1 *gin.RouterGroup) {

	v1.GET("/lsdir", GetLsEtcdKey)

	v1.GET("/key", getEtcdKeyValue) // 获取一个key的具体值

	v1.GET("/key/format", getValueToFormat) // 格式化为json或toml

	v1.GET("/members", getEtcdMembers) // 获取节点列表
	v1.GET("/server", getEtcdServerList) // 获取etcd服务列表

	v1.GET("/logs", getLogsList) // 查询日志

	v1.GET("/users", getUserList)       // 获取用户列表
	v1.GET("/logtypes", getLogTypeList) // 获取日志类型列表
}

func GetLsEtcdKey(c *gin.Context) {
	go saveLog(c, "ls dir")

	key := c.Query("key")
	var err error
	defer func() {
		if err != nil {
			logger.Log.Errorw("Get key list error", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
		}
	}()

	etcdCli, exists := c.Get("EtcdServer")
	if exists == false {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Etcd client is empty",
		})
		return
	}
	cli := etcdCli.(*etcdv3.Etcd3Client)

	resp, err := cli.LsDir(key)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 获取key的值
func getEtcdKeyValue(c *gin.Context) {
	go saveLog(c, "Get the value of the key")

	key := c.Query("key")
	var err error
	defer func() {
		if err != nil {
			logger.Log.Errorw("Get the value of the key value incorrectly", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
		}
	}()

	etcdCli, exists := c.Get("EtcdServer")
	if exists == false {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Etcd client is empty",
		})
		return
	}
	cli := etcdCli.(*etcdv3.Etcd3Client)

	val, err := cli.Value(key)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, val)
}

// 获取服务节点
func getEtcdMembers(c *gin.Context) {
	go saveLog(c, "Get etcd cluster information")

	var err error
	defer func() {
		if err != nil {
			logger.Log.Errorw("Get service node error", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
		}
	}()

	etcdCli, exists := c.Get("EtcdServer")
	if exists == false {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Etcd client is empty",
		})
		return
	}
	cli := etcdCli.(*etcdv3.Etcd3Client)

	members, err := cli.Members()
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, members)
}

// PostReq 添加和修改时的body
type PostReq struct {
	*etcdv3.Node
	EtcdName string `json:"etcd_name"`
}







// 获取key前缀，下的值为指定格式 josn toml
func getValueToFormat(c *gin.Context) {
	go saveLog(c, "Format display key")

	format := c.Query("format")
	key := c.Query("key")

	var err error
	defer func() {
		if err != nil {
			logger.Log.Errorw("Save key error", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
		}
	}()

	etcdCli, exists := c.Get("EtcdServer")
	if exists == false {
		err = errors.New("Etcd client is empty")
		return
	}
	cli := etcdCli.(*etcdv3.Etcd3Client)

	list, err := cli.GetRecursiveValue(key)
	if err != nil {
		return
	}

	// js, _ := json.Marshal(list)
	// log.Println(string(js))

	switch format {
	case "json":
		resp, err := etcdv3.NodeJsonFormat(key, list)
		if err != nil {
			return
		}
		respJs, _ := json.MarshalIndent(resp, "", "    ")
		c.JSON(http.StatusOK, string(respJs))
		return
	case "toml":

	default:
		err = errors.New("Unsupported format")
	}
}

// 获取etcd服务列表
func getEtcdServerList(c *gin.Context) {
	go saveLog(c, "Get the etcd service column table")

	cfg := config.GetCfg()
	if cfg == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Configured as nil",
		})
		return
	}
	list := cfg.Server
	if list == nil {
		list = make([]*config.EtcdServer, 0)
		c.JSON(http.StatusOK, list)
		return
	}
	// 当前用户角色
	userRole := ""
	userRoleIn, exists := c.Get("userRole")
	if exists == true {
		userRole = userRoleIn.(string)
	}

	// log.Println(userRole)

	// 只返回有权限服务列表
	list1 := make([]*config.EtcdServer, 0)
	for _, s := range list {
		if s.Roles == nil || len(s.Roles) == 0 {
			list1 = append(list1, s)
		} else {
			for _, r := range s.Roles {
				if r == userRole {
					list1 = append(list1, s)
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, list1)
}

// 获取用户列表
func getUserList(c *gin.Context) {
	us := make([]map[string]string, 0)
	cfg := config.GetCfg()
	if cfg != nil {
		for _, v := range cfg.Users {
			us = append(us, map[string]string{
				"name": v.Username,
				"role": v.Role,
			})
		}
	}

	c.JSON(http.StatusOK, us)
}

// 获取操作类型列表
func getLogTypeList(c *gin.Context) {
	c.JSON(http.StatusOK, []string{
		"Get the key list",
		"Format display key",
		"Get the value of the key",
		"Get etcd cluster information",
		"Delete key",
		"Save key",
		"Get the etcd service column table",
	})
}

type LogLine struct {
	Date  string  `json:"date"`
	User  string  `json:"user"`
	Role  string  `json:"role"`
	Msg   string  `json:"msg"`
	Ts    float64 `json:"ts"`
	Level string  `json:"level"`
}

// 查看日志
func getLogsList(c *gin.Context) {
	page := c.Query("page")
	pageSize := c.Query("page_size")
	dateStr := c.Query("date")
	querUser := c.Query("user")
	queryLogType := c.Query("log_type")

	var err error
	defer func() {
		if err != nil {
			logger.Log.Errorw("View log errors", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
		}
	}()

	// 计算开始和结束行
	pageNum, _ := strconv.Atoi(page)
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSizeNum, _ := strconv.Atoi(pageSize)
	if pageSizeNum <= 0 {
		pageSizeNum = 10
	}
	startLine := (pageNum - 1) * pageSizeNum
	endLine := pageNum * pageSizeNum

	fileName := fmt.Sprintf("%slogs/%s.log", common.GetRootDir(), dateStr)
	// fmt.Println(fileName)
	// 判断文件是否存在
	if exists, err := common.PathExists(fileName); exists == false || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("No logs for [%s]", dateStr),
		})
		return
	}
	// 读取指定行
	file, err := os.Open(fileName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Read log file error",
		})
		return
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	lineCount := 1
	list := make([]*LogLine, 0) // 最终数组
	for fileScanner.Scan() {
		logTxt := fileScanner.Text()
		if logTxt == "" {
			continue
		}
		// 解析日志
		oneLog := new(LogLine)
		err = json.Unmarshal([]byte(logTxt), oneLog)
		if err != nil {
			logger.Log.Errorw("Parse log file error", "err", err)
			continue
		}
		// 只看info类型日志
		if oneLog.Level != "info" {
			continue
		}

		if lineCount > startLine && lineCount <= endLine {
			// 判断用户和日志类型参数
			if querUser != "" && oneLog.User != querUser {
				continue
			}
			if queryLogType != "" && oneLog.Msg != queryLogType {
				continue
			}

			oneLog.Date = time.Unix(int64(oneLog.Ts), 0).In(time.Local).Format("2006-01-02 15:04:05")
			list = append(list, oneLog)
		}

		lineCount++
	}
	err = nil

	c.JSON(http.StatusOK, gin.H{
		"list":  list,
		"total": lineCount - 1,
	})
}

// 记录访问日志
func saveLog(c *gin.Context, msg string) {
	//user := c.MustGet(gin.AuthUserKey).(string) // 用户
	user := "admin"
	userRole := "" // 角色
	userRoleIn, exists := c.Get("userRole")
	if exists == true && userRoleIn != nil {
		userRole = userRoleIn.(string)
	}
	// 存储日志
	logger.Log.Infow(msg, "user", user, "role", userRole)
}
