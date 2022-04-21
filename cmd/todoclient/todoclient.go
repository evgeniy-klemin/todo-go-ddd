package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/evgeniy-klemin/todo/client"
)

const Count = 10000
const Workers = 6

func main() {
	start := time.Now()

	commands := make(chan string)

	wg := sync.WaitGroup{}
	for i := 0; i < Workers; i++ {
		wg.Add(1)
		go getItem(commands, &wg)
	}

	for i := 0; i < Count; i++ {
		commands <- "17349de3-5eb7-4899-a6c6-19dbc44257a7"
	}

	close(commands)

	wg.Wait()

	duration := time.Since(start)
	log.Println("duration: ", duration)
	log.Println("1req: ", duration.Microseconds()/Count, " mcs")
	log.Println("rate: ", Count*1000/duration.Milliseconds(), " cps")
}

func getItem(commands <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		uuid := <-commands
		if uuid == "" {
			return
		}
		ctx := context.Background()
		cwr, _ := client.NewClientWithResponses("http://localhost:8080")
		resp, err := cwr.GetItemsItemIdWithResponse(ctx, client.ItemId(uuid))
		if err != nil {
			log.Fatal(err)
			return
		}
		if resp.StatusCode() != http.StatusOK {
			log.Printf("error: %s", string(resp.Body))
		}
		itemBytes, _ := json.MarshalIndent(resp.JSON200, "", "\t")
		log.Print("Item:\n", string(itemBytes))
		//log.Printf("Item: %+v", resp.JSON200)
	}
}
