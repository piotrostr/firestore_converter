package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	opt := option.WithCredentialsFile("/Users/piotrostr/Downloads/production-schedule-3033a-fa40b1f07c55.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	cols := client.Collections(ctx)
	allCols, err := cols.GetAll()
	if err != nil {
		log.Fatalln(err)
	}

	if len(allCols) == 0 {
		log.Println("No collections found")
		return
	}

	res := make(map[string]any)

	for len(allCols) > 0 {
		col := allCols[len(allCols)-1]
		allCols = allCols[:len(allCols)-1]

		colRes := make(map[string]any)
		docsIterator := col.Documents(ctx)
		for {
			doc, err := docsIterator.Next()
			if err == iterator.Done {
				break
			}
			colRes[doc.Ref.ID] = doc.Data()
			nestedCollectionsIterator := doc.Ref.Collections(ctx)
			allNestedCols, err := nestedCollectionsIterator.GetAll()
			if err != nil {
				log.Fatalln(err)
			}
			allCols = append(allCols, allNestedCols...)
		}
		res[col.ID] = colRes
	}

	defer client.Close()

	dumpJSON(res)
}

func dumpJSON(data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	f, err := os.Create("dump.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	_, err = f.Write(jsonData)
	if err != nil {
		log.Fatalln(err)
	}
}
