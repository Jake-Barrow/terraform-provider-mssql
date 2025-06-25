# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.3]

### Changed

- Updated docs

## [0.4.2]

### Changed

- Updated version number
- Tried to setup github action for terraform

## [0.4.1]

### Added

- New property `ignore_deletion` to resources `mssql_user` and `mssql_database_schema` to allow deletion to be ignored when running `terraform destroy`.

## [0.4.0]

### Added

- New resource `mssql_entraid_login` for managing Entra ID (formerly Azure AD) logins
- New data source `mssql_entraid_login` for retrieving Entra ID login information

### Changed

- Optimized resource creation and update operations

## [0.3.9]

### Changed

- Upgraded to go version 1.23
- Upgraded dependencies

### Fixed

- Minor fixes in acceptance tests

## [0.3.8]

### Added

- New resource `mssql_database_sqlscript` for executing and managing SQL scripts within databases

### Changed

- Added optional parameter `type` ('E'|'X') when creating database users from Entra ID (formerly Azure AD) when object_id is specified
- Improved resource handling to trigger recreation
- Password updates for users (contained database) are now performed in-place instead of requiring resource recreation
- Migrated authentication from ADAL (Azure Active Directory Authentication Library) to MSAL (Microsoft Authentication Library)

### Fixed
- Fixed issue where resource property updates were not properly triggering resource recreation when the SQL Server-side update operation failed

## [0.3.7]

### Fixed

- removed unused docs

## [0.3.6] - 2024 08-26

### Changed

- Renamed `mssql_external_datasource` to `mssql_azure_external_datasource`

### Changed

- Improve `mssql_azure_external_datasource` to check mssql version
- Added import for `mssql_azure_external_datasource`
- Added AccTest for Validation Password

## [0.3.5] - 2024-03-25

### Added

- Support to create the mssql login with SID
- New resource mssql_database_permission
- New resource mssql_database_role
- New resource mssql_database_schema
- New resource mssql_database_masterkey
- New resource mssql_database_credential
- New resource mssql_external_datasource
- New datasource mssql_login
- New datasource mssql_user
- New datasource mssql_database_permission
- New datasource mssql_database_role
- New datasource mssql_database_schema
- New datasource mssql_database_credential
- New datasource mssql_external_datasource

## [0.3.0] - 2023-12-29

### Changed

- Make minimum terraform version 1.5. Versions less than this are no longer supported ([endoflife.date](https://endoflife.date/terraform))
- Upgraded to go version 1.21.
- Upgraded dependencies.
- Replaced github.com/denisenkom/go-mssqldb with github.com/microsoft/go-mssqldb.
- Upgraded terraform dependencies.
- Improve Makefile.

## [0.2.7] - 2022-12-16

### Fixed

- Fix concurrency issue on user create/update. [PR #52](https://github.com/betr-io/terraform-provider-mssql/pull/52). Closes [#31](https://github.com/betr-io/terraform-provider-mssql/issues/31]. Thanks to [Isabel Andrade](https://github.com/beandrad) for the PR.
- Fix role reorder update issue. [PR #53](https://github.com/betr-io/terraform-provider-mssql/pull/53). Closes [#46](https://github.com/betr-io/terraform-provider-mssql/issues/46). Thanks to [Paul Brittain](https://github.com/paulbrittain) for the PR.

## [0.2.6] - 2022-11-25

### Added

- Support two of the auth forms available through the new [fedauth](https://github.com/denisenkom/go-mssqldb#azure-active-directory-authentication): `ActiveDirectoryDefault` and `ActiveDirectoryManagedIdentity` (because user-assigned identity) as these are the most useful variants. [PR #42](https://github.com/betr-io/terraform-provider-mssql/pull/42). Closes [#30](https://github.com/betr-io/terraform-provider-mssql/issues/30). Thanks to [Bittrance](https://github.com/bittrance) for the PR.
- Improve docs on managed identities. [PR #39](https://github.com/betr-io/terraform-provider-mssql/pull/36). Thanks to [Alexander Guth](https://github.com/alxy) for the PR.

## [0.2.5] - 2022-06-03

### Added

- Add SID as output attribute to the `mssql_user` resource. [PR #36](https://github.com/betr-io/terraform-provider-mssql/pull/36). Closes [#35](https://github.com/betr-io/terraform-provider-mssql/issues/35). Thanks to [rjbell](https://github.com/rjbell) for the PR.

### Changed

- Treat `password` attribute of `mssql_user` as sensitive. Closes [#37](https://github.com/betr-io/terraform-provider-mssql/issues/37).
- Fully qualify package name with Github repository. [PR #38](https://github.com/betr-io/terraform-provider-mssql/pull/38). Thanks to [Ewan Noble](https://github.com/EwanNoble) for the PR.
- Upgraded to go version 1.18
- Upgraded dependencies.
- Upgraded dependencies in test fixtures.

### Fixed

- Only get sql logins if user is not external. [PR #33](https://github.com/betr-io/terraform-provider-mssql/pull/33). Closes [#32](https://github.com/betr-io/terraform-provider-mssql/issues/32). Thanks to [Alexander Guth](https://github.com/alxy) for the PR.

## [0.2.4] - 2021-11-15

Thanks to [Richard Lavey](https://github.com/rlaveycal) ([PR #24](https://github.com/betr-io/terraform-provider-mssql/pull/24)).

### Fixed

- Race condition with String_Split causes failure ([#23](https://github.com/betr-io/terraform-provider-mssql/issues/23))

## [0.2.3] - 2021-09-16

Thanks to [Matthis Holleville](https://github.com/matthisholleville) ([PR #17](https://github.com/betr-io/terraform-provider-mssql/pull/17)), and [bruno-motacardoso](https://github.com/bruno-motacardoso) ([PR #14](https://github.com/betr-io/terraform-provider-mssql/pull/14)).

### Changed

- Add string split function, which should allow the provider to work on SQL Server 2014 (#17).
- Improved documentation (#14).

## [0.2.2] - 2021-08-24

### Changed

- Upgraded to go version 1.17.
- Upgraded dependencies.
- Upgraded dependencies in test fixtures.

## [0.2.1] - 2021-04-30

Thanks to [Anders Båtstrand](https://github.com/anderius) ([PR #8](https://github.com/betr-io/terraform-provider-mssql/pull/8), [PR #9](https://github.com/betr-io/terraform-provider-mssql/pull/9))

### Changed

- Upgrade go-mssqldb to support go version 1.16.

### Fixed

- Cannot create user because of conflicting collation. ([#6](https://github.com/betr-io/terraform-provider-mssql/issues/6))

## [0.2.0] - 2021-04-06

When it is not possible to give AD role: _Directory Readers_ to the Sql Server Identity or an AD Group, use *object_id* to add external user.

Thanks to [Brice Messeca](https://github.com/smag-bmesseca) ([PR #1](https://github.com/betr-io/terraform-provider-mssql/pull/1))

### Added

- Optional object_id attribute to mssql_user

## [0.1.1] - 2020-11-17

Update documentation and examples.

## [0.1.0] - 2020-11-17

Initial release.

### Added

- Resource `mssql_login` to manipulate logins to a SQL Server.
- Resource `mssql_user` to manipulate users in a SQL Server database.

### Fixed
- Fixed issue where resource property updates (including password, secrets, default_language, default_database) were not properly triggering resource recreation when the SQL Server-side update operation failed

[Unreleased]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.7...HEAD
[0.2.7]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/betr-io/terraform-provider-mssql/releases/tag/v0.1.0
