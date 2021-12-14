package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestVersionerNoTransformWhenIrrelevant(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/versions.yaml", `
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

func TestVersionerTransformsDeploymentAsVersionFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/versions.yaml", `
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
      - image: gcr.io/foo/bar:baz
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
      - image: oh/cool@6a92cd1fcdc8d8cdec60f33dda4db2cb1fcdcacf3410a8e05b3741f44a9b5998
        name: the-container
      - image: baz
        name: xyz
`)
}

func TestVersionerTransformsPodAsVersionFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/versions.yaml", `
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
	rm := th.LoadAndRunTransformer(`
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`, `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: elasticsearch
    name: magna-carta
  - image: gcr.io/foo/bar:baz
    name: the-container
  - image: baz
    name: xyz
`)
	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
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

func TestVersionerTransformsPodTemplateAsVersionFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/versions.yaml", `
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
	rm := th.LoadAndRunTransformer(`
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`, `
apiVersion: v1
kind: PodTemplate
metadata:
  name: nginx
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
	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: PodTemplate
metadata:
  name: nginx
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

func TestVersionerTransformsReplicationControllerAsVersionFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/versions.yaml", `
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
	rm := th.LoadAndRunTransformer(`
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`, `
apiVersion: v1
kind: ReplicationController
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
	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: ReplicationController
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

func TestVersionerTransformsReplicaSetAsVersionFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/versions.yaml", `
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
	rm := th.LoadAndRunTransformer(`
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`, `
apiVersion: apps/v1
kind: ReplicaSet
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
	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: ReplicaSet
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

func TestVersionerTransformsStatefulSetAsVersionFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/versions.yaml", `
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
	rm := th.LoadAndRunTransformer(`
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`, `
apiVersion: apps/v1
kind: StatefulSet
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
	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: StatefulSet
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

func TestVersionerTransformsCronJobAsVersionFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).BuildGoPlugin("yanana.tokyo", "v1", "Versioner")
	defer th.Reset()
	th.WriteF("/versions.yaml", `
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
	rm := th.LoadAndRunTransformer(`
apiVersion: yanana.tokyo/v1
kind: Versioner
metadata:
  name: notImportantHere
versionsFilePath: versions.yaml
environment: staging
`, `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: nginx
spec:
  jobTemplate:
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
	th.AssertActualEqualsExpected(rm, `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: nginx
spec:
  jobTemplate:
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
