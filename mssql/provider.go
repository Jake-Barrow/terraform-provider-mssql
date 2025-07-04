package mssql

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Jake-Barrow/terraform-provider-mssql/mssql/model"
	"github.com/Jake-Barrow/terraform-provider-mssql/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type mssqlProvider struct {
	factory model.ConnectorFactory
	logger  *zerolog.Logger
}

const (
	providerLogFile = "terraform-provider-mssql.log"
)

var (
	defaultTimeout = schema.DefaultTimeout(30 * time.Second)
)

func New(version, commit string) func() *schema.Provider {
	return func() *schema.Provider {
		return Provider(sql.GetFactory())
	}
}

func Provider(factory model.ConnectorFactory) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"debug": {
				Type:        schema.TypeBool,
				Description: fmt.Sprintf("Enable provider debug logging (logs to file %s)", providerLogFile),
				Optional:    true,
				Default:     false,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"mssql_login": resourceLogin(),
			"mssql_user": resourceUser(),
			"mssql_database_permissions": resourceDatabasePermissions(),
			"mssql_database_role": resourceDatabaseRole(),
			"mssql_database_schema": resourceDatabaseSchema(),
			"mssql_database_masterkey": resourceDatabaseMasterkey(),
			"mssql_database_credential": resourceDatabaseCredential(),
			"mssql_azure_external_datasource": resourceAzureExternalDatasource(),
			"mssql_database_sqlscript": resourceDatabaseSQLScript(),
			"mssql_entraid_login": resourceEntraIDLogin(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"mssql_login": dataSourceLogin(),
			"mssql_user": dataSourceUser(),
			"mssql_database_permissions": dataSourceDatabasePermissions(),
			"mssql_database_role": dataSourceDatabaseRole(),
			"mssql_database_schema": dataSourceDatabaseSchema(),
			"mssql_database_credential": datasourceDatabaseCredential(),
			"mssql_azure_external_datasource": datasourceAzureExternalDatasource(),
			"mssql_entraid_login": dataSourceEntraIDLogin(),
		},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return providerConfigure(ctx, data, factory)
		},
	}
}

func providerConfigure(ctx context.Context, data *schema.ResourceData, factory model.ConnectorFactory) (model.Provider, diag.Diagnostics) {
	isDebug := data.Get("debug").(bool)
	logger := newLogger(isDebug)

	logger.Info().Msg("Created provider")

	return mssqlProvider{factory: factory, logger: logger}, nil
}

func (p mssqlProvider) GetConnector(prefix string, data *schema.ResourceData) (interface{}, error) {
	return p.factory.GetConnector(prefix, data)
}

func (p mssqlProvider) ResourceLogger(resource, function string) zerolog.Logger {
	return p.logger.With().Str("resource", resource).Str("func", function).Logger()
}

func (p mssqlProvider) DataSourceLogger(datasource, function string) zerolog.Logger {
	return p.logger.With().Str("datasource", datasource).Str("func", function).Logger()
}

func newLogger(isDebug bool) *zerolog.Logger {
	var writer io.Writer = nil
	logLevel := zerolog.Disabled
	if isDebug {
		f, err := os.OpenFile(providerLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Err(err).Msg("error opening file")
		}
		writer = f
		logLevel = zerolog.DebugLevel
	}
	logger := zerolog.New(writer).Level(logLevel).With().Timestamp().Logger()
	return &logger
}
