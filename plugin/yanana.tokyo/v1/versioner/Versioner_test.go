package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

func TestVersionerNoTransformWhenIrrelevant(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()
	tc.BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")
	th.WriteF("/app/versions.yaml", `
environments:
  staging:
    foo:
      name: foo/bar
      tag: baz
    the-container:
      tag: new-v1
`)
	rm := th.LoadAndRunTransformer(`
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`, `
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
	th.AssertActualEqualsExpected(rm, `
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
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()
	tc.BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")
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
      digest: 6a92cd1fcdc8d8cdec60f33dda4db2cb1fcdcacf3410a8e05b3741f44a9b5998
`)
	rm := th.LoadAndRunTransformer(`
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`, `
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
      - image: foo:bar
        name: the-container
      - image: baz
        name: xyz
`)
	th.AssertActualEqualsExpected(rm, `
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
      - image: foo@6a92cd1fcdc8d8cdec60f33dda4db2cb1fcdcacf3410a8e05b3741f44a9b5998
        name: the-container
      - image: baz
        name: xyz
`)
}

