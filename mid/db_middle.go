package mid

import (
	"net/http"
	"runtime"

	"github.com/94peter/sterna"
	"github.com/94peter/sterna/api"
	"github.com/94peter/sterna/api/mid"
	"github.com/94peter/sterna/db"
	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/util"
	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
)

type DBMidDI interface {
	log.LoggerDI
	db.MongoDI
}

type DBMiddle string

func NewGinDBMid(service string) mid.GinMiddle {
	return &dbMiddle{
		name: service,
	}
}

type dbMiddle struct {
	name string
	di   DBMidDI
}

func (lm *dbMiddle) GetName() string {
	return lm.name
}

var (
	NotGetDIError  = api.NewApiErrorWithKey(http.StatusInternalServerError, "can not get di", "DB_000")
	InvalidDIError = api.NewApiErrorWithKey(http.StatusInternalServerError, "invalid di", "DB_001")
	MongoDBError   = func(msg string) api.ApiError {
		return api.NewApiErrorWithKey(http.StatusInternalServerError, "new mongodb client fail: "+msg, "DB_002")
	}
)

func (m *dbMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		servDi := util.GetCtxVal(c.Request, sterna.CtxServDiKey)
		if servDi == nil {
			m.outputErr(c, NotGetDIError)
			c.Abort()
			return
		}

		if dbdi, ok := servDi.(DBMidDI); ok {
			uuid := uuid.New().String()
			l := dbdi.NewLogger(uuid)

			dbclt, err := dbdi.NewMongoDBClient(c.Request.Context(), "")
			if err != nil {
				m.outputErr(c, MongoDBError(err.Error()))
				c.Abort()
				return
			}
			defer dbclt.Close()

			c.Request = util.SetCtxKeyVal(c.Request, db.CtxMongoKey, dbclt)
			c.Request = util.SetCtxKeyVal(c.Request, log.CtxLogKey, l)

			c.Next()
			runtime.GC()
		} else {
			m.outputErr(c, InvalidDIError)
			c.Abort()
			return
		}
	}
}

func (lm *dbMiddle) outputErr(c *gin.Context, err api.ApiError) {
	api.GinOutputErr(c, lm.name, err)
}
