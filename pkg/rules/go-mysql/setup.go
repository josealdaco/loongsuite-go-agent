// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package go_mysql

import (
	"context"
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
	"os"
	"strings"
	"time"
	_ "unsafe"
)

type gomysqlInnerEnabler struct {
	enabled bool
}

func (r gomysqlInnerEnabler) Enable() bool {
	return r.enabled
}

var gomysqlEnabler = gomysqlInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_GOMYSQL_ENABLED") != "false"}

var GoMySQLInstrumentation = BuildGoMySQLInstrumenter()

type gomysqlContext struct {
	ctx      context.Context
	endpoint string
	database string
}

func (g *gomysqlContext) setContext(ctx context.Context) {
	g.ctx = ctx
}
func (g *gomysqlContext) GetContext() context.Context {
	return g.ctx
}
func (g *gomysqlContext) GetEndpoint() string {
	return g.endpoint
}
func (g *gomysqlContext) GetDatabase() string {
	return g.database
}
func (g *gomysqlContext) setEndpoint(endpoint string) {
	g.endpoint = endpoint
}
func (g *gomysqlContext) setDatabase(database string) {
	g.database = database
}

type hookedContext struct {
	ctx            context.Context
	gomysqlcontext *gomysqlContext
	cmd            string
	opType         string
	database       string
	args           []interface{}
}

var gomysqlCTX = &gomysqlContext{}

//go:linkname onBeforeDialContext github.com/go-mysql-org/go-mysql/client.onBeforeDialContext
func onBeforeDialContext(call api.CallContext, ctx context.Context, network, addr, user, password, dbName string, dialer client.Dialer, options ...client.Option) {
	if !gomysqlEnabler.Enable() {
		return
	}
	if gomysqlCTX.GetEndpoint() == "" {
		gomysqlCTX.setEndpoint(addr)
	}
	if gomysqlCTX.ctx == nil {
		gomysqlCTX.setContext(ctx)
	}
	newCTX := GoMySQLInstrumentation.Start(ctx, &gomysqlRequest{
		endpoint:  addr,
		database:  dbName,
		startTime: time.Now(),
	})
	call.SetData(&hookedContext{
		gomysqlcontext: gomysqlCTX,
		ctx:            newCTX,
	})
}

//go:linkname onExitDialContext github.com/go-mysql-org/go-mysql/client.onExitDialContext
func onExitDialContext(call api.CallContext, conn *client.Conn, err error) {
	if !gomysqlEnabler.Enable() {
		return
	}
	data, ok := call.GetData().(*hookedContext)
	if !ok {
		fmt.Println("onExitDialContext: data assertion failed")
		return
	}
	c, ok := data.ctx.(context.Context)
	if !ok {
		fmt.Println("onExitDialContext: context assertion failed")
		return
	}
	ctx, ok := c.(context.Context)
	if !ok {
		fmt.Println("onExitDialContext: context type assertion failed")
		return
	}
	fmt.Println("Dial Context Value: ", ctx)
	endpoint := data.gomysqlcontext.endpoint
	database := data.gomysqlcontext.database
	GoMySQLInstrumentation.End(ctx, &gomysqlRequest{
		endpoint: endpoint,
		database: database,
		ctx:      ctx,
	}, conn, err)
}

//go:linkname onBeforeExecute github.com/go-mysql-org/go-mysql/client.onBeforeExecute
func onBeforeExecute(call api.CallContext, conn *client.Conn, command string, args ...interface{}) {
	if !gomysqlEnabler.Enable() {
		return
	}
	fmt.Println("Execute called with cmd:", command, "args:", args)
	if gomysqlCTX.GetContext() == nil {
		gomysqlCTX.setContext(context.Background())
	}
	newCTX := GoMySQLInstrumentation.Start(gomysqlCTX.GetContext(), &gomysqlRequest{
		args:     args,
		endpoint: gomysqlCTX.GetEndpoint(),
		cmd:      command,
		opType:   calOp(command),
	})
	call.SetData(&hookedContext{
		gomysqlcontext: gomysqlCTX,
		ctx:            newCTX,
		args:           args,
		cmd:            command,
		opType:         calOp(command),
		database:       conn.GetDB(),
	})

}

//go:linkname onExitExecute github.com/go-mysql-org/go-mysql/client.onExitExecute
func onExitExecute(call api.CallContext, result *mysql.Result, err error) {
	if !gomysqlEnabler.Enable() {
		return
	}
	data, ok := call.GetData().(*hookedContext)
	if !ok {
		fmt.Println("onExitExecute: data assertion failed")
		return
	}
	c, ok := data.ctx.(context.Context)
	if !ok {
		fmt.Println("onExitExecute: context assertion failed")
		return
	}
	ctx, ok := c.(context.Context)
	if !ok {
		fmt.Println("onExitExecute: context type assertion failed")
		return
	}
	fmt.Println("Data base name from Exec: ", data.database)
	GoMySQLInstrumentation.End(ctx, &gomysqlRequest{
		ctx:      ctx,
		args:     data.args,
		endpoint: gomysqlCTX.GetEndpoint(),
		cmd:      data.cmd,
		opType:   data.opType,
		database: data.database,
	}, result, err)
}

func calOp(sql string) string {
	sqls := strings.Split(sql, " ")
	var op string
	if len(sqls) > 0 {
		op = sqls[0]
	}
	return op
}
