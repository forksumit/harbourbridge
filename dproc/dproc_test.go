package dproc

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudspannerecosystem/harbourbridge/internal"
	"github.com/cloudspannerecosystem/harbourbridge/profiles"
	"github.com/cloudspannerecosystem/harbourbridge/schema"
	"github.com/cloudspannerecosystem/harbourbridge/spanner/ddl"
	"github.com/stretchr/testify/assert"
)

func TestGetDataprocRequestParams(t *testing.T) {
	type paramsStruct struct {
		sourceProfile profiles.SourceProfile
		targetProfile profiles.TargetProfile
		srcSchema     string
		srcTable      string
		primaryKeys   string
		location      string
		subnet        string
	}

	sourceProfileMySQL, _ := profiles.NewSourceProfile(fmt.Sprintf("config=%v", filepath.Join("..", "test_data", "mysql_shard_dataproc.cfg")), "mysql")
	sourceProfileMySQL.Driver, _ = sourceProfileMySQL.ToLegacyDriver("mysql")

	sourceProfileOracle, _ := profiles.NewSourceProfile(fmt.Sprintf("config=%v", filepath.Join("..", "test_data", "mysql_shard_dataproc.cfg")), "oracle")
	sourceProfileOracle.Driver, _ = sourceProfileMySQL.ToLegacyDriver("oracle")

	targetProfile, _ := profiles.NewTargetProfile("project=test-project,instance=sp-test-instance,dbName=sp-test-db")

	tableId := "t1"
	colDefs := make(map[string]schema.Column)
	colIds := []string{"c1", "c2", "c3", "c4"}
	colDefs["c1"] = schema.Column{Name: "pk1"}
	colDefs["c2"] = schema.Column{Name: "pk2"}
	colDefs["c3"] = schema.Column{Name: "col3"}
	colDefs["c4"] = schema.Column{Name: "col4"}

	testCases := []struct {
		name          string
		params        paramsStruct
		want          DataprocRequestParams
		errorExpected bool
	}{
		{
			name: "empty/invalid source profile and target profile",
			params: paramsStruct{
				sourceProfile: profiles.SourceProfile{},
				targetProfile: profiles.TargetProfile{},
				srcSchema:     "test_schema",
				srcTable:      "test_table",
				primaryKeys:   "pk1,pk2",
				location:      "us-central1",
				subnet:        "projects/test-project/regions/us-central1/subnetworks/test-subnet",
			},
			want:          DataprocRequestParams{},
			errorExpected: true,
		},
		{
			name: "valid oracle source profile and target profile",
			params: paramsStruct{
				sourceProfile: sourceProfileOracle,
				targetProfile: targetProfile,
				srcSchema:     "test_schema",
				srcTable:      "test_table",
				primaryKeys:   "pk1,pk2",
				location:      "us-central1",
				subnet:        "projects/test-project/regions/us-central1/subnetworks/test-subnet",
			},
			want:          DataprocRequestParams{},
			errorExpected: true,
		},
		{
			name: "valid mysql source profile and target profile",
			params: paramsStruct{
				sourceProfile: sourceProfileMySQL,
				targetProfile: targetProfile,
				srcSchema:     "test_schema",
				srcTable:      "test_table",
				primaryKeys:   "pk1,pk2",
				location:      "us-central1",
				subnet:        "projects/test-project/regions/us-central1/subnetworks/test-subnet",
			},
			want: DataprocRequestParams{
				Project: "test-project",
				TemplateArgs: []string{"--template",
					"JDBCTOSPANNER",
					"--templateProperty",
					"project.id=test-project",
					"--templateProperty",
					"jdbctospanner.jdbc.url=jdbc:mysql://0.0.0.0:3306/test?user=test&password=test",
					"--templateProperty",
					"jdbctospanner.jdbc.driver.class.name=com.mysql.jdbc.Driver",
					"--templateProperty",
					"jdbctospanner.sql=select t.pk1, t.pk2, t.col3, t.col4 from test_schema.test_table as t",
					"--templateProperty",
					"jdbctospanner.output.instance=sp-test-instance",
					"--templateProperty",
					"jdbctospanner.output.database=sp-test-db",
					"--templateProperty",
					"jdbctospanner.output.table=test_table",
					"--templateProperty",
					"jdbctospanner.output.primaryKey=pk1,pk2",
					"--templateProperty",
					"jdbctospanner.output.saveMode=Append",
					"--templateProperty",
					"jdbctospanner.output.batch.size=500",
					"--templateProperty",
					"jdbctospanner.jdbc.fetchsize=500"},
				JarFileUris: []string{"file:///usr/lib/spark/external/spark-avro.jar",
					"gs://dataproc-templates-binaries/latest/java/dataproc-templates.jar",
					"gs://dataproc-templates/jars/mysql-connector-java.jar"},
				SubnetworkUri: "projects/test-project/regions/us-central1/subnetworks/test-subnet",
				Location:      "us-central1",
			},
			errorExpected: false,
		},
		{
			name: "valid mysql source profile and target profile with missing primary key in source",
			params: paramsStruct{
				sourceProfile: sourceProfileMySQL,
				targetProfile: targetProfile,
				srcSchema:     "test_schema",
				srcTable:      "test_table",
				primaryKeys:   "",
				location:      "us-central1",
				subnet:        "projects/test-project/regions/us-central1/subnetworks/test-subnet",
			},
			want: DataprocRequestParams{
				Project: "test-project",
				TemplateArgs: []string{"--template",
					"JDBCTOSPANNER",
					"--templateProperty",
					"project.id=test-project",
					"--templateProperty",
					"jdbctospanner.jdbc.url=jdbc:mysql://0.0.0.0:3306/test?user=test&password=test",
					"--templateProperty",
					"jdbctospanner.jdbc.driver.class.name=com.mysql.jdbc.Driver",
					"--templateProperty",
					"jdbctospanner.sql=select uuid() as synth_id, t.pk1, t.pk2, t.col3, t.col4 from test_schema.test_table as t",
					"--templateProperty",
					"jdbctospanner.output.instance=sp-test-instance",
					"--templateProperty",
					"jdbctospanner.output.database=sp-test-db",
					"--templateProperty",
					"jdbctospanner.output.table=test_table",
					"--templateProperty",
					"jdbctospanner.output.primaryKey=synth_id",
					"--templateProperty",
					"jdbctospanner.output.saveMode=Append",
					"--templateProperty",
					"jdbctospanner.output.batch.size=500",
					"--templateProperty",
					"jdbctospanner.jdbc.fetchsize=500"},
				JarFileUris: []string{"file:///usr/lib/spark/external/spark-avro.jar",
					"gs://dataproc-templates-binaries/latest/java/dataproc-templates.jar",
					"gs://dataproc-templates/jars/mysql-connector-java.jar"},
				SubnetworkUri: "projects/test-project/regions/us-central1/subnetworks/test-subnet",
				Location:      "us-central1",
			},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		conv := internal.MakeConv()
		conv.SrcSchema[tableId] = schema.Table{Name: tc.params.srcTable, Schema: tc.params.srcSchema, ColIds: colIds, ColDefs: colDefs}
		conv.SpSchema[tableId] = ddl.CreateTable{Name: tc.params.srcTable, ColIds: colIds}
		got, err := GetDataprocRequestParams(conv, tc.params.sourceProfile, tc.params.targetProfile, tableId, tc.params.primaryKeys, tc.params.location, tc.params.subnet)
		assert.Equal(t, got, tc.want, tc.name)
		assert.Equal(t, tc.errorExpected, err != nil)
	}
}
