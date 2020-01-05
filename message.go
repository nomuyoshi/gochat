package main

import (
	"time"
)

type message struct {
	Name    string
	Message string
	SentAt  time.Time
}
