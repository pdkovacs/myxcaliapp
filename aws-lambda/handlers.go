package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	xcalistores3 "xcalistore-s3"
)

type eventHandlerFn func(parsedEvent map[string]any) (any, error)

func HandleListDrawingsRequest(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	return handle(ctx, event, func(parsedEvent map[string]any) (any, error) {
		titles, listErr := store.ListDrawingTitles(ctx)
		if listErr != nil {
			fmt.Printf("failed to list drawing titles: %#v\n", listErr)
			return nil, listErr
		}
		return titles, nil
	})
}

func HandleGetDrawingRequest(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	return handle(ctx, event, func(parsedEvent map[string]any) (any, error) {
		rawPathParameters := parsedEvent["pathParameters"]
		typedPathParams, ok := rawPathParameters.(map[string]string)
		if !ok {
			errMsg := "'pathParameters' event property is not of type map[string]string"
			fmt.Print(errMsg)
			return nil, fmt.Errorf("%s", errMsg)
		}
		title := typedPathParams["title"]
		content, getContentErr := store.GetDrawing(ctx, title)
		if getContentErr != nil {
			fmt.Printf("failed to get drawing content for %s: %#v", title, getContentErr)
			return nil, getContentErr
		}
		fmt.Printf("content of length %d found for %s", len(content), title)
		response, createResponseErr := createResponse(false, "", content)
		if createResponseErr != nil {
			fmt.Printf("failed to createResponse for %s: %#v", title, createResponseErr)
			return nil, createResponseErr
		}

		return response, nil
	})
}

func HandlePutDrawingRequest(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	return handle(ctx, event, func(parsedEvent map[string]any) (any, error) {
		rawPathParameters := parsedEvent["pathParameters"]
		typedPathParams, ok := rawPathParameters.(map[string]string)
		if !ok {
			errMsg := "'pathParameters' event property is not of type map[string]string"
			fmt.Print(errMsg)
			return nil, fmt.Errorf("%s", errMsg)
		}
		title := typedPathParams["title"]
		body := parsedEvent["body"]
		content, bodyIsString := body.(string)
		if !bodyIsString {
			msg := "body for %s isn't string: %#v"
			fmt.Printf(msg+"\n", title, body)
			return nil, fmt.Errorf(msg, title, body)
		}
		fmt.Printf("received content for %s of length  %d: ", title, len(content))
		contentReader := strings.NewReader(content)
		putDrawingErr := store.PutDrawing(ctx, title, contentReader)
		if putDrawingErr != nil {
			fmt.Printf("failed to store drawing %s: %v", title, putDrawingErr)
			return nil, fmt.Errorf("failed to store drawing %s: %v", title, putDrawingErr)
		}
		return nil, nil
	})
}

func handle(ctx context.Context, event json.RawMessage, eventHandler eventHandlerFn) (LambdaResponseToAPIGW, error) {
	var response LambdaResponseToAPIGW

	bucketName := os.Getenv("DRAWINGS_BUCKET_NAME")
	if len(bucketName) == 0 {
		fmt.Printf("failed to obtain bucket-name from Lambda Context")
		return response, fmt.Errorf("failed to obtain bucket-name from Lambda Context")
	}

	store, createStoreErr := xcalistores3.NewStore(ctx, bucketName)
	if createStoreErr != nil {
		fmt.Printf("failed to create store: %v\n", createStoreErr)
		return response, createStoreErr
	}
	sessMan := SessionManager{store}

	parsedEvent, sessionId, errorResponse, parseCheckErr := parseEventCheckCreateSession(sessMan, ctx, event)

	if parseCheckErr != nil {
		fmt.Printf("responding with internal error: %#v", parseCheckErr)
		return response, parseCheckErr
	}

	if errorResponse != nil {
		fmt.Printf("responding with authn error: %#v", errorResponse)
		return *errorResponse, nil
	}

	body, eventHandlerErr := eventHandler(parsedEvent)
	if eventHandlerErr != nil {
		fmt.Printf("responding with internal error: %#v", eventHandlerErr)
	}

	payloadResponse, createRespErr := createResponse(false, sessionId, body)
	if createRespErr != nil {
		fmt.Printf("failed to create response: %v\n", createRespErr)
		return response, createRespErr
	}

	return *payloadResponse, nil
}
