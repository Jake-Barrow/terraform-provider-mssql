package mssql

import (
	"context"
	"strings"

	"github.com/Jake-Barrow/terraform-provider-mssql/mssql/model"
	"github.com/Jake-Barrow/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceDatabaseSchema() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseSchemaCreate,
		ReadContext:   resourceDatabaseSchemaRead,
		UpdateContext: resourceDatabaseSchemaUpdate,
		DeleteContext: resourceDatabaseSchemaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseSchemaImport,
		},
		Schema: map[string]*schema.Schema{
			serverProp: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: getServerSchema(serverProp),
				},
			},
			ignoreDeletionProp: {
				Type:     schema.TypeBool,
				Optional: true,
        Default:  false,
			},
			databaseProp: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  defaultDatabaseDefault,
			},
			schemaNameProp: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			ownerNameProp: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  defaultDboPropDefault,
				DiffSuppressFunc: func(k, old, new string, data *schema.ResourceData) bool {
					return (old == "" && new == defaultDboPropDefault) || (old == defaultDboPropDefault && new == "")
				},
			},
			schemaIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			ownerIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: defaultTimeout,
			Read:   defaultTimeout,
			Update: defaultTimeout,
			Delete: defaultTimeout,
		},
	}
}

type DatabaseSchemaConnector interface {
	CreateDatabaseSchema(ctx context.Context, database string, schemaName string, ownerName string) error
	GetDatabaseSchema(ctx context.Context, database, schemaName string) (*model.DatabaseSchema, error)
	UpdateDatabaseSchema(ctx context.Context, database string, schemaName string, ownerName string) error
	DeleteDatabaseSchema(ctx context.Context, database, schemaName string) error
	DatabaseExists(ctx context.Context, database string) (bool, error)
}

func resourceDatabaseSchemaCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "schema", "create")
	logger.Debug().Msgf("Create %s", getDatabaseSchemaID(data))

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	ownerName := data.Get(ownerNameProp).(string)

	connector, err := getDatabaseSchemaConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateDatabaseSchema(ctx, database, schemaName, ownerName); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create schema [%s].[%s]", database, schemaName))
	}

	data.SetId(getDatabaseSchemaID(data))

	logger.Info().Msgf("created schema [%s].[%s]", database, schemaName)

	return resourceDatabaseSchemaRead(ctx, data, meta)
}

func resourceDatabaseSchemaRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "schema", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)

	connector, err := getDatabaseSchemaConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if database exists
	exists, err := connector.DatabaseExists(ctx, database)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to check if database [%s] exists", database))
	}
	if !exists {
		logger.Info().Msgf("Database [%s] does not exist", database)
		data.SetId("")
		return nil
	}

	sqlschema, err := connector.GetDatabaseSchema(ctx, database, schemaName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to get schema [%s].[%s]", database, schemaName))
	}

	if sqlschema == nil {
		logger.Info().Msgf("schema [%s].[%s] does not exist", database, schemaName)
		data.SetId("")
	} else {
		if err = data.Set(schemaIdProp, sqlschema.SchemaID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(schemaNameProp, sqlschema.SchemaName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(ownerNameProp, sqlschema.OwnerName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(ownerIdProp, sqlschema.OwnerId); err != nil {
			return diag.FromErr(err)
		}
	}

	logger.Info().Msgf("read schema [%s].[%s]", database, schemaName)

	return nil
}

func resourceDatabaseSchemaDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
  ignoreDeletion := data.Get(ignoreDeletionProp).(bool)

  if ignoreDeletion {
    return nil
  }

	logger := loggerFromMeta(meta, "schema", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)

	connector, err := getDatabaseSchemaConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteDatabaseSchema(ctx, database, schemaName); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete schema [%s].[%s]", database, schemaName))
	}

	data.SetId("")

	logger.Info().Msgf("deleted schema [%s].[%s]", database, schemaName)

	return nil
}

func resourceDatabaseSchemaUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "schema", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	ownerName := data.Get(ownerNameProp).(string)

	// Store old values for all properties that might change
	oldValues := make(map[string]interface{})
	for _, prop := range []string{ownerNameProp} {
		if data.HasChange(prop) {
			oldValue, _ := data.GetChange(prop)
			oldValues[prop] = oldValue
		}
	}

	connector, err := getDatabaseSchemaConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.UpdateDatabaseSchema(ctx, database, schemaName, ownerName); err != nil {
		// If update fails, revert all changed values in the state
		for prop, oldValue := range oldValues {
			if err := data.Set(prop, oldValue); err != nil {
				logger.Error().Err(err).Msgf("Failed to revert %s state after update error", prop)
			}
		}
		return diag.FromErr(errors.Wrapf(err, "unable to update schema [%s].[%s]", database, schemaName))
	}

	data.SetId(getDatabaseSchemaID(data))

	logger.Info().Msgf("updated schema [%s].[%s]", database, schemaName)

	return resourceDatabaseSchemaRead(ctx, data, meta)
}

func resourceDatabaseSchemaImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "schema", "import")
	logger.Debug().Msgf("Import %s", data.Id())

	server, u, err := serverFromId(data.Id())
	if err != nil {
		return nil, err
	}
	if err := data.Set(serverProp, server); err != nil {
		return nil, err
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) != 4 {
		return nil, errors.New("invalid ID")
	}
	if err = data.Set(databaseProp, parts[1]); err != nil {
		return nil, err
	}
	if err = data.Set(schemaNameProp, parts[3]); err != nil {
		return nil, err
	}

	data.SetId(getDatabaseSchemaID(data))

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)

	connector, err := getDatabaseSchemaConnector(meta, data)
	if err != nil {
		return nil, err
	}

	sqlschema, err := connector.GetDatabaseSchema(ctx, database, schemaName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get schema [%s].[%s]", database, schemaName)
	}

	if sqlschema == nil {
		return nil, errors.Errorf("schema [%s].[%s] does not exist", database, schemaName)
	}

	if err = data.Set(schemaIdProp, sqlschema.SchemaID); err != nil {
		return nil, err
	}
	if err = data.Set(ownerNameProp, sqlschema.OwnerName); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getDatabaseSchemaConnector(meta interface{}, data *schema.ResourceData) (DatabaseSchemaConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(DatabaseSchemaConnector), nil
}
