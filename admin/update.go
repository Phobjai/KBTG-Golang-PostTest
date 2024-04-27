package admin

import (
	"net/http"

	"github.com/Phobjai/assessment-tax/initdb"
	"github.com/labstack/echo/v4"
)

// UpdatePersonalDeduction handles the request to update personal deductions.
func UpdatePersonalDeduction(c echo.Context) error {
	var req DeductionUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Err{"Invalid request"})
	}

	if req.Amount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{"Amount cannot be greater than 100000"})
	}

	if _, err := initdb.DB.Exec("UPDATE admin_config SET deduction = $1", req.Amount); err != nil {
		return c.JSON(http.StatusInternalServerError, Err{"Failed to update database"})
	}

	return c.JSON(http.StatusOK, DeductionUpdateResponse{PersonalDeduction: req.Amount})
}
