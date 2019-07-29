# Versioner Kustomize Plugin

This is a plugin for kustomize to manage containers' versions declaratively by an external file.

## Run Example

```
docker run -it --rm \
  -v $PWD:/app -w /app \
  yanana/kustomize-plugin-versioner \
  buld --enable_alpha_plugins examples
```
