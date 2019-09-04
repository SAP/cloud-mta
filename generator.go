package main

//go:generate go run ./tools/embed.go -source=./validations/schema/schema_v3.2.yaml -target=./validations/mta_schema.go -name=schemaDef -package=validate
//go:generate go run ./tools/embed.go -source=./validations/schema/mtaext-schema_v3.2.yaml -target=./validations/mtaext_schema.go -name=extSchemaDef -package=validate
//go:generate go run ./tools/embed.go -source=./configs/version.yaml -target=./internal/version/version_cfg.go -name=VersionConfig -package=version
