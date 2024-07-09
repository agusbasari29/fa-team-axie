package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// import "log"

type Data struct {
	Cookie string `json:"cookie"`
	Auth   string `json:"auth"`
}

type GraphQLRequest struct {
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
}

// Define structs matching your GraphQL response structure
type Address struct {
	Ethereum string `json:"ethereum"`
	Tomo     string `json:"tomo"`
	Loom     string `json:"loom"`
	Ronin    string `json:"ronin"`
}

type Settings struct {
	UnsubscribeLandDelegationEmail bool `json:"unsubscribeLandDelegationEmail"`
	UnsubscribeNotificationEmail   bool `json:"unsubscribeNotificationEmail"`
}

type Referral struct {
	Code    string `json:"code"`
	Address string `json:"address"`
	AddedAt string `json:"addedAt"`
}

type ProfileBrief struct {
	AccountID          string   `json:"accountId"`
	Addresses          Address  `json:"addresses"`
	Email              string   `json:"email"`
	Activated          bool     `json:"activated"`
	Name               string   `json:"name"`
	Settings           Settings `json:"settings"`
	Referral           Referral `json:"referral"`
	IsScholar          bool     `json:"isScholar"`
	FortuneSlipBalance int      `json:"fortuneSlipBalance"`
}

type GetProfileBriefResponse struct {
	Data struct {
		Profile ProfileBrief `json:"profile"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type Quest struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type UserQuests struct {
	DailyResetAt     string  `json:"dailyResetAt"`
	DailyTotalPoints int     `json:"dailyTotalPoints"`
	Quests           []Quest `json:"quests"`
}

type GetUserQuestsResponse struct {
	Data struct {
		UserQuests UserQuests `json:"userQuests"`
	} `json:"data"`
	Error []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type WinBattleResponse struct {
	Data struct {
		VerifyQuest bool `json:"verifyQuest"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type GetWeeklyQuestPointsLeaderboardResponse struct {
	Data struct {
		Leaderboard struct {
			TotalParticipants int             `json:"totalParticipants"`
			TotalScore        json.RawMessage `json:"totalScore"`
			Typename          string          `json:"__typename"`
		} `json:"leaderboard"`
		UserLeaderboardRank struct {
			Rank     int             `json:"rank"`
			Score    json.RawMessage `json:"score"`
			Typename string          `json:"__typename"`
		} `json:"userLeaderboardRank"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

const (
	url         = "https://graphql-gateway.axieinfinity.com/graphql"
	contentType = "application/json"
	green       = "\033[32m"
	red         = "\033[31m"
	yellow      = "\033[33m"
	blue        = "\033[34m"
	purple      = "\033[35m"
	cyan        = "\033[36m"
	white       = "\033[37m"
	gray        = "\033[90m"
	red2        = "\033[91m"
	green2      = "\033[92m"
	yellow2     = "\033[93m"
	blue2       = "\033[94m"
	purple2     = "\033[95m"
	cyan2       = "\033[96m"
	white2      = "\033[97m"
	reset       = "\033[0m"
)

func main() {
	data := GetData()

	for _, d := range data {
		profileName := GetProfileBrief(d).Data.Profile.Name
		user := GetProfileBrief(d).Data.Profile.Addresses.Ronin
		if len(profileName) < 15 {
			fmt.Printf("%s%s%s\t\t", cyan, profileName, reset)
		} else {
			fmt.Printf("%s%s%s\t", cyan, profileName, reset)
		}
		Quest := GetUserQuests(d).Data.UserQuests.Quests
		for _, q := range Quest {
			if q.Type == "Win1ClassicBattle" && q.Status == "Complete" {
				fmt.Printf("Classic : %sVERIFIED%s\t", green, reset)
			} else if q.Type == "Win1ClassicBattle" && q.Status == "Open" {
				verifyQuest := Win1ClassicBattle(d).Data.VerifyQuest
				if !verifyQuest {
					fmt.Printf("Classic : %sBELUM SELESAI%s\t", red, reset)
				} else {
					fmt.Printf("Classic : %sVERIFIED%s\t", yellow, reset)
				}
			}
			if q.Type == "Win1OriginsBattle" && q.Status == "Complete" {
				fmt.Printf("Origin : %sVERIFIED%s\t", green, reset)
			} else if q.Type == "Win1OriginsBattle" && q.Status == "Open" {
				verifyQuest := Win1OriginsBattle(d).Data.VerifyQuest
				if !verifyQuest {
					fmt.Printf("Origin : %sBELUM SELESAI%s\t", red, reset)
				} else {
					fmt.Printf("Origin : %sVERIFIED%s\t", yellow, reset)
				}
			}

		}
		Score, _ := parseScore(GetWeeklyQuestPointsLeaderboard(d, user).Data.UserLeaderboardRank.Score)
		fmt.Printf("Score : %s%d%s\t", purple, Score, reset)
		Rank := GetWeeklyQuestPointsLeaderboard(d, user).Data.UserLeaderboardRank.Rank
		fmt.Printf("Rank : %s%d%s\n", purple, Rank, reset)

		// fmt.Printf("=====================================\n")
	}
}

func GetData() []Data {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dataStr := os.Getenv("DATA")
	if dataStr == "" {
		log.Fatal("DATA not set")
	}

	var data []Data
	err = json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func GetRequest(data *Data, requestPayload interface{}) []byte {
	// Encode the request payload as JSON
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		log.Fatalf("Failed to encode request payload: %v", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Authorization", "Bearer "+data.Auth)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(requestBody)))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Cookie", data.Cookie)
	req.Header.Set("Host", "graphql-gateway.axieinfinity.com")
	req.Header.Set("Origin", "https://app.axieinfinity.com")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://app.axieinfinity.com/")
	req.Header.Set("Sec-Ch-Ua", `"Not/A)Brand";v="8", "Chromium";v="126"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.6478.127 Safari/537.36")

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	var reader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer reader.(*gzip.Reader).Close()
	default:
		reader = resp.Body
	}

	// Read the response body
	body, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	return body
}

func GetProfileBrief(data Data) GetProfileBriefResponse {
	query := `query GetProfileBrief {
		profile {
		  ...ProfileBrief
		  __typename
		}
	  }
	  
	  fragment ProfileBrief on AccountProfile {
		accountId
		addresses {
		  ...Addresses
		  __typename
		}
		email
		activated
		name
		settings {
		  unsubscribeLandDelegationEmail
		  unsubscribeNotificationEmail
		  __typename
		}
		referral {
		  ...AccountReferralFragement
		  __typename
		}
		isScholar
		fortuneSlipBalance
		__typename
	  }
	  
	  fragment Addresses on NetAddresses {
		ethereum
		tomo
		loom
		ronin
		__typename
	  }
	  
	  fragment AccountReferralFragement on AccountReferral {
		code
		address
		addedAt
		__typename
	  }`

	requestPayload := GraphQLRequest{
		OperationName: "GetProfileBrief",
		Variables:     map[string]interface{}{},
		Query:         query,
	}

	body := GetRequest(&data, requestPayload)
	var response GetProfileBriefResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Check for GraphQL errors
	if len(response.Errors) > 0 {
		for _, err := range response.Errors {
			fmt.Printf("GraphQL Error: %s\n", err.Message)
		}
	}

	return response
}

func GetUserQuests(data Data) GetUserQuestsResponse {
	query := `query GetUserQuests {
		  userQuests {
			    quests {
			      type
			      status
			    }
		    }
		  }`

	requestPayload := GraphQLRequest{
		OperationName: "GetUserQuests",
		Variables:     map[string]interface{}{},
		Query:         query,
	}

	body := GetRequest(&data, requestPayload)
	var response GetUserQuestsResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Check for GraphQL errors
	if len(response.Error) > 0 {
		for _, err := range response.Error {
			fmt.Printf("GraphQL Error: %s\n", err.Message)
		}
	}

	return response
}

func Win1ClassicBattle(data Data) WinBattleResponse {
	query := `mutation VerifyQuest($questType: QuestType!) {
		verifyQuest(questType: $questType)
	  }`

	// Create the GraphQL request payload
	requestPayload := GraphQLRequest{
		OperationName: "VerifyQuest",
		Variables:     map[string]interface{}{"questType": "Win1ClassicBattle"},
		Query:         query,
	}

	body := GetRequest(&data, requestPayload)
	var response WinBattleResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Check for GraphQL errors
	if len(response.Errors) > 0 {
		for _, err := range response.Errors {
			fmt.Printf("GraphQL Error: %s\n", err.Message)
		}
	}

	return response
}

func Win1OriginsBattle(data Data) WinBattleResponse {
	query := `mutation VerifyQuest($questType: QuestType!) {
		verifyQuest(questType: $questType)
	  }`

	requestPayload := GraphQLRequest{
		OperationName: "VerifyQuest",
		Variables:     map[string]interface{}{"questType": "Win1OriginsBattle"},
		Query:         query,
	}

	body := GetRequest(&data, requestPayload)
	var response WinBattleResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Check for GraphQL errors
	if len(response.Errors) > 0 {
		for _, err := range response.Errors {
			fmt.Printf("GraphQL Error: %s\n", err.Message)
		}
	}

	return response
}

func GetWeeklyQuestPointsLeaderboard(data Data, user string) GetWeeklyQuestPointsLeaderboardResponse {
	query := `query GetWeeklyQuestPointsLeaderboard($user: String!, $includeUserRank: Boolean!) {
		leaderboard(type: WeeklyQuestPoints) {
		  totalParticipants
		  totalScore
		  __typename
		}
		userLeaderboardRank(user: $user, type: WeeklyQuestPoints) @include(if: $includeUserRank) {
		  rank
		  score
		  __typename
		}
	  }`

	// Create the GraphQL request payload
	requestPayload := GraphQLRequest{
		OperationName: "GetWeeklyQuestPointsLeaderboard",
		Variables: map[string]interface{}{
			"user":            user,
			"includeUserRank": true,
		},
		Query: query,
	}

	body := GetRequest(&data, requestPayload)
	var response GetWeeklyQuestPointsLeaderboardResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Check for GraphQL errors
	if len(response.Errors) > 0 {
		for _, err := range response.Errors {
			fmt.Printf("GraphQL Error: %s\n", err.Message)
		}
	}

	return response
}

func parseScore(scoreRaw json.RawMessage) (int, error) {
	var score int
	var scoreStr string
	// Try to unmarshal as int
	if err := json.Unmarshal(scoreRaw, &score); err == nil {
		return score, nil
	}
	// If that fails, try to unmarshal as string and convert to int
	if err := json.Unmarshal(scoreRaw, &scoreStr); err == nil {
		return strconv.Atoi(scoreStr)
	}
	return 0, fmt.Errorf("invalid score format")
}

// // Define HTTP handler function
// log.Print("starting server...")
// http.HandleFunc("/", handler)

// // Determine port for HTTP service.
// port := os.Getenv("PORT")
// if port == "" {
// 	port = "3000"
// 	log.Printf("defaulting to port %s", port)
// }

// // Start HTTP server.
// log.Printf("listening on port %s", port)
// if err := http.ListenAndServe(":"+port, nil); err != nil {
// 	log.Fatal(err)
// }

// func handler(w http.ResponseWriter, r *http.Request) {
// 	// name := os.Getenv("NAME")
// 	// if name == "" {
// 	// 	name = "World"
// 	// }
// 	// fmt.Fprintf(w, "Hello %s!\n", name)
// }
