package main

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/gocolly/colly"
)

type Results struct {
	Name         string
	Availability string
}

func getAvailabilityFromSmxSite() []Results {
	c := colly.NewCollector()
	c.SetRequestTimeout(120 * time.Second)
	results := make([]Results, 0)

	// Callbacks

	c.OnHTML("a.product-card", func(e *colly.HTMLElement) {
		e.ForEach("div.product-card__info", func(i int, h *colly.HTMLElement) {

			item := Results{}
			item.Name = e.ChildText("div.product-card__name")
			item.Availability = e.ChildText("div.product-card__availability")
			results = append(results, item)

		})

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Got this error:", e)
	})

	// c.OnScraped(func(r *colly.Response) {
	// 	fmt.Println("Finished", r.Request.URL)
	// 	js, err := json.MarshalIndent(results, "", "    ")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println("Writing data to file")
	// 	if err := os.WriteFile("results.json", js, 0664); err == nil {
	// 		fmt.Println("Data written to file successfully")
	// 	}

	// })

	c.Visit("https://shop.steprevolution.com")
	return results
}

func sendEmail(someMessage string) {
	// Sender data.
	from := "GabrielWohlford@gmail.com"
	password := "password"

	// Receiver email address.
	to := []string{
		"GabrielWohlford@gmail.com",
	}

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message.
	message := []byte(someMessage)

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}

func (r Results) toString() string {
	return "Product: " + r.Name + "Status:" + r.Availability
}

func resultsToString(r []Results) string {
	stringResults := ""
	for _, element := range r {
		stringResults += element.toString() + "\n"
	}
	return stringResults
}

func anythingAvailable(r *[]Results) bool {
	results := *r
	available := false
	for i := 0; i < len(results); i++ {
		if results[i].Availability != "Sold Out" {
			available = true
		}
	}
	return available
}

func crawl() {
	ticker := time.NewTicker(30 * time.Second)
	done := make(chan bool)

	//get initial results
	baseResults := getAvailabilityFromSmxSite()
	//check results if anything is available
	thingsAvailable := anythingAvailable(&baseResults)
	resultsMessage := resultsToString(baseResults)
	if thingsAvailable {
		sendEmail(resultsMessage)
	}

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				//run ticker till results are differnt from baseResults
				//different set new base results and repeat the process
				fmt.Println("Ticker running")
				newResults := getAvailabilityFromSmxSite()
				newResultsMessage := resultsToString(newResults)
				if newResultsMessage != resultsMessage {
					sendEmail("Updated: " + newResultsMessage)
					resultsMessage = newResultsMessage
				}
			}
		}
	}()

	for {
		var first string

		// Taking input from user
		println("type: <stop> to end the crawl")
		fmt.Scanln(&first)
		if first == "stop" {
			ticker.Stop()
			done <- true
			println("ALL DONE")
		}

	}
}

func main() {

	crawl()
}
