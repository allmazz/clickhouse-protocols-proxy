package controller

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/allmazz/clickhouse-protocol-proxy/clickhouse-protocol-proxy/internal/config"
	"github.com/allmazz/clickhouse-protocol-proxy/clickhouse-protocol-proxy/internal/transport/native"

	"context"
	"net/http"
	"strings"
	"time"
)

var updateStatements = []string{"insert", "update", "alter", "create", "delete", "drop", "detach", "attach"}

type Controller struct {
	cfg *config.Config
	log *zap.Logger

	router *gin.Engine
	server *http.Server
	native *native.Native
}

func New(cfg *config.Config, log *zap.Logger) *Controller {
	ctrl := &Controller{
		cfg:    cfg,
		log:    log,
		router: gin.Default(),
		native: native.New(cfg, log.With(zap.Field{Type: zapcore.StringType, Key: "component", String: "native"})),
	}
	ctrl.bind()
	return ctrl
}

func (c *Controller) bind() {
	c.router.Use(ginzap.Ginzap(c.log, time.RFC3339, true))
	c.router.Use(ginzap.RecoveryWithZap(c.log, true))
	c.router.Any("/", func(ctx *gin.Context) {
		username := ctx.GetHeader("X-ClickHouse-User")
		password := ctx.GetHeader("X-ClickHouse-Key")
		user := ctx.Request.URL.User
		if user != nil {
			username = user.Username()
			password, _ = user.Password()
		}
		if username == "" || password == "" {
			username = ctx.Query("user")
			password = ctx.Query("password")
		}

		database := ctx.GetHeader("X-ClickHouse-Database")
		if database == "" {
			database = ctx.Query("database")
		}

		query := ctx.Param("query")
		if query == "" {
			data, err := ctx.GetRawData()
			if err != nil {
				ctx.String(400, "empty query")
				return
			}
			query = string(data)
			if query == "" {
				ctx.String(400, "empty query")
				return
			}
		}

		if ctx.Request.Method == "GET" {
			lwQuery := strings.ToLower(query)
			for _, keyword := range updateStatements {
				if strings.Contains(lwQuery, keyword) {
					ctx.String(500, "Cannot execute query in readonly mode. For queries over HTTP, method GET implies readonly.")
					return
				}
			}
		}

		q := &native.Query{
			Database: database,
			Username: username,
			Password: password,
			Query:    query,
			Params:   map[string]any{},
		}

		for name, value := range ctx.Request.URL.Query() {
			if strings.HasPrefix(name, "param_") {
				q.Params[strings.TrimPrefix(name, "param_")] = value
			}
		}

		res, err := c.native.Query(ctx, q)
		if err != nil {
			if err.Error() != "EOF" {
				ctx.String(400, err.Error())
			}
			ctx.String(200, "")
			return
		}
		ctx.JSON(200, res)
	})
}

func (c *Controller) Run() chan error {
	c.server = &http.Server{
		Addr:    c.cfg.Server.Addr,
		Handler: c.router,
	}

	ch := make(chan error)
	go func() {
		ch <- c.server.ListenAndServe()
	}()
	c.log.Info("the server started")

	return ch
}

func (c *Controller) Stop(ctx context.Context) error {
	return c.server.Shutdown(ctx)
}
