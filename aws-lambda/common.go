package awslambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type LambdaResponseToAPIGW struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers"`
	IsBase64Encoded   bool                `json:"isBase64Encoded"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
}

func createResponse(challange bool, session string, body any) (*LambdaResponseToAPIGW, error) {
	var respStruct *LambdaResponseToAPIGW
	var headers map[string]string
	bodyToSend := ""

	if challange && len(session) > 0 {
		return respStruct, fmt.Errorf("invalid arguments: either challange or session, not both")
	}

	if challange {
		return &LambdaResponseToAPIGW{
			StatusCode:        401,
			Headers:           map[string]string{"WWW-Authenticate": "Basic"},
			IsBase64Encoded:   false,
			MultiValueHeaders: nil,
			Body:              "",
		}, nil
	}

	if len(session) > 0 {
		cookieToSet := &http.Cookie{
			Name:     sessionCookieName,
			Value:    session,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}

		headers = map[string]string{"Set-Cookie": cookieToSet.String()}
	}

	if body != nil {
		bodyToSendInBytes, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return respStruct, marshalErr
		}
		bodyToSend = string(bodyToSendInBytes)
	}

	respStruct = &LambdaResponseToAPIGW{
		StatusCode:        200,
		Headers:           headers,
		IsBase64Encoded:   false,
		MultiValueHeaders: nil,
		Body:              bodyToSend,
	}

	return respStruct, nil
}

// parseEventCheckCreateSession parses the event and, after checking the "Cookie" and "authentication" header for credentials returns as the map-typed first parameter.
// The second return value is the session value the browser needs to set, the third parameter is an error response (most with a WWW-Authenticate challange) if any
// the last parameter is an internal processing error if any.
func parseEventCheckCreateSession(sessMan SessionManager, ctx context.Context, event json.RawMessage) (map[string]interface{}, string, *LambdaResponseToAPIGW, error) {
	var response *LambdaResponseToAPIGW
	var parsedEvent map[string]interface{}

	if eventParseErr := json.Unmarshal(event, &parsedEvent); eventParseErr != nil {
		log.Printf("Failed to unmarshal event: %v", eventParseErr)
		return nil, "", response, eventParseErr
	}

	fmt.Printf("parsedEvent: %#v\n", parsedEvent)
	fmt.Printf("cookies: %#v\n", parsedEvent["cookies"])
	fmt.Printf("headers: %#v\n", parsedEvent["headers"])
	fmt.Printf("multiValueHeaders: %#v\n", parsedEvent["multiValueHeaders"])

	headers, headersCastOk := parsedEvent["headers"].(map[string]any)
	if !headersCastOk {
		fmt.Printf("failed to cast headers:\n")
		return nil, "", response, fmt.Errorf("failed to cast headers")
	}

	var incomingCookieValue string
	if headers["Cookie"] != nil {
		fmt.Printf("cookies received: %#v\n", headers["Cookie"])
		cookieString, cookiesCastOk := headers["Cookie"].(string)
		if !cookiesCastOk {
			fmt.Printf("failed to cast cookies: %#v\n", headers["Cookie"])
			return nil, "", response, fmt.Errorf("failed to cast cookies")
		}
		cookies := strings.Split(cookieString, ";")
		for _, cookie := range cookies {
			cookieParts := strings.Split(cookie, "=")
			fmt.Printf("checking cookie name: %v\n", cookieParts[0])
			if cookieParts[0] == sessionCookieName {
				incomingCookieValue = cookieParts[1]
			}
		}
		if len(incomingCookieValue) == 0 {
			fmt.Printf("cookie named %s not found\n", sessionCookieName)
		}
	}

	sessionId, createSessErr := sessMan.checkCreateSession(ctx, incomingCookieValue, headers)
	if createSessErr != nil {
		fmt.Printf("failed to create session: %v\n", createSessErr)

		var challange *Challange
		if errors.As(createSessErr, &challange) {
			fmt.Printf("preparing challange for client...\n")
			response, createRespErr := createResponse(true, "", nil)
			if createRespErr != nil {
				fmt.Printf("failed to create response: %v\n", createRespErr)
				return nil, "", response, createRespErr
			}
			fmt.Printf("returning response with challange: %v...\n", response)
			return nil, "", response, nil
		}

		return nil, "", response, createSessErr
	}

	if len(sessionId) == 0 {
		return parsedEvent, "", response, nil
	}

	return parsedEvent, sessionId, nil, nil
}
