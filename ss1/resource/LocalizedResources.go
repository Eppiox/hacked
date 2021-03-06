package resource

// LocalizedResources associates a language with a resource provider under a specific identifier.
type LocalizedResources struct {
	// ID is the identifier of the provider. This could be a filename for instance.
	ID string
	// Language specifies for which language the provider has resources.
	Language Language
	// Provider is the actual container of the resources.
	Provider Provider
}
