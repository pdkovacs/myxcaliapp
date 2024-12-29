package awslambda

import (
	"context"
	"fmt"
	"slices"
)

const (
	sessionCookieName = "xcaliapp-session"
)

type SessionStore interface {
	GetAllowedCredentials(ctx context.Context) (string, error)
	CreateSession(ctx context.Context) (string, error)
	ListSessions(ctx context.Context) ([]string, error)
}

type SessionManager struct {
	store SessionStore
}

type Challange struct{}

func (challange *Challange) Error() string {
	return "WWW-Authenticate" // "Basic"
}

// checkCreateSession checks the input headers for session or authorization header.
//
// If the expected session-id is found, returns ("", nil).
// If an invalid session is found return ("", &Challange).
// If no session is found, but a valid authorization header is found, (<new-session-id>, nil) is returned.
// If an invalid or no authorization header is found ("", &Challange) is returned
func (manager *SessionManager) checkCreateSession(ctx context.Context, receivedSessionId string, headers map[string]any) (string, error) {
	fmt.Printf("checkCreateSession: received headers: %#v\n", headers)

	if len(receivedSessionId) > 0 {
		fmt.Printf("session: %v\n", receivedSessionId)
		allowedSessions, listSessionErr := manager.store.ListSessions(ctx)
		if listSessionErr != nil {
			return "", fmt.Errorf("failed to list sessions: %w", listSessionErr)
		}
		fmt.Printf("allowed sessions: %#v\n", allowedSessions)
		if slices.Contains(allowedSessions, receivedSessionId) {
			fmt.Printf("receivedSessionId %s has been established to be allowed\n", receivedSessionId)
			return "", nil
		}
		fmt.Printf("receivedSessionId %s doesn't match any of the allowed sessionIds %#v\n", receivedSessionId, allowedSessions)
		return "", &Challange{}
	}

	authrHeader, authrHeaderFound := headers["authorization"]
	if !authrHeaderFound {
		fmt.Printf("Authorization header not found\n")
		return "", &Challange{}
	}

	allowedCred, getAllowedCredErr := manager.store.GetAllowedCredentials(ctx)
	if getAllowedCredErr != nil {
		return "", fmt.Errorf("failed to get allowed credentials: %w", getAllowedCredErr)
	}

	if allowedCred == authrHeader {
		return "", nil
	}

	sessionId, createSessIdErr := manager.store.CreateSession(ctx)
	if createSessIdErr != nil {
		return "", fmt.Errorf("store error while create session: %w", createSessIdErr)
	}

	return sessionId, nil
}
