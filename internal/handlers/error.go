package handlers

import (
	"errors"
	"fmt"

	apperrors "jejak/internal/errors"
	"jejak/internal/helpers"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// respondOK sends a 200 response with the standard envelope.
func respondOK(c fiber.Ctx, data any, message string) error {
	return c.Status(fiber.StatusOK).JSON(helpers.Response{
		Data:    data,
		Message: message,
	})
}

// respondCreated sends a 201 response with the standard envelope.
func respondCreated(c fiber.Ctx, data any, message string) error {
	return c.Status(fiber.StatusCreated).JSON(helpers.Response{
		Data:    data,
		Message: message,
	})
}

// respondOKWithMeta sends a 200 response with data, message, and pagination meta.
func respondOKWithMeta(c fiber.Ctx, data any, message string, meta any) error {
	return c.Status(fiber.StatusOK).JSON(helpers.Response{
		Data:    data,
		Message: message,
		Meta:    meta,
	})
}

// respondValidation sends a 400 response.
// For validator.ValidationErrors it expands field-level messages into the errors array.
func respondValidation(c fiber.Ctx, err error) error {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errs := make([]string, len(ve))
		for i, fe := range ve {
			errs[i] = fmt.Sprintf("%s: failed on '%s'", fe.Field(), fe.Tag())
		}
		return c.Status(fiber.StatusBadRequest).JSON(helpers.Response{
			Data:    nil,
			Message: "Validation failed",
			Errors:  errs,
		})
	}
	return c.Status(fiber.StatusBadRequest).JSON(helpers.Response{
		Data:    nil,
		Message: "Invalid request body",
	})
}

// respondError maps an error to the appropriate HTTP response.
// If err is an *HttpError, the embedded status code and message are used.
// Otherwise a generic 500 Internal Server Error is returned.
func respondError(c fiber.Ctx, err error) error {
	var httpErr *apperrors.HttpError
	if errors.As(err, &httpErr) {
		return c.Status(httpErr.Code).JSON(helpers.Response{
			Data:    nil,
			Message: httpErr.Message,
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(helpers.Response{
		Data:    nil,
		Message: "Internal server error",
	})
}

// respondNoContent sends a 204 No Content response.
func respondNoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// buildErrResponse returns a simple error envelope for inline JSON responses.
func buildErrResponse(message string) fiber.Map {
	return fiber.Map{"message": message}
}
