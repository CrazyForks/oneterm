package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/spf13/cast"

	"github.com/veops/oneterm/internal/acl"
	"github.com/veops/oneterm/internal/model"
	"github.com/veops/oneterm/internal/service"
	"github.com/veops/oneterm/pkg/config"
	"github.com/veops/oneterm/pkg/errors"
)

var (
	gatewayService = service.NewGatewayService()

	gatewayPreHooks = []preHook[*model.Gateway]{
		// Validate public key
		func(ctx *gin.Context, data *model.Gateway) {
			if err := gatewayService.ValidatePublicKey(data); err != nil {
				ctx.AbortWithError(http.StatusBadRequest, &errors.ApiError{Code: errors.ErrWrongPk, Data: nil})
				return
			}
		},
		// Encrypt sensitive data
		func(ctx *gin.Context, data *model.Gateway) {
			gatewayService.EncryptSensitiveData(data)
		},
	}

	gatewayPostHooks = []postHook[*model.Gateway]{
		// Attach asset count
		func(ctx *gin.Context, data []*model.Gateway) {
			if err := gatewayService.AttachAssetCount(ctx, data); err != nil {
				return
			}
		},
		// Decrypt sensitive data
		func(ctx *gin.Context, data []*model.Gateway) {
			gatewayService.DecryptSensitiveData(data)
		},
	}

	gatewayDcs = []deleteCheck{
		// Check dependencies
		func(ctx *gin.Context, id int) {
			assetName, err := gatewayService.CheckAssetDependencies(ctx, id)
			if err == nil && assetName == "" {
				return
			}
			code := lo.Ternary(err == nil, http.StatusBadRequest, http.StatusInternalServerError)
			err = lo.Ternary[error](err == nil, &errors.ApiError{Code: errors.ErrHasDepency, Data: map[string]any{"name": assetName}}, err)
			ctx.AbortWithError(code, err)
		},
	}
)

// CreateGateway godoc
//
//	@Tags		gateway
//	@Param		gateway	body		model.Gateway	true	"gateway"
//	@Success	200		{object}	HttpResponse
//	@Router		/gateway [post]
func (c *Controller) CreateGateway(ctx *gin.Context) {
	doCreate(ctx, true, &model.Gateway{}, config.RESOURCE_GATEWAY, gatewayPreHooks...)
}

// DeleteGateway godoc
//
//	@Tags		gateway
//	@Param		id	path		int	true	"gateway id"
//	@Success	200	{object}	HttpResponse
//	@Router		/gateway/:id [delete]
func (c *Controller) DeleteGateway(ctx *gin.Context) {
	doDelete(ctx, true, &model.Gateway{}, config.RESOURCE_GATEWAY, gatewayDcs...)
}

// UpdateGateway godoc
//
//	@Tags		gateway
//	@Param		id		path		int				true	"gateway id"
//	@Param		gateway	body		model.Gateway	true	"gateway"
//	@Success	200		{object}	HttpResponse
//	@Router		/gateway/:id [put]
func (c *Controller) UpdateGateway(ctx *gin.Context) {
	doUpdate(ctx, true, &model.Gateway{}, config.RESOURCE_GATEWAY, gatewayPreHooks...)
}

// GetGateways godoc
//
//	@Tags		gateway
//	@Param		page_index	query		int		true	"gateway id"
//	@Param		page_size	query		int		true	"gateway id"
//	@Param		search		query		string	false	"name or host or account or port"
//	@Param		id			query		int		false	"gateway id"
//	@Param		ids			query		string	false	"gateway ids"
//	@Param		name		query		string	false	"gateway name"
//	@Param		info		query		bool	false	"is info mode"
//	@Param		type		query		int		false	"account type"
//	@Success	200			{object}	HttpResponse{data=ListData{list=[]model.Gateway}}
//	@Router		/gateway [get]
func (c *Controller) GetGateways(ctx *gin.Context) {
	currentUser, _ := acl.GetSessionFromCtx(ctx)
	info := cast.ToBool(ctx.Query("info"))

	// Build base query using service layer
	db := gatewayService.BuildQuery(ctx)

	// Apply authorization filter if needed
	if info && !acl.IsAdmin(currentUser) {
		// Use V2 authorization system for asset filtering
		authV2Service := service.NewAuthorizationV2Service()
		_, assetIds, _, err := authV2Service.GetAuthorizationScopeByACL(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, &errors.ApiError{Code: errors.ErrInternal, Data: map[string]any{"err": err}})
			return
		}

		// Apply gateway filter by asset IDs
		db = gatewayService.FilterByAssetIds(db, assetIds)
	}

	doGet(ctx, !info, db, config.RESOURCE_GATEWAY, gatewayPostHooks...)
}
