package mid

import (
	"net/http"
	"reflect"

	"github.com/94peter/sterna"
	"github.com/94peter/sterna/api"
	"github.com/94peter/sterna/api/mid"
	"github.com/94peter/sterna/db"
	"github.com/94peter/sterna/util"
	"github.com/gin-gonic/gin"
)

type DIMiddle string

func NewGinDiMid(clt db.RedisClient, env string, di interface{}, service string) mid.GinMiddle {
	return &diMiddle{
		service: service,
		clt:     clt,
		env:     env,
		di:      di,
	}
}

type diMiddle struct {
	service string
	clt     db.RedisClient
	env     string
	di      interface{}
}

func (lm *diMiddle) outputErr(c *gin.Context, err error) {
	api.GinOutputErr(c, lm.service, err)
}

func (lm *diMiddle) GetName() string {
	return "di"
}

func (am *diMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-DiKey")
		if key == "" {
			am.outputErr(c, api.NewApiError(http.StatusInternalServerError, "missing X-Dikey"))
			c.Abort()
			return
		}

		val := reflect.ValueOf(am.di)
		if val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}
		newValue := reflect.New(val.Type()).Interface()
		confByte, err := am.clt.Get(key)
		if err != nil {
			am.outputErr(c, api.NewApiError(http.StatusInternalServerError, err.Error()))
			c.Abort()
			return
		}
		sterna.InitConfByByte(confByte, newValue)
		c.Request = util.SetCtxKeyVal(c.Request, sterna.CtxServDiKey, newValue)

		c.Next()
	}
}

func GetGinHost(c *gin.Context) string {
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return host
}
