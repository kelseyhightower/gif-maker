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
	"context"
	"fmt"
	"image"
	"image/gif"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/spanner"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type Event struct {
	ID        string
	Message   string
	Timestamp time.Time
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	span := traceClient.SpanFromRequest(r)
	defer span.Finish()
	err := r.ParseMultipartForm(100000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gifs := make([]string, 0)

	form := r.MultipartForm
	images := form.File["images"]

	for i, _ := range images {
		f, err := images[i].Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			log.Println("error decoding input image:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpfile, err := ioutil.TempFile("", "gif")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tmpfile.Close()

		err = gif.Encode(tmpfile, img, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gifs = append(gifs, tmpfile.Name())
	}

	animatedGIF := &gif.GIF{}
	for _, g := range gifs {
		f, err := os.Open(g)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		img, err := gif.Decode(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		animatedGIF.Image = append(animatedGIF.Image, img.(*image.Paletted))
		animatedGIF.Delay = append(animatedGIF.Delay, 0)
	}

	storageSpan := span.NewChild("upload-to-cloud-storage")
	storageCtx := context.Background()

	result := storageClient.Bucket(bucket).Object("animated.gif").NewWriter(storageCtx)
	result.ContentType = "image/gif"
	result.ACL = []storage.ACLRule{{storage.AllUsers, storage.RoleReader}}

	err = gif.EncodeAll(result, animatedGIF)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Close()
	storageSpan.Finish()

	// Log an event.
	event := Event{
		ID:        uuid.New().String(),
		Message:   fmt.Sprintf("Animated GIF created.", result.Attrs().MediaLink),
		Timestamp: time.Now(),
	}

	m, err := spanner.InsertStruct("Events", event)
	if err != nil {
		log.Println(err)
	}

	databaseSpan := span.NewChild("log-event-to-spanner")
	ctx := context.Background()
	_, err = spannerClient.Apply(ctx, []*spanner.Mutation{m})
	if err != nil {
		log.Println(err)
	}
	databaseSpan.Finish()

	fmt.Fprintf(w, result.Attrs().MediaLink)
}
