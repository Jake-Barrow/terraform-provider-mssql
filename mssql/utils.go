package mssql

import (
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/rs/zerolog"
  "github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func getLoginID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  loginName := data.Get(loginNameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s", host, port, loginName)
}

func getUserID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  username := data.Get(usernameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, username)
}

func getDatabasePermissionsID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  principalId := data.Get(principalIdProp).(int)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%d/%s", host, port, database, principalId, "permissions")
}

func getRoleID(data *schema.ResourceData) string {
  host := data.Get(serverProp + ".0.host").(string)
  port := data.Get(serverProp + ".0.port").(string)
  database := data.Get(databaseProp).(string)
  roleName := data.Get(roleNameProp).(string)
  return fmt.Sprintf("sqlserver://%s:%s/%s/%s", host, port, database, roleName)
}

func loggerFromMeta(meta interface{}, resource, function string) zerolog.Logger {
  return meta.(model.Provider).ResourceLogger(resource, function)
}

func toStringSlice(values []interface{}) []string {
  result := make([]string, len(values))
  for i, v := range values {
    result[i] = v.(string)
  }
  return result
}
