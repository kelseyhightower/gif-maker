// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"fmt"
	"image/gif"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
)

var html = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Kubernetes Pod</title>
</head>
<body>
  <h3>Pod Info</h3>
  <ul>
    <li>Hostname: %s</li>
  </ul>
  <h3>Certificate Details</h3>
  <ul>
    <li>Issuer: %s</li>
    <li>Serial: %s</li>
    <li>NotBefore: %s</li>
    <li>NotAfter: %s</li>
  </ul>
</body>
</html>
`

type Event struct {
	ID        string
	Message   string
	Timestamp time.Time
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	span := traceClient.SpanFromRequest(r)
	defer span.Finish()

	gifs := []gif.GIF{}

	animatedGIF := &gif.GIF{}
	for _, g := range gifs {
		animatedGIF.Image = append(animatedGIF.Image, g.Image[0])
		animatedGIF.Delay = append(animatedGIF.Delay, 0)
	}

	var b bytes.Buffer
	err := gif.EncodeAll(&b, animatedGIF)
	if err != nil {
		log.Println(err)
	}

	// Log an event.
	event := Event{
		ID:        uuid.New().String(),
		Message:   "Animated GIF created.",
		Timestamp: time.Now(),
	}

	m, err := spanner.InsertStruct("Event", event)
	if err != nil {
		log.Println(err)
	}

	ctx := context.Background()
	_, err = spannerClient.Apply(ctx, []*spanner.Mutation{m})
	if err != nil {
		log.Println(err)
	}

	fmt.Fprintf(w, html, hostname)
}
