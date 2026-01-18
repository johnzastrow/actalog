package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// DocumentRepository implements domain.DocumentRepository
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create creates a new document
func (r *DocumentRepository) Create(document *domain.Document) error {
	now := time.Now()
	document.CreatedAt = now
	document.UpdatedAt = now

	query := rebindQuery(`INSERT INTO documents (organization_id, name, description, document_type, url, is_required, expires_after_days, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			document.OrganizationID,
			document.Name,
			document.Description,
			document.DocumentType,
			document.URL,
			document.IsRequired,
			document.ExpiresAfterDays,
			document.IsActive,
			document.CreatedAt,
			document.UpdatedAt,
		).Scan(&document.ID)
	}

	result, err := r.db.Exec(query,
		document.OrganizationID,
		document.Name,
		document.Description,
		document.DocumentType,
		document.URL,
		document.IsRequired,
		document.ExpiresAfterDays,
		document.IsActive,
		document.CreatedAt,
		document.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	document.ID = id
	return nil
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(id int64) (*domain.Document, error) {
	query := rebindQuery(`SELECT id, organization_id, name, description, document_type, url, is_required, expires_after_days, is_active, created_at, updated_at
		FROM documents WHERE id = ?`)

	document := &domain.Document{}
	err := r.db.QueryRow(query, id).Scan(
		&document.ID,
		&document.OrganizationID,
		&document.Name,
		&document.Description,
		&document.DocumentType,
		&document.URL,
		&document.IsRequired,
		&document.ExpiresAfterDays,
		&document.IsActive,
		&document.CreatedAt,
		&document.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return document, nil
}

// GetByOrganizationID retrieves all documents for an organization
func (r *DocumentRepository) GetByOrganizationID(orgID int64, includeInactive bool) ([]*domain.Document, error) {
	var query string
	if includeInactive {
		query = rebindQuery(`SELECT id, organization_id, name, description, document_type, url, is_required, expires_after_days, is_active, created_at, updated_at
			FROM documents WHERE organization_id = ? ORDER BY name`)
	} else {
		query = rebindQuery(`SELECT id, organization_id, name, description, document_type, url, is_required, expires_after_days, is_active, created_at, updated_at
			FROM documents WHERE organization_id = ? AND is_active = true ORDER BY name`)
	}

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []*domain.Document
	for rows.Next() {
		doc := &domain.Document{}
		err := rows.Scan(
			&doc.ID,
			&doc.OrganizationID,
			&doc.Name,
			&doc.Description,
			&doc.DocumentType,
			&doc.URL,
			&doc.IsRequired,
			&doc.ExpiresAfterDays,
			&doc.IsActive,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, rows.Err()
}

// GetRequiredByOrganizationID retrieves all required documents for an organization
func (r *DocumentRepository) GetRequiredByOrganizationID(orgID int64) ([]*domain.Document, error) {
	query := rebindQuery(`SELECT id, organization_id, name, description, document_type, url, is_required, expires_after_days, is_active, created_at, updated_at
		FROM documents WHERE organization_id = ? AND is_active = true AND is_required = true ORDER BY name`)

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []*domain.Document
	for rows.Next() {
		doc := &domain.Document{}
		err := rows.Scan(
			&doc.ID,
			&doc.OrganizationID,
			&doc.Name,
			&doc.Description,
			&doc.DocumentType,
			&doc.URL,
			&doc.IsRequired,
			&doc.ExpiresAfterDays,
			&doc.IsActive,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, rows.Err()
}

// Update updates a document
func (r *DocumentRepository) Update(document *domain.Document) error {
	document.UpdatedAt = time.Now()

	query := rebindQuery(`UPDATE documents SET name = ?, description = ?, document_type = ?, url = ?, is_required = ?, expires_after_days = ?, is_active = ?, updated_at = ?
		WHERE id = ?`)

	_, err := r.db.Exec(query,
		document.Name,
		document.Description,
		document.DocumentType,
		document.URL,
		document.IsRequired,
		document.ExpiresAfterDays,
		document.IsActive,
		document.UpdatedAt,
		document.ID,
	)
	return err
}

// Delete deletes a document
func (r *DocumentRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM documents WHERE id = ?`)
	_, err := r.db.Exec(query, id)
	return err
}

// UserDocumentRepository implements domain.UserDocumentRepository
type UserDocumentRepository struct {
	db *sql.DB
}

// NewUserDocumentRepository creates a new user document repository
func NewUserDocumentRepository(db *sql.DB) *UserDocumentRepository {
	return &UserDocumentRepository{db: db}
}

// Create creates a new user document
func (r *UserDocumentRepository) Create(userDoc *domain.UserDocument) error {
	now := time.Now()
	userDoc.CreatedAt = now
	userDoc.UpdatedAt = now

	query := rebindQuery(`INSERT INTO user_documents (user_id, document_id, status, completed_at, expires_at, verified_by_user_id, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			userDoc.UserID,
			userDoc.DocumentID,
			userDoc.Status,
			userDoc.CompletedAt,
			userDoc.ExpiresAt,
			userDoc.VerifiedByUserID,
			userDoc.Notes,
			userDoc.CreatedAt,
			userDoc.UpdatedAt,
		).Scan(&userDoc.ID)
	}

	result, err := r.db.Exec(query,
		userDoc.UserID,
		userDoc.DocumentID,
		userDoc.Status,
		userDoc.CompletedAt,
		userDoc.ExpiresAt,
		userDoc.VerifiedByUserID,
		userDoc.Notes,
		userDoc.CreatedAt,
		userDoc.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	userDoc.ID = id
	return nil
}

// GetByID retrieves a user document by ID
func (r *UserDocumentRepository) GetByID(id int64) (*domain.UserDocument, error) {
	query := rebindQuery(`SELECT ud.id, ud.user_id, ud.document_id, ud.status, ud.completed_at, ud.expires_at, ud.verified_by_user_id, ud.notes, ud.created_at, ud.updated_at,
		d.name as document_name, d.document_type, u.name as user_name, u.email as user_email,
		v.name as verified_by_name, v.email as verified_by_email
		FROM user_documents ud
		JOIN documents d ON ud.document_id = d.id
		JOIN users u ON ud.user_id = u.id
		LEFT JOIN users v ON ud.verified_by_user_id = v.id
		WHERE ud.id = ?`)

	userDoc := &domain.UserDocument{}
	err := r.db.QueryRow(query, id).Scan(
		&userDoc.ID,
		&userDoc.UserID,
		&userDoc.DocumentID,
		&userDoc.Status,
		&userDoc.CompletedAt,
		&userDoc.ExpiresAt,
		&userDoc.VerifiedByUserID,
		&userDoc.Notes,
		&userDoc.CreatedAt,
		&userDoc.UpdatedAt,
		&userDoc.DocumentName,
		&userDoc.DocumentType,
		&userDoc.UserName,
		&userDoc.UserEmail,
		&userDoc.VerifiedByName,
		&userDoc.VerifiedByEmail,
	)
	if err != nil {
		return nil, err
	}
	return userDoc, nil
}

// GetByUserID retrieves all documents for a user
func (r *UserDocumentRepository) GetByUserID(userID int64) ([]*domain.UserDocument, error) {
	query := rebindQuery(`SELECT ud.id, ud.user_id, ud.document_id, ud.status, ud.completed_at, ud.expires_at, ud.verified_by_user_id, ud.notes, ud.created_at, ud.updated_at,
		d.name as document_name, d.document_type
		FROM user_documents ud
		JOIN documents d ON ud.document_id = d.id
		WHERE ud.user_id = ? ORDER BY d.name`)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userDocs []*domain.UserDocument
	for rows.Next() {
		ud := &domain.UserDocument{}
		err := rows.Scan(
			&ud.ID,
			&ud.UserID,
			&ud.DocumentID,
			&ud.Status,
			&ud.CompletedAt,
			&ud.ExpiresAt,
			&ud.VerifiedByUserID,
			&ud.Notes,
			&ud.CreatedAt,
			&ud.UpdatedAt,
			&ud.DocumentName,
			&ud.DocumentType,
		)
		if err != nil {
			return nil, err
		}
		userDocs = append(userDocs, ud)
	}
	return userDocs, rows.Err()
}

// GetByUserAndDocument retrieves a specific user document
func (r *UserDocumentRepository) GetByUserAndDocument(userID, documentID int64) (*domain.UserDocument, error) {
	query := rebindQuery(`SELECT id, user_id, document_id, status, completed_at, expires_at, verified_by_user_id, notes, created_at, updated_at
		FROM user_documents WHERE user_id = ? AND document_id = ?`)

	userDoc := &domain.UserDocument{}
	err := r.db.QueryRow(query, userID, documentID).Scan(
		&userDoc.ID,
		&userDoc.UserID,
		&userDoc.DocumentID,
		&userDoc.Status,
		&userDoc.CompletedAt,
		&userDoc.ExpiresAt,
		&userDoc.VerifiedByUserID,
		&userDoc.Notes,
		&userDoc.CreatedAt,
		&userDoc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return userDoc, nil
}

// GetByUserAndOrganization retrieves all user documents for a specific organization
func (r *UserDocumentRepository) GetByUserAndOrganization(userID, orgID int64) ([]*domain.UserDocument, error) {
	query := rebindQuery(`SELECT ud.id, ud.user_id, ud.document_id, ud.status, ud.completed_at, ud.expires_at, ud.verified_by_user_id, ud.notes, ud.created_at, ud.updated_at,
		d.name as document_name, d.document_type
		FROM user_documents ud
		JOIN documents d ON ud.document_id = d.id
		WHERE ud.user_id = ? AND d.organization_id = ? ORDER BY d.name`)

	rows, err := r.db.Query(query, userID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userDocs []*domain.UserDocument
	for rows.Next() {
		ud := &domain.UserDocument{}
		err := rows.Scan(
			&ud.ID,
			&ud.UserID,
			&ud.DocumentID,
			&ud.Status,
			&ud.CompletedAt,
			&ud.ExpiresAt,
			&ud.VerifiedByUserID,
			&ud.Notes,
			&ud.CreatedAt,
			&ud.UpdatedAt,
			&ud.DocumentName,
			&ud.DocumentType,
		)
		if err != nil {
			return nil, err
		}
		userDocs = append(userDocs, ud)
	}
	return userDocs, rows.Err()
}

// GetPendingByUserAndOrganization retrieves pending documents for a user in an organization
func (r *UserDocumentRepository) GetPendingByUserAndOrganization(userID, orgID int64) ([]*domain.UserDocument, error) {
	query := rebindQuery(`SELECT ud.id, ud.user_id, ud.document_id, ud.status, ud.completed_at, ud.expires_at, ud.verified_by_user_id, ud.notes, ud.created_at, ud.updated_at,
		d.name as document_name, d.document_type
		FROM user_documents ud
		JOIN documents d ON ud.document_id = d.id
		WHERE ud.user_id = ? AND d.organization_id = ? AND ud.status = 'pending' ORDER BY d.name`)

	rows, err := r.db.Query(query, userID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userDocs []*domain.UserDocument
	for rows.Next() {
		ud := &domain.UserDocument{}
		err := rows.Scan(
			&ud.ID,
			&ud.UserID,
			&ud.DocumentID,
			&ud.Status,
			&ud.CompletedAt,
			&ud.ExpiresAt,
			&ud.VerifiedByUserID,
			&ud.Notes,
			&ud.CreatedAt,
			&ud.UpdatedAt,
			&ud.DocumentName,
			&ud.DocumentType,
		)
		if err != nil {
			return nil, err
		}
		userDocs = append(userDocs, ud)
	}
	return userDocs, rows.Err()
}

// GetExpiringSoon retrieves documents expiring within daysAhead days
func (r *UserDocumentRepository) GetExpiringSoon(daysAhead int) ([]*domain.UserDocument, error) {
	expirationDate := time.Now().AddDate(0, 0, daysAhead)

	var query string
	if currentDriver == "postgres" {
		query = rebindQuery(`SELECT ud.id, ud.user_id, ud.document_id, ud.status, ud.completed_at, ud.expires_at, ud.verified_by_user_id, ud.notes, ud.created_at, ud.updated_at,
			d.name as document_name, d.document_type, u.name as user_name, u.email as user_email
			FROM user_documents ud
			JOIN documents d ON ud.document_id = d.id
			JOIN users u ON ud.user_id = u.id
			WHERE ud.status = 'completed' AND ud.expires_at IS NOT NULL AND ud.expires_at <= ? AND ud.expires_at > NOW()
			ORDER BY ud.expires_at`)
	} else {
		query = rebindQuery(`SELECT ud.id, ud.user_id, ud.document_id, ud.status, ud.completed_at, ud.expires_at, ud.verified_by_user_id, ud.notes, ud.created_at, ud.updated_at,
			d.name as document_name, d.document_type, u.name as user_name, u.email as user_email
			FROM user_documents ud
			JOIN documents d ON ud.document_id = d.id
			JOIN users u ON ud.user_id = u.id
			WHERE ud.status = 'completed' AND ud.expires_at IS NOT NULL AND ud.expires_at <= ? AND ud.expires_at > datetime('now')
			ORDER BY ud.expires_at`)
	}

	rows, err := r.db.Query(query, expirationDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userDocs []*domain.UserDocument
	for rows.Next() {
		ud := &domain.UserDocument{}
		err := rows.Scan(
			&ud.ID,
			&ud.UserID,
			&ud.DocumentID,
			&ud.Status,
			&ud.CompletedAt,
			&ud.ExpiresAt,
			&ud.VerifiedByUserID,
			&ud.Notes,
			&ud.CreatedAt,
			&ud.UpdatedAt,
			&ud.DocumentName,
			&ud.DocumentType,
			&ud.UserName,
			&ud.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		userDocs = append(userDocs, ud)
	}
	return userDocs, rows.Err()
}

// Update updates a user document
func (r *UserDocumentRepository) Update(userDoc *domain.UserDocument) error {
	userDoc.UpdatedAt = time.Now()

	query := rebindQuery(`UPDATE user_documents SET status = ?, completed_at = ?, expires_at = ?, verified_by_user_id = ?, notes = ?, updated_at = ? WHERE id = ?`)

	_, err := r.db.Exec(query,
		userDoc.Status,
		userDoc.CompletedAt,
		userDoc.ExpiresAt,
		userDoc.VerifiedByUserID,
		userDoc.Notes,
		userDoc.UpdatedAt,
		userDoc.ID,
	)
	return err
}

// MarkCompleted marks a user document as completed
func (r *UserDocumentRepository) MarkCompleted(id int64, verifiedByUserID *int64) error {
	now := time.Now()

	// Get the user document first
	userDoc, err := r.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user document: %w", err)
	}

	// Get the document to check for expires_after_days
	var expiresAfterDays *int
	docQuery := rebindQuery(`SELECT expires_after_days FROM documents WHERE id = ?`)
	err = r.db.QueryRow(docQuery, userDoc.DocumentID).Scan(&expiresAfterDays)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	var expiresAt *time.Time
	if expiresAfterDays != nil {
		exp := now.AddDate(0, 0, *expiresAfterDays)
		expiresAt = &exp
	}

	query := rebindQuery(`UPDATE user_documents SET status = 'completed', completed_at = ?, expires_at = ?, verified_by_user_id = ?, updated_at = ? WHERE id = ?`)

	_, err = r.db.Exec(query, now, expiresAt, verifiedByUserID, now, id)
	return err
}

// Delete deletes a user document
func (r *UserDocumentRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM user_documents WHERE id = ?`)
	_, err := r.db.Exec(query, id)
	return err
}
