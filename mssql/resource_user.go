package mssql

import (
	"context"
	"strings"

	"github.com/Jake-Barrow/terraform-provider-mssql/mssql/model"
	"github.com/Jake-Barrow/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserImport,
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
			usernameProp: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			objectIdProp: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, data *schema.ResourceData) bool {
					return old == new
				},
			},
			typeStrProp: {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringInSlice([]string{"E", "X"}, false),
				ConflictsWith: []string{loginNameProp, passwordProp},
				RequiredWith:  []string{objectIdProp},
			},
			loginNameProp: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{passwordProp, objectIdProp},
				ValidateFunc:  validate.SQLIdentifier,
			},
			passwordProp: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validate.SQLIdentifierPassword,
			},
			sidStrProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			authenticationTypeProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			principalIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			defaultSchemaProp: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultDboPropDefault,
				ValidateFunc: validate.SQLIdentifier,
			},
			defaultLanguageProp: {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, data *schema.ResourceData) bool {
					return data.Get(authenticationTypeProp) == "INSTANCE" || old == new
				},
			},
			rolesProp: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

type UserConnector interface {
	CreateUser(ctx context.Context, database string, user *model.User) error
	GetUser(ctx context.Context, database, username string) (*model.User, error)
	UpdateUser(ctx context.Context, database string, user *model.User) error
	DeleteUser(ctx context.Context, database, username string) error
	DatabaseExists(ctx context.Context, database string) (bool, error)
}

func resourceUserCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "user", "create")
	logger.Debug().Msgf("Create %s", getUserID(data))

	database := data.Get(databaseProp).(string)
	username := data.Get(usernameProp).(string)
	objectId := data.Get(objectIdProp).(string)
	loginName := data.Get(loginNameProp).(string)
	password := data.Get(passwordProp).(string)
	typeStr := data.Get(typeStrProp).(string)
	defaultSchema := data.Get(defaultSchemaProp).(string)
	defaultLanguage := data.Get(defaultLanguageProp).(string)
	roles := data.Get(rolesProp).(*schema.Set).List()

	connector, err := getUserConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	user := &model.User{
		Username:        username,
		ObjectId:        objectId,
		LoginName:       loginName,
		Password:        password,
		TypeStr:         typeStr,
		DefaultSchema:   defaultSchema,
		DefaultLanguage: defaultLanguage,
		Roles:           toStringSlice(roles),
	}
	if err = connector.CreateUser(ctx, database, user); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create user [%s].[%s]", database, username))
	}

	data.SetId(getUserID(data))

	logger.Info().Msgf("created user [%s].[%s]", database, username)

	return resourceUserRead(ctx, data, meta)
}

func resourceUserRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "user", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	username := data.Get(usernameProp).(string)

	connector, err := getUserConnector(meta, data)
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

	user, err := connector.GetUser(ctx, database, username)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read user [%s].[%s]", database, username))
	}
	if user == nil {
		logger.Info().Msgf("No user found for [%s].[%s]", database, username)
		data.SetId("")
	} else {
		if err = data.Set(loginNameProp, user.LoginName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(sidStrProp, user.SIDStr); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(typeStrProp, user.TypeStr); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(authenticationTypeProp, user.AuthType); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(principalIdProp, user.PrincipalID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(defaultSchemaProp, user.DefaultSchema); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(defaultLanguageProp, user.DefaultLanguage); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(rolesProp, user.Roles); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceUserUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "user", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	database := data.Get(databaseProp).(string)
	username := data.Get(usernameProp).(string)
	defaultSchema := data.Get(defaultSchemaProp).(string)
	defaultLanguage := data.Get(defaultLanguageProp).(string)
	roles := data.Get(rolesProp).(*schema.Set).List()

	// Store old values for all properties that might change
	oldValues := make(map[string]interface{})
	for _, prop := range []string{passwordProp, defaultSchemaProp, defaultLanguageProp} {
		if data.HasChange(prop) {
			oldValue, _ := data.GetChange(prop)
			oldValues[prop] = oldValue
		}
	}
	// Handle roles separately since it's a Set type
	if data.HasChange(rolesProp) {
		oldValue, _ := data.GetChange(rolesProp)
		if oldSet, ok := oldValue.(*schema.Set); ok {
			oldValues[rolesProp] = oldSet.List()
		}
	}

	connector, err := getUserConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	user := &model.User{
		Username:        username,
		DefaultSchema:   defaultSchema,
		DefaultLanguage: defaultLanguage,
		Roles:           toStringSlice(roles),
	}

	// Only include password in the update if it has changed
	if data.HasChange(passwordProp) {
		user.Password = data.Get(passwordProp).(string)
	}

	if err = connector.UpdateUser(ctx, database, user); err != nil {
		// If update fails, revert all changed values in the state
		for prop, oldValue := range oldValues {
			if err := data.Set(prop, oldValue); err != nil {
				logger.Error().Err(err).Msgf("Failed to revert %s state after update error", prop)
			}
		}
		return diag.FromErr(errors.Wrapf(err, "unable to update user [%s].[%s]", database, username))
	}

	data.SetId(getUserID(data))

	logger.Info().Msgf("updated user [%s].[%s]", database, username)

	return resourceUserRead(ctx, data, meta)
}

func resourceUserDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ignoreDeletion := data.Get(ignoreDeletionProp).(bool)

  if ignoreDeletion {
    return nil
  }

  logger := loggerFromMeta(meta, "user", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)
	username := data.Get(usernameProp).(string)

	connector, err := getUserConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteUser(ctx, database, username); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete user [%s].[%s]", database, username))
	}

	logger.Info().Msgf("deleted user [%s].[%s]", database, username)

	// d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
	data.SetId("")

	return nil
}

func resourceUserImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "user", "import")
	logger.Debug().Msgf("Import %s", data.Id())

	server, u, err := serverFromId(data.Id())
	if err != nil {
		return nil, err
	}
	if err = data.Set(serverProp, server); err != nil {
		return nil, err
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) != 4 {
		return nil, errors.New("invalid ID")
	}
	if err = data.Set(databaseProp, parts[1]); err != nil {
		return nil, err
	}
	if err = data.Set(usernameProp, parts[3]); err != nil {
		return nil, err
	}

	data.SetId(getUserID(data))

	database := data.Get(databaseProp).(string)
	username := data.Get(usernameProp).(string)

  data.Set(ignoreDeletionProp, false)

	connector, err := getUserConnector(meta, data)
	if err != nil {
		return nil, err
	}

	login, err := connector.GetUser(ctx, database, username)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read user [%s].[%s] for import", database, username)
	}

	if login == nil {
		return nil, errors.Errorf("no user [%s].[%s] found for import", database, username)
	}

	if err = data.Set(authenticationTypeProp, login.AuthType); err != nil {
		return nil, err
	}
	if err = data.Set(typeStrProp, login.TypeStr); err != nil {
		return nil, err
	}
	if err = data.Set(principalIdProp, login.PrincipalID); err != nil {
		return nil, err
	}
	if err = data.Set(defaultSchemaProp, login.DefaultSchema); err != nil {
		return nil, err
	}
	if err = data.Set(defaultLanguageProp, login.DefaultLanguage); err != nil {
		return nil, err
	}
	if err = data.Set(rolesProp, login.Roles); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getUserConnector(meta interface{}, data *schema.ResourceData) (UserConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(UserConnector), nil
}
