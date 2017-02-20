//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package execution

import (
	"encoding/json"

	"github.com/couchbase/query/auth"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/value"
)

type GrantRole struct {
	base
	plan *plan.GrantRole
}

func NewGrantRole(plan *plan.GrantRole, context *Context) *GrantRole {
	rv := &GrantRole{
		base: newRedirectBase(),
		plan: plan,
	}

	rv.output = rv
	return rv
}

func (this *GrantRole) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitGrantRole(this)
}

func (this *GrantRole) Copy() Operator {
	return &GrantRole{this.base.copy(), this.plan}
}

func (this *GrantRole) RunOnce(context *Context, parent value.Value) {
	this.once.Do(func() {
		defer context.Recover() // Recover from any panic
		this.switchPhase(_EXECTIME)
		defer this.switchPhase(_NOTIME)
		defer close(this.itemChannel) // Broadcast that I have stopped
		defer this.notify()           // Notify that I have stopped

		if context.Readonly() {
			return
		}

		planRoles := this.plan.Node().Roles()
		roleList := make([]*auth.Role, len(planRoles))
		for i, planRole := range planRoles {
			roleList[i] = &auth.Role{Name: planRole.Role, Keyspace: planRole.Bucket}
		}

		err := expression.DoGrantRole(this.plan.Node().Users(), roleList, context)

		if err != nil {
			context.Error(err)
		}
	})
}

func (this *GrantRole) MarshalJSON() ([]byte, error) {
	r := this.plan.MarshalBase(func(r map[string]interface{}) {
		this.marshalTimes(r)
	})
	return json.Marshal(r)
}