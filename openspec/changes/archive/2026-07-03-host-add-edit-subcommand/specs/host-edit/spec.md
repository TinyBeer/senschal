## ADDED Requirements

### Requirement: Edit SSH host configuration by alias

The system SHALL provide a `seneschal host edit <alias>` command that opens the SSH TOML configuration file of the specified host in the user's preferred terminal editor, and SHALL re-validate the configuration after the editor exits.

- **WHEN** the user runs `seneschal host edit <alias>`
- **THEN** the system locates the SSH config file at `<SSHConfigDir>/<alias>.toml`
- **AND** if the file does not exist, the system SHALL return an error indicating the alias was not found
- **AND** the system SHALL open the file in the user's terminal editor via `pkg/util.EditFile`
- **AND** after the editor exits, the system SHALL re-read and parse the TOML file
- **AND** if parsing succeeds, the system SHALL log a success message
- **AND** if parsing fails (invalid TOML, missing required fields), the system SHALL return an error with the parse failure details

#### Scenario: Edit existing host config
- **WHEN** the user runs `seneschal host edit myserver`
- **AND** the file `~/.seneschal/conf/ssh/myserver.toml` exists and is valid TOML
- **THEN** the system opens the file in the user's editor
- **AND** after the editor exits, the system reads and validates the config
- **AND** the system logs success

#### Scenario: Edit non-existent alias
- **WHEN** the user runs `seneschal host edit nonexistent`
- **AND** the file `~/.seneschal/conf/ssh/nonexistent.toml` does not exist
- **THEN** the system returns an error indicating the host alias was not found
- **AND** the config is not opened in the editor

#### Scenario: Save invalid TOML during edit
- **WHEN** the user runs `seneschal host edit myserver`
- **AND** the editor opens the TOML file
- **AND** the user saves invalid TOML content (e.g., malformed syntax)
- **AND** the editor exits
- **THEN** the system returns an error with TOML parse failure details
- **AND** the user is informed the file was saved but contains errors

#### Scenario: Editor exits with non-zero code
- **WHEN** the user runs `seneschal host edit myserver`
- **AND** the editor exits with a non-zero exit code
- **AND** the file was saved with valid content
- **THEN** the system reads and validates the config successfully
- **AND** the system logs a warning about the non-zero exit code
