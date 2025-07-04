package mssql

import (
	"bytes"
	"context"
	sql2 "database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/Jake-Barrow/terraform-provider-mssql/mssql/model"
	"github.com/Jake-Barrow/terraform-provider-mssql/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var runLocalAccTests bool
var testAccProvider *schema.Provider
var testAccProviders map[string]func() (*schema.Provider, error)

func init() {
	_, runLocalAccTests = os.LookupEnv("TF_ACC_LOCAL")
	testAccProvider = Provider(sql.GetFactory())
	testAccProviders = map[string]func() (*schema.Provider, error){
		"mssql": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := testAccProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	var keys []string
	_, azure := os.LookupEnv("TF_ACC")
	_, local := os.LookupEnv("TF_ACC_LOCAL")
	if local || azure {
		keys = append(keys, "MSSQL_USERNAME", "MSSQL_PASSWORD")
	}
	if azure {
		keys = append(keys, "MSSQL_TENANT_ID", "MSSQL_CLIENT_ID", "MSSQL_CLIENT_SECRET", "TF_ACC_SQL_SERVER", "TF_ACC_AZURE_USER_CLIENT_ID", "TF_ACC_AZURE_USER_CLIENT_SECRET")
	}
	for _, key := range keys {
		if v := os.Getenv(key); v == "" {
			t.Fatalf("Environment variable %s must be set for acceptance tests", key)
		}
	}
}

type Check struct {
	name, op string
	expected interface{}
}

type TestConnector interface {
	GetLogin(name string) (*model.Login, error)
	GetUser(database, name string) (*model.User, error)
	GetDatabasePermissions(database, name string) (*model.DatabasePermissions, error)
	GetDatabaseRole(database, name string) (*model.DatabaseRole, error)
	GetDatabaseSchema(database, name string) (*model.DatabaseSchema, error)
	GetDatabaseCredential(database, name string) (*model.DatabaseCredential, error)
	GetAzureExternalDatasource(database, name string) (*model.AzureExternalDatasource, error)
	GetDatabaseMasterkey(database string) (*model.DatabaseMasterkey, error)
	DataBaseExecuteScript(database string, sqlscript string) error
	GetEntraIDLogin(name string) (*model.EntraIDLogin, error)
	GetSystemUser() (string, error)
	GetCurrentUser(database string) (string, string, error)
}

type testConnector struct {
	c interface{}
}

func getTestConnector(a map[string]string) (TestConnector, error) {
	prefix := serverProp + ".0."

	connector := &sql.Connector{
		Host:    a[prefix+"host"],
		Port:    a[prefix+"port"],
		Timeout: 60 * time.Second,
	}

	if username, ok := a[prefix+"login.0.username"]; ok {
		connector.Login = &sql.LoginUser{
			Username: username,
			Password: a[prefix+"login.0.password"],
		}
	}

	if tenantId, ok := a[prefix+"azure_login.0.tenant_id"]; ok {
		connector.AzureLogin = &sql.AzureLogin{
			TenantID:     tenantId,
			ClientID:     a[prefix+"azure_login.0.client_id"],
			ClientSecret: a[prefix+"azure_login.0.client_secret"],
		}
	}

	return testConnector{c: connector}, nil
}

func getTestLoginConnector(a map[string]string) (TestConnector, error) {
	prefix := serverProp + ".0."
	connector := &sql.Connector{
		Host:    a[prefix+"host"],
		Port:    a[prefix+"port"],
		Timeout: 60 * time.Second,
	}
	if password, ok := a[passwordProp]; ok {
		connector.Login = &sql.LoginUser{
			Username: a[loginNameProp],
			Password: password,
		}
	}

	return testConnector{c: connector}, nil
}

func getTestUserConnector(a map[string]string, username, password string) (TestConnector, error) {
	prefix := serverProp + ".0."
	connector := &sql.Connector{
		Host:    a[prefix+"host"],
		Port:    a[prefix+"port"],
		Timeout: 60 * time.Second,
	}
	connector.Login = &sql.LoginUser{
		Username: username,
		Password: password,
	}
	if database, ok := a[databaseProp]; ok {
		connector.Database = database
	}

	return testConnector{c: connector}, nil
}

func getTestExternalConnector(a map[string]string, tenantId, clientId, clientSecret string) (TestConnector, error) {
	prefix := serverProp + ".0."
	connector := &sql.Connector{
		Host:    a[prefix+"host"],
		Port:    a[prefix+"port"],
		Timeout: 60 * time.Second,
	}
	connector.AzureLogin = &sql.AzureLogin{
		TenantID:     tenantId,
		ClientID:     clientId,
		ClientSecret: clientSecret,
	}
	if database, ok := a[databaseProp]; ok {
		connector.Database = database
	}

	return testConnector{c: connector}, nil
}

func (t testConnector) GetLogin(name string) (*model.Login, error) {
	return t.c.(LoginConnector).GetLogin(context.Background(), name)
}

func (t testConnector) GetUser(database, name string) (*model.User, error) {
	return t.c.(UserConnector).GetUser(context.Background(), database, name)
}

func (t testConnector) GetDatabasePermissions(database, name string) (*model.DatabasePermissions, error) {
	return t.c.(DatabasePermissionsConnector).GetDatabasePermissions(context.Background(), database, name)
}

func (t testConnector) GetDatabaseRole(database string, roleName string) (*model.DatabaseRole, error) {
	return t.c.(DatabaseRoleConnector).GetDatabaseRole(context.Background(), database, roleName)
}

func (t testConnector) GetDatabaseSchema(database string, schemaName string) (*model.DatabaseSchema, error) {
	return t.c.(DatabaseSchemaConnector).GetDatabaseSchema(context.Background(), database, schemaName)
}

func (t testConnector) GetDatabaseCredential(database, credentialName string) (*model.DatabaseCredential, error) {
	return t.c.(DatabaseCredentialConnector).GetDatabaseCredential(context.Background(), database, credentialName)
}

func (t testConnector) GetAzureExternalDatasource(database, datasourceName string) (*model.AzureExternalDatasource, error) {
	return t.c.(AzureExternalDatasourceConnector).GetAzureExternalDatasource(context.Background(), database, datasourceName)
}

func (t testConnector) GetDatabaseMasterkey(database string) (*model.DatabaseMasterkey, error) {
	return t.c.(DatabaseMasterkeyConnector).GetDatabaseMasterkey(context.Background(), database)
}

func (t testConnector) DataBaseExecuteScript(database, sqlscript string) error {
	return t.c.(DatabaseSQLScriptConnector).DataBaseExecuteScript(context.Background(), database, sqlscript)
}

func (t testConnector) GetEntraIDLogin(name string) (*model.EntraIDLogin, error) {
	return t.c.(EntraIDLoginConnector).GetEntraIDLogin(context.Background(), name)
}

func (t testConnector) GetSystemUser() (string, error) {
	var user string
	err := t.c.(*sql.Connector).QueryRowContext(context.Background(), "SELECT SYSTEM_USER;", func(row *sql2.Row) error {
		return row.Scan(&user)
	})
	return user, err
}

func (t testConnector) GetCurrentUser(database string) (string, string, error) {
	if database == "" {
		database = "master"
	}
	t.c.(*sql.Connector).Database = database
	var current, system string
	err := t.c.(*sql.Connector).QueryRowContext(context.Background(), "SELECT CURRENT_USER, SYSTEM_USER;", func(row *sql2.Row) error {
		return row.Scan(&current, &system)
	})
	return current, system, err
}

func templateToString(name, text string, data interface{}) (string, error) {
	t, err := template.New(name).Parse(text)
	if err != nil {
		return "", err
	}
	var doc bytes.Buffer
	if err = t.Execute(&doc, data); err != nil {
		return "", err
	}
	return doc.String(), nil
}

func testAccImportStateId(resource string, azure bool) func(state *terraform.State) (string, error) {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no record ID is set")
		}
		return rs.Primary.ID + "?azure=" + strconv.FormatBool(azure), nil
	}
}
