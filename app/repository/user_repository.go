package repository

import (
	"context"
	"database/sql"
	"fmt"

	"sistem-prestasi-mhs/app/models"

	"github.com/google/uuid"
)

const (
	// Definisikan kolom sekali untuk konsistensi
	userCols = `id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at`
	roleCols = `id, name, description, created_at`
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ============================================================================
// CRUD OPERATIONS
// ============================================================================

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		u.ID, u.Username, u.Email, u.PasswordHash, u.FullName, u.RoleID, u.IsActive,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) Update(ctx context.Context, u *models.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, full_name = $3, role_id = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		u.Username, u.Email, u.FullName, u.RoleID, u.IsActive, u.ID,
	).Scan(&u.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("user not found")
	}
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete
	query := "UPDATE users SET is_active = false WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// ============================================================================
// RETRIEVAL OPERATIONS
// ============================================================================

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE id = $1", userCols)
	return r.fetchOneUser(ctx, query, id)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE username = $1 AND is_active = true", userCols)
	return r.fetchOneUser(ctx, query, username)
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`, userCols)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := r.scanUser(rows, &u); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

// ============================================================================
// ROLE & PERMISSIONS
// ============================================================================

func (r *UserRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	query := fmt.Sprintf("SELECT %s FROM roles WHERE id = $1", roleCols)

	var role models.Role
	err := r.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// GetPermissions retrieves permission names based on UserID (Join Users -> Roles -> Perms)
func (r *UserRepository) GetPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN users u ON u.role_id = rp.role_id
		WHERE u.id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// ============================================================================
// INTERNAL HELPERS
// ============================================================================

func (r *UserRepository) fetchOneUser(ctx context.Context, query string, args ...interface{}) (*models.User, error) {
	var u models.User
	err := r.scanUser(r.db.QueryRowContext(ctx, query, args...), &u)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) scanUser(scanner interface {
	Scan(dest ...interface{}) error
}, u *models.User) error {
	return scanner.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
}
