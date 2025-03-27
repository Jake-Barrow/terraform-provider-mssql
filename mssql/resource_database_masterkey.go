package mssql

import (
	"context"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/ValeruS/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceDatabaseMasterkey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseMasterkeyCreate,
		ReadContext:   resourceDatabaseMasterkeyRead,
		UpdateContext: resourceDatabaseMasterkeyUpdate,
		DeleteContext: resourceDatabaseMasterkeyDelete,
		Schema: map[string]*schema.Schema{
			serverProp: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: getServerSchema(serverProp),
				},
			},
			databaseProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			passwordProp: {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validate.SQLIdentifierPassword,
			},
			keynameProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			keyguidProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			principalIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			symmetrickeyidProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			keylengthProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			keyalgorithmProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			algorithmdescProp: {
				Type:     schema.TypeString,
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

type DatabaseMasterkeyConnector interface {
	CreateDatabaseMasterkey(ctx context.Context, database, password string) error
	GetDatabaseMasterkey(ctx context.Context, database string) (*model.DatabaseMasterkey, error)
	UpdateDatabaseMasterkey(ctx context.Context, database, password string) error
	DeleteDatabaseMasterkey(ctx context.Context, database string) error
	DatabaseExists(ctx context.Context, database string) (bool, error)
}

func resourceDatabaseMasterkeyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databasemasterkey", "create")
	logger.Debug().Msgf("Create %s", getDatabaseMasterkeyID(data))

	database := data.Get(databaseProp).(string)
	password := data.Get(passwordProp).(string)

	connector, err := getDatabaseMasterkeyConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateDatabaseMasterkey(ctx, database, password); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create database master key on database [%s]", database))
	}

	data.SetId(getDatabaseMasterkeyID(data))

	logger.Info().Msgf("created database master key on database [%s]", database)

	return resourceDatabaseMasterkeyRead(ctx, data, meta)
}

func resourceDatabaseMasterkeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databasemasterkey", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)

	connector, err := getDatabaseMasterkeyConnector(meta, data)
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

	masterkey, err := connector.GetDatabaseMasterkey(ctx, database)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read database master key on database [%s]", database))
	}

	if masterkey == nil {
		logger.Info().Msgf("No database master key found on database [%s]", database)
		data.SetId("")
	} else {
		if err = data.Set(keynameProp, masterkey.KeyName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(principalIdProp, masterkey.PrincipalID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(symmetrickeyidProp, masterkey.SymmetricKeyID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(keylengthProp, masterkey.KeyLength); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(keyalgorithmProp, masterkey.KeyAlgorithm); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(algorithmdescProp, masterkey.AlgorithmDesc); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(keyguidProp, masterkey.KeyGuid); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceDatabaseMasterkeyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databasemasterkey", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	database := data.Get(databaseProp).(string)
	password := data.Get(passwordProp).(string)

	connector, err := getDatabaseMasterkeyConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.UpdateDatabaseMasterkey(ctx, database, password); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to update database key on database [%s]", database))
	}

	data.SetId(getDatabaseMasterkeyID(data))

	logger.Info().Msgf("updated database master key on database [%s]", database)

	return resourceDatabaseMasterkeyRead(ctx, data, meta)
}

func resourceDatabaseMasterkeyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databasemasterkey", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)

	connector, err := getDatabaseMasterkeyConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteDatabaseMasterkey(ctx, database); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete database master key on database [%s]", database))
	}

	data.SetId("")

	logger.Info().Msgf("deleted database master key on database [%s]", database)

	return nil
}

func getDatabaseMasterkeyConnector(meta interface{}, data *schema.ResourceData) (DatabaseMasterkeyConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(DatabaseMasterkeyConnector), nil
}
