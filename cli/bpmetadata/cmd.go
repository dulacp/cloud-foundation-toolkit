package bpmetadata

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var mdFlags struct {
	path     string
	nested   bool
	force    bool
	display  bool
	validate bool
}

const (
	readmeFileName     = "README.md"
	tfVersionsFileName = "versions.tf"
	tfRolesFileName    = "test/setup/iam.tf"
	tfServicesFileName = "test/setup/main.tf"
	iconFilePath       = "assets/icon.png"
	modulesPath        = "modules"
	examplesPath       = "examples"
	metadataFileName   = "metadata.yaml"
	metadataApiVersion = "blueprints.cloud.google.com/v1alpha1"
	metadataKind       = "BlueprintMetadata"
)

func init() {
	viper.AutomaticEnv()

	Cmd.Flags().BoolVarP(&mdFlags.display, "display", "d", false, "Generate the display metadata used for UI rendering.")
	Cmd.Flags().BoolVarP(&mdFlags.force, "force", "f", false, "Force the generation of fresh metadata.")
	Cmd.Flags().StringVarP(&mdFlags.path, "path", "p", ".", "Path to the blueprint for generating metadata.")
	Cmd.Flags().BoolVar(&mdFlags.nested, "nested", true, "Flag for generating metadata for nested blueprint, if any.")
	Cmd.Flags().BoolVarP(&mdFlags.validate, "validate", "v", false, "Validate metadata against the schema definition.")
}

var Cmd = &cobra.Command{
	Use:   "metadata",
	Short: "Generates blueprint metatda",
	Long:  `Generates metadata.yaml for specified blueprint`,
	Args:  cobra.NoArgs,
	RunE:  generate,
}

// The top-level command function that generates metadata based on the provided flags
func generate(cmd *cobra.Command, args []string) error {
	wdPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working dir: %w", err)
	}

	// validate metadata if there is an argument passed into the command
	if mdFlags.validate {
		if err := validateMetadata(mdFlags.path, wdPath); err != nil {
			return err
		}

		return nil
	}

	currBpPath := mdFlags.path
	if !path.IsAbs(mdFlags.path) {
		currBpPath = path.Join(wdPath, mdFlags.path)
	}

	var allBpPaths []string
	_, err = os.Stat(path.Join(currBpPath, readmeFileName))

	// throw an error and exit if root level readme.md doesn't exist
	if err != nil {
		return fmt.Errorf("Top-level module does not have a readme. Details: %w\n", err)
	}

	allBpPaths = append(allBpPaths, currBpPath)
	var errors []string

	// if nested, check if modules/ exists and create paths
	// for submodules
	if mdFlags.nested {
		modulesPathforBp := path.Join(currBpPath, modulesPath)
		_, err = os.Stat(modulesPathforBp)
		if os.IsNotExist(err) {
			Log.Info("sub-modules do not exist for this blueprint")
		} else {
			moduleDirs, err := util.WalkTerraformDirs(modulesPathforBp)
			if err != nil {
				errors = append(errors, err.Error())
			} else {
				allBpPaths = append(allBpPaths, moduleDirs...)
			}
		}
	}

	for _, modPath := range allBpPaths {
		// check if module path has readme.md
		_, err := os.Stat(path.Join(modPath, readmeFileName))

		// log info if a sub-module doesn't have a readme.md and continue
		if err != nil {
			Log.Info("Skipping metadata for sub-module identified as an internal module", "Path:", modPath)
			continue
		}

		err = generateMetadataForBpPath(modPath)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}

	return nil
}

func generateMetadataForBpPath(bpPath string) error {
	//try to read existing metadata.yaml
	bpObj, err := UnmarshalMetadata(bpPath)
	if err != nil && !mdFlags.force {
		return err
	}

	// create metadata details
	bpMetaObj, err := CreateBlueprintMetadata(bpPath, bpObj)
	if err != nil {
		return fmt.Errorf("error creating metadata for blueprint at path: %s. Details: %w", bpPath, err)
	}

	// write metadata to disk
	err = WriteMetadata(bpMetaObj, bpPath)
	if err != nil {
		return fmt.Errorf("error writing metadata to disk for blueprint at path: %s. Details: %w", bpPath, err)
	}

	return nil
}

func CreateBlueprintMetadata(bpPath string, bpMetadataObj *BlueprintMetadata) (*BlueprintMetadata, error) {
	// verfiy that the blueprint path is valid & get repo details
	repoDetails, err := getRepoDetailsByPath(bpPath, bpMetadataObj.Spec.Source)
	if err != nil {
		return nil, err
	}

	// start creating blueprint metadata
	bpMetadataObj.ResourceMeta = yaml.ResourceMeta{
		TypeMeta: yaml.TypeMeta{
			APIVersion: metadataApiVersion,
			Kind:       metadataKind,
		},
		ObjectMeta: yaml.ObjectMeta{
			NameMeta: yaml.NameMeta{
				Name:      repoDetails.Name,
				Namespace: "",
			},
			Labels:      bpMetadataObj.ResourceMeta.ObjectMeta.Labels,
			Annotations: map[string]string{"config.kubernetes.io/local-config": "true"},
		},
	}

	// start creating the Spec node
	readmeContent, err := os.ReadFile(path.Join(bpPath, readmeFileName))
	if err != nil {
		return nil, fmt.Errorf("error reading blueprint readme markdown: %w", err)
	}

	// create blueprint info
	err = bpMetadataObj.Spec.BlueprintInfo.create(bpPath, readmeContent)
	if err != nil {
		return nil, fmt.Errorf("error creating blueprint info: %w", err)
	}

	// create blueprint interfaces i.e. variables & outputs
	err = bpMetadataObj.Spec.BlueprintInterface.create(bpPath)
	if err != nil {
		return nil, fmt.Errorf("error creating blueprint interfaces: %w", err)
	}

	// get blueprint requirements
	rolesCfgPath := path.Join(repoDetails.Source.RootPath, tfRolesFileName)
	svcsCfgPath := path.Join(repoDetails.Source.RootPath, tfServicesFileName)
	requirements, err := getBlueprintRequirements(rolesCfgPath, svcsCfgPath)
	if err != nil {
		return nil, fmt.Errorf("error creating blueprint requirements: %w", err)
	}

	bpMetadataObj.Spec.BlueprintRequirements = *requirements

	// create blueprint content i.e. documentation, icons, etc.
	bpMetadataObj.Spec.BlueprintContent.create(bpPath, repoDetails.Source.RootPath, readmeContent)

	return bpMetadataObj, nil
}

func (i *BlueprintInfo) create(bpPath string, readmeContent []byte) error {
	title, err := getMdContent(readmeContent, 1, 1, "", false)
	if err != nil {
		return err
	}

	i.Title = title.literal
	repoDetails, err := getRepoDetailsByPath(bpPath, i.Source)
	if err != nil {
		return err
	}

	i.Source = &BlueprintRepoDetail{
		Repo:       repoDetails.Source.Path,
		SourceType: "git",
	}

	versionInfo, err := getBlueprintVersion(path.Join(bpPath, tfVersionsFileName))
	if err == nil {
		i.Version = versionInfo.moduleVersion
		i.ActuationTool = BlueprintActuationTool{
			Version: versionInfo.requiredTfVersion,
			Flavor:  "Terraform",
		}
	}

	// create descriptions
	i.Description = &BlueprintDescription{}
	tagline, err := getMdContent(readmeContent, -1, -1, "Tagline", true)
	if err == nil {
		i.Description.Tagline = tagline.literal
	}

	detailed, err := getMdContent(readmeContent, -1, -1, "Detailed", true)
	if err == nil {
		i.Description.Detailed = detailed.literal
	}

	preDeploy, err := getMdContent(readmeContent, -1, -1, "PreDeploy", true)
	if err == nil {
		i.Description.PreDeploy = preDeploy.literal
	}

	var archListToSet []string
	architecture, err := getMdContent(readmeContent, -1, -1, "Architecture", true)
	if err == nil {
		for _, li := range architecture.listItems {
			archListToSet = append(archListToSet, li.text)
		}

		i.Description.Architecture = archListToSet
	}

	// create icon
	iPath := path.Join(repoDetails.Source.RootPath, iconFilePath)
	exists, _ := fileExists(iPath)
	if exists {
		i.Icon = iconFilePath
	}

	d, err := getDeploymentDuration(readmeContent, "Deployment Duration")
	if err == nil {
		i.DeploymentDuration = *d
	}

	c, err := getCostEstimate(readmeContent, "Cost")
	if err == nil {
		i.CostEstimate = *c
	}

	return nil
}

func (i *BlueprintInterface) create(bpPath string) error {
	interfaces, err := getBlueprintInterfaces(bpPath)
	if err != nil {
		return err
	}

	i.Variables = interfaces.Variables
	i.Outputs = interfaces.Outputs

	return nil
}

func (c *BlueprintContent) create(bpPath string, rootPath string, readmeContent []byte) {
	var docListToSet []BlueprintListContent
	documentation, err := getMdContent(readmeContent, -1, -1, "Documentation", true)
	if err == nil {
		for _, li := range documentation.listItems {
			doc := BlueprintListContent{
				Title: li.text,
				URL:   li.url,
			}

			docListToSet = append(docListToSet, doc)
		}

		c.Documentation = docListToSet
	}

	// create architecture
	a, err := getArchitctureInfo(readmeContent, "Architecture")
	if err == nil {
		c.Architecture = *a
	}

	// create sub-blueprints
	modPath := path.Join(bpPath, modulesPath)
	modContent, err := getModules(modPath)
	if err == nil {
		c.SubBlueprints = modContent
	}

	// create examples
	exPath := path.Join(rootPath, examplesPath)
	exContent, err := getExamples(exPath)
	if err == nil {
		c.Examples = exContent
	}
}

func WriteMetadata(obj *BlueprintMetadata, bpPath string) error {
	// marshal and write the file
	yFile, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(bpPath, metadataFileName), yFile, 0644)
}

func UnmarshalMetadata(bpPath string) (*BlueprintMetadata, error) {
	bpObj := BlueprintMetadata{}
	metaFilePath := path.Join(bpPath, metadataFileName)

	// return empty metadata if file does not exist or if the file is not read
	if _, err := os.Stat(metaFilePath); errors.Is(err, os.ErrNotExist) {
		return &bpObj, nil
	}

	f, err := os.ReadFile(metaFilePath)
	if err != nil {
		return &bpObj, fmt.Errorf("unable to read metadata from the existing file: %w", err)
	}

	err = yaml.Unmarshal(f, &bpObj)
	if err != nil {
		return &bpObj, err
	}

	currVersion := bpObj.ResourceMeta.TypeMeta.APIVersion
	currKind := bpObj.ResourceMeta.TypeMeta.Kind

	//validate GVK for current metadata
	if currVersion != metadataApiVersion {
		return &bpObj, fmt.Errorf("found incorrect version for the metadata: %s. Supported version is: %s", currVersion, metadataApiVersion)
	}

	if currKind != metadataKind {
		return &bpObj, fmt.Errorf("found incorrect kind for the metadata: %s. Supported kind is %s", currKind, metadataKind)
	}

	return &bpObj, nil
}
