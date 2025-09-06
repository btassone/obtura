package models

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/btassone/obtura/pkg/database"
)

// User represents a user in the system
type User struct {
	ID              int64      `db:"id"`
	Name            string     `db:"name"`
	Email           string     `db:"email"`
	Password        string     `db:"password"`
	Role            string     `db:"role"`
	Avatar          *string    `db:"avatar"`
	Bio             *string    `db:"bio"`
	Active          bool       `db:"active"`
	EmailVerifiedAt *time.Time `db:"email_verified_at"`
	RememberToken   *string    `db:"remember_token"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}

// TableName returns the table name
func (u *User) TableName() string {
	return "users"
}

// PrimaryKey returns the primary key field
func (u *User) PrimaryKey() string {
	return "id"
}

// UserRepository handles user database operations
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, name, email, password, role, avatar, bio, active, 
		       email_verified_at, remember_token, created_at, updated_at, deleted_at
		FROM users 
		WHERE email = ? AND deleted_at IS NULL
		LIMIT 1
	`
	
	row := r.db.QueryRow(query, email)
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Avatar,
		&user.Bio,
		&user.Active,
		&user.EmailVerifiedAt,
		&user.RememberToken,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	
	return user, err
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id int64) (*User, error) {
	user := &User{}
	query := `
		SELECT id, name, email, password, role, avatar, bio, active, 
		       email_verified_at, remember_token, created_at, updated_at, deleted_at
		FROM users 
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`
	
	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Avatar,
		&user.Bio,
		&user.Active,
		&user.EmailVerifiedAt,
		&user.RememberToken,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	
	return user, err
}

// Create creates a new user
func (r *UserRepository) Create(user *User) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	
	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	query := `
		INSERT INTO users (name, email, password, role, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(
		query,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	user.ID = id
	return nil
}

// Update updates a user
func (r *UserRepository) Update(user *User) error {
	user.UpdatedAt = time.Now()
	
	query := `
		UPDATE users 
		SET name = ?, email = ?, role = ?, avatar = ?, bio = ?, 
		    active = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	
	_, err := r.db.Exec(
		query,
		user.Name,
		user.Email,
		user.Role,
		user.Avatar,
		user.Bio,
		user.Active,
		user.UpdatedAt,
		user.ID,
	)
	
	return err
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(userID int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	query := `
		UPDATE users 
		SET password = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	
	_, err = r.db.Exec(query, hashedPassword, time.Now(), userID)
	return err
}

// CheckPassword checks if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}