// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package gensql

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Beredskapsvakt struct {
	// Created by Vaktor Plan
	ID    uuid.UUID
	Ident string
	Plan  json.RawMessage
}
