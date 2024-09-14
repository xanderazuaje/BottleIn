// backend/docs/docs.go
package docs

import "github.com/swaggo/swag"

var docTemplate = `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample Go backend server.",
        "title": "Go Backend API",
        "version": "1.0"
    },
    "host": "localhost:8080",
    "basePath": "/api",
    "paths": {
        "/hello": {
            "get": {
                "description": "Returns a hello message",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "hello"
                ],
                "summary": "Say Hello",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Response"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.Response": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string",
                    "example": "Hello from Go Backend!"
                }
            }
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:         "1.0",
	Host:            "localhost:8080",
	BasePath:        "/api",
	SwaggerTemplate: docTemplate,
	LeftDelim:       "{{",
	RightDelim:      "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
