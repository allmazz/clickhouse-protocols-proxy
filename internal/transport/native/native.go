package native

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"

	"github.com/allmazz/clickhouse-protocol-proxy/clickhouse-protocol-proxy/internal/config"

	"context"
	"fmt"
	"reflect"
	"sync"
)

type Native struct {
	cfg *config.Config
	log *zap.Logger

	connPool  map[string]clickhouse.Conn
	connMutex map[string]*sync.Mutex
}

func New(cfg *config.Config, log *zap.Logger) *Native {
	return &Native{
		cfg:       cfg,
		log:       log,
		connPool:  map[string]clickhouse.Conn{},
		connMutex: map[string]*sync.Mutex{},
	}
}

func (n *Native) Query(ctx context.Context, q *Query) (*Response, error) {
	conn, err := n.getConn(ctx, q)
	if err != nil {
		return nil, err
	}

	params := []any{}
	for k, v := range q.Params {
		params = append(params, clickhouse.Named(k, v))
	}
	rows, err := conn.Query(ctx, q.Query, params...)
	if err != nil {
		return nil, err
	}

	columns := make([]MetaElement, len(rows.Columns()))
	rowStructFields := make([]reflect.StructField, len(rows.Columns()))
	for i, columnName := range rows.Columns() {
		columnType := rows.ColumnTypes()[i]
		rowStructFields[i] = reflect.StructField{
			Name: fmt.Sprintf("V%d", i),
			Type: columnType.ScanType(),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%[1]v" ch:"%[1]v"`, columnName)),
		}
		columns[i] = MetaElement{
			Name: columnName,
			Type: columnType.DatabaseTypeName(),
		}
	}
	rowStructType := reflect.StructOf(rowStructFields)
	rowsStructs := make([]interface{}, 0)

	for rows.Next() {
		rowStruct := reflect.New(rowStructType).Elem()
		err = rows.ScanStruct(rowStruct.Addr().Interface())
		if err != nil {
			return nil, err
		}
		rowsStructs = append(rowsStructs, rowStruct.Interface())
	}

	resp := n.mapResponse(columns, rowsStructs)
	return &resp, nil
}

func (n *Native) getConn(ctx context.Context, q *Query) (clickhouse.Conn, error) {
	if q.Database == "" {
		q.Database = "default"
	}
	key := q.Database + "_" + q.Username

	if conn, ok := n.connPool[key]; ok {
		return conn, nil
	}

	mutex := n.connMutex[key]
	if mutex == nil {
		mutex = &sync.Mutex{}
		n.connMutex[key] = mutex
	}
	mutex.Lock()
	defer mutex.Unlock()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr:        n.cfg.Target.Hosts,
		Protocol:    clickhouse.Native,
		Settings:    n.cfg.Target.Settings,
		DialTimeout: n.cfg.Target.DialTimeout,

		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		MaxIdleConns:     n.cfg.Target.MaxConnectionsPerUser,
		MaxOpenConns:     n.cfg.Target.MaxConnectionsPerUser,
		ConnMaxLifetime:  n.cfg.Target.MaxConnectionLifetime,
		ReadTimeout:      n.cfg.Target.ReadTimeout,

		Auth: clickhouse.Auth{
			Database: q.Database,
			Username: q.Username,
			Password: q.Password,
		},

		Debug: n.cfg.Target.Debug,
		Debugf: func(format string, v ...any) {
			n.log.Debug(fmt.Sprintf(format, v...))
		},

		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: config.ServiceName, Version: config.Version},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	n.connPool[key] = conn
	return conn, nil
}
