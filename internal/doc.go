package internal

// This directory contains the OpenAPI 3.0 specification which defines our
// server. The file petstore.gen.go is automatically generated from the schema

// Run oapi-codegen to regenerate the petstore boilerplate
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=client --generate types,client -o ../client/todoclient.gen.go ../docs/todo.yaml
