package repositories

import (
	"jejak/internal/models"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

// FindByEmailOrUsername implements [UserRepository].
func (u *UserRepositoryImpl) FindByEmailOrUsername(identifier string) (*models.User, error) {
	var user *models.User
	if err := u.db.Where("email ILIKE ? OR username ILIKE ?", identifier, identifier).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// FindByUsername implements [UserRepository].
func (u *UserRepositoryImpl) FindByUsername(username string) (*models.User, error) {
	var user *models.User
	if err := u.db.Where("username ILIKE ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// FindByIDIncludePassword implements UserRepository.
func (u *UserRepositoryImpl) FindByIDIncludePassword(id string) (*models.User, error) {
	var user *models.User
	if err := u.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// Create implements UserRepository.
func (u *UserRepositoryImpl) Create(user *models.User) error {
	return u.db.Create(user).Error
}

// Delete implements UserRepository.
func (u *UserRepositoryImpl) Delete(id string) error {
	return u.db.Delete(&models.User{}, "id = ?", id).Error
}

// Update implements UserRepository.
func (u *UserRepositoryImpl) Update(user *models.User) error {
	return u.db.Save(user).Error
}

// FindAll implements UserRepository.
func (u *UserRepositoryImpl) FindAll() ([]models.User, error) {
	var users []models.User
	if err := u.db.Omit("password").Preload("Teams").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindByEmail implements UserRepository.
func (u *UserRepositoryImpl) FindByEmail(email string) (*models.User, error) {
	var user *models.User
	if err := u.db.Where("email ILIKE ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID implements UserRepository.
func (u *UserRepositoryImpl) FindByID(id string) (*models.User, error) {
	var user *models.User
	if err := u.db.Omit("password").Preload("Teams").Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepositoryImpl) ListDistinctRoles() ([]string, error) {
	type roleRow struct {
		Role string `gorm:"column:role"`
	}

	var rows []roleRow
	err := u.db.Raw("SELECT DISTINCT unnest(roles) AS role FROM users WHERE roles IS NOT NULL ORDER BY role ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Role != "" {
			roles = append(roles, row.Role)
		}
	}

	return roles, nil
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// allowedFilterCols is the allowlist of columns that may be used as filter keys.
var allowedFilterCols = map[string]bool{
	"roles": true,
}

// applyFilters appends WHERE clauses to query based on the filters map.
// Only columns in allowedFilterCols are accepted to prevent injection.
func applyFilters(query *gorm.DB, filters map[string][]string) *gorm.DB {
	for col, values := range filters {
		if !allowedFilterCols[col] || len(values) == 0 {
			continue
		}

		hasNull := false
		realValues := make([]string, 0, len(values))
		for _, v := range values {
			if v == "__NULL__" {
				hasNull = true
			} else {
				realValues = append(realValues, v)
			}
		}

		switch col {
		case "roles":
			if hasNull && len(realValues) > 0 {
				query = query.Where("(roles && ? OR roles IS NULL OR roles = '{}')", pq.Array(realValues))
			} else if hasNull {
				query = query.Where("roles IS NULL OR roles = '{}'")
			} else {
				query = query.Where("roles && ?", pq.Array(realValues))
			}
		}
	}
	return query
}

// Count implements UserRepository.Count
func (u *UserRepositoryImpl) Count(search string, filters map[string][]string, total *int64) error {
	query := u.db.Model(&models.User{})

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ?", like, like)
	}

	query = applyFilters(query, filters)

	return query.Count(total).Error
}

// FindPaginated implements UserRepository.FindPaginated
func (u *UserRepositoryImpl) FindPaginated(search string, limit, offset int, sortBy, sortOrder string, filters map[string][]string) ([]models.User, error) {
	query := u.db.Model(&models.User{}).Omit("password")

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ?", like, like)
	}

	query = applyFilters(query, filters)

	validSortFields := map[string]string{
		"no":         "created_at",
		"created_at": "created_at",
		"email":      "email",
		"username":   "username",
	}
	field, ok := validSortFields[sortBy]
	if !ok {
		field = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	var users []models.User
	if err := query.Preload("Teams").Order(field + " " + sortOrder).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
