package sql

import (
	"context"
	"database/sql"

	"github.com/Jake-Barrow/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetDatabaseSchema(ctx context.Context, database, schemaName string) (*model.DatabaseSchema, error) {
	cmd := `SELECT
				dp1.schema_id,
				dp1.name,
				dp1.principal_id,
				CASE
					WHEN @@VERSION LIKE 'Microsoft SQL Azure%'
						AND @database = 'master'
						AND (@ownerName = 'dbo' OR @ownerName = '') THEN ''
					ELSE dp2.name
				END AS name
			FROM [sys].[schemas] dp1
			INNER JOIN [sys].[database_principals] dp2
				ON dp1.principal_id = dp2.principal_id
			WHERE dp1.name = @schemaName`
	var sqlschema model.DatabaseSchema
	err := c.
		setDatabase(&database).
		QueryRowContext(ctx, cmd,
			func(r *sql.Row) error {
				return r.Scan(&sqlschema.SchemaID, &sqlschema.SchemaName, &sqlschema.OwnerId, &sqlschema.OwnerName)
			},
			sql.Named("database", database),
			sql.Named("schemaName", schemaName),
			sql.Named("ownerName", sqlschema.OwnerName),
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &sqlschema, nil
}

func (c *Connector) CreateDatabaseSchema(ctx context.Context, database, schemaName string, ownerName string) error {
	cmd := `DECLARE @sql nvarchar(max);
			IF @ownerName = 'dbo' OR @ownerName = ''
				BEGIN
					SET @sql = 'CREATE SCHEMA ' + @schemaName
				END
			ELSE
				BEGIN
					SET @sql = 'CREATE SCHEMA ' + @schemaName + ' AUTHORIZATION ' + @ownerName
				END
			EXEC (@sql);`

	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("schemaName", schemaName),
			sql.Named("ownerName", ownerName),
			sql.Named("database", database),
		)
}

func (c *Connector) DeleteDatabaseSchema(ctx context.Context, database, schemaName string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE @sql NVARCHAR(max)
			DECLARE @user_name NVARCHAR(max) = (SELECT USER_NAME())
			IF @@VERSION LIKE 'Microsoft SQL Azure%' AND @database = 'master'
				BEGIN
					SET @stmt = 'IF EXISTS (SELECT 1 FROM [sys].[schemas] WHERE [name] = ' + QuoteName(@schemaName, '''') + ') ' +
								'DROP SCHEMA ' + @schemaName
				END
			ELSE
				BEGIN
					SET @sql =  'IF EXISTS (SELECT 1 FROM [sys].[schemas] dp1 INNER JOIN [sys].[database_principals] dp2 ON dp1.principal_id = dp2.principal_id AND dp1.name = ' + QuoteName(@schemaName, '''') + ') ' +
								'ALTER AUTHORIZATION ON SCHEMA:: [' + @schemaName + '] TO [' + @user_name + ']'
					EXEC sp_executesql @sql;
					SET @stmt = 'IF EXISTS (SELECT 1 FROM [sys].[schemas] WHERE [name] = ' + QuoteName(@schemaName, '''') + ') ' +
								'DROP SCHEMA ' + @schemaName
				END
			EXEC (@stmt)`

	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("database", database),
			sql.Named("schemaName", schemaName),
		)
}

func (c *Connector) UpdateDatabaseSchema(ctx context.Context, database string, schemaName string, ownerName string) error {
	cmd := `DECLARE @sql NVARCHAR(max)
			IF @ownerName = 'dbo' OR @ownerName = ''
				BEGIN
					SET @ownerName = (SELECT USER_NAME())
				END
			SET @sql = 'ALTER AUTHORIZATION ON SCHEMA:: [' + @schemaName + '] TO [' + @ownerName + ']'
			EXEC (@sql)`

	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("schemaName", schemaName),
			sql.Named("ownerName", ownerName),
		)
}
