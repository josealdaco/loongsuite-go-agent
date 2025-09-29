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
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"strings"
	"time"
)

type gomysqlRequest struct {
	opType    string
	args      []interface{}
	endpoint  string
	cmd       string
	database  string
	ctx       context.Context
	startTime time.Time
}

type gomysqlAttrsGetter struct {
}

func (m gomysqlAttrsGetter) GetSystem(request *gomysqlRequest) string {
	return "mysql"
}

func (m gomysqlAttrsGetter) GetServerAddress(request *gomysqlRequest) string {
	return request.endpoint
}

func (m gomysqlAttrsGetter) GetStatement(request *gomysqlRequest) string {
	if len(request.args) == 0 {
		return request.cmd
	}

	// Replace ? placeholders with actual values
	result := request.cmd
	for _, arg := range request.args {
		placeholder := "?"
		replacement := fmt.Sprintf("%v", arg)

		// Find the i-th occurrence of ? and replace it
		idx := strings.Index(result, placeholder)
		if idx != -1 {
			result = result[:idx] + replacement + result[idx+1:]
		}
	}

	return result
}

func (m gomysqlAttrsGetter) GetOperation(request *gomysqlRequest) string {
	return request.opType
}

func (m gomysqlAttrsGetter) GetCollection(request *gomysqlRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (m gomysqlAttrsGetter) GetParameters(request *gomysqlRequest) []any {
	return nil
}

func (m gomysqlAttrsGetter) GetDbNamespace(request *gomysqlRequest) string {
	return request.database
}

func (m gomysqlAttrsGetter) GetBatchSize(request *gomysqlRequest) int {
	return 0
}

func BuildGoMySQLInstrumenter() instrumenter.Instrumenter[*gomysqlRequest, interface{}] {
	builder := instrumenter.Builder[*gomysqlRequest, any]{}
	getter := gomysqlAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[*gomysqlRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[*gomysqlRequest]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GOMYSQL_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(db.DbClientMetrics("sql.go-mysql")).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[*gomysqlRequest, any, db.DbClientAttrsGetter[*gomysqlRequest]]{Base: db.DbClientCommonAttrsExtractor[*gomysqlRequest, any, db.DbClientAttrsGetter[*gomysqlRequest]]{Getter: getter}}).
		BuildInstrumenter()
}
