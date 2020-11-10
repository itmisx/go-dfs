package pkg

import (
	"bytes"
	"encoding/json"
	"go-dfs/internal/lang"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

// Node global snowflake node
var Node *snowflake.Node

func init() {
	var err error
	rand.Seed(time.Now().UnixNano())
	Node, err = snowflake.NewNode(int64(rand.Intn(1024)))
	if err != nil {
		panic("")
	}

}

// Helper Helper method
type Helper struct {
}

// PostJSON 发送post 请求 timeout 单位 秒
func (h Helper) PostJSON(url string, data interface{}, header map[string]string, timeout time.Duration) (res []byte, err error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return res, err
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return res, err
	}
	request.Header.Set("Content-Type", "application/json")
	for key, value := range header {
		request.Header.Set(key, value)
	}
	client := &http.Client{}
	client.Timeout = time.Second * timeout
	resp, err := client.Do(request)
	if err != nil {
		return res, err
	}
	defer func() { _ = resp.Body.Close() }()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}
	return respData, nil
}

// AjaxReturn ajax return
func (h Helper) AjaxReturn(c *gin.Context, code int64, data interface{}) {
	l := c.Request.Header.Get("Accept-Language")
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"data": data,
		"msg":  lang.T(l, code),
	})
}

//UUID Generate a snowflake ID.
func (h Helper) UUID() int64 {
	return Node.Generate().Int64()
}
