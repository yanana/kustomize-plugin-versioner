package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestVersionerNoTransformWhenIrrelevant(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()

	th.WriteF("/app/versions.yaml", `
environments:
  staging:
    foo:
      name: foo/bar
      tag: baz
    the-container:
      tag: new-v1
`)

	th.WriteF("/app/versioner.yaml", `
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`)

	th.WriteF("/app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - image: elasticsearch
        name: elasticsearch
`)

	th.WriteK("/app", `
resources:
  - deployment.yaml
transformers:
  - versioner.yaml
`)

	m := th.Run("/app", th.MakeOptionsPluginsEnabled())

	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - image: elasticsearch
        name: elasticsearch
`)
}

func TestVersionerTransformAsVersionFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/app/versions.yaml", `
environments:
  production:
    magna-carta:
      name: magna/carta
      tag: 1
    the-container:
      tag: old-v1
  staging:
    magna-carta:
      name: magna/carta
      tag: 2
    the-container:
      name: oh/cool
      digest: 6a92cd1fcdc8d8cdec60f33dda4db2cb1fcdcacf3410a8e05b3741f44a9b5998
`)

	th.WriteF("/app/versioner.yaml", `
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`)

	th.WriteF("/app/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - image: elasticsearch
        name: magna-carta
      - image: gcr.io/foo/bar:baz
        name: the-container
      - image: baz
        name: xyz
`)

	th.WriteK("/app", `
resources:
  - deployment.yaml
transformers:
  - versioner.yaml
`)

	m := th.Run("/app", th.MakeOptionsPluginsEnabled())

	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - image: magna/carta:2
        name: magna-carta
      - image: oh/cool@6a92cd1fcdc8d8cdec60f33dda4db2cb1fcdcacf3410a8e05b3741f44a9b5998
        name: the-container
      - image: baz
        name: xyz
`)
}
