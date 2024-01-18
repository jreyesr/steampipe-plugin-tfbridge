package tfbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	tfaddr "github.com/hashicorp/terraform-registry-address"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type discoveryResponse struct {
	Providers string `json:"providers.v1"`
}

type pluginVersionResponse struct {
	DownloadURL string `json:"download_url"`
}

// This function contains some parts of code that is Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
// see https://github.com/hashicorp/terraform/blob/e26d07dda41a74a009b1b750471395bf8773601c/internal/providercache/cached_provider.go#L106
/*
isProviderBinary checks if the passed file could be the actual provider binary inside of a downloaded (decompressed) directory, which may hold extra files.
README and LICENSE have been spotted in addition to the actual provider binary

For a provider named e.g. tf.example.com/awesomecorp/happycloud, we
expect an executable file whose name starts with
"terraform-provider-happycloud", followed by zero or more additional
characters. If there _are_ additional characters then the first one
must be an underscore or a period
*/
func isProviderBinary(ctx context.Context, fname string, provider tfaddr.Provider) (bool, error) {
	expectedNameStart := "terraform-provider-" + provider.Type
	expectedPeriod := expectedNameStart + "."
	expectedUnderscore := expectedNameStart + "_"

	fileInfo, err := os.Stat(fname)
	if err != nil {
		plugin.Logger(ctx).Info("findProviderBinary", "error", err, "filename", fname)
		return false, err
	}

	if fileInfo.IsDir() {
		return false, nil
	}
	// There's no platform-independent way of checking the executableness of a file, so we don't try
	// IOW, we are just searching for a real file (non-dir) that starts with the magic words
	if fileInfo.Name() == expectedNameStart || strings.HasPrefix(fileInfo.Name(), expectedPeriod) || strings.HasPrefix(fileInfo.Name(), expectedUnderscore) {
		return true, nil
	}
	return false, nil
}

/*
DownloadProvider uses the Terraform Registry API to download a provider binary into the temp directory.
*/
func DownloadProvider(ctx context.Context, name, version string, d *plugin.TableMapData) (path string, err error) {
	provider, err := tfaddr.ParseProviderSource(name)
	if err != nil {
		return
	}

	// NOTE: Around august 2023, Hashicorp changed the Terraform Registry's ToS so the only allowed use
	// is "for use with, or in support of, HashiCorp Terraform"
	// Since this isn't TF, we probably can't download providers from there
	// Instead, providers are downloaded from the open-source OpenTofu project (https://opentofu.org/),
	// which doesn't seem to have such limitations
	// See https://github.com/opentffoundation/roadmap/issues/24#issuecomment-1699535216
	// Note that this also covers the case where the user of the Steampipe plugin explicitly
	// passes "registry.terraform.io/org/provider" as the source,
	// since even in those cases we can't use the official TF registry
	if provider.Hostname == tfaddr.DefaultProviderRegistryHost {
		newDefaultRegistry := svchost.Hostname("registry.opentofu.org")
		plugin.Logger(ctx).Info("resolveServiceLicensingFix", "oldProvider", provider.Hostname, "newProvider", newDefaultRegistry)
		provider.Hostname = newDefaultRegistry
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
	plugin.Logger(ctx).Debug("getProviderInfo", "url", pluginVersionInfoUrl)
	resp, err := http.Get(pluginVersionInfoLocation)
	if err != nil {
		return
	}
	// extra-friendly error message for 404, since that one should be most common
	if resp.StatusCode == 404 {
		err = fmt.Errorf("download: Terraform provider %s, version %s does not exist! Please check that it is available for your OS and arch", provider.ForDisplay(), version)
		return
	}
	if resp.StatusCode != 200 {
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
	for _, p := range paths {
		canBe, detectErr := isProviderBinary(ctx, p, provider)
		if err != nil {
			err = detectErr
			return
		}
		if canBe {
			path = p
			return
		}
	}

	// if we reached here, we couldn't find a binary candidate
	err = fmt.Errorf("couldn't find binary for provider %s among files %v", provider.ForDisplay(), paths)
	return
}
