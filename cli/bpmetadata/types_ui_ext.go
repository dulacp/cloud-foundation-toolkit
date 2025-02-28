package bpmetadata

type ExtensionType string

const (
	Undefined ExtensionType = "UNDEFINED_TYPE"

	// General formats.
	EmailAddress    ExtensionType = "EMAIL_ADDRESS"
	MultiLineString ExtensionType = "MULTI_LINE_STRING"

	// GCE related.
	GCEDiskImage       ExtensionType = "GCE_DISK_IMAGE"
	GCEDiskType        ExtensionType = "GCE_DISK_TYPE"
	GCEDiskSize        ExtensionType = "GCE_DISK_SIZE"
	GCEMachineType     ExtensionType = "GCE_MACHINE_TYPE"
	GCENetwork         ExtensionType = "GCE_NETWORK"
	GCEZone            ExtensionType = "GCE_ZONE"
	GCESubnetwork      ExtensionType = "GCE_SUBNETWORK"
	GCERegion          ExtensionType = "GCE_REGION"
	GCEGPUType         ExtensionType = "GCE_GPU_TYPE"
	GCEGPUCount        ExtensionType = "GCE_GPU_COUNT"
	GCEExternalIP      ExtensionType = "GCE_EXTERNAL_IP"
	GCEIPForwarding    ExtensionType = "GCE_IP_FORWARDING"
	GCEFirewall        ExtensionType = "GCE_FIREWALL"
	GCEFirewallRange   ExtensionType = "GCE_FIREWALL_RANGE"
	GCEGenericResource ExtensionType = "GCE_GENERIC_RESOURCE"

	// GCS related.
	GCSBucket ExtensionType = "GCS_BUCKET"

	// IAM related.
	IAMServiceAccount ExtensionType = "IAM_SERVICE_ACCOUNT"
)

// An extension for variables defined as part of DisplayVariable. The
// extension defines Google-specifc metadata necessary for choosing an
// appropriate input widget or adding restrictions to GCP-specific resources.
type GooglePropertyExtension struct {
	Type ExtensionType `json:"type" yaml:"type" jsonschema:"enum=EMAIL_ADDRESS,enum=MULTI_LINE_STRING,enum=GCE_DISK_IMAGE,enum=GCE_DISK_TYPE,enum=GCE_DISK_SIZE,enum=GCE_MACHINE_TYPE,enum=GCE_NETWORK,enum=GCE_ZONE,enum=GCE_SUBNETWORK,enum=GCE_REGION,enum=GCE_GPU_TYPE,enum=GCE_GPU_COUNT,enum=GCE_EXTERNAL_IP,enum=GCE_IP_FORWARDING,enum=GCE_FIREWALL,enum=GCE_FIREWALL_RANGE,enum=GCE_GENERIC_RESOURCE,enum=GCS_BUCKET,enum=IAM_SERVICE_ACCOUNT"`

	// Some properties (e.g. GCE_MACHINE_TYPE) require a zone context in order to
	// determine the set of allowable values. This field references another
	// property from the schema, which must have type GCE_ZONE.
	ZoneProperty string `json:"zoneProperty,omitempty" yaml:"zoneProperty,omitempty"`

	// Property-specific extensions.
	GCEMachineType    GCEMachineTypeExtension     `json:"gceMachineType,omitempty" yaml:"gceMachineType,omitempty"`
	GCEDiskSize       GCEDiskSizeExtension        `json:"gceDiskSize,omitempty" yaml:"gceDiskSize,omitempty"`
	GCESubnetwork     GCESubnetworkExtension      `json:"gceSubnetwork,omitempty" yaml:"gceSubnetwork,omitempty"`
	GCEResource       GCEGenericResourceExtension `json:"gceResource,omitempty" yaml:"gceResource,omitempty"`
	GCEGPUType        GCEGPUTypeExtension         `json:"gceGpuType,omitempty" yaml:"gceGpuType,omitempty"`
	GCEGPUCount       GCEGPUCountExtension        `json:"gceGpuCount,omitempty" yaml:"gceGpuCount,omitempty"`
	GCENetwork        GCENetworkExtension         `json:"gceNetwork,omitempty" yaml:"gceNetwork,omitempty"`
	GCEExternalIP     GCEExternalIPExtension      `json:"gceExternalIp,omitempty" yaml:"gceExternalIp,omitempty"`
	GCEIPForwarding   GCEIPForwardingExtension    `json:"gceIpForwarding,omitempty" yaml:"gceIpForwarding,omitempty"`
	GCEFirewall       GCEFirewallExtension        `json:"gceFirewall,omitempty" yaml:"gceFirewall,omitempty"`
	GCEFirewallRange  GCEFirewallRangeExtension   `json:"gceFirewallRange,omitempty" yaml:"gceFirewallRange,omitempty"`
	GCEZone           GCELocationExtension        `json:"gceZone,omitempty" yaml:"gceZone,omitempty"`
	GCERegion         GCELocationExtension        `json:"gceRegion,omitempty" yaml:"gceRegion,omitempty"`
	IAMServiceAccount IAMServiceAccountExtension  `json:"iamServiceAccount,omitempty" yaml:"iamServiceAccount,omitempty"`
}

type GCELocationExtension struct {
	AllowlistedZones   []string `json:"allowlistedZones,omitempty" yaml:"allowlistedZones,omitempty"`
	AllowlistedRegions []string `json:"allowlistedRegions,omitempty" yaml:"allowlistedRegions,omitempty"`
}

type GCEMachineTypeExtension struct {
	// Minimum cpu. Used to filter the list of selectable machine types.
	MinCPU int `json:"minCpu,omitempty" yaml:"minCpu,omitempty"`

	// Minimum ram. Used to filter the list of selectable machine types.
	MinRAMGB int `json:"minRamGb,omitempty" yaml:"minRamGb,omitempty"`

	// If true, custom machine types will not be selectable.
	// More info:
	// https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type
	DisallowCustomMachineTypes bool `json:"disallowCustomMachineTypes,omitempty" yaml:"disallowCustomMachineTypes,omitempty"`
}

type GCEGPUTypeExtension struct {
	MachineType string `json:"machineType" yaml:"machineType"`
	GPUType     string `json:"gpuType,omitempty" yaml:"gpuType,omitempty"`
}

type GCEGPUCountExtension struct {
	// This field references another variable from the schema,
	// which must have type GCEMachineType.
	MachineTypeVariable string `json:"machineTypeVariable" yaml:"machineTypeVariable"`
}

type GCEDiskSizeExtension struct {
	// The allowable range of disk sizes depends on the disk type. This field
	// references another variable from the schema, which must have type GCEDiskType.
	DiskTypeVariable string `json:"diskTypeVariable" yaml:"diskTypeVariable"`
}

type GCENetworkExtension struct {
	// AllowSharedVpcs indicates this solution can receive
	// shared VPC selflinks (fully qualified compute links).
	AllowSharedVPCs bool `json:"allowSharedVpcs,omitempty" yaml:"allowSharedVpcs,omitempty"`
	// Used to indicate to which machine type this network interface will be
	// attached to.
	MachineTypeVariable string `json:"machineTypeVariable" yaml:"machineTypeVariable"`
}

type ExternalIPType string

const (
	IPEphemeral ExternalIPType = "EPHEMERAL"
	IPStaic     ExternalIPType = "STATIC"
)

type GCEExternalIPExtension struct {
	// NetworkVariable is used to indicate the network variable in the schema
	// this external IP belongs to.
	NetworkVariable string `json:"networkVariable" yaml:"networkVariable"`

	// Type specifies if the external IP is ephemeral or static.
	// Defaults to ephemeral if not specified.
	Type ExternalIPType `json:"externalIpType,omitempty" yaml:"externalIpType,omitempty"`
}

type GCEIPForwardingExtension struct {
	// NetworkVariable is used to indicate the network variable in the schema
	// this external IP belongs to.
	NetworkVariable string `json:"networkVariable" yaml:"networkVariable"`
	NotConfigurable bool   `json:"notConfigurable,omitempty" yaml:"notConfigurable,omitempty"`
}

type GCEFirewallExtension struct {
	// NetworkVariable is used to indicate the network variable in the schema
	// this external IP belongs to.
	NetworkVariable string `json:"networkVariable" yaml:"networkVariable"`
}

type GCEFirewallRangeExtension struct {
	// FirewallVariable is used to indicate the firewall variable with the type
	// GCEFirewall in the schema to which this firewall range belongs to.
	FirewallVariable string `json:"firewallVariable" yaml:"firewallVariable"`
}

type GCESubnetworkExtension struct {
	// Subnetwork variable requires a network context in order to determine the
	// set of available subnetworks. This field references another
	// variable from the schema, which must have type GCENetwork.
	NetworkVariable string `json:"networkVariable" yaml:"networkVariable"`
}

type GCEGenericResourceExtension struct {
	// GCE resource type to be fetched. This field references another
	// property from the schema, which must have type GCEGenericResource.
	ResourceVariable string `json:"resourceVariable" yaml:"resourceVariable"`
}

type IAMServiceAccountExtension struct {
	// List of IAM roles that to  grant to a new SA, or the roles to filter
	// existing SAs with.
	Roles []string `json:"roles" yaml:"roles"`
}
