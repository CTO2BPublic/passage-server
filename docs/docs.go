// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "url": "https://cto2b.eu",
            "email": "tomas@cto2b.eu"
        },
        "license": {
            "name": "Apache 2.0"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/access/requests": {
            "get": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "List all access requests",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Access requests"
                ],
                "summary": "List access requests",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.AccessRequest"
                            }
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Create new access request",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Access requests"
                ],
                "summary": "Create access request",
                "parameters": [
                    {
                        "description": "Access request definition",
                        "name": "role",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.AccessRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controllers.ResponseSuccess"
                        }
                    }
                }
            }
        },
        "/access/requests/{ID}": {
            "delete": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Delete access request by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Access requests"
                ],
                "summary": "Delete access request",
                "parameters": [
                    {
                        "type": "string",
                        "default": "xxxx-xxxx-xxxx",
                        "description": "AccessRequest id",
                        "name": "ID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.AccessRequest"
                            }
                        }
                    }
                }
            }
        },
        "/access/requests/{ID}/approve": {
            "post": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Approve access requests. All providers assigned to role will ensure user access",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Access requests"
                ],
                "summary": "Approve access request",
                "parameters": [
                    {
                        "type": "string",
                        "default": "xxxx-xxxx-xxxx",
                        "description": "AccessRequest id",
                        "name": "ID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controllers.ResponseSuccess"
                        }
                    }
                }
            }
        },
        "/access/requests/{ID}/expire": {
            "post": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Expire user access. All providers assigned to role will ensure access expiration",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Access requests"
                ],
                "summary": "Expire access request",
                "parameters": [
                    {
                        "type": "string",
                        "default": "xxxx-xxxx-xxxx",
                        "description": "AccessRequest id",
                        "name": "ID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controllers.ResponseSuccess"
                        }
                    }
                }
            }
        },
        "/access/roles": {
            "get": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Create a new which can be later used in access requests",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Access roles"
                ],
                "summary": "List roles",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.AccessRole"
                            }
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Create a new which can be later used in access requests",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Access roles"
                ],
                "summary": "Create role",
                "parameters": [
                    {
                        "description": "Role definition",
                        "name": "role",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.AccessRole"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controllers.ResponseSuccess"
                        }
                    }
                }
            }
        },
        "/healthz": {
            "get": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Healthy",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "API health"
                ],
                "summary": "Healthy",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Health"
                            }
                        }
                    }
                }
            }
        },
        "/livez": {
            "get": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Liveness",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "API health"
                ],
                "summary": "Liveness",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Health"
                            }
                        }
                    }
                }
            }
        },
        "/readyz": {
            "get": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Readyness",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "API health"
                ],
                "summary": "Readyness",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Health"
                            }
                        }
                    }
                }
            }
        },
        "/user/profile": {
            "get": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Returns curent user's profile",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "User profile",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.UserProfile"
                        }
                    }
                }
            }
        },
        "/user/profile/settings": {
            "put": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Updates current user's settings",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Update user settings",
                "parameters": [
                    {
                        "description": "User profiles settings",
                        "name": "role",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UserProfileSettings"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controllers.ResponseSuccessCreated"
                        }
                    }
                }
            }
        },
        "/userinfo": {
            "get": {
                "security": [
                    {
                        "JWT": []
                    }
                ],
                "description": "Returns information about authenticated user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "User info",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.ClaimsMap"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "controllers.ResponseSuccess": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "integer",
                    "example": 201
                },
                "title": {
                    "type": "string",
                    "example": "Record successfully created"
                },
                "type": {
                    "type": "string",
                    "example": "/status/success"
                }
            }
        },
        "controllers.ResponseSuccessCreated": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "integer",
                    "example": 201
                },
                "title": {
                    "type": "string",
                    "example": "Record successfully created"
                },
                "type": {
                    "type": "string",
                    "example": "/status/success"
                }
            }
        },
        "models.AccessRequest": {
            "type": "object",
            "properties": {
                "details": {
                    "$ref": "#/definitions/models.AccessRequestDetails"
                },
                "id": {
                    "type": "string"
                },
                "roleRef": {
                    "$ref": "#/definitions/models.AccessRoleRef"
                }
            }
        },
        "models.AccessRequestDetails": {
            "type": "object",
            "properties": {
                "attributes": {
                    "type": "object",
                    "additionalProperties": true
                },
                "justification": {
                    "type": "string",
                    "example": "Need to access k8s namespace"
                },
                "ttl": {
                    "type": "string",
                    "example": "72h"
                }
            }
        },
        "models.AccessRole": {
            "type": "object",
            "properties": {
                "approvalRuleRef": {
                    "$ref": "#/definitions/models.ApprovalRuleRef"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "string",
                    "example": "3b7af992-5a30-4ce1-821b-cac8194a230b"
                },
                "name": {
                    "type": "string"
                },
                "providers": {
                    "description": "Multiple access mappings for the role",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.ProviderConfig"
                    }
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "models.AccessRoleRef": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "example": "SRE-PU-ACCESS"
                }
            }
        },
        "models.ApprovalRuleRef": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                }
            }
        },
        "models.ClaimsMap": {
            "type": "object"
        },
        "models.CredentialRef": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                }
            }
        },
        "models.Health": {
            "type": "object",
            "properties": {
                "healthy": {
                    "type": "boolean"
                }
            }
        },
        "models.ProviderConfig": {
            "type": "object",
            "properties": {
                "credentialRef": {
                    "$ref": "#/definitions/models.CredentialRef"
                },
                "name": {
                    "type": "string"
                },
                "parameters": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "provider": {
                    "type": "string"
                },
                "runAsync": {
                    "type": "boolean"
                }
            }
        },
        "models.UserProfile": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "settings": {
                    "$ref": "#/definitions/models.UserProfileSettings"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "models.UserProfileSettings": {
            "type": "object",
            "properties": {
                "providerUsernames": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "JWT": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "passage-server",
	Description:      "powerful, open-source access control management solution built in Go",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
