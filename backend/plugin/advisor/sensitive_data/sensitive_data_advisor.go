package sensitive_data

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	advisor "github.com/bytebase/bytebase/backend/plugin/advisor"
	"github.com/bytebase/bytebase/backend/plugin/advisor/db"
	"github.com/bytebase/bytebase/backend/store"
	"github.com/bytebase/bytebase/backend/utils"
	"github.com/bytebase/bytebase/common"
	"github.com/bytebase/bytebase/plugin/db/util"
)

const (
	adviserType = advisor.SensitiveDataIssue
	adviserName = "Sensitive Data Advisor"
	adviserDesc = "Detect sensitive data in database schemas and enforce approval flows for sensitive data changes."
)

var (
	adviserTypeValue = string(adviserType)
	adviserConfigSchema = map[string]interface{}{
		"type": "object",
		"properties": {
			"enabled": {
				"type": "boolean",
				"default": true,
			},
			"defaultSensitivityLevel": {
				"type": "string",
				"default": string(storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_LOW),
			},
		},
		"required": ["enabled"],
	}
)

func init() {
	advisor.Register(storepb.PluginType_SENSITIVE_DATA_ADVISOR, advisorTypeValue, &SensitiveDataAdvisor{})
}

// SensitiveDataAdvisor is the sensitive data advisor.
type SensitiveDataAdvisor struct {
	advisorBase
}

// advisorBase is the base advisor implementation.
type advisorBase struct {
	advisor.UnimplementedAdvisor
}

// Init initializes the sensitive data advisor.
func (a *SensitiveDataAdvisor) Init(ctx context.Context, sp store.Provider, config string) error {
	// Parse config
	var advisorConfig struct {
		Enabled                    bool                          `json:"enabled"`
		DefaultSensitivityLevel    storepb.SensitiveDataLevel   `json:"defaultSensitivityLevel"`
	}
	if config != "" {
		if err := json.Unmarshal([]byte(config), &advisorConfig); err != nil {
			return fmt.Errorf("failed to parse sensitive data advisor config: %w", err)
		}
	} else {
		// Use default config
		advisorConfig.Enabled = true
		advisorConfig.DefaultSensitivityLevel = storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_LOW
	}

	// Validate default sensitivity level
	if advisorConfig.DefaultSensitivityLevel == storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		return fmt.Errorf("invalid default sensitivity level: %v", advisorConfig.DefaultSensitivityLevel)
	}

	// Store config in context or advisor struct for later use
	// For now, we'll just return nil to indicate successful initialization
	return nil
}

// Advise is the main method that detects sensitive data in database schemas.
func (a *SensitiveDataAdvisor) Advise(ctx context.Context, sp store.Provider, request advisor.AdviseRequest) ([]advisor.AdviseResult, error) {
	// Check if the advisor is enabled
	// For now, we'll assume it's enabled

	// Check if the request contains a schema change
	if request.SchemaChange == "" {
		// No schema change provided, return no results
		return nil, nil
	}

	// Parse the schema change to determine which tables/columns are being modified
	tables, columns, err := parseSchemaChange(request.SchemaChange)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema change: %w", err)
	}

	// Retrieve all sensitive data rules from the store
	rules, err := sp.ListSensitiveDataRules(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sensitive data rules: %w", err)
	}

	// Detect sensitive data in the modified tables/columns
	sensitiveDataChanges, err := detectSensitiveDataChanges(ctx, sp, rules, tables, columns)
	if err != nil {
		return nil, fmt.Errorf("failed to detect sensitive data changes: %w", err)
	}

	// If there are no sensitive data changes, return no results
	if len(sensitiveDataChanges) == 0 {
		return nil, nil
	}

	// Create advise results for each sensitive data change
	var results []advisor.AdviseResult
	for _, change := range sensitiveDataChanges {
		// Determine the issue severity based on the sensitivity level
		severity := advisor.SeverityNone
		switch change.SensitivityLevel {
		case storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_HIGH:
			severity = advisor.SeverityCritical
		case storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_MEDIUM:
			severity = advisor.SeverityMajor
		case storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_LOW:
			severity = advisor.SeverityMinor
		}

		// Create the advise result
		result := advisor.AdviseResult{
			AdvisorType:     advisorTypeValue,
			Severity:        severity,
			Title:            fmt.Sprintf("Sensitive Data Change Detected: %s.%s", change.TableName, change.FieldName),
			Content:          fmt.Sprintf("The change to %s.%s involves sensitive data with level %s.", change.TableName, change.FieldName, change.SensitivityLevel),
			Source:           request.SchemaChange,
			StartLine:        change.StartLine,
			EndLine:          change.EndLine,
		}

		// Add custom data to the result (e.g., sensitive data level, rule ID)
		customData := map[string]interface{}{
			"sensitivityLevel": change.SensitivityLevel,
			"ruleID":           change.RuleID,
		}
		customDataBytes, _ := json.Marshal(customData)
		result.CustomData = string(customDataBytes)

		results = append(results, result)
	}

	return results, nil
}

// GetType returns the type of the advisor.
func (a *SensitiveDataAdvisor) GetType() string {
	return advisorTypeValue
}

// GetName returns the name of the advisor.
func (a *SensitiveDataAdvisor) GetName() string {
	return adviserName
}

// GetDesc returns the description of the advisor.
func (a *SensitiveDataAdvisor) GetDesc() string {
	return adviserDesc
}

// GetConfigSchema returns the config schema of the advisor.
func (a *SensitiveDataAdvisor) GetConfigSchema() map[string]interface{} {
	return advisorConfigSchema
}

// parseSchemaChange parses a schema change statement to determine which tables/columns are being modified.
func parseSchemaChange(schemaChange string) (tables []string, columns []string, err error) {
	// This is a simplified implementation that extracts tables and columns from common DDL statements.
	// In a real implementation, we would use a proper SQL parser to accurately parse the schema change.

	// Convert the schema change to lowercase for easier parsing
	lowerCaseSchemaChange := strings.ToLower(schemaChange)

	// Extract tables
	// For CREATE TABLE statements
	createTableRegex := regexp.MustCompile(`create\s+table\s+([a-zA-Z0-9_]+)`)
	matches := createTableRegex.FindStringSubmatch(lowerCaseSchemaChange)
	if len(matches) > 1 {
		tables = append(tables, matches[1])
	}

	// For ALTER TABLE statements
	alterTableRegex := regexp.MustCompile(`alter\s+table\s+([a-zA-Z0-9_]+)`)
	matches = alterTableRegex.FindStringSubmatch(lowerCaseSchemaChange)
	if len(matches) > 1 {
		tables = append(tables, matches[1])
	}

	// Extract columns
	// For CREATE TABLE statements (column definitions)
	columnRegex := regexp.MustCompile(`([a-zA-Z0-9_]+)\s+[a-zA-Z0-9]+`)
	matches = columnRegex.FindAllStringSubmatch(lowerCaseSchemaChange, -1)
	for _, match := range matches {
		if len(match) > 1 {
			columns = append(columns, match[1])
		}
	}

	// For ALTER TABLE statements (ADD COLUMN)
	addColumnRegex := regexp.MustCompile(`add\s+column\s+([a-zA-Z0-9_]+)`)
	matches = addColumnRegex.FindAllStringSubmatch(lowerCaseSchemaChange, -1)
	for _, match := range matches {
		if len(match) > 1 {
			columns = append(columns, match[1])
		}
	}

	return tables, columns, nil
}

// detectSensitiveDataChanges detects sensitive data changes in the specified tables and columns.
func detectSensitiveDataChanges(ctx context.Context, sp store.Provider, rules []*store.SensitiveDataRule, tables []string, columns []string) ([]*store.SensitiveDataChange, error) {
	var sensitiveDataChanges []*store.SensitiveDataChange

	// For each table and column, check if it matches any sensitive data rule
	for _, table := range tables {
		for _, column := range columns {
			// Check each rule to see if it matches the current table and column
			for _, rule := range rules {
				// Check if the rule applies to the current table
				if rule.TableName != "" && rule.TableName != table {
					// The rule is specific to a different table, skip it
					continue
				}

				// Check if the column matches any of the rule's criteria
				matched := false
				for _, field := range rule.Fields {
					// Check if the field name matches
					if field.FieldName == column {
						matched = true
						break
					}

					// Check if the data type matches
					// For now, we'll assume that the data type is not specified in the schema change
					// In a real implementation, we would parse the data type from the schema change

					// Check if the regular expression matches
					if field.Regex != "" {
						regex, err := regexp.Compile(field.Regex)
						if err != nil {
							// Invalid regular expression, skip this field
							continue
						}

						if regex.MatchString(column) {
							matched = true
							break
						}
					}
				}

				if matched {
					// The column matches the rule, create a sensitive data change
					change := &store.SensitiveDataChange{
						TableName:          table,
						FieldName:          column,
						SensitivityLevel:   rule.Level,
						RuleID:             rule.ID,
						StartLine:          1,  // In a real implementation, we would determine the actual start line
						EndLine:            1,  // In a real implementation, we would determine the actual end line
					}

					sensitiveDataChanges = append(sensitiveDataChanges, change)
					break // No need to check other rules for this column
				}
			}
		}
	}

	return sensitiveDataChanges, nil
}
