terraform {
  required_version = "~> 1.5"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.47"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.85"
    }
    mssql = {
      source  = "Jake-Barrow/mssql"
      version = "~> 0.2"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.10"
    }
  }
}

provider "azuread" {}

provider "azurerm" {
  features {}
}

provider "mssql" {
  debug = "true"
}

provider "random" {}

variable "prefix" {
  description = "A prefix used when naming Azure resources"
  type        = string
}

variable "sql_servers_group" {
  description = "The name of an Azure AD group assigned the role 'Directory Reader'. The Azure SQL Server will be added to this group to enable external logins."
  type        = string
  default     = "SQL Servers"
}

variable "location" {
  description = "The location of the Azure resources."
  type        = string
  default     = "East US"
}

variable "tenant_id" {
  description = "The tenant id of the Azure AD tenant"
  type        = string
}

variable "local_ip_addresses" {
  description = "The external IP addresses of the machines running the acceptance tests. This is necessary to allow access to the Azure SQL Server resource."
  type        = list(string)
}

#
# Creates an Azure SQL Database running in a temporary resource group on Azure.
#

# Random names and secrets
resource "random_string" "random" {
  length  = 16
  upper   = false
  special = false
}

locals {
  prefix = "${var.prefix}-${substr(random_string.random.result, 0, 4)}"
}

# An Azure AD group assigned the role 'Directory Readers'. The Azure SQL Server needs to be assigned to this group to enable external logins.
data "azuread_group" "sql_servers" {
  display_name = var.sql_servers_group
}

# An Azure AD service principal used as Azure Administrator for the Azure SQL Server resource
resource "azuread_application" "sa" {
  display_name = "${local.prefix}-sa"
  web {
    homepage_url = "https://test.example.com"
  }
}

resource "azuread_service_principal" "sa" {
  client_id = azuread_application.sa.client_id
}

resource "azuread_service_principal_password" "sa" {
  service_principal_id = azuread_service_principal.sa.object_id
}

# An Azure AD service principal used to test creating an external login to the Azure SQL server resource
resource "azuread_application" "user" {
  display_name = "${local.prefix}-user"
  web {
    homepage_url = "https://test.example.com"
  }
}

resource "azuread_service_principal" "user" {
  client_id = azuread_application.user.client_id
}

resource "azuread_service_principal_password" "user" {
  service_principal_id = azuread_service_principal.user.id
}

# Temporary resource group
resource "azurerm_resource_group" "rg" {
  name     = "${lower(var.prefix)}-${random_string.random.result}"
  location = var.location
}

# An Azure SQL Server
resource "azurerm_mssql_server" "sql_server" {
  name                = "${lower(local.prefix)}-sql-server"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location

  version                      = "12.0"
  administrator_login          = "SuperAdministrator"
  administrator_login_password = azuread_service_principal_password.sa.value

  azuread_administrator {
    tenant_id      = var.tenant_id
    object_id      = azuread_service_principal.sa.client_id
    login_username = azuread_service_principal.sa.display_name
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azuread_group_member" "sql" {
  group_object_id  = data.azuread_group.sql_servers.id
  member_object_id = azurerm_mssql_server.sql_server.identity[0].principal_id
}

resource "azurerm_mssql_firewall_rule" "sql_server_fw_rule" {
  count            = length(var.local_ip_addresses)
  name             = "AllowIP ${count.index}"
  server_id        = azurerm_mssql_server.sql_server.id
  start_ip_address = var.local_ip_addresses[count.index]
  end_ip_address   = var.local_ip_addresses[count.index]
}

# The Azure SQL Database used in tests
resource "azurerm_mssql_database" "db" {
  name      = "testdb"
  server_id = azurerm_mssql_server.sql_server.id
  sku_name  = "Basic"
}

resource "time_sleep" "wait_15_seconds" {
  depends_on = [azurerm_mssql_database.db]

  create_duration = "15s"
}


#
# Creates a login and user in the SQL Server
#
resource "random_password" "server" {
  keepers = {
    login_name = "testlogin"
    username   = "testuser"
  }
  length  = 32
  special = true
}

resource "mssql_login" "server" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    login {
      username = azurerm_mssql_server.sql_server.administrator_login
      password = azurerm_mssql_server.sql_server.administrator_login_password
    }
  }
  login_name = random_password.server.keepers.login_name
  password   = random_password.server.result

  depends_on = [time_sleep.wait_15_seconds]
}

resource "mssql_user" "server" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    login {
      username = azurerm_mssql_server.sql_server.administrator_login
      password = azurerm_mssql_server.sql_server.administrator_login_password
    }
  }
  database   = azurerm_mssql_database.db.name
  username   = random_password.server.keepers.username
  login_name = mssql_login.server.login_name
}

output "instance" {
  value = {
    login_name = mssql_login.server.login_name,
    password   = mssql_login.server.password
  }
  sensitive = true
}


#
# Creates a user with login in the SQL Server database
#

resource "random_password" "database" {
  keepers = {
    username = "testuser2"
  }
  length  = 32
  special = true
}

resource "mssql_user" "database" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    login {
      username = azurerm_mssql_server.sql_server.administrator_login
      password = azurerm_mssql_server.sql_server.administrator_login_password
    }
  }
  database = azurerm_mssql_database.db.name
  username = "${local.prefix}-user"
  password = random_password.database.result
}

output "database" {
  value = {
    username = mssql_user.database.username,
    password = mssql_user.database.password
  }
  sensitive = true
}


#
# Creates a login and user from Azure AD in the SQL Server
#

resource "mssql_user" "external" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    azure_login {
      tenant_id     = var.tenant_id
      client_id     = azuread_service_principal.sa.client_id
      client_secret = azuread_service_principal_password.sa.value
    }
  }
  database = azurerm_mssql_database.db.name
  username = azuread_service_principal.user.display_name
}

output "external" {
  value = {
    tenant_id     = var.tenant_id
    client_id     = azuread_service_principal.user.client_id
    client_secret = azuread_service_principal_password.user.value
  }
  sensitive = true
}

resource "mssql_database_role" "example" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    azure_login {
      tenant_id     = var.tenant_id
      client_id     = azuread_service_principal.sa.client_id
      client_secret = azuread_service_principal_password.sa.value
    }
  }
  database  = "master"
  role_name = "testrole"
}

resource "mssql_database_role" "example_authorization" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    azure_login {
      tenant_id     = var.tenant_id
      client_id     = azuread_service_principal.sa.client_id
      client_secret = azuread_service_principal_password.sa.value
    }
  }
  database   = "master"
  role_name  = "testrole"
  owner_name = mssql_user.external.username
}

resource "mssql_database_permissions" "example" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    azure_login {
      tenant_id     = var.tenant_id
      client_id     = azuread_service_principal.sa.client_id
      client_secret = azuread_service_principal_password.sa.value
    }
  }
  database = "example"
  username = "username"
  permissions = [
    "EXECUTE",
    "UPDATE",
    "INSERT",
  ]
}

resource "mssql_database_schema" "example" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    azure_login {
      tenant_id     = var.tenant_id
      client_id     = azuread_service_principal.sa.client_id
      client_secret = azuread_service_principal_password.sa.value
    }
  }
  database    = "master"
  schema_name = "testschema"
}

resource "mssql_database_schema" "example_authorization" {
  server {
    host = azurerm_mssql_server.sql_server.fully_qualified_domain_name
    azure_login {
      tenant_id     = var.tenant_id
      client_id     = azuread_service_principal.sa.client_id
      client_secret = azuread_service_principal_password.sa.value
    }
  }
  database    = "my-database"
  schema_name = "testschema"
  owner_name  = mssql_user.external.username
}
