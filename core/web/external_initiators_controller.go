package web

import (
	"errors"
	"net/http"

	"chainlink/core/auth"
	"chainlink/core/services"
	"chainlink/core/store/models"
	"chainlink/core/store/presenters"

	"github.com/gin-gonic/gin"
)

// ExternalInitiatorsController manages external initiators
type ExternalInitiatorsController struct {
	App services.Application
}

// Create builds and saves a new service agreement record.
func (eic *ExternalInitiatorsController) Create(c *gin.Context) {
	eir := &models.ExternalInitiatorRequest{}
	if !eic.App.GetStore().Config.Dev() && !eic.App.GetStore().Config.FeatureExternalInitiators() {
		jsonAPIError(c,
			http.StatusMethodNotAllowed,
			errors.New("The External Initiator feature is disabled by configuration"),
		)
		return
	}

	eia := auth.NewToken()
	if err := c.ShouldBindJSON(eir); err != nil {
		jsonAPIError(c, http.StatusUnprocessableEntity, err)
	} else if ei, err := models.NewExternalInitiator(eia, eir); err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
	} else if err := services.ValidateExternalInitiator(eir, eic.App.GetStore()); err != nil {
		jsonAPIError(c, http.StatusBadRequest, err)
	} else if err := eic.App.GetStore().CreateExternalInitiator(ei); err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
	} else {
		resp := presenters.NewExternalInitiatorAuthentication(*ei, *eia)
		jsonAPIResponseWithStatus(c, resp, "external initiator authentication", http.StatusCreated)
	}
}

// Destroy deletes an ExternalInitiator
func (eic *ExternalInitiatorsController) Destroy(c *gin.Context) {
	if !eic.App.GetStore().Config.Dev() {
		jsonAPIError(c, http.StatusMethodNotAllowed, errors.New("External Initiators are currently under development and not yet usable outside of development mode"))
		return
	}

	id := c.Param("AccessKey")
	if err := eic.App.GetStore().DeleteExternalInitiator(id); err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
	} else {
		jsonAPIResponseWithStatus(c, nil, "external initiator", http.StatusNoContent)
	}
}
