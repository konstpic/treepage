package models

import (
	"time"

	"github.com/konstpic/treepage/backend/pkg/embeddings"
	"github.com/lib/pq"
)

type OIDCProvider struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name            string    `gorm:"size:128;uniqueIndex" json:"name"`
	ProviderType    string    `gorm:"size:32;default:generic" json:"provider_type"`
	IssuerURL       string    `json:"issuer_url"`
	ClientID        string    `gorm:"size:256" json:"client_id"`
	ClientSecretRef string    `gorm:"size:128" json:"-"`
	RedirectURL     string    `json:"redirect_url"`
	Scopes          string    `json:"scopes"`
	RoleClaim       string    `gorm:"size:128" json:"role_claim"`
	GroupClaim      string    `gorm:"size:128" json:"group_claim"`
	SyncGroups      bool      `json:"sync_groups"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (OIDCProvider) TableName() string { return "oidc_providers" }

type Role struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:64;uniqueIndex" json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        string `gorm:"size:128;uniqueIndex" json:"code"`
	Description string `json:"description"`
}

func (Permission) TableName() string { return "permissions" }

type User struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email          string     `gorm:"size:320;uniqueIndex" json:"email"`
	DisplayName    string     `gorm:"size:256" json:"display_name"`
	PasswordHash   string     `gorm:"size:255" json:"-"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	ExternalID     string     `gorm:"size:256" json:"external_id,omitempty"`
	OIDCProviderID *string    `gorm:"column:oidc_provider_id;type:uuid" json:"oidc_provider_id,omitempty"`
	IsActive       bool       `json:"is_active"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Roles          []Role     `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

func (User) TableName() string { return "users" }

type Group struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:128;uniqueIndex" json:"name"`
	ExternalID  string    `gorm:"size:256" json:"external_id,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Group) TableName() string { return "groups" }

type GroupMember struct {
	GroupID string `gorm:"type:uuid;primaryKey" json:"group_id"`
	UserID  string `gorm:"type:uuid;primaryKey" json:"user_id"`
}

func (GroupMember) TableName() string { return "group_members" }

type SpaceGroup struct {
	SpaceID string `gorm:"type:uuid;primaryKey" json:"space_id"`
	GroupID string `gorm:"type:uuid;primaryKey" json:"group_id"`
	RoleID  string `gorm:"type:uuid" json:"role_id"`
}

func (SpaceGroup) TableName() string { return "space_groups" }

type RefreshToken struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"type:uuid;index" json:"user_id"`
	TokenHash string     `gorm:"size:128;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type Space struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug        string    `gorm:"size:128;uniqueIndex" json:"slug"`
	Name        string    `gorm:"size:256" json:"name"`
	Description string    `json:"description,omitempty"`
	IsPublic    bool      `json:"is_public"`
	CreatedBy   *string   `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Space) TableName() string { return "spaces" }

type SpaceMember struct {
	SpaceID string `gorm:"type:uuid;primaryKey" json:"space_id"`
	UserID  string `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID  string `gorm:"type:uuid" json:"role_id"`
}

func (SpaceMember) TableName() string { return "space_members" }

type Repository struct {
	ID                   string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SpaceID              string     `gorm:"type:uuid;index" json:"space_id"`
	Name                 string     `gorm:"size:256" json:"name"`
	URL                  string     `json:"url"`
	Branch               string     `gorm:"size:128;default:main" json:"branch"`
	Provider             string     `gorm:"type:git_provider;default:generic" json:"provider"`
	AccessTokenRef       string     `gorm:"size:128" json:"access_token_ref,omitempty"`
	DocsPath             string     `gorm:"size:512;default:docs" json:"docs_path"`
	SyncMode             string     `gorm:"size:32;default:scheduled" json:"sync_mode"`
	SyncIntervalSeconds  int        `gorm:"default:300" json:"sync_interval_seconds"`
	WebhookSecretRef     string     `gorm:"size:128" json:"webhook_secret_ref,omitempty"`
	LastSyncAt           *time.Time `json:"last_sync_at,omitempty"`
	LastSyncStatus       string     `gorm:"size:32" json:"last_sync_status,omitempty"`
	LastSyncError        string     `json:"last_sync_error,omitempty"`
	Enabled              bool       `json:"enabled"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (Repository) TableName() string { return "repositories" }

type Document struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SpaceID     string         `gorm:"type:uuid;index" json:"space_id"`
	RepositoryID *string       `gorm:"type:uuid;index" json:"repository_id,omitempty"`
	Slug        string         `gorm:"size:512" json:"slug"`
	Title       string         `gorm:"size:512" json:"title"`
	Path        string         `json:"path"`
	Content     string         `json:"content"`
	ContentHTML string         `json:"content_html,omitempty"`
	Tags        pq.StringArray `gorm:"type:text[]" json:"tags"`
	AuthorID    *string        `gorm:"type:uuid" json:"author_id,omitempty"`
	AuthorName  string         `gorm:"size:256" json:"author_name,omitempty"`
	CommitSHA         string         `gorm:"size:64" json:"commit_sha,omitempty"`
	SyncedContentHash string         `gorm:"size:64" json:"synced_content_hash,omitempty"`
	HasPendingChanges bool           `json:"has_pending_changes"`
	LastSyncedAt      *time.Time     `json:"last_synced_at,omitempty"`
	IsPublished       bool           `json:"is_published"`
	WorkflowState     string         `gorm:"size:32;default:published" json:"workflow_state"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (Document) TableName() string { return "documents" }

type DocumentVersion struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID    string    `gorm:"type:uuid;index" json:"document_id"`
	VersionNumber int       `json:"version_number"`
	Title         string    `gorm:"size:512" json:"title"`
	Content       string    `json:"content"`
	CommitSHA     string    `gorm:"size:64" json:"commit_sha,omitempty"`
	AuthorID      *string   `gorm:"type:uuid" json:"author_id,omitempty"`
	AuthorName    string    `gorm:"size:256" json:"author_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func (DocumentVersion) TableName() string { return "document_versions" }

type DocumentTranslation struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID string    `gorm:"type:uuid;uniqueIndex:idx_doc_tr_locale" json:"document_id"`
	Locale     string    `gorm:"size:8;uniqueIndex:idx_doc_tr_locale" json:"locale"`
	SourceHash string    `gorm:"size:64" json:"source_hash"`
	Title      string    `gorm:"size:512" json:"title"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

func (DocumentTranslation) TableName() string { return "document_translations" }

type BookTranslation struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BookID      string    `gorm:"type:uuid;uniqueIndex:idx_book_tr_locale" json:"book_id"`
	Locale      string    `gorm:"size:8;uniqueIndex:idx_book_tr_locale" json:"locale"`
	SourceHash  string    `gorm:"size:64" json:"source_hash"`
	Title       string    `gorm:"size:512" json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

func (BookTranslation) TableName() string { return "book_translations" }

type AuditLog struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       *string   `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action       string    `gorm:"size:128" json:"action"`
	ResourceType string    `gorm:"size:64" json:"resource_type"`
	ResourceID   *string   `gorm:"type:uuid" json:"resource_id,omitempty"`
	Metadata     string    `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

type SystemSetting struct {
	Key       string    `gorm:"primaryKey;size:128" json:"key"`
	Value     string    `gorm:"type:jsonb;not null" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy *string   `gorm:"type:uuid" json:"updated_by,omitempty"`
}

func (SystemSetting) TableName() string { return "system_settings" }

type Book struct {
	ID              string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SpaceID         string     `gorm:"type:uuid;index;not null" json:"space_id"`
	Slug            string     `gorm:"size:256;not null" json:"slug"`
	Title           string     `gorm:"size:512;not null" json:"title"`
	Description     string     `json:"description,omitempty"`
	RootPath        string     `gorm:"size:512;not null" json:"root_path"`
	Audience        string     `gorm:"size:64;default:developer" json:"audience"`
	Focus           string     `json:"focus,omitempty"`
	Status          string     `gorm:"size:32;default:draft" json:"status"`
	SourceHash      string     `gorm:"size:64" json:"source_hash,omitempty"`
	OutlineJSON     string     `gorm:"type:jsonb;not null;default:'[]'" json:"outline_json"`
	ContentMarkdown string     `gorm:"type:text" json:"content_markdown,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	Enhanced        bool       `json:"enhanced"`
	CreatedBy       *string    `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	GeneratedAt     *time.Time `json:"generated_at,omitempty"`
}

func (Book) TableName() string { return "books" }

type SyncJob struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RepositoryID   string     `gorm:"type:uuid;index" json:"repository_id"`
	Status         string     `gorm:"size:32;default:pending" json:"status"`
	TriggerType    string     `gorm:"size:32;default:manual" json:"trigger_type"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	FilesProcessed   int        `json:"files_processed"`
	ConflictsSkipped int        `json:"conflicts_skipped"`
	ErrorMessage     string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (SyncJob) TableName() string { return "sync_jobs" }

type UserFavorite struct {
	UserID     string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	DocumentID string    `gorm:"type:uuid;primaryKey" json:"document_id"`
	CreatedAt  time.Time `json:"created_at"`
	Document   *Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
}

func (UserFavorite) TableName() string { return "user_favorites" }

type UserRecentView struct {
	UserID     string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	DocumentID string    `gorm:"type:uuid;primaryKey" json:"document_id"`
	SpaceID    string    `gorm:"type:uuid" json:"space_id"`
	ViewedAt   time.Time `json:"viewed_at"`
	Document   *Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Space      *Space    `gorm:"foreignKey:SpaceID" json:"space,omitempty"`
}

func (UserRecentView) TableName() string { return "user_recent_views" }

type Notification struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string     `gorm:"type:uuid;index" json:"user_id"`
	Type         string     `gorm:"size:64" json:"type"`
	Title        string     `gorm:"size:512" json:"title"`
	Body         string     `json:"body"`
	ResourceType *string    `gorm:"size:64" json:"resource_type,omitempty"`
	ResourceID   *string    `gorm:"type:uuid" json:"resource_id,omitempty"`
	ReadAt       *time.Time `json:"read_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }

type DocumentAttachment struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID string    `gorm:"type:uuid;index" json:"document_id"`
	Filename   string    `gorm:"size:512" json:"filename"`
	StorageKey string    `gorm:"size:1024;uniqueIndex" json:"storage_key"`
	MimeType   string    `gorm:"size:128" json:"mime_type"`
	SizeBytes  int64     `json:"size_bytes"`
	UploadedBy *string   `gorm:"type:uuid" json:"uploaded_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (DocumentAttachment) TableName() string { return "document_attachments" }

type PageACLRule struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SpaceID     string    `gorm:"type:uuid;index" json:"space_id"`
	PathPrefix  string    `gorm:"size:512" json:"path_prefix"`
	SubjectType string    `gorm:"size:16" json:"subject_type"`
	SubjectID   string    `gorm:"type:uuid" json:"subject_id"`
	Role        string    `gorm:"size:32" json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

func (PageACLRule) TableName() string { return "page_acl_rules" }

type DocumentComment struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID string         `gorm:"type:uuid;index" json:"document_id"`
	ParentID   *string        `gorm:"type:uuid" json:"parent_id,omitempty"`
	AuthorID   *string        `gorm:"type:uuid" json:"author_id,omitempty"`
	Body       string         `json:"body"`
	Mentions   pq.StringArray `gorm:"type:uuid[]" json:"mentions"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	ResolvedAt *time.Time     `json:"resolved_at,omitempty"`
	AuthorName string         `gorm:"-" json:"author_name,omitempty"`
}

func (DocumentComment) TableName() string { return "document_comments" }

type SearchQueryLog struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      *string   `gorm:"type:uuid" json:"user_id,omitempty"`
	QueryText   string    `gorm:"size:512" json:"query_text"`
	ResultCount int       `json:"result_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func (SearchQueryLog) TableName() string { return "search_query_log" }

type DocumentViewStats struct {
	DocumentID   string     `gorm:"type:uuid;primaryKey" json:"document_id"`
	ViewCount    int64      `json:"view_count"`
	LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
}

func (DocumentViewStats) TableName() string { return "document_view_stats" }

type DocumentChunk struct {
	ID          string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID  string              `gorm:"type:uuid;index" json:"document_id"`
	ChunkIndex  int                 `json:"chunk_index"`
	Content     string              `json:"content"`
	ContentHash string              `gorm:"size:64" json:"content_hash"`
	Embedding   embeddings.Vector   `gorm:"type:jsonb" json:"embedding,omitempty"`
}

func (DocumentChunk) TableName() string { return "document_chunks" }

type RAGFeedback struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     *string   `gorm:"type:uuid" json:"user_id,omitempty"`
	Question   string    `json:"question"`
	Answer     string    `json:"answer,omitempty"`
	Helpful    bool      `json:"helpful"`
	Confidence float32   `json:"confidence,omitempty"`
	Sources    []byte    `gorm:"type:jsonb" json:"sources,omitempty"`
	Citations  []byte    `gorm:"type:jsonb" json:"citations,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (RAGFeedback) TableName() string { return "rag_feedback" }

type RAGLearnedSynonym struct {
	Term      string    `gorm:"size:128;primaryKey" json:"term"`
	Synonyms  []string  `gorm:"type:text[]" json:"synonyms"`
	HitCount  int       `json:"hit_count"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (RAGLearnedSynonym) TableName() string { return "rag_learned_synonyms" }
