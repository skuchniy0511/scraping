package main

import (
	"context"
	"encoding/base64"
	"strconv"

	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type UserData struct {
	Name     string `json:"name"`
	Company  string `json:"company"`
	Title    string `json:"title"`
	Location string `json:"location"`
	Bio      string `json:"bio"`
	Links    string `json:"links"`
	Tags     string `json:"tags"`
}

func main() {
	// Define the base URL for the API
	baseURL := "https://api.example.com/data"
	ctx := context.Background()

	// Define the range of user numbers to query
	startUserNum := 1
	endUserNum := 40000

	// Authenticate with the Google Sheets API
	sh, err := GetClient()
	if err != nil {
		return
	}

	jsonData := UserData{}
	// Loop over the range of user numbers
	for userNum := startUserNum; userNum <= endUserNum; userNum++ {
		// Construct the URL for this user number
		url := fmt.Sprintf("%s?user=%d", baseURL, userNum)

		// Send a GET request to the API for this user number
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		// Read the response body as HTML
		htmlBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		// Parse the HTML into JSON

		err = json.Unmarshal(htmlBytes, &jsonData)
		if err != nil {
			panic(err)
		}
		if jsonData.Links != "" {

			// Construct the row data for this user number
			sheetId := ""
			spreadsheetId := " <SPREADSHEETID>"

			// Convert sheet ID to sheet name.
			response1, err := sh.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
			if err != nil || response1.HTTPStatusCode != 200 {
				log.Fatal(err)
				return
			}

			sheetName := ""
			sId, err := strconv.ParseInt(sheetId, 10, 64);
			if err != nil {
				return
			}
			for _, v := range response1.Sheets {
				prop := v.Properties
				if prop.SheetId == sId {
					sheetName = prop.Title
					break
				}
			}

			//Append value to the sheet.
			row := &sheets.ValueRange{
				Values: [][]interface{}{{jsonData}},
			}

			response2, err := sh.Spreadsheets.Values.Append(spreadsheetId, sheetName, row).ValueInputOption("USERS").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
			if err != nil || response2.HTTPStatusCode != 200 {
				log.Fatal(err)
				return
			}
		} else {
			break
		}
	}
}

func GetClient() (*sheets.Service, error) {
	// create api context
	ctx := context.Background()

	// get bytes from base64 encoded google service accounts key
	credBytes, err := base64.StdEncoding.DecodeString(os.Getenv("KEY_JSON_BASE64"))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// authenticate and get configuration
	config, err := google.JWTConfigFromJSON(credBytes, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// create client with config and context
	client := config.Client(ctx)

	// create new service using client
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return srv, nil
}
