// Package store is the implementation for managing Bytebase's own metadata in a PostgreSQL database.
package store

import (
	"context"
	"database/sql"
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/store/model"
)

// Store provides database access to all raw objects.
type Store struct {
	dbConnManager *DBConnectionManager
	enableCache   bool

	// Cache.
	Secret               string
	userIDCache          *lru.Cache[int, *UserMessage]
	userEmailCache       *lru.Cache[string, *UserMessage]
	instanceCache        *lru.Cache[string, *InstanceMessage]
	databaseCache        *lru.Cache[string, *DatabaseMessage]
	projectCache         *lru.Cache[string, *ProjectMessage]
	policyCache          *lru.Cache[string, *PolicyMessage]
	issueCache           *lru.Cache[int, *IssueMessage]
	issueByPipelineCache *lru.Cache[int, *IssueMessage]
	pipelineCache        *lru.Cache[int, *PipelineMessage]
	settingCache         *lru.Cache[storepb.SettingName, *SettingMessage]
	idpCache             *lru.Cache[string, *IdentityProviderMessage]
	databaseGroupCache   *lru.Cache[string, *DatabaseGroupMessage]
	rolesCache           *lru.Cache[string, *RoleMessage]
	groupCache           *lru.Cache[string, *GroupMessage]
	sheetCache           *lru.Cache[int, *SheetMessage]

	// Large objects.
	sheetStatementCache *lru.Cache[int, string]
	dbMetadataCache     *lru.Cache[string, *model.DatabaseMetadata]
}

// SensitiveDataStore provides database access to sensitive data rules and approval flows.
type SensitiveDataStore interface {
	// SensitiveDataRule
	CreateSensitiveDataRule(ctx context.Context, rule *SensitiveDataRuleMessage) (*SensitiveDataRuleMessage, error)
	ListSensitiveDataRules(ctx context.Context, filter *SensitiveDataRuleFilter) ([]*SensitiveDataRuleMessage, error)
	GetSensitiveDataRule(ctx context.Context, id int) (*SensitiveDataRuleMessage, error)
	UpdateSensitiveDataRule(ctx context.Context, id int, updater func(*SensitiveDataRuleMessage) (*SensitiveDataRuleMessage, error)) (*SensitiveDataRuleMessage, error)
	DeleteSensitiveDataRule(ctx context.Context, id int) error

	// ApprovalFlow
	CreateApprovalFlow(ctx context.Context, flow *ApprovalFlowMessage) (*ApprovalFlowMessage, error)
	ListApprovalFlows(ctx context.Context, filter *ApprovalFlowFilter) ([]*ApprovalFlowMessage, error)
	GetApprovalFlow(ctx context.Context, id int) (*ApprovalFlowMessage, error)
	UpdateApprovalFlow(ctx context.Context, id int, updater func(*ApprovalFlowMessage) (*ApprovalFlowMessage, error)) (*ApprovalFlowMessage, error)
	DeleteApprovalFlow(ctx context.Context, id int) error
}

// SensitiveDataStore implements SensitiveDataStore.
func (s *Store) SensitiveDataStore() SensitiveDataStore {
	return s
}

// WorkflowStore provides database access to workflows.
type WorkflowStore interface {
	// ... existing WorkflowStore methods ...
}

// WorkflowStore implements WorkflowStore.
func (s *Store) WorkflowStore() WorkflowStore {
	return s
}

// WorkflowRunStore provides database access to workflow runs.
type WorkflowRunStore interface {
	// ... existing WorkflowRunStore methods ...
}

// WorkflowRunStore implements WorkflowRunStore.
func (s *Store) WorkflowRunStore() WorkflowRunStore {
	return s
}

// WorkflowTaskStore provides database access to workflow tasks.
type WorkflowTaskStore interface {
	// ... existing WorkflowTaskStore methods ...
}

// WorkflowTaskStore implements WorkflowTaskStore.
func (s *Store) WorkflowTaskStore() WorkflowTaskStore {
	return s
}

// WorkflowTaskRunStore provides database access to workflow task runs.
type WorkflowTaskRunStore interface {
	// ... existing WorkflowTaskRunStore methods ...
}

// WorkflowTaskRunStore implements WorkflowTaskRunStore.
func (s *Store) WorkflowTaskRunStore() WorkflowTaskRunStore {
	return s
}

// WorkflowTaskInstanceStore provides database access to workflow task instances.
type WorkflowTaskInstanceStore interface {
	// ... existing WorkflowTaskInstanceStore methods ...
}

// WorkflowTaskInstanceStore implements WorkflowTaskInstanceStore.
func (s *Store) WorkflowTaskInstanceStore() WorkflowTaskInstanceStore {
	return s
}

// WorkflowTaskDefinitionStore provides database access to workflow task definitions.
type WorkflowTaskDefinitionStore interface {
	// ... existing WorkflowTaskDefinitionStore methods ...
}

// WorkflowTaskDefinitionStore implements WorkflowTaskDefinitionStore.
func (s *Store) WorkflowTaskDefinitionStore() WorkflowTaskDefinitionStore {
	return s
}

// WorkflowTaskDefinitionVersionStore provides database access to workflow task definition versions.
type WorkflowTaskDefinitionVersionStore interface {
	// ... existing WorkflowTaskDefinitionVersionStore methods ...
}

// WorkflowTaskDefinitionVersionStore implements WorkflowTaskDefinitionVersionStore.
func (s *Store) WorkflowTaskDefinitionVersionStore() WorkflowTaskDefinitionVersionStore {
	return s
}

// MemberStore provides database access to members.
type MemberStore interface {
	// ... existing MemberStore methods ...
}

// MemberStore implements MemberStore.
func (s *Store) MemberStore() MemberStore {
	return s
}

// MemberRoleStore provides database access to member roles.
type MemberRoleStore interface {
	// ... existing MemberRoleStore methods ...
}

// MemberRoleStore implements MemberRoleStore.
func (s *Store) MemberRoleStore() MemberRoleStore {
	return s
}

// GroupMemberStore provides database access to group members.
type GroupMemberStore interface {
	// ... existing GroupMemberStore methods ...
}

// GroupMemberStore implements GroupMemberStore.
func (s *Store) GroupMemberStore() GroupMemberStore {
	return s
}

// GroupStore provides database access to groups.
type GroupStore interface {
	// ... existing GroupStore methods ...
}

// GroupStore implements GroupStore.
func (s *Store) GroupStore() GroupStore {
	return s
}

// DatabaseRoleStore provides database access to database roles.
type DatabaseRoleStore interface {
	// ... existing DatabaseRoleStore methods ...
}

// DatabaseRoleStore implements DatabaseRoleStore.
func (s *Store) DatabaseRoleStore() DatabaseRoleStore {
	return s
}

// InstanceRoleStore provides database access to instance roles.
type InstanceRoleStore interface {
	// ... existing InstanceRoleStore methods ...
}

// InstanceRoleStore implements InstanceRoleStore.
func (s *Store) InstanceRoleStore() InstanceRoleStore {
	return s
}

// EnvironmentRoleStore provides database access to environment roles.
type EnvironmentRoleStore interface {
	// ... existing EnvironmentRoleStore methods ...
}

// EnvironmentRoleStore implements EnvironmentRoleStore.
func (s *Store) EnvironmentRoleStore() EnvironmentRoleStore {
	return s
}

// ProjectRoleStore provides database access to project roles.
type ProjectRoleStore interface {
	// ... existing ProjectRoleStore methods ...
}

// ProjectRoleStore implements ProjectRoleStore.
func (s *Store) ProjectRoleStore() ProjectRoleStore {
	return s
}

// WorkspaceRoleStore provides database access to workspace roles.
type WorkspaceRoleStore interface {
	// ... existing WorkspaceRoleStore methods ...
}

// WorkspaceRoleStore implements WorkspaceRoleStore.
func (s *Store) WorkspaceRoleStore() WorkspaceRoleStore {
	return s
}

// CustomRoleStore provides database access to custom roles.
type CustomRoleStore interface {
	// ... existing CustomRoleStore methods ...
}

// CustomRoleStore implements CustomRoleStore.
func (s *Store) CustomRoleStore() CustomRoleStore {
	return s
}

// IntegrationStore provides database access to integrations.
type IntegrationStore interface {
	// ... existing IntegrationStore methods ...
}

// IntegrationStore implements IntegrationStore.
func (s *Store) IntegrationStore() IntegrationStore {
	return s
}

// APIKeyStore provides database access to API keys.
type APIKeyStore interface {
	// ... existing APIKeyStore methods ...
}

// APIKeyStore implements APIKeyStore.
func (s *Store) APIKeyStore() APIKeyStore {
	return s
}

// AnnouncementStore provides database access to announcements.
type AnnouncementStore interface {
	// ... existing AnnouncementStore methods ...
}

// AnnouncementStore implements AnnouncementStore.
func (s *Store) AnnouncementStore() AnnouncementStore {
	return s
}

// FeatureStore provides database access to features.
type FeatureStore interface {
	// ... existing FeatureStore methods ...
}

// FeatureStore implements FeatureStore.
func (s *Store) FeatureStore() FeatureStore {
	return s
}

// DataClassificationStore provides database access to data classifications.
type DataClassificationStore interface {
	// ... existing DataClassificationStore methods ...
}

// DataClassificationStore implements DataClassificationStore.
func (s *Store) DataClassificationStore() DataClassificationStore {
	return s
}

// SemanticTypeStore provides database access to semantic types.
type SemanticTypeStore interface {
	// ... existing SemanticTypeStore methods ...
}

// SemanticTypeStore implements SemanticTypeStore.
func (s *Store) SemanticTypeStore() SemanticTypeStore {
	return s
}

// SecretStore provides database access to secrets.
type SecretStore interface {
	// ... existing SecretStore methods ...
}

// SecretStore implements SecretStore.
func (s *Store) SecretStore() SecretStore {
	return s
}

// AuditLogStore provides database access to audit logs.
type AuditLogStore interface {
	// ... existing AuditLogStore methods ...
}

// AuditLogStore implements AuditLogStore.
func (s *Store) AuditLogStore() AuditLogStore {
	return s
}

// ActivityStore provides database access to activities.
type ActivityStore interface {
	// ... existing ActivityStore methods ...
}

// ActivityStore implements ActivityStore.
func (s *Store) ActivityStore() ActivityStore {
	return s
}

// LockStore provides database access to locks.
type LockStore interface {
	// ... existing LockStore methods ...
}

// LockStore implements LockStore.
func (s *Store) LockStore() LockStore {
	return s
}

// ApprovalStore provides database access to approvals.
type ApprovalStore interface {
	// ... existing ApprovalStore methods ...
}

// ApprovalStore implements ApprovalStore.
func (s *Store) ApprovalStore() ApprovalStore {
	return s
}

// ApprovalTemplateStore provides database access to approval templates.
type ApprovalTemplateStore interface {
	// ... existing ApprovalTemplateStore methods ...
}

// ApprovalTemplateStore implements ApprovalTemplateStore.
func (s *Store) ApprovalTemplateStore() ApprovalTemplateStore {
	return s
}

// ReleaseConfigStore provides database access to release configs.
type ReleaseConfigStore interface {
	// ... existing ReleaseConfigStore methods ...
}

// ReleaseConfigStore implements ReleaseConfigStore.
func (s *Store) ReleaseConfigStore() ReleaseConfigStore {
	return s
}

// ReleaseStore provides database access to releases.
type ReleaseStore interface {
	// ... existing ReleaseStore methods ...
}

// ReleaseStore implements ReleaseStore.
func (s *Store) ReleaseStore() ReleaseStore {
	return s
}

// MigrationSourceStore provides database access to migration sources.
type MigrationSourceStore interface {
	// ... existing MigrationSourceStore methods ...
}

// MigrationSourceStore implements MigrationSourceStore.
func (s *Store) MigrationSourceStore() MigrationSourceStore {
	return s
}

// SchemaSnapshotStore provides database access to schema snapshots.
type SchemaSnapshotStore interface {
	// ... existing SchemaSnapshotStore methods ...
}

// SchemaSnapshotStore implements SchemaSnapshotStore.
func (s *Store) SchemaSnapshotStore() SchemaSnapshotStore {
	return s
}

// SQLReviewAnnotationStore provides database access to SQL review annotations.
type SQLReviewAnnotationStore interface {
	// ... existing SQLReviewAnnotationStore methods ...
}

// SQLReviewAnnotationStore implements SQLReviewAnnotationStore.
func (s *Store) SQLReviewAnnotationStore() SQLReviewAnnotationStore {
	return s
}

// SQLReviewRuleSetStore provides database access to SQL review rule sets.
type SQLReviewRuleSetStore interface {
	// ... existing SQLReviewRuleSetStore methods ...
}

// SQLReviewRuleSetStore implements SQLReviewRuleSetStore.
func (s *Store) SQLReviewRuleSetStore() SQLReviewRuleSetStore {
	return s
}

// SQLReviewRuleStore provides database access to SQL review rules.
type SQLReviewRuleStore interface {
	// ... existing SQLReviewRuleStore methods ...
}

// SQLReviewRuleStore implements SQLReviewRuleStore.
func (s *Store) SQLReviewRuleStore() SQLReviewRuleStore {
	return s
}

// WebhookStore provides database access to webhooks.
type WebhookStore interface {
	// ... existing WebhookStore methods ...
}

// WebhookStore implements WebhookStore.
func (s *Store) WebhookStore() WebhookStore {
	return s
}

// TagStore provides database access to tags.
type TagStore interface {
	// ... existing TagStore methods ...
}

// TagStore implements TagStore.
func (s *Store) TagStore() TagStore {
	return s
}

// SheetStore provides database access to sheets.
type SheetStore interface {
	// ... existing SheetStore methods ...
}

// SheetStore implements SheetStore.
func (s *Store) SheetStore() SheetStore {
	return s
}

// LGTMStore provides database access to LGTMs.
type LGTMStore interface {
	// ... existing LGTMStore methods ...
}

// LGTMStore implements LGTMStore.
func (s *Store) LGTMStore() LGTMStore {
	return s
}

// MigrationHistoryStore provides database access to migration histories.
type MigrationHistoryStore interface {
	// ... existing MigrationHistoryStore methods ...
}

// MigrationHistoryStore implements MigrationHistoryStore.
func (s *Store) MigrationHistoryStore() MigrationHistoryStore {
	return s
}

// PipelineStore provides database access to pipelines.
type PipelineStore interface {
	// ... existing PipelineStore methods ...
}

// PipelineStore implements PipelineStore.
func (s *Store) PipelineStore() PipelineStore {
	return s
}

// TaskRunStore provides database access to task runs.
type TaskRunStore interface {
	// ... existing TaskRunStore methods ...
}

// TaskRunStore implements TaskRunStore.
func (s *Store) TaskRunStore() TaskRunStore {
	return s
}

// TaskStore provides database access to tasks.
type TaskStore interface {
	// ... existing TaskStore methods ...
}

// TaskStore implements TaskStore.
func (s *Store) TaskStore() TaskStore {
	return s
}

// StageStore provides database access to stages.
type StageStore interface {
	// ... existing StageStore methods ...
}

// StageStore implements StageStore.
func (s *Store) StageStore() StageStore {
	return s
}

// IssueStore provides database access to issues.
type IssueStore interface {
	// ... existing IssueStore methods ...
}

// IssueStore implements IssueStore.
func (s *Store) IssueStore() IssueStore {
	return s
}

// PrincipalStore provides database access to principals.
type PrincipalStore interface {
	// ... existing PrincipalStore methods ...
}

// PrincipalStore implements PrincipalStore.
func (s *Store) PrincipalStore() PrincipalStore {
	return s
}

// IdpStore provides database access to identity providers.
type IdpStore interface {
	// ... existing IdpStore methods ...
}

// IdpStore implements IdpStore.
func (s *Store) IdpStore() IdpStore {
	return s
}

// SettingStore provides database access to settings.
type SettingStore interface {
	// ... existing SettingStore methods ...
}

// SettingStore implements SettingStore.
func (s *Store) SettingStore() SettingStore {
	return s
}

// PolicyStore provides database access to policies.
type PolicyStore interface {
	// ... existing PolicyStore methods ...
}

// PolicyStore implements PolicyStore.
func (s *Store) PolicyStore() PolicyStore {
	return s
}

// RoleStore provides database access to roles.
type RoleStore interface {
	// ... existing RoleStore methods ...
}

// RoleStore implements RoleStore.
func (s *Store) RoleStore() RoleStore {
	return s
}

// DatabaseStore provides database access to databases.
type DatabaseStore interface {
	// ... existing DatabaseStore methods ...
}

// DatabaseStore implements DatabaseStore.
func (s *Store) DatabaseStore() DatabaseStore {
	return s
}

// InstanceStore provides database access to instances.
type InstanceStore interface {
	// ... existing InstanceStore methods ...
}

// InstanceStore implements InstanceStore.
func (s *Store) InstanceStore() InstanceStore {
	return s
}

// EnvironmentStore provides database access to environments.
type EnvironmentStore interface {
	// ... existing EnvironmentStore methods ...
}

// EnvironmentStore implements EnvironmentStore.
func (s *Store) EnvironmentStore() EnvironmentStore {
	return s
}

// ProjectStore provides database access to projects.
type ProjectStore interface {
	// ... existing ProjectStore methods ...
}

// ProjectStore implements ProjectStore.
func (s *Store) ProjectStore() ProjectStore {
	return s
}

// Cache provides database access to cache.
type Cache interface {
	// ... existing Cache methods ...
}

// Cache implements Cache.
func (s *Store) Cache() Cache {
	return s
}

// New creates a new instance of Store.
// pgURL can be either a direct PostgreSQL URL or a file path containing the URL.
func New(ctx context.Context, pgURL string, enableCache bool) (*Store, error) {
	userIDCache, err := lru.New[int, *UserMessage](32768)
	if err != nil {
		return nil, err
	}
	userEmailCache, err := lru.New[string, *UserMessage](32768)
	if err != nil {
		return nil, err
	}
	instanceCache, err := lru.New[string, *InstanceMessage](32768)
	if err != nil {
		return nil, err
	}
	databaseCache, err := lru.New[string, *DatabaseMessage](32768)
	if err != nil {
		return nil, err
	}
	projectCache, err := lru.New[string, *ProjectMessage](32768)
	if err != nil {
		return nil, err
	}
	policyCache, err := lru.New[string, *PolicyMessage](128)
	if err != nil {
		return nil, err
	}
	issueCache, err := lru.New[int, *IssueMessage](256)
	if err != nil {
		return nil, err
	}
	issueByPipelineCache, err := lru.New[int, *IssueMessage](256)
	if err != nil {
		return nil, err
	}
	pipelineCache, err := lru.New[int, *PipelineMessage](256)
	if err != nil {
		return nil, err
	}
	settingCache, err := lru.New[storepb.SettingName, *SettingMessage](64)
	if err != nil {
		return nil, err
	}
	idpCache, err := lru.New[string, *IdentityProviderMessage](4)
	if err != nil {
		return nil, err
	}
	databaseGroupCache, err := lru.New[string, *DatabaseGroupMessage](1024)
	if err != nil {
		return nil, err
	}
	rolesCache, err := lru.New[string, *RoleMessage](64)
	if err != nil {
		return nil, err
	}
	sheetCache, err := lru.New[int, *SheetMessage](64)
	if err != nil {
		return nil, err
	}
	sheetStatementCache, err := lru.New[int, string](10)
	if err != nil {
		return nil, err
	}
	dbMetadataCache, err := lru.New[string, *model.DatabaseMetadata](128)
	if err != nil {
		return nil, err
	}
	groupCache, err := lru.New[string, *GroupMessage](1024)
	if err != nil {
		return nil, err
	}

	// Initialize database connection (handles both direct URL and file-based)
	dbConnManager := NewDBConnectionManager(pgURL)
	if err := dbConnManager.Initialize(ctx); err != nil {
		return nil, err
	}

	s := &Store{
		dbConnManager: dbConnManager,
		enableCache:   enableCache,

		// Cache.
		userIDCache:          userIDCache,
		userEmailCache:       userEmailCache,
		instanceCache:        instanceCache,
		databaseCache:        databaseCache,
		projectCache:         projectCache,
		policyCache:          policyCache,
		issueCache:           issueCache,
		issueByPipelineCache: issueByPipelineCache,
		pipelineCache:        pipelineCache,
		settingCache:         settingCache,
		idpCache:             idpCache,
		databaseGroupCache:   databaseGroupCache,
		rolesCache:           rolesCache,
		sheetCache:           sheetCache,
		sheetStatementCache:  sheetStatementCache,
		dbMetadataCache:      dbMetadataCache,
		groupCache:           groupCache,
	}

	return s, nil
}

// Close closes underlying db.
func (s *Store) Close() error {
	return s.dbConnManager.Close()
}

func (s *Store) GetDB() *sql.DB {
	return s.dbConnManager.GetDB()
}

func getInstanceCacheKey(instanceID string) string {
	return instanceID
}

func getPolicyCacheKey(resourceType storepb.Policy_Resource, resource string, policyType storepb.Policy_Type) string {
	return fmt.Sprintf("policies/%s/%s/%s", resourceType, resource, policyType)
}

func getDatabaseCacheKey(instanceID, databaseName string) string {
	return fmt.Sprintf("%s/%s", instanceID, databaseName)
}

func getDatabaseGroupCacheKey(projectID, resourceID string) string {
	return fmt.Sprintf("%s/%s", projectID, resourceID)
}
