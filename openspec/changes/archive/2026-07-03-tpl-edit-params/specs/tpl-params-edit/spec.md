## ADDED Requirements

### Requirement: Interactive edit of template settings

The system SHALL provide a `joy tpl edit <template_name>` subcommand that opens the specified template's `setting.toml` file in the user's configured editor (using `$VISUAL > $EDITOR > vi` priority) and waits for the editor to close before returning.

#### Scenario: Successfully edit existing template settings

- **WHEN** user runs `seneschal joy tpl edit my-template`
- **THEN** the system locates `~/.seneschal/tpl/my-template/setting.toml`
- **THEN** the system opens the file in the user's default editor
- **THEN** the system waits for the editor to exit
- **THEN** the system prints a confirmation message

#### Scenario: Template not found

- **WHEN** user runs `seneschal joy tpl edit nonexistent-template`
- **THEN** the system SHALL return an error `template "nonexistent-template" not found`
- **THEN** the exit code SHALL be non-zero

#### Scenario: Template has no setting file

- **WHEN** user runs `seneschal joy tpl edit my-template` and the template directory exists but has no `setting.toml`
- **THEN** the system SHALL report an error that the setting file is missing

#### Scenario: No arguments provided

- **WHEN** user runs `seneschal joy tpl edit` without arguments
- **THEN** the system SHALL display usage help and return an error
