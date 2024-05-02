package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)
//------------------------------------------------------------------------------
/*
worker() goroutine that takes in a worker ID, and  incomingJobs & results resultsChannel from main()
incomingJobs takes in string type input & resultsChannel channel takes in a map.
*/
func worker(workerID int, incomingJobs <-chan string, resultsChannel chan<- map[string]int) {
	pageLengths := make(map[string]int)

	for url := range incomingJobs {
		// download page content
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println("Error fetching URL:", url, err)
			continue
		}
		//close out the resp thread after its done. Good to avoid resource leaks.
		defer resp.Body.Close()

		// Read the page content
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			continue
		}
        // Calculate the length of the page content
		length := len(body)
		//Create a map of the url to length
		pageLengths[url] = length

		fmt.Printf("Worker %d calculated length of %s: %d\n", workerID, url, length)

	} //end of for loop

	// Send back the map of URL to page length
	resultsChannel <- pageLengths
} //end of worker function
//------------------------------------------------------------------------------
func main() {
	start := time.Now() // Record the start time

	//create a slice of urls to download for checking length
	urls := []string{
		"https://www.yahoo.com/",
		"http://www.cnn.com",
		"http://www.python.org",
		"http://www.jython.org",
		"http://www.pypy.org",
		"http://www.perl.org",
		"http://www.cisco.com",
		"http://www.facebook.com",
		"http://www.twitter.com",
		"http://www.macrumors.com/",
		"http://arstechnica.com/",
		"http://www.reuters.com/",
		"http://abcnews.go.com/",
		"http://www.cnbc.com/",
		"http://www.twilio.com/",
	}

	/*
	Create channels for communication between main() and worker() goroutines

	jobs_channel: This is a channel used for passing URLs from the main goroutine to worker goroutines. 
	      It's a channel of type string, meaning it will transmit URL strings. The numJobs parameter specifies the capacity of the channel,
		  allowing numJobs URL strings to be buffered in the channel before blocking. This helps in controlling the number of URLs that
		  can be concurrently processed by the workers.
    
	results_channel: This is a channel used for passing results (map of URL to page length) from worker goroutines back to the main goroutine. 
	         It's a channel of type map[string]int, meaning it will transmit maps where keys are URL strings and values are page lengths.
			The numWorkers parameter specifies the capacity of the channel, allowing numWorkers results to be buffered in the channel before blocking.
			This helps in controlling the number of results that can be concurrently processed by the main goroutine.

	Note the channels below are Buffered channels. Read more Buffered vs Unbuffered channels here:
	    https://stackoverflow.com/questions/23233381/whats-the-difference-between-c-makechan-int-and-c-makechan-int-1
	*/
	const numWorkers = 5
	jobs_channel := make(chan string, 5)
	results_channel := make(chan map[string]int, numWorkers)

	// Launch worker goroutines - think like spawning threads of worker()
	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs_channel, results_channel)
	}

	// Send jobs to workers
	for _, url := range urls {
		jobs_channel <- url
	}
	close(jobs_channel) // Close the jobs channel to signal that no more jobs will be sent

	// Collect results from workers
	totalPageLengths := make(map[string]int)
	for i := 0; i < numWorkers; i++ {
		pageLengths := <-results_channel
		for url, length := range pageLengths {
			totalPageLengths[url] = length
		}
	}

	// Print the map of URL to page length
	fmt.Println("URLs and their corresponding page lengths:")
	for url, length := range totalPageLengths {
		fmt.Printf("%s: %d\n", url, length)
	}

	elapsed := time.Since(start) // Calculate the elapsed time
	fmt.Println("Time taken:", elapsed)
}//end of main()
//------------------------------------------------------------------------------