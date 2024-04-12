package npm

import (
	"fmt"
	"net/http"

	"code.gitea.io/gitea/modules/json"
	"code.gitea.io/gitea/modules/log"
	npm_module "code.gitea.io/gitea/modules/packages/npm"
	"code.gitea.io/gitea/modules/proxy"
	"code.gitea.io/gitea/services/context"
)

func newHTTPClient(httpTransport *http.Transport) http.Client {
	if httpTransport == nil {
		httpTransport = &http.Transport{
			Proxy: proxy.Proxy(),
		}
	}

	return http.Client{
		Transport: httpTransport,
	}
}

func ProxyPackageMetadata(ctx *context.Context, packageName string) error {
	// TODO: check cache in case of error
	return RemoteProxyPackageMetadata(ctx, packageName)
}

type InvalidProxyResponseError struct {
	StatusCode int
	Status     string
}

func (err InvalidProxyResponseError) Error() string {
	return fmt.Sprintf("Invalid response received from remote registry: %d - %s", err.StatusCode, err.Status)
}

func RemoteProxyPackageMetadata(ctx *context.Context, packageName string) error {
	client := newHTTPClient(nil)
	registryResponse, e := client.Get("https://registry.npmjs.org/" + packageName)
	if e != nil {
		return e
	}

	if registryResponse.StatusCode < 200 || registryResponse.StatusCode >= 300 {
		return InvalidProxyResponseError{
			StatusCode: registryResponse.StatusCode,
			Status:     registryResponse.Status,
		}
	}

	var content npm_module.PackageMetadata
	if e := json.NewDecoder(registryResponse.Body).Decode(&content); e != nil {
		log.Error("Decoding JSON failed from NPM registry: %v", e)
	}

	ctx.JSON(http.StatusOK, content)
	return nil
}
