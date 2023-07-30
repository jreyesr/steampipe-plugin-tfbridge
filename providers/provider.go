// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providers

import (
	"github.com/zclconf/go-cty/cty"

	"github.com/jreyesr/steampipe-plugin-tfbridge/configschema"
	// "github.com/jreyesr/steampipe-plugin-tfbridge/states"
	"github.com/jreyesr/steampipe-plugin-tfbridge/tfdiags"
)

// Interface represents the set of methods required for a complete resource
// provider plugin.
type Interface interface {
	// GetSchema returns the complete schema for the provider.
	GetProviderSchema() GetProviderSchemaResponse

	// ValidateProviderConfig allows the provider to validate the configuration.
	// The ValidateProviderConfigResponse.PreparedConfig field is unused. The
	// final configuration is not stored in the state, and any modifications
	// that need to be made must be made during the Configure method call.
	ValidateProviderConfig(ValidateProviderConfigRequest) ValidateProviderConfigResponse

	// ValidateDataResourceConfig allows the provider to validate the data source
	// configuration values.
	ValidateDataResourceConfig(ValidateDataResourceConfigRequest) ValidateDataResourceConfigResponse

	// Configure configures and initialized the provider.
	ConfigureProvider(ConfigureProviderRequest) ConfigureProviderResponse

	// Stop is called when the provider should halt any in-flight actions.
	//
	// Stop should not block waiting for in-flight actions to complete. It
	// should take any action it wants and return immediately acknowledging it
	// has received the stop request. Terraform will not make any further API
	// calls to the provider after Stop is called.
	//
	// The error returned, if non-nil, is assumed to mean that signaling the
	// stop somehow failed and that the user should expect potentially waiting
	// a longer period of time.
	Stop() error

	// ReadDataSource returns the data source's current state.
	ReadDataSource(ReadDataSourceRequest) ReadDataSourceResponse

	// Close shuts down the plugin process if applicable.
	Close() error
}

// GetProviderSchemaResponse is the return type for GetProviderSchema, and
// should only be used when handling a value for that method. The handling of
// of schemas in any other context should always use ProviderSchema, so that
// the in-memory representation can be more easily changed separately from the
// RCP protocol.
type GetProviderSchemaResponse struct {
	// Provider is the schema for the provider itself.
	Provider Schema

	// ProviderMeta is the schema for the provider's meta info in a module
	ProviderMeta Schema

	// ResourceTypes map the resource type name to that type's schema.
	ResourceTypes map[string]Schema

	// DataSources maps the data source name to that data source's schema.
	DataSources map[string]Schema

	// Diagnostics contains any warnings or errors from the method call.
	Diagnostics tfdiags.Diagnostics

	// ServerCapabilities lists optional features supported by the provider.
	ServerCapabilities ServerCapabilities
}

// Schema pairs a provider or resource schema with that schema's version.
// This is used to be able to upgrade the schema in UpgradeResourceState.
//
// This describes the schema for a single object within a provider. Type
// "Schemas" (plural) instead represents the overall collection of schemas
// for everything within a particular provider.
type Schema struct {
	Version int64
	Block   *configschema.Block
}

// ServerCapabilities allows providers to communicate extra information
// regarding supported protocol features. This is used to indicate availability
// of certain forward-compatible changes which may be optional in a major
// protocol version, but cannot be tested for directly.
type ServerCapabilities struct {
	// PlanDestroy signals that this provider expects to receive a
	// PlanResourceChange call for resources that are to be destroyed.
	PlanDestroy bool

	// The GetProviderSchemaOptional capability indicates that this
	// provider does not require calling GetProviderSchema to operate
	// normally, and the caller can used a cached copy of the provider's
	// schema.
	GetProviderSchemaOptional bool
}

type ValidateProviderConfigRequest struct {
	// Config is the raw configuration value for the provider.
	Config cty.Value
}

type ValidateProviderConfigResponse struct {
	// PreparedConfig is unused and will be removed with support for plugin protocol v5.
	PreparedConfig cty.Value
	// Diagnostics contains any warnings or errors from the method call.
	Diagnostics tfdiags.Diagnostics
}

type ValidateDataResourceConfigRequest struct {
	// TypeName is the name of the data source type to validate.
	TypeName string

	// Config is the configuration value to validate, which may contain unknown
	// values.
	Config cty.Value
}

type ValidateDataResourceConfigResponse struct {
	// Diagnostics contains any warnings or errors from the method call.
	Diagnostics tfdiags.Diagnostics
}

type ConfigureProviderRequest struct {
	// Terraform version is the version string from the running instance of
	// terraform. Providers can use TerraformVersion to verify compatibility,
	// and to store for informational purposes.
	TerraformVersion string

	// Config is the complete configuration value for the provider.
	Config cty.Value
}

type ConfigureProviderResponse struct {
	// Diagnostics contains any warnings or errors from the method call.
	Diagnostics tfdiags.Diagnostics
}

type ReadDataSourceRequest struct {
	// TypeName is the name of the data source type to Read.
	TypeName string

	// Config is the complete configuration for the requested data source.
	Config cty.Value

	// ProviderMeta is the configuration for the provider_meta block for the
	// module and provider this resource belongs to. Its use is defined by
	// each provider, and it should not be used without coordination with
	// HashiCorp. It is considered experimental and subject to change.
	ProviderMeta cty.Value
}

type ReadDataSourceResponse struct {
	// State is the current state of the requested data source.
	State cty.Value

	// Diagnostics contains any warnings or errors from the method call.
	Diagnostics tfdiags.Diagnostics
}
