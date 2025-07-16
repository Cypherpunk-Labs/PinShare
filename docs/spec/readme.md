REF: https://medium.com/@MikeMwita/generating-go-code-from-openapi-specification-document-ae225e49e970
REF: https://github.com/oapi-codegen/oapi-codegen

go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

oapi-codegen --package=main  --generate types,server ./openapi.spec.yaml > generated_oapi/server.go

Then we need to add impl.go and main.go to finish off as per opai=codegen docs.
ERROR: is relative, but relative import paths are not supported in module mode

seems like the generate flag was wrong and server is too bare, could find more options in https://github.com/oapi-codegen/oapi-codegen/blob/main/configuration-schema.json

oapi-codegen --package=main  --generate types,std-http-server,embedded-spec,models ../openapi.spec.yaml > server.go

oapi-codegen --package=api  --generate types,std-http-server,embedded-spec,models ./docs/spec/basemetadata.openapi.spec.yaml > internal/api/basemetadata.go