package handlers

import (
	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/services"
	"jejak/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	service  *services.UserService
	validate *validator.Validate
}

func NewUserHandler(service *services.UserService, validate *validator.Validate) *UserHandler {
	return &UserHandler{
		service:  service,
		validate: validate,
	}
}

func (h *UserHandler) Me(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return respondError(c, apperrors.ErrInvalidToken)
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, user, "User retrieved successfully")
}

func (h *UserHandler) UpdateMyPassword(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return respondError(c, apperrors.ErrInvalidToken)
	}

	var req dto.UpdateMyPasswordRequest

	if err := c.Bind().Body(&req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.validate.Struct(&req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.service.UpdateMyPassword(userID, &req); err != nil {
		return respondError(c, err)
	}

	return respondOK(c, nil, "Password updated successfully")
}

func (h *UserHandler) GetAllUsers(c fiber.Ctx) error {
	roles, _ := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusForbidden, "You are not authorized to view users"))
	}

	sortBy := c.Query("sort_by", "no")
	sortOrder := c.Query("sort_order", "asc")
	search := c.Query("search")
	page := fiber.Query[int](c, "page", 1)
	perPage := fiber.Query[int](c, "per_page", 10)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	// optional column filters
	filters := map[string][]string{}
	keys := []string{"roles"}
	for _, key := range keys {
		values := c.RequestCtx().QueryArgs().PeekMulti(key)
		if len(values) > 0 {
			strs := make([]string, len(values))
			for i, v := range values {
				strs[i] = string(v)
			}
			filters[key] = strs
		}
	}

	users, total, err := h.service.GetAllPaginated(search, page, perPage, sortBy, sortOrder, filters)
	if err != nil {
		return respondError(c, err)
	}

	meta := utils.NewPaginationMeta(total, page, perPage)

	return respondOKWithMeta(c, users, "Users retrieved successfully", meta)
}

func (h *UserHandler) GetRoleOptions(c fiber.Ctx) error {
	roles, _ := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusForbidden, "You are not authorized to view roles"))
	}

	roleOptions, err := h.service.GetRoleOptions()
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, roleOptions, "Roles retrieved successfully")
}

func (h *UserHandler) GetUserByID(c fiber.Ctx) error {
	id := c.Params("id")
	roles, _ := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusForbidden, "You are not authorized to view user"))
	}

	user, err := h.service.GetUserByID(id)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, user, "User retrieved successfully")
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	roles, _ := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusForbidden, "You are not authorized to create user"))
	}

	var req dto.CreateUserRequest

	if err := c.Bind().Body(&req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.validate.Struct(&req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.service.CreateUser(&req); err != nil {
		return respondError(c, err)
	}

	return respondCreated(c, nil, "User created successfully")
}

func (h *UserHandler) UpdateUser(c fiber.Ctx) error {
	roles, _ := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusForbidden, "You are not authorized to update user"))
	}

	id := c.Params("id")
	var req dto.UpdateUserRequest

	if err := c.Bind().Body(&req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.validate.Struct(&req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.service.UpdateUser(id, &req); err != nil {
		return respondError(c, err)
	}

	return respondOK(c, nil, "User updated successfully")
}

func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	roles, _ := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusForbidden, "You are not authorized to delete user"))
	}

	id := c.Params("id")

	if err := h.service.DeleteUser(id); err != nil {
		return respondError(c, err)
	}

	return respondNoContent(c)
}

// fiber:context-methods migrated
