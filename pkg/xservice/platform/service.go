package platform

import (
	"context"
	"errors"
	"message-core/pkg/xhttp"
	"os"
)

var (
	httpClient xhttp.Client
	baseUrl    string
)

func NewClien() {
	httpClient = xhttp.NewClient()
	baseUrl = os.Getenv("PLATFORM_BASE_URL")
}

func ValidationUser(
	ctx context.Context,
	req GatewayValidationRequest,
) (err error) {
	userCache, _ := GetUserCache(ctx, req.UserName, req.Password)
	if userCache.UserState == "Validated" {
		// if valid user from cache set key rule cache
		go SetRuleCache(context.Background(), req.UserName, userCache)
		return nil
	}

	if userCache.UserState == "Invalid" {
		return errors.New("Invalid user name or password")
	}

	// http://host.docker.internal
	var resp PlatformBaseResponse
	path := baseUrl + "/api/internal/v1/topics/validation"
	xopt := xhttp.RequestOption{GroupPath: "api/internal/v1/topics/validation"}
	if _, err = httpClient.PostJSON(ctx, path, &req, &resp, xopt); err != nil {
		return
	}
	if resp.StatusCode != 200 {
		go SetUserCache(
			context.Background(),
			req.UserName,
			req.Password,
			UserCacheModel{
				UserState: "Invalid",
			})
		return errors.New("Error when validate user from patform.")
	}

	go SetUserCache(
		context.Background(),
		req.UserName,
		req.Password,
		UserCacheModel{
			UserState: "Validated",
			Rules:     resp.Data.RulesDevices,
		})

	return
}
