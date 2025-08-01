package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"

	"github.com/veops/oneterm/internal/acl"
	"github.com/veops/oneterm/internal/model"
	"github.com/veops/oneterm/internal/service"
	myErrors "github.com/veops/oneterm/pkg/errors"
)

// UpsertAuthorization godoc
//
//	@Tags		authorization
//	@Param		authorization	body		model.Authorization	true	"authorization"
//	@Success	200				{object}	HttpResponse
//	@Router		/authorization [post]
func (c *Controller) UpsertAuthorization(ctx *gin.Context) {
	auth := &model.Authorization{}
	err := ctx.ShouldBindBodyWithJSON(auth)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, &myErrors.ApiError{Code: myErrors.ErrInvalidArgument, Data: map[string]any{"err": err}})
		return
	}

	if !service.DefaultAuthService.HasPermAuthorization(ctx, auth, acl.GRANT) {
		err = &myErrors.ApiError{Code: myErrors.ErrNoPerm, Data: map[string]any{"perm": acl.GRANT}}
		ctx.AbortWithError(http.StatusForbidden, err)
		return
	}

	// Use transaction processing
	err = service.DefaultAuthService.UpsertAuthorizationWithTx(ctx, auth)
	if err != nil {
		if ctx.IsAborted() {
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, &myErrors.ApiError{Code: myErrors.ErrInternal, Data: map[string]any{"err": err}})
		return
	}

	ctx.JSON(http.StatusOK, HttpResponse{
		Data: map[string]any{
			"id": auth.GetId(),
		},
	})
}

// DeleteAccount godoc
//
//	@Tags		authorization
//	@Param		id	path		int	true	"authorization id"
//	@Success	200	{object}	HttpResponse
//	@Router		/authorization/:id [delete]
func (c *Controller) DeleteAuthorization(ctx *gin.Context) {
	authId := cast.ToInt(ctx.Param("id"))

	// Get authorization information through service
	auth, err := service.DefaultAuthService.GetAuthorizationById(ctx, authId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithError(http.StatusBadRequest, &myErrors.ApiError{Code: myErrors.ErrInvalidArgument, Data: map[string]any{"err": err}})
		} else {
			ctx.AbortWithError(http.StatusInternalServerError, &myErrors.ApiError{Code: myErrors.ErrInternal, Data: map[string]any{"err": err}})
		}
		return
	}

	if !service.DefaultAuthService.HasPermAuthorization(ctx, auth, acl.GRANT) {
		ctx.AbortWithError(http.StatusForbidden, &myErrors.ApiError{Code: myErrors.ErrNoPerm, Data: map[string]any{"perm": acl.GRANT}})
		return
	}

	// Delete authorization
	if err := service.DefaultAuthService.DeleteAuthorization(ctx, auth); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, &myErrors.ApiError{Code: myErrors.ErrInternal, Data: map[string]any{"err": err}})
		return
	}

	ctx.JSON(http.StatusOK, HttpResponse{
		Data: map[string]any{
			"id": auth.GetId(),
		},
	})
}

// GetAuthorizations godoc
//
//	@Tags		authorization
//	@Param		page_index	query		int	true	"page_index"
//	@Param		page_size	query		int	true	"page_size"
//	@Param		node_id		query		int	false	"node id"
//	@Param		asset_id	query		int	false	"asset id"
//	@Param		account_id	query		int	false	"account id"
//	@Success	200			{object}	HttpResponse{data=ListData{list=[]model.Account}}
//	@Router		/authorization [get]
func (c *Controller) GetAuthorizations(ctx *gin.Context) {
	nodeId := cast.ToInt(ctx.Query("node_id"))
	assetId := cast.ToInt(ctx.Query("asset_id"))
	accountId := cast.ToInt(ctx.Query("account_id"))

	auth := &model.Authorization{
		AssetId:   assetId,
		AccountId: accountId,
		NodeId:    nodeId,
	}

	if !service.DefaultAuthService.HasPermAuthorization(ctx, auth, acl.GRANT) {
		ctx.AbortWithError(http.StatusForbidden, &myErrors.ApiError{Code: myErrors.ErrNoPerm, Data: map[string]any{"perm": acl.GRANT}})
		return
	}

	auths, count, err := service.DefaultAuthService.GetAuthorizations(ctx, nodeId, assetId, accountId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, &myErrors.ApiError{Code: myErrors.ErrInternal, Data: map[string]any{"err": err}})
		return
	}

	// Convert to slice of any type
	authsAny := make([]any, 0, len(auths))
	for _, a := range auths {
		authsAny = append(authsAny, a)
	}

	result := &ListData{
		Count: count,
		List:  authsAny,
	}

	ctx.JSON(http.StatusOK, HttpResponse{
		Data: result,
	})
}

func getIdsByAuthorizationIds(ctx *gin.Context) (nodeIds, assetIds, accountIds []int) {
	authorizationIds, ok := ctx.Value(kAuthorizationIds).([]*model.AuthorizationIds)
	if !ok || len(authorizationIds) == 0 {
		return
	}
	return assetService.GetIdsByAuthorizationIds(ctx, authorizationIds)
}
