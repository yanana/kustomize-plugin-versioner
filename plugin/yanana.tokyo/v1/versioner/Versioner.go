package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/yaml"
)

type Image struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Tag    string `json:"tag,omitempty" yaml:"tag,omitempty"`
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

type Versions map[string]Image

type plugin struct {
	Environment      string `json:"environment" yaml:"environment"`
	VersionsFilePath string `json:"versionsFilePath" yaml:"versionsFilePath"`
	Versions         Versions
	loaderRoot       string
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func readVersionsFile(ldr ifc.Loader, versionsFilePath string, env string, versions *Versions) error {
	p := filepath.Join(ldr.Root(), filepath.Clean(versionsFilePath))
	data, err := ldr.Load(p)
	if err != nil {
		return err
	}
	e := &struct {
		Environments map[string]Versions `json:"environments,omitempty" yaml:"environments,omitempty"`
	}{}
	if err := yaml.Unmarshal(data, e); err != nil {
		return err
	}
	if _, found := e.Environments[env]; !found {
		return errors.Errorf("versions for the environment %s was not found in %s", env, p)
	}
	*versions = e.Environments[env]

	return nil
}

func (p *plugin) unmarshal(ldr ifc.Loader, data []byte) error {
	if err := yaml.Unmarshal(data, p); err != nil {
		return err
	}
	if err := readVersionsFile(ldr, p.VersionsFilePath, p.Environment, &p.Versions); err != nil {
		return err
	}
	return nil
}

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {
	p.loaderRoot = ldr.Root()

	return p.unmarshal(ldr, c)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		// Replace images in "containers" and "initContainers".
		if err := p.findAndReplaceImage(r.Map()); err != nil && r.OrgId().Kind != `CustomResourceDefinition` {
			return err
		}
	}
	return nil
}

func (p *plugin) mutateImage(original string, image Image) (string, error) {
	name, tag := split(original)
	if image.Name != "" {
		name = image.Name
	}
	if image.Tag != "" {
		tag = ":" + image.Tag
	}
	if image.Digest != "" {
		tag = "@" + image.Digest
	}
	return name + tag, nil
}

// findAndReplaceImage replaces the image name and
// tags inside one object.
// It searches the object for container session
// then loops though all images inside containers
// session, finds matched ones and update the
// image name and tag name
func (p *plugin) findAndReplaceImage(obj map[string]interface{}) error {
	paths := []string{"containers", "initContainers"}
	updated := false
	for _, path := range paths {
		containers, found := obj[path]
		if !found {
			continue
		}
		if _, err := p.updateContainers(containers); err != nil {
			return err
		}
		updated = true
	}
	if !updated {
		return p.findContainers(obj)
	}
	return nil
}

func (p *plugin) updateContainers(in interface{}) (interface{}, error) {
	containers, ok := in.([]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"containers path is not of type []interface{} but %T", in)
	}
	for c := range containers {
		container := containers[c].(map[string]interface{})
		containerImage, found := container["image"]
		if !found {
			continue
		}
		imageName := containerImage.(string)
		containerName, found := container["name"]
		if !found {
			continue
		}
		name := containerName.(string)
		image, found := p.Versions[name]
		if !found {
			continue
		}
		newImage, err := p.mutateImage(imageName, image)
		if err != nil {
			return nil, err
		}
		container["image"] = newImage
	}
	return containers, nil
}

func (p *plugin) findContainers(obj map[string]interface{}) error {
	for key := range obj {
		switch typedV := obj[key].(type) {
		case map[string]interface{}:
			err := p.findAndReplaceImage(typedV)
			if err != nil {
				return err
			}
		case []interface{}:
			for i := range typedV {
				item := typedV[i]
				typedItem, ok := item.(map[string]interface{})
				if ok {
					err := p.findAndReplaceImage(typedItem)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// split separates and returns the name and tag parts
// from the image string using either colon `:` or at `@` separators.
// Note that the returned tag keeps its separator.
func split(imageName string) (name string, tag string) {
	// check if image name contains a domain
	// if domain is present, ignore domain and check for `:`
	ic := -1
	if slashIndex := strings.Index(imageName, "/"); slashIndex < 0 {
		ic = strings.LastIndex(imageName, ":")
	} else {
		lastIc := strings.LastIndex(imageName[slashIndex:], ":")
		// set ic only if `:` is present
		if lastIc > 0 {
			ic = slashIndex + lastIc
		}
	}
	ia := strings.LastIndex(imageName, "@")
	if ic < 0 && ia < 0 {
		return imageName, ""
	}

	i := ic
	if ic < 0 {
		i = ia
	}

	name = imageName[:i]
	tag = imageName[i:]
	return
}
