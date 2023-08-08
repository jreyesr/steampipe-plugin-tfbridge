package tfbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	tfaddr "github.com/hashicorp/terraform-registry-address"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type discoveryResponse struct {
	Providers string `json:"providers.v1"`
}

type pluginVersionResponse struct {
	DownloadURL string `json:"download_url"`
}

/*
DownloadProvider uses the Terraform Registry API to download a provider binary into the temp directory.
*/
func DownloadProvider(ctx context.Context, name, version string, d *plugin.TableMapData) (path string, err error) {
	provider, err := tfaddr.ParseProviderSource(name)
	if err != nil {
		return
	}
	hostnameUrl, err := url.Parse(fmt.Sprintf("https://%s", provider.Hostname.String()))
	if err != nil {
		return
	}

	providerUrl, err := disco.New().DiscoverServiceURL(provider.Hostname, "providers.v1")
	if err != nil {
		return
	}
	// mess with the providerUrl so it's referenced to the hostname
	providerUrl = hostnameUrl.ResolveReference(providerUrl)
	plugin.Logger(ctx).Info("resolveService", "hostnameUrl", hostnameUrl, "providerUrl", providerUrl)

	pluginVersionInfoLocation := fmt.Sprintf("%s%s/%s/%s/download/%s/%s",
		providerUrl,
		provider.Namespace,
		provider.Type,
		version,
		runtime.GOOS,
		runtime.GOARCH,
	)
	pluginVersionInfoUrl, _ := url.Parse(pluginVersionInfoLocation)
	resp, err := http.Get(pluginVersionInfoLocation)
	if err != nil {
		return
	}
	// extra-friendly error message for 404, since that one should be most common
	if resp.StatusCode == 404 {
		err = fmt.Errorf("download: Terraform provider %s, version %s does not exist! Please check that it is available for your OS and arch", provider.ForDisplay(), version)
		return
	}
	if resp.StatusCode != 200 || resp.Header.Get("content-type") != "application/json" {
		err = fmt.Errorf("get version response invalid: code %d, contenttype %s", resp.StatusCode, resp.Header.Get("content-type"))
		return
	}
	defer resp.Body.Close()

	var pluginVersion pluginVersionResponse
	err = json.NewDecoder(resp.Body).Decode(&pluginVersion)
	if err != nil {
		return
	}
	plugin.Logger(ctx).Info("getVersionInfo", "sourceUrl", pluginVersionInfoLocation, "downloadUrl", pluginVersion.DownloadURL)

	pluginDownloadUrl, err := url.Parse(pluginVersion.DownloadURL)
	if err != nil {
		return
	}
	// "If this [i.e. download_url] is a relative URL then it will be resolved relative to the URL that returned the containing JSON object."
	pluginDownloadUrl = pluginVersionInfoUrl.ResolveReference(pluginDownloadUrl)

	paths, err := d.GetSourceFiles(pluginDownloadUrl.String())
	if err != nil {
		return
	}
	if len(paths) != 1 {
		err = fmt.Errorf("unexpected number of files: %d %v, expected 1 file", len(paths), paths)
		return
	}

	path = paths[0]
	// path = "/home/reyes/code/steampipe-plugin-tfbridge/terraform-provider-dns_v3.2.4_x5"
	return
}
