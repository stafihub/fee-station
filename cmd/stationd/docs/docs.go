// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "tk",
            "email": "tpkeeper@qq.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1/station/poolInfo": {
            "get": {
                "description": "get pool info",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v1"
                ],
                "summary": "get pool info",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/utils.Rsp"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/station_handlers.RspPoolInfo"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/v1/station/swapInfo": {
            "get": {
                "description": "get swap info",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "v1"
                ],
                "summary": "get swap info",
                "parameters": [
                    {
                        "type": "string",
                        "description": "token symbol",
                        "name": "symbol",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "tx hash hex string",
                        "name": "txHash",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/utils.Rsp"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/station_handlers.RspSwapInfo"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "station_handlers.PoolInfo": {
            "type": "object",
            "properties": {
                "poolAddress": {
                    "description": "base58,bech32 or hex",
                    "type": "string"
                },
                "swapRate": {
                    "description": "decimals 6",
                    "type": "string"
                },
                "symbol": {
                    "type": "string"
                }
            }
        },
        "station_handlers.RspPoolInfo": {
            "type": "object",
            "properties": {
                "poolInfoList": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/station_handlers.PoolInfo"
                    }
                },
                "swapMaxLimit": {
                    "description": "decimals 12",
                    "type": "string"
                },
                "swapMinLimit": {
                    "description": "decimals 12",
                    "type": "string"
                }
            }
        },
        "station_handlers.RspSwapInfo": {
            "type": "object",
            "properties": {
                "swapStatus": {
                    "type": "integer"
                }
            }
        },
        "utils.Rsp": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "object"
                },
                "message": {
                    "type": "string"
                },
                "status": {
                    "type": "integer"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "localhost:8083",
	BasePath:    "/feeStation/api",
	Schemes:     []string{},
	Title:       "drop API",
	Description: "drop api document.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
