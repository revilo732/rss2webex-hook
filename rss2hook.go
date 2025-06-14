// rss2hook is a simple utility which will make HTTP POST
// requests to remote web-hooks when new items appear in an RSS feed.
//
// Steve
//

package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron"
)

// RSSEntry describes a single RSS feed and the corresponding hook
// to POST to.
type RSSEntry struct {
	// The URL of the RSS/Atom feed
	feed string

	// The end-point to make the webhook request to.
	hook string
}

// Loaded contains the loaded feeds + hooks, as read from the specified
// configuration file
var Loaded []RSSEntry

// Timeout is the (global) timeout we use when loading remote RSS
// feeds.
var Timeout time.Duration

// loadConfig loads the named configuration file and populates our
// `Loaded` list of RSS-feeds & Webhook addresses

//ADDON2 
func loadConfig(filename string) {
    file, err := os.Open(filename)
    if err != nil {
        fmt.Printf("Error opening %s - %s\n", filename, err.Error())
        return
    }
    defer file.Close()

    //
    // Process it line by line.
    //
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        // Get the next line, and strip leading/trailing space
        tmp := scanner.Text()
        tmp = strings.TrimSpace(tmp)

        // Skip lines that begin with a comment.
        if (tmp != "") && (!strings.HasPrefix(tmp, "#")) {
            // Change the regex pattern to split on the last '=' character
            parser := regexp.MustCompile("^(.*)=(.*)$")
            match := parser.FindStringSubmatch(tmp)

            // OK we found a suitable entry.
            if len(match) == 3 {
                feed := strings.TrimSpace(match[1])
                hook := strings.TrimSpace(match[2])

                // Append the new entry to our list
                entry := RSSEntry{feed: feed, hook: hook}
                Loaded = append(Loaded, entry)
            }
        }
    }
}
//ADDON2vege


// fetchFeed fetches the contents of the specified URL.
func fetchFeed(url string) (string, error) {

	//HTTP2 kikapcs
	// Ensure we setup a timeout for our fetch
	client := &http.Client{
    Transport: &http.Transport{
        ForceAttemptHTTP2: false, // Disable HTTP/2
    },
    Timeout: Timeout,
}

	// We'll only make a GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// We ensure we identify ourself.
	req.Header.Set("User-Agent", "BasicAgent")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the body returned
	output, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// isNew returns TRUE if this feed-item hasn't been notified about
// previously.
func isNew(parent string, item *gofeed.Item) bool {

	hasher := sha1.New()
	hasher.Write([]byte(parent))
	hasher.Write([]byte(item.GUID))
	hashBytes := hasher.Sum(nil)

	// Hexadecimal conversion
	hexSha1 := hex.EncodeToString(hashBytes)

	if _, err := os.Stat(os.Getenv("HOME") + "/.rss2hook/seen/" + hexSha1); os.IsNotExist(err) {
		return true
	}
	return false
}

// recordSeen ensures that we won't re-announce a given feed-item.
func recordSeen(parent string, item *gofeed.Item) {

	hasher := sha1.New()
	hasher.Write([]byte(parent))
	hasher.Write([]byte(item.GUID))
	hashBytes := hasher.Sum(nil)

	// Hexadecimal conversion
	hexSha1 := hex.EncodeToString(hashBytes)

	dir := os.Getenv("HOME") + "/.rss2hook/seen"
	os.MkdirAll(dir, os.ModePerm)

	_ = ioutil.WriteFile(dir+"/"+hexSha1, []byte(item.Link), 0644)

}

// checkFeeds is our work-horse.
//
// For each available feed it looks for new entries, and when founds
// triggers `notify` upon the resulting entry
func checkFeeds() {

	//
	// For each thing we're monitoring
	//
	for _, monitor := range Loaded {

		// Fetch the feed-contents
		content, err := fetchFeed(monitor.feed)

		if err != nil {
			fmt.Printf("Error fetching %s - %s\n",
				monitor.feed, err.Error())
			continue
		}

		// Now parse the feed contents into a set of items
		fp := gofeed.NewParser()
		feed, err := fp.ParseString(content)
		if err != nil {
			fmt.Printf("Error parsing %s contents: %s\n", monitor.feed, err.Error())
			continue
		}

		// For each entry in the feed
		for _, i := range feed.Items {

			// If we've not already notified about this one.
			if isNew(monitor.feed, i) {

				// Trigger the notification
				err := notify(monitor.hook, i)

				// and if that notification succeeded
				// then record this item as having been
				// processed successfully.
				if err == nil {
					recordSeen(monitor.feed, i)
				}
			}
		}
	}
}

// notify actually submits the specified item to the remote webhook.
//
// The RSS-item is submitted as a JSON-object.
    //ADDON START
	func notify(hook string, item *gofeed.Item) error {
		// Convert the item to a string format, including title and description for the markdown content.
		data := fmt.Sprintf("[**%s**](%s)\n%s", item.Title, item.Link, item.Description)
	
		// Create the payload as a map and marshal it into JSON
		payload := map[string]string{"markdown": data}
		jsonValue, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("notify: Failed to encode JSON: %s\n", err.Error())
			return err
		}
	
		// Create a new HTTP request
		req, err := http.NewRequest("POST", hook, bytes.NewBuffer(jsonValue))
		if err != nil {
			fmt.Printf("notify: Failed to create request: %s\n", err.Error())
			return err
		}
	
		// Set the Content-Type header
		req.Header.Set("Content-Type", "application/json")
	
		// Send the request using the HTTP client
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("notify: Failed to POST to %s - %s\n", hook, err.Error())
			return err
		}
		defer res.Body.Close()
	
		// Check the response status
		status := res.StatusCode
		if status != 200 {
			// Read the response body to get more error information
			body, _ := ioutil.ReadAll(res.Body)
			fmt.Printf("notify: Warning - Status code was not 200: %d, Response body: %s\n", status, string(body))
		}
		return nil
	}
	//ADDON END

// main is our entry-point
func main() {

	// Parse the command-line flags
	config := flag.String("config", "", "The path to the configuration-file to read")
	timeout := flag.Duration("timeout", 5*time.Second, "The timeout used for fetching the remote feeds")
	flag.Parse()

	// Setup the default timeout.
	Timeout = *timeout

	if *config == "" {
		fmt.Printf("Please specify a configuration-file to read\n")
		return
	}

	//
	// Load the configuration file
	//
	loadConfig(*config)

	//
	// Show the things we're monitoring
	//
	for _, ent := range Loaded {
		fmt.Printf("Monitoring feed %s\nPosting to %s\n\n",
			ent.feed, ent.hook)
	}

	//
	// Make the initial scan of feeds immediately to avoid waiting too
	// long for the first time.
	//
	checkFeeds()

	//
	// Now repeat that every five minutes.
	//
	c := cron.New()
	c.AddFunc("@every 5m", func() { checkFeeds() })
	c.Start()

	//
	// Now we can loop waiting to be terminated via ctrl-c, etc.
	//
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()
	<-done
}
