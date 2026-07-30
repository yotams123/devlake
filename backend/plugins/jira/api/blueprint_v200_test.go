/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"testing"

	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/srvhelper"
	"github.com/apache/incubator-devlake/plugins/jira/models"
	"github.com/stretchr/testify/assert"
)

func TestMakeDataSourcePipelinePlanV200PassesScopeConfigId(t *testing.T) {
	const scopeConfigId uint64 = 42
	subtaskMetas := []plugin.SubTaskMeta{
		{Name: "collectIssues", EnabledByDefault: true, DomainTypes: []string{plugin.DOMAIN_TYPE_TICKET}},
	}

	plan, err := makeDataSourcePipelinePlanV200(
		subtaskMetas,
		[]*srvhelper.ScopeDetail[models.JiraBoard, models.JiraScopeConfig]{
			{
				// ScopeConfig.ID takes priority over the scope's own ScopeConfigId
				Scope: models.JiraBoard{
					Scope:   common.Scope{ConnectionId: 1, ScopeConfigId: 1},
					BoardId: 100,
				},
				ScopeConfig: &models.JiraScopeConfig{
					ScopeConfig: common.ScopeConfig{Model: common.Model{ID: scopeConfigId}, Entities: []string{plugin.DOMAIN_TYPE_TICKET}},
				},
			},
		},
		&models.JiraConnection{},
	)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(plan))
	assert.Equal(t, 1, len(plan[0]))
	// the resolved scopeConfigId must be threaded into the task options so
	// PrepareTaskData can reliably load the scope config at runtime instead
	// of depending on the _tool_jira_boards fallback lookup.
	assert.EqualValues(t, scopeConfigId, plan[0][0].Options["scopeConfigId"])
}

func TestMakeDataSourcePipelinePlanV200FallsBackToScopeScopeConfigId(t *testing.T) {
	const scopeConfigId uint64 = 7
	subtaskMetas := []plugin.SubTaskMeta{
		{Name: "collectIssues", EnabledByDefault: true, DomainTypes: []string{plugin.DOMAIN_TYPE_TICKET}},
	}

	plan, err := makeDataSourcePipelinePlanV200(
		subtaskMetas,
		[]*srvhelper.ScopeDetail[models.JiraBoard, models.JiraScopeConfig]{
			{
				// no ScopeConfig resolved (nil) -> fall back to the scope's own ScopeConfigId
				Scope: models.JiraBoard{
					Scope:   common.Scope{ConnectionId: 1, ScopeConfigId: scopeConfigId},
					BoardId: 100,
				},
				ScopeConfig: nil,
			},
		},
		&models.JiraConnection{},
	)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(plan))
	assert.Equal(t, 1, len(plan[0]))
	assert.EqualValues(t, scopeConfigId, plan[0][0].Options["scopeConfigId"])
}
