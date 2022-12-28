package mid

import (
	"net/http"
	"reflect"

	"github.com/94peter/sterna"
	"github.com/94peter/sterna/api"
	"github.com/94peter/sterna/api/mid"
	"github.com/94peter/sterna/mystorage"
	"github.com/94peter/sterna/util"
	"github.com/gin-gonic/gin"
)

type DevDIMiddle string

func NewGinDevDiMid(fileStorage mystorage.Storage, di interface{}, service string) mid.GinMiddle {
	return &devDiMiddle{
		service: service,
		storage: fileStorage,
		di:      di,
	}
}

type devDiMiddle struct {
	service string
	storage mystorage.Storage
	di      interface{}
}

func (lm *devDiMiddle) outputErr(c *gin.Context, err error) {
	api.GinOutputErr(c, lm.service, err)
}

func (lm *devDiMiddle) GetName() string {
	return "di"
}

func (am *devDiMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		val := reflect.ValueOf(am.di)
		if val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}
		newValue := reflect.New(val.Type()).Interface()
		confByte, err := am.storage.Get("config.yml")
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
