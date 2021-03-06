// Copyright 2020 The SQLFlow Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tensorflow

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"sqlflow.org/sqlflow/pkg/database"
	"sqlflow.org/sqlflow/pkg/ir"
	pb "sqlflow.org/sqlflow/pkg/proto"
)

func mockSession() *pb.Session {
	return &pb.Session{DbConnStr: database.GetTestingMySQLURL()}
}
func TestTrainCodegen(t *testing.T) {
	a := assert.New(t)
	tir := ir.MockTrainStmt(false)
	_, err := Train(tir, mockSession())
	a.NoError(err)

	pir := ir.MockPredStmt(tir)

	sess := &pb.Session{
		Token:            "",
		DbConnStr:        "",
		ExitOnSubmit:     false,
		UserId:           "",
		HiveLocation:     "/sqlflowtmp",
		HdfsNamenodeAddr: "192.168.1.1:8020",
		HdfsUser:         "sqlflow_admin",
		HdfsPass:         "sqlflow_pass",
	}
	code, err := Pred(pir, sess)
	a.NoError(err)

	r, _ := regexp.Compile(`hdfs_user="(.*)"`)
	a.Equal(r.FindStringSubmatch(code)[1], "sqlflow_admin")
	r, _ = regexp.Compile(`hdfs_pass="(.*)"`)
	a.Equal(r.FindStringSubmatch(code)[1], "sqlflow_pass")
}

func TestTrainWithModelRepoImage(t *testing.T) {
	a := assert.New(t)
	tir := ir.MockTrainStmt(false)
	tir.ModelImage = "myRepo/MyDNNClassifier:v1.0"
	code, err := Train(tir, mockSession())
	a.NoError(err)
	r, _ := regexp.Compile(`model_repo_image="(.*)"`)
	a.Equal(r.FindStringSubmatch(code)[1], tir.ModelImage)
}

func TestTrainWithOptimizer(t *testing.T) {
	a := assert.New(t)
	tir := ir.MockTrainStmt(false)
	a.NotContains(tir.Attributes, "model.optimizer")
	_, err := Train(tir, mockSession())
	a.NoError(err)
	a.NotContains(tir.Attributes, "model.optimizer")

	tir.Attributes["model.optimizer"] = "RMSprop"
	a.NoError(InitializeAttributes(tir))
	_, err = Train(tir, mockSession())
	a.NoError(err)
	a.Equal(tir.Attributes["model.optimizer"], "RMSprop()")

	tir.Attributes["not_optimizer.learning_rate"] = 123
	tir.Attributes["model.optimizer"] = "RMSprop"
	a.Error(InitializeAttributes(tir))

	tir = ir.MockTrainStmt(false)
	tir.Attributes["optimizer.learning_rate"] = 0.002
	a.NoError(InitializeAttributes(tir))
	_, err = Train(tir, mockSession())
	a.NoError(err)
	a.Equal(tir.Attributes["model.optimizer"], "Adagrad(learning_rate=0.002, )")
	a.NotContains(tir.Attributes, "optimizer.learning_rate")

	tir.Attributes["model.optimizer"] = "RMSprop"
	tir.Attributes["optimizer.learning_rate"] = 0.002
	a.NoError(InitializeAttributes(tir))
	_, err = Train(tir, mockSession())
	a.NoError(err)
	a.Equal(tir.Attributes["model.optimizer"], "RMSprop(learning_rate=0.002, )")
}
