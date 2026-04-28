package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"

	"github.com/rs/zerolog/log"
)

// HandleAssetGet godocs
//
//	@Summary	Get Item by Asset ID
//	@Tags		Items
//	@Produce	json
//	@Param		id	path		string	true	"Asset ID"
//	@Success	200	{object}	repo.PaginationResult[repo.EntitySummary]{}
//	@Router		/v1/assets/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleAssetGet() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())
		assetIDParam := chi.URLParam(r, "id")
		assetIDParam = strings.ReplaceAll(assetIDParam, "-", "") // Remove dashes
		// Convert the asset ID to an int64
		assetID, err := strconv.ParseInt(assetIDParam, 10, 64)
		if err != nil {
			return err
		}
		pageParam := r.URL.Query().Get("page")
		var page int64 = -1
		if pageParam != "" {
			page, err = strconv.ParseInt(pageParam, 10, 32)
			if err != nil {
				return server.JSON(w, http.StatusBadRequest, "Invalid page number")
			}
		}

		pageSizeParam := r.URL.Query().Get("pageSize")
		var pageSize int64 = -1
		if pageSizeParam != "" {
			pageSize, err = strconv.ParseInt(pageSizeParam, 10, 32)
			if err != nil {
				return server.JSON(w, http.StatusBadRequest, "Invalid page size")
			}
		}

		items, err := ctrl.repo.Entities.QueryByAssetID(r.Context(), ctx.GID, repo.AssetID(assetID), int(page), int(pageSize))
		if err != nil {
			log.Err(err).Msg("failed to get item")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		return server.JSON(w, http.StatusOK, items)
	}
}

func (ctrl *V1Controller) HandleAssetMovementHistoryGet() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())
		assetIDParam := strings.ReplaceAll(chi.URLParam(r, "id"), "-", "")
		assetID, err := strconv.ParseInt(assetIDParam, 10, 64)
		if err != nil {
			return server.JSON(w, http.StatusBadRequest, "invalid asset id")
		}

		limit := 10
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil {
				limit = parsed
			}
		}

		rows, err := ctrl.repo.Entities.GetAssetMovementHistory(r.Context(), ctx.GID, assetID, limit)
		if err != nil {
			log.Err(err).Msg("failed to query asset movement history")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusOK, rows)
	}
}

func (ctrl *V1Controller) HandleAssetMovementHistoryCreate() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())

		var req repo.AssetMovementCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return server.JSON(w, http.StatusBadRequest, "invalid json payload")
		}

		assetIDParam := strings.ReplaceAll(chi.URLParam(r, "id"), "-", "")
		assetID, err := strconv.ParseInt(assetIDParam, 10, 64)
		if err != nil {
			return server.JSON(w, http.StatusBadRequest, "invalid asset id")
		}
		req.AssetID = assetID

		created, err := ctrl.repo.Entities.CreateAssetMovement(r.Context(), ctx.GID, ctx.UID, req)
		if err != nil {
			log.Err(err).Msg("failed to create asset movement")
			if strings.Contains(err.Error(), "override required") {
				return server.JSON(w, http.StatusConflict, err.Error())
			}
			if strings.Contains(err.Error(), "not found") {
				return server.JSON(w, http.StatusNotFound, err.Error())
			}
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		return server.JSON(w, http.StatusCreated, created)
	}
}
