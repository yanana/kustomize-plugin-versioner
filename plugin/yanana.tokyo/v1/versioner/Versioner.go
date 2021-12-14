package main

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ resmap.TransformerPlugin = (*plugin)(nil)

type Image struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Tag    string `json:"tag,omitempty" yaml:"tag,omitempty"`
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

// A map of Images keyed by image name.
type versions map[string]Image

type Filter struct {
	Name  string `json:"name" yaml:"name"`
	Image Image  `json:"image,omitempty" yaml:"image,omitempty"`
}

var _ kio.Filter = Filter{}

// targetFss is a types.FsSlice whose elements are FieldSpecs
// that should be manipulated.
var targetFss = types.FsSlice{
	{
		Gvk: resid.Gvk{
			Group: "apps",
			Kind:  "Deployment",
		},
		Path: "spec/template/spec/containers",
	},
	{
		Gvk: resid.Gvk{
			Group: "",
			Kind:  "Pod",
		},
		Path: "spec/containers",
	},
	{
		Gvk: resid.Gvk{
			Group: "",
			Kind:  "PodTemplate",
		},
		Path: "template/spec/containers",
	},
	{
		Gvk: resid.Gvk{
			Group: "",
			Kind:  "ReplicationController",
		},
		Path: "spec/template/spec/containers",
	},
	{
		Gvk: resid.Gvk{
			Group: "apps",
			Kind:  "ReplicaSet",
		},
		Path: "spec/template/spec/containers",
	},
	{
		Gvk: resid.Gvk{
			Group: "apps",
			Kind:  "StatefulSet",
		},
		Path: "spec/template/spec/containers",
	},
	{
		Gvk: resid.Gvk{
			Group: "apps",
			Kind:  "DaemonSet",
		},
		Path: "spec/template/spec/containers",
	},
	{
		Gvk: resid.Gvk{
			Group: "batch",
			Kind:  "Job",
		},
		Path: "spec/template/spec/containers",
	},
	{
		Gvk: resid.Gvk{
			Group: "batch",
			Kind:  "CronJob",
		},
		Path: "spec/jobTemplate/spec/template/spec/containers",
	},
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(f.filter)).Filter(nodes)
	return nodes, err
}

func (f Filter) filter(node *yaml.RNode) (*yaml.RNode, error) {
	meta, err := node.GetMeta()
	if err != nil {
		return nil, err
	}

	if meta.Kind == `CustomResourceDefinition` {
		return node, nil
	}

	if err := node.PipeE(fsslice.Filter{
		FsSlice:  targetFss,
		SetValue: updateFn(f.Name, f.Image),
	}); err != nil {
		return nil, err
	}

	return node, nil
}

func updateFn(name string, image Image) filtersutil.SetFn {
	return func(node *yaml.RNode) error {
		return node.PipeE(imageUpdater{
			Name:  name,
			Image: image,
		})
	}
}

type imageUpdater struct {
	Name  string
	Image Image
}

func (u imageUpdater) Filter(node *yaml.RNode) (*yaml.RNode, error) {
	switch node.YNode().Kind {
	case yaml.SequenceNode:
		if err := node.VisitElements(func(node *yaml.RNode) error {
			if node.YNode().Kind == yaml.MappingNode {
				nameF := node.Field("name")
				if nameF == nil {
					return nil
				}
				name, err := nameF.Value.String()
				if err != nil {
					return err
				}
				name = strings.TrimSpace(name)
				if name != u.Name {
					return nil
				}
				imageF := node.Field("image")
				if imageF == nil {
					return nil
				}
				image, err := imageF.Value.String()
				if err != nil {
					return err
				}
				image = strings.TrimSpace(image)
				modImage := newImage(image, u.Image)
				setter := yaml.FieldSetter{
					Name:        "image",
					StringValue: modImage,
				}
				return node.PipeE(setter)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return node, nil
}

type plugin struct {
	Environment      string `json:"environment" yaml:"environment"`
	VersionsFilePath string `json:"versionsFilePath" yaml:"versionsFilePath"`
	Versions         versions
	loaderRoot       string
}

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) error {
	p.loaderRoot = h.Loader().Root()

	return p.unmarshal(h.Loader(), config)
}

// noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

// readVersionsFile reads the versions file and parse it into plugin.Version.
func (p *plugin) readVersionsFile(ldr ifc.Loader) error {
	path := filepath.Join(ldr.Root(), filepath.Clean(p.VersionsFilePath))
	data, err := ldr.Load(path)
	if err != nil {
		return err
	}
	e := &struct {
		Environments map[string]versions `json:"environments,omitempty" yaml:"environments,omitempty"`
	}{}
	if err := yaml.Unmarshal(data, e); err != nil {
		return err
	}
	if _, found := e.Environments[p.Environment]; !found {
		return errors.Errorf("versions for the environment %s was not found in %s", p.Environment, path)
	}
	versions := versions{}
	for name, image := range e.Environments[p.Environment] {
		versions[name] = image //.toTypesImage()
	}
	p.Versions = versions

	return nil
}

func (p *plugin) unmarshal(ldr ifc.Loader, data []byte) error {
	if err := yaml.Unmarshal(data, p); err != nil {
		return err
	}
	if err := p.readVersionsFile(ldr); err != nil {
		return err
	}

	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	for name, image := range p.Versions {
		if err := m.ApplyFilter(Filter{
			Name:  name,
			Image: image,
		}); err != nil {
			return err
		}
	}
	return nil
}

func newImage(original string, image Image) string {
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
	return name + tag
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
